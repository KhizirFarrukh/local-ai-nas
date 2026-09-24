package api

import (
	"encoding/json"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// attackServer is a real HTTP server with a real tus server over a fresh
// files area holding fixtureFiles; it returns the client and the area.
func attackServer(t *testing.T) (client, string) {
	t.Helper()
	svc, area := testFilesDir(t)
	return client{t, testutil.NewServer(t, New(Options{Files: svc, Uploads: testUploads(t, svc)})).URL}, area
}

// problemOK checks that an answer is a 4xx problem.
func problemOK(t *testing.T, what string, st int, hdr http.Header, b []byte) apperr.Problem {
	t.Helper()
	var p apperr.Problem
	if st < 400 || st >= 500 || hdr.Get("Content-Type") != apperr.ContentType || json.Unmarshal(b, &p) != nil {
		t.Errorf("%s: %d %s, want a 4xx problem", what, st, b)
	}
	return p
}

// TestAttackMaliciousNames: every endpoint that creates a name refuses
// names that break the name rules, with invalid_name and the rule.
func TestAttackMaliciousNames(t *testing.T) {
	c, area := attackServer(t)
	names := map[string]string{
		"CON": storage.RuleReservedName, "aux.txt": storage.RuleReservedName, "Com1.log": storage.RuleReservedName,
		"a:b": storage.RuleForbiddenChar, "a<b": storage.RuleForbiddenChar, "a|b": storage.RuleForbiddenChar,
		"a?b": storage.RuleForbiddenChar, "a*b": storage.RuleForbiddenChar, `a"b`: storage.RuleForbiddenChar,
		"trailing.": storage.RuleTrailingChar, "trailing ": storage.RuleTrailingChar,
		"tab\there": storage.RuleControlChar, "bell\x07": storage.RuleControlChar, "del\x7f": storage.RuleControlChar,
		strings.Repeat("n", 300): storage.RuleNameTooLong,
		storage.TempPrefix + "x": storage.RuleReservedName,
		"a／b":                    storage.RuleLookalikeSep,
		"a∕b":                    storage.RuleLookalikeSep,
		"‥":                      "", // reads as ".." (outside_root at the resolver, dot_name as a name)
		"line\nbreak":            storage.RuleControlChar,
		"new\rline":              storage.RuleControlChar,
		"esc\x1b[31mred\x1b[0m":  storage.RuleControlChar,
		"space-and-dot. ":        storage.RuleTrailingChar,
		"LPT9":                   storage.RuleReservedName,
		"CONIN$":                 storage.RuleReservedName,
		strings.Repeat("é", 128): storage.RuleNameTooLong, // 256 bytes
		"ok-but-\x00-nul":        "",
	}
	js := func(v string) string { b, _ := json.Marshal(v); return string(b) }
	jsonType := "application/json"
	for name, rule := range names {
		attempts := []struct {
			what                string
			method, path, ctype string
			body                string
			headers             map[string]string
		}{
			{"new folder", "POST", "/api/v1/files/folders", jsonType, `{"path":` + js("/"+name) + `}`, nil},
			{"upload", "PUT", "/api/v1/files/content?path=" + url.QueryEscape("/"+name), "application/octet-stream", "x", nil},
			{"rename", "POST", "/api/v1/files/operations/rename", jsonType, `{"path":"/readme.md","new_name":` + js(name) + `}`, nil},
			{"move", "POST", "/api/v1/files/operations/move", jsonType, `{"from":"/readme.md","to":` + js("/"+name) + `}`, nil},
			{"copy", "POST", "/api/v1/files/operations/copy", jsonType, `{"from":"/readme.md","to":` + js("/"+name) + `}`, nil},
			{"tus", "POST", UploadsPath, "", "", map[string]string{"Tus-Resumable": "1.0.0", "Upload-Length": "1", "Upload-Metadata": "target_path " + b64("/"+name)}},
		}
		for _, a := range attempts {
			var body []byte
			if a.body != "" {
				body = []byte(a.body)
			}
			st, hdr, b := c.do(a.method, a.path, a.ctype, body, a.headers)
			p := problemOK(t, a.what+" "+strings.ToValidUTF8(name, "?"), st, hdr, b)
			if p.Code != "invalid_name" && p.Code != "outside_root" {
				t.Errorf("%s %q: code %q, want invalid_name", a.what, name, p.Code)
			}
			if rule != "" && p.Rule != rule {
				t.Errorf("%s %q: rule %q, want %q", a.what, name, p.Rule, rule)
			}
		}
	}
	// Nothing was created.
	got, err := testutil.ReadFiles(area)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(fixtureFiles) {
		t.Errorf("the files area changed: %v", got)
	}
}

// TestAttackHeaderInjection: a file whose name holds CR/LF, quotes, or
// semicolons (possible on Linux when made outside the API) is downloaded
// without any injected header and with a Content-Disposition that gives
// the exact name back.
func TestAttackHeaderInjection(t *testing.T) {
	c, area := attackServer(t)
	names := []string{"semi;colon,comma=x.txt"}
	if runtime.GOOS != "windows" { // Windows cannot hold these names
		names = append(names,
			"evil\r\nSet-Cookie: pwned=1\r\n.txt",
			`quote"; filename="evil.exe`,
			"lf\nX-Injected: 1.txt")
	}
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(area, name), []byte("payload"), 0o600); err != nil {
			t.Fatal(err)
		}
		st, hdr, b := c.do("GET", "/api/v1/files/content?path="+url.QueryEscape("/"+name), "", nil, nil)
		if st != http.StatusOK || string(b) != "payload" {
			t.Errorf("%q: %d %q", name, st, b)
			continue
		}
		if hdr.Get("Set-Cookie") != "" || hdr.Get("X-Injected") != "" {
			t.Errorf("%q: an injected header came through: %v", name, hdr)
		}
		cd := hdr.Get("Content-Disposition")
		if strings.ContainsAny(cd, "\r\n") {
			t.Errorf("%q: Content-Disposition holds a line break: %q", name, cd)
		}
		if _, params, err := mime.ParseMediaType(cd); err != nil || params["filename"] != name {
			t.Errorf("%q: Content-Disposition %q gives %q (%v)", name, cd, params["filename"], err)
		}
		// The listing carries the name as JSON, escaped.
		st, _, b = c.do("GET", "/api/v1/files/items?path=/", "", nil, nil)
		if st != http.StatusOK || !json.Valid(b) {
			t.Errorf("the listing with %q: %d %s", name, st, b)
		}
	}
}

// TestAttackOversizedInputs: inputs far over any limit are refused with
// 4xx (or Go's own 431 for oversized headers), and the server keeps
// working.
func TestAttackOversizedInputs(t *testing.T) {
	c, _ := attackServer(t)
	long := strings.Repeat("n", 300)
	deep := strings.Repeat("/a", 10000)
	longPath := strings.Repeat("/"+strings.Repeat("p", 200), 21) // 4221 bytes
	for _, p := range []string{"/" + long, "/docs/" + long, deep, longPath} {
		for _, req := range [][2]string{
			{"GET", "/api/v1/files/items?path=" + url.QueryEscape(p)},
			{"GET", "/api/v1/files/content?path=" + url.QueryEscape(p)},
			{"DELETE", "/api/v1/files/items?path=" + url.QueryEscape(p)},
			{"PUT", "/api/v1/files/content?path=" + url.QueryEscape(p)},
		} {
			st, hdr, b := c.do(req[0], req[1], "application/octet-stream", []byte("x"), nil)
			problemOK(t, req[0]+" "+p[:min(len(p), 40)]+"…", st, hdr, b)
		}
	}
	st, hdr, b := c.do("POST", "/api/v1/files/folders", "application/json", []byte(`{"path":"/`+strings.Repeat("a", 2<<20)+`"}`), nil)
	if p := problemOK(t, "2 MiB body", st, hdr, b); p.Code != "too_large" {
		t.Errorf("2 MiB body: %+v", p)
	}
	// Headers over Go's limit: Go answers 431 itself.
	st, _, _ = c.do("GET", "/api/v1/system/health", "", nil, map[string]string{"X-Big": strings.Repeat("h", 2<<20)})
	if st != http.StatusRequestHeaderFieldsTooLarge {
		t.Errorf("2 MiB header: %d, want 431", st)
	}
	st, hdr, b = c.do("POST", UploadsPath, "", nil, map[string]string{
		"Tus-Resumable": "1.0.0", "Upload-Length": "1", "Upload-Metadata": "target_path " + b64("/"+strings.Repeat("m", 500<<10)),
	})
	problemOK(t, "500 KiB tus metadata", st, hdr, b)
	if st, _, _ := c.do("GET", "/api/v1/system/health", "", nil, nil); st != http.StatusOK {
		t.Errorf("the server does not answer after the attacks: %d", st)
	}
}

// TestAttackTusMetadata: abusive tus headers and metadata are refused
// with 4xx problems, and nothing is stored.
func TestAttackTusMetadata(t *testing.T) {
	c, area := attackServer(t)
	create := func(h map[string]string) (int, http.Header, []byte) {
		all := map[string]string{"Tus-Resumable": "1.0.0"}
		for k, v := range h {
			all[k] = v
		}
		return c.do("POST", UploadsPath, "", nil, all)
	}
	cases := map[string]map[string]string{
		"invalid base64":     {"Upload-Length": "1", "Upload-Metadata": "target_path !!!not-base64!!!"},
		"no value":           {"Upload-Length": "1", "Upload-Metadata": "target_path"},
		"traversal":          {"Upload-Length": "1", "Upload-Metadata": "target_path " + b64("/../u0002/x")},
		"NUL in the target":  {"Upload-Length": "1", "Upload-Metadata": "target_path " + b64("/a\x00b")},
		"relative target":    {"Upload-Length": "1", "Upload-Metadata": "target_path " + b64("x.bin")},
		"policy injection":   {"Upload-Length": "1", "Upload-Metadata": "target_path " + b64("/p.bin") + ",on_conflict " + b64("fail;rm -rf /")},
		"checksum not hex":   {"Upload-Length": "1", "Upload-Metadata": "target_path " + b64("/s.bin") + ",sha256 " + b64(strings.Repeat("z", 64))},
		"negative length":    {"Upload-Length": "-1", "Upload-Metadata": "target_path " + b64("/n.bin")},
		"length not numeric": {"Upload-Length": "ten", "Upload-Metadata": "target_path " + b64("/n.bin")},
		"length overflow":    {"Upload-Length": "99999999999999999999999", "Upload-Metadata": "target_path " + b64("/n.bin")},
		"both lengths":       {"Upload-Length": "1", "Upload-Defer-Length": "1", "Upload-Metadata": "target_path " + b64("/n.bin")},
		"concatenation":      {"Upload-Length": "1", "Upload-Concat": "partial", "Upload-Metadata": "target_path " + b64("/n.bin")},
		"old tus version":    {"Tus-Resumable": "0.2.2", "Upload-Length": "1", "Upload-Metadata": "target_path " + b64("/n.bin")},
		"photos area":        {"Upload-Length": "1", "Upload-Metadata": "target_path " + b64("/../../photos/u0001/p.jpg")},
	}
	for name, h := range cases {
		st, hdr, b := create(h)
		problemOK(t, name, st, hdr, b)
	}
	// Abusive offsets on a real upload.
	st, hdr, _ := create(map[string]string{"Upload-Length": "4", "Upload-Metadata": "target_path " + b64("/ok.bin")})
	if st != http.StatusCreated {
		t.Fatalf("create: %d", st)
	}
	loc := strings.TrimPrefix(hdr.Get("Location"), c.url)
	for _, off := range []string{"-5", "ten", "99999999999999999999999", "5"} {
		st, hdr, b := c.do("PATCH", loc, "application/offset+octet-stream", []byte("ab"), map[string]string{"Tus-Resumable": "1.0.0", "Upload-Offset": off})
		problemOK(t, "offset "+off, st, hdr, b)
	}
	// More data than the declared length.
	st, hdr, b := c.do("PATCH", loc, "application/offset+octet-stream", []byte("abcdef"), map[string]string{"Tus-Resumable": "1.0.0", "Upload-Offset": "0"})
	problemOK(t, "more than Upload-Length", st, hdr, b)
	got, err := testutil.ReadFiles(area)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(fixtureFiles) {
		t.Errorf("the files area changed: %v", got)
	}
}

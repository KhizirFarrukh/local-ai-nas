package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/KhizirFarrukh/local-ai-nas/internal/health"
	"github.com/KhizirFarrukh/local-ai-nas/internal/logging"
	"github.com/KhizirFarrukh/local-ai-nas/internal/testutil"
)

// runAsMainEnv makes the test binary act as the real program, so the smoke
// test runs main() in a separate process without a separate build.
const runAsMainEnv = "LOCALAINAS_TEST_RUN_MAIN"

func TestMain(m *testing.M) {
	if os.Getenv(runAsMainEnv) == "1" {
		os.Args = append([]string{"local-ai-nas"}, os.Args[1:]...)
		main()
		return
	}
	os.Exit(m.Run())
}

// isolateEnv removes LOCALAINAS_* variables of the developer's environment
// for the test (t.Setenv restores them afterwards). On Windows it also
// moves the default config file location into an empty temp directory.
func isolateEnv(t *testing.T) {
	t.Helper()
	for _, kv := range os.Environ() {
		if name, _, _ := strings.Cut(kv, "="); strings.HasPrefix(name, "LOCALAINAS_") && name != runAsMainEnv {
			t.Setenv(name, "")
			_ = os.Unsetenv(name)
		}
	}
	if runtime.GOOS == "windows" {
		t.Setenv("ProgramData", t.TempDir())
	}
}

// lockedBuffer is a bytes.Buffer that is safe for concurrent use.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func runCmd(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(t.Context(), args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestRunCommands(t *testing.T) {
	isolateEnv(t)
	tests := []struct {
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{[]string{"version"}, 0, "local-ai-nas dev\n", ""},
		{[]string{"--version"}, 0, "local-ai-nas dev\n", ""},
		{[]string{"help"}, 0, "Usage:", ""},
		{nil, 2, "", "Usage:"},
		{[]string{"frobnicate"}, 2, "", `unknown command "frobnicate"`},
		{[]string{"serve", "-h"}, 0, "", "-storage-root"},
		{[]string{"serve", "--no-such-flag"}, 2, "", "flag provided but not defined"},
		{[]string{"serve", "extra"}, 2, "", `unexpected argument "extra"`},
		{[]string{"serve"}, 1, "", "config: storage.root: required"},
		{[]string{"serve", "--storage-root", "relative/dir"}, 1, "", "must be an absolute path"},
		{[]string{"migrate"}, 2, "", `"up" or "status"`},
		{[]string{"migrate", "sideways"}, 2, "", `"up" or "status"`},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			code, stdout, stderr := runCmd(t, tt.args...)
			if code != tt.wantCode {
				t.Errorf("exit code = %d, want %d (stderr: %s)", code, tt.wantCode, stderr)
			}
			if !strings.Contains(stdout, tt.wantStdout) {
				t.Errorf("stdout = %q, want it to contain %q", stdout, tt.wantStdout)
			}
			if !strings.Contains(stderr, tt.wantStderr) {
				t.Errorf("stderr = %q, want it to contain %q", stderr, tt.wantStderr)
			}
		})
	}
}

func TestServeInvalidConfigNamesKey(t *testing.T) {
	isolateEnv(t)
	code, _, stderr := runCmd(t, "serve", "--storage-root", testutil.StorageRoot(t), "--log-level", "loud")
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	want := `config: log.level (from flag --log-level): must be debug, info, warn, or error, got "loud"`
	if !strings.Contains(stderr, want) {
		t.Errorf("stderr = %q, want it to contain %q", stderr, want)
	}
}

// freeAddr returns a loopback address with a port that was free a moment
// ago. The config does not accept port 0.
func freeAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	return addr
}

// waitHealthy polls the health endpoint until it answers or the deadline
// passes, and returns the last response's report and headers.
func waitHealthy(t *testing.T, addr string, exited func() bool) (health.Report, http.Header) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if exited != nil && exited() {
			t.Fatal("the server exited before it became healthy")
		}
		resp, err := http.Get("http://" + addr + "/api/v1/system/health")
		if err == nil {
			var rep health.Report
			derr := json.NewDecoder(resp.Body).Decode(&rep)
			_ = resp.Body.Close()
			if derr != nil {
				t.Fatalf("health body: %v", derr)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("health status %d: %+v", resp.StatusCode, rep)
			}
			return rep, resp.Header
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("the server did not become healthy within 15s")
	return health.Report{}, nil
}

func TestServeInProcess(t *testing.T) {
	isolateEnv(t)
	root := testutil.StorageRoot(t)
	addr := freeAddr(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var stderr lockedBuffer
	done := make(chan int, 1)
	go func() {
		done <- run(ctx, []string{"serve", "--storage-root", root, "--server-bind", addr}, io.Discard, &stderr)
	}()
	exited := func() bool { return len(done) > 0 }

	rep, hdr := waitHealthy(t, addr, exited)
	if rep.Status != "ok" || rep.Version != "dev" || len(rep.Checks) != 1 || rep.Checks[0].Name != "database" {
		t.Errorf("health report = %+v", rep)
	}
	if id := hdr.Get(logging.RequestIDHeader); len(id) != 32 {
		t.Errorf("X-Request-ID = %q, want a generated ID", id)
	}

	// A method the route does not allow gets 405 from the mux.
	resp, err := http.Post("http://"+addr+"/api/v1/system/health", "text/plain", strings.NewReader("x"))
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("POST health = %d, want 405", resp.StatusCode)
	}

	cancel()
	select {
	case code := <-done:
		if code != 0 {
			t.Errorf("serve exit code = %d, want 0; stderr:\n%s", code, stderr.String())
		}
	case <-time.After(15 * time.Second):
		t.Fatal("serve did not stop after the context was cancelled")
	}

	logFile, err := os.ReadFile(filepath.Join(root, ".local-ai-nas", "logs", logging.LogFileName))
	if err != nil {
		t.Fatalf("log file: %v", err)
	}
	for _, want := range []string{`"msg":"starting"`, `"msg":"database migration applied"`, `"msg":"listening"`, `"msg":"request"`, `"route":"GET /api/v1/system/health"`, `"msg":"shutting down"`, `"msg":"stopped"`} {
		if !strings.Contains(string(logFile), want) {
			t.Errorf("log file lacks %s", want)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".local-ai-nas", "db", "nas.db")); err != nil {
		t.Errorf("database file: %v", err)
	}
}

func TestServePortInUse(t *testing.T) {
	isolateEnv(t)
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	code, _, stderr := runCmd(t, "serve", "--storage-root", testutil.StorageRoot(t), "--server-bind", l.Addr().String())
	if code != 1 || !strings.Contains(stderr, "cannot listen") {
		t.Errorf("exit code %d, stderr %q; want 1 and \"cannot listen\"", code, stderr)
	}
}

func TestMigrateCommand(t *testing.T) {
	isolateEnv(t)
	root := testutil.StorageRoot(t)
	steps := []struct {
		action string
		want   *regexp.Regexp
	}{
		{"status", regexp.MustCompile(`(?m)^00001\s+pending\s+-$`)},
		{"up", regexp.MustCompile(`(?m)^applied 00001 `)},
		{"up", regexp.MustCompile(`(?m)^the database is up to date$`)},
		{"status", regexp.MustCompile(`(?m)^00001\s+applied\s+\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}Z$`)},
	}
	for _, s := range steps {
		code, stdout, stderr := runCmd(t, "migrate", "--storage-root", root, s.action)
		if code != 0 || !s.want.MatchString(stdout) {
			t.Errorf("migrate %s: exit %d, stdout %q, stderr %q; want a match for %s", s.action, code, stdout, stderr, s.want)
		}
	}
}

func TestUnknownRootEntriesAreReported(t *testing.T) {
	isolateEnv(t)
	root := testutil.StorageRoot(t)
	if err := testutil.WriteFiles(root, map[string]string{"notes.txt": "mine"}); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runCmd(t, "migrate", "--storage-root", root, "status")
	if code != 0 || !strings.Contains(stderr, "left alone") || !strings.Contains(stderr, "notes.txt") {
		t.Errorf("exit %d, stderr %q; want a warning naming notes.txt", code, stderr)
	}
	files, err := testutil.ReadFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	if files["notes.txt"] != "mine" {
		t.Error("the unknown entry was changed")
	}
}

func TestLayoutErrorStopsStartup(t *testing.T) {
	isolateEnv(t)
	root := testutil.StorageRoot(t)
	if err := testutil.WriteFiles(root, map[string]string{"files": "a file where the area should be"}); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runCmd(t, "serve", "--storage-root", root)
	if code != 1 || !strings.Contains(stderr, "is not a directory") {
		t.Errorf("exit %d, stderr %q; want 1 and a layout error", code, stderr)
	}
}

func TestRelocatedDatabase(t *testing.T) {
	isolateEnv(t)
	root := testutil.StorageRoot(t)
	dbDir := filepath.Join(testutil.StorageRoot(t), "db")
	code, _, stderr := runCmd(t, "migrate", "--storage-root", root, "--storage-db-dir", dbDir, "up")
	if code != 0 {
		t.Fatalf("migrate up: exit %d, stderr %q", code, stderr)
	}
	if _, err := os.Stat(filepath.Join(dbDir, "nas.db")); err != nil {
		t.Errorf("database not in the configured directory: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".local-ai-nas", "db")); !os.IsNotExist(err) {
		t.Error("the default database directory exists although the database was moved")
	}
}

func TestOverlappingInternalDataStopsStartup(t *testing.T) {
	isolateEnv(t)
	root := testutil.StorageRoot(t)
	code, _, stderr := runCmd(t, "serve", "--storage-root", root, "--storage-logs-dir", filepath.Join(root, "files"))
	if code != 1 || !strings.Contains(stderr, "the logs directory (storage.logs_dir) and the files area are the same directory") {
		t.Errorf("exit %d, stderr %q; want 1 and an overlap error", code, stderr)
	}
}

func TestGracefulShutdownWaitsForInFlight(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		_, _ = io.WriteString(w, "finished")
	})}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	errc := make(chan error, 1)
	go func() { errc <- serve(ctx, srv, ln, 10*time.Second, slog.New(slog.DiscardHandler)) }()

	type result struct {
		body string
		err  error
	}
	resc := make(chan result, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/")
		if err != nil {
			resc <- result{err: err}
			return
		}
		b, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		resc <- result{string(b), err}
	}()

	<-started
	cancel() // shutdown begins while the request is in flight
	time.Sleep(100 * time.Millisecond)
	close(release)

	if r := <-resc; r.err != nil || r.body != "finished" {
		t.Errorf("in-flight request = %q, %v; want it to finish", r.body, r.err)
	}
	if err := <-errc; err != nil {
		t.Errorf("serve = %v, want a clean shutdown", err)
	}
}

func TestGracefulShutdownTimeout(t *testing.T) {
	started := make(chan struct{})
	block := make(chan struct{})
	defer close(block)
	srv := &http.Server{Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		close(started)
		<-block
	})}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	errc := make(chan error, 1)
	go func() { errc <- serve(ctx, srv, ln, 200*time.Millisecond, slog.New(slog.DiscardHandler)) }()
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String() + "/")
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
	<-started
	cancel()
	select {
	case err := <-errc:
		if err == nil || !strings.Contains(err.Error(), "graceful shutdown") {
			t.Errorf("serve = %v, want a graceful-shutdown timeout error", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("serve did not give up after the shutdown timeout")
	}
}

func TestServeReturnsListenerError(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_ = ln.Close() // Serve fails at once on a closed listener
	err = serve(t.Context(), &http.Server{}, ln, time.Second, slog.New(slog.DiscardHandler))
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		t.Errorf("serve = %v, want the listener error", err)
	}
}

// TestSmokeBinary runs the program as a separate process (this test
// binary acting as main), checks the health endpoint, and stops it: with
// SIGTERM where signals exist (graceful, exit 0), by killing it on Windows.
func TestSmokeBinary(t *testing.T) {
	isolateEnv(t)
	root := testutil.StorageRoot(t)
	addr := freeAddr(t)
	cmd := exec.Command(os.Args[0], "serve", "--storage-root", root, "--server-bind", addr)
	cmd.Env = append(os.Environ(), runAsMainEnv+"=1")
	var stderr lockedBuffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	// done is closed when the process has exited; waitErr is set before.
	// (A buffered channel's length cannot tell "exited" once the result
	// has been received, which made the cleanup below wait forever.)
	done := make(chan struct{})
	var waitErr error
	go func() {
		waitErr = cmd.Wait()
		close(done)
	}()
	exited := func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}
	defer func() {
		if !exited() {
			_ = cmd.Process.Kill()
			<-done
		}
	}()

	rep, _ := waitHealthy(t, addr, exited)
	if rep.Status != "ok" {
		t.Errorf("health = %+v", rep)
	}

	if runtime.GOOS == "windows" {
		return // no SIGTERM on Windows; the deferred Kill stops the process
	}
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
		if waitErr != nil {
			t.Errorf("process exit: %v; stderr:\n%s", waitErr, stderr.String())
		}
		if !strings.Contains(stderr.String(), `"msg":"stopped"`) {
			t.Errorf("no graceful stop in the log:\n%s", stderr.String())
		}
	case <-time.After(15 * time.Second):
		t.Fatal("the process did not stop after SIGTERM")
	}
}

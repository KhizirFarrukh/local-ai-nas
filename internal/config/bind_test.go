package config

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// fakeNetwork replaces the resolver and the container check for a test.
func fakeNetwork(t *testing.T, hosts map[string][]string, container bool) {
	t.Helper()
	origLookup, origContainer := lookupHost, inContainer
	lookupHost = func(_ context.Context, host string) ([]string, error) {
		if addrs, ok := hosts[host]; ok {
			return addrs, nil
		}
		return nil, errors.New("no such host")
	}
	inContainer = func() bool { return container }
	t.Cleanup(func() { lookupHost, inContainer = origLookup, origContainer })
}

// TestBindGuard is the S01.6-T06 acceptance test: until S03, the server
// listens on loopback only.
func TestBindGuard(t *testing.T) {
	fakeNetwork(t, map[string][]string{
		"loop.test":  {"127.0.0.1", "::1"},
		"lan.test":   {"192.168.1.20"},
		"mixed.test": {"127.0.0.1", "10.0.0.5"},
	}, false)
	root := absRoot(t)
	tests := []struct {
		bind   string
		accept bool
	}{
		{"127.0.0.1:8080", true},
		{"127.3.4.5:8080", true},
		{"[::1]:8080", true},
		{"localhost:8080", true},
		{"LocalHost:8080", true},
		{"loop.test:8080", true},
		{"0.0.0.0:8080", false},
		{"[::]:8080", false},
		{"192.168.1.10:8080", false},
		{"10.0.0.1:8080", false},
		{"[fe80::1]:8080", false},
		{"lan.test:8080", false},
		{"mixed.test:8080", false},
		{"unknown.test:8080", false},
	}
	for _, tt := range tests {
		_, err := Load(Sources{Flags: map[string]string{"storage.root": root, "server.bind": tt.bind}})
		switch {
		case tt.accept && err != nil:
			t.Errorf("bind %s refused: %v", tt.bind, err)
		case !tt.accept && err == nil:
			t.Errorf("bind %s accepted, want it refused", tt.bind)
		case !tt.accept:
			var ce *Error
			if !errors.As(err, &ce) || ce.Key != "server.bind" || !strings.Contains(err.Error(), "S03") {
				t.Errorf("bind %s: error %q should name server.bind and point to S03", tt.bind, err)
			}
		}
	}
}

// TestContainerBind: any address is allowed only inside a container and
// only with server.allow_container_bind.
func TestContainerBind(t *testing.T) {
	root := absRoot(t)
	load := func(allow string) error {
		flags := map[string]string{"storage.root": root, "server.bind": "0.0.0.0:8080"}
		if allow != "" {
			flags["server.allow_container_bind"] = allow
		}
		_, err := Load(Sources{Flags: flags})
		return err
	}
	fakeNetwork(t, nil, true)
	if err := load("true"); err != nil {
		t.Errorf("in a container with the setting: %v", err)
	}
	if err := load(""); err == nil || !strings.Contains(err.Error(), "S03") {
		t.Errorf("in a container without the setting: %v, want the loopback refusal", err)
	}
	fakeNetwork(t, nil, false)
	if err := load("true"); err == nil || !strings.Contains(err.Error(), "only inside a container") {
		t.Errorf("outside a container with the setting: %v, want a refusal that explains the setting", err)
	}
}

func TestBoolSetting(t *testing.T) {
	root := absRoot(t)
	for _, tt := range []struct {
		src     Sources
		want    bool
		wantErr string
	}{
		{Sources{Flags: map[string]string{"storage.root": root, "server.allow_container_bind": "true"}}, true, ""},
		{Sources{Flags: map[string]string{"storage.root": root}, LookupEnv: env(map[string]string{"LOCALAINAS_SERVER_ALLOW_CONTAINER_BIND": "false"})}, false, ""},
		{Sources{Flags: map[string]string{"storage.root": root, "server.allow_container_bind": "yes"}}, false, "true or false"},
	} {
		l, err := Load(tt.src)
		switch {
		case tt.wantErr != "":
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("%v: error %v, want %q", tt.src.Flags, err, tt.wantErr)
			}
		case err != nil:
			t.Errorf("%v: %v", tt.src.Flags, err)
		case l.Server.AllowContainerBind != tt.want:
			t.Errorf("%v: AllowContainerBind = %v, want %v", tt.src.Flags, l.Server.AllowContainerBind, tt.want)
		}
	}
}

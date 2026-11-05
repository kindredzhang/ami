package workspacedeps

import (
	"strings"
	"testing"
)

func TestLayoutPaths(t *testing.T) {
	home := Home("/data", "codex")
	cases := map[string]string{
		DepsRoot("/data"):           "/data/.ami/deps",
		home:                        "/data/.ami/deps/codex",
		ShimDir("/data"):            "/data/.ami/deps/bin",
		LocksDir("/data"):           "/data/.ami/deps/.locks",
		StatePath(home):             "/data/.ami/deps/codex/state.json",
		VersionsDir(home):           "/data/.ami/deps/codex/versions",
		CurrentDir(home):            "/data/.ami/deps/codex/current",
		lockPath(home, "codex"):     "/data/.ami/deps/.locks/codex.lock",
		Home("/Users/me/ws/", "uv"): "/Users/me/ws/.ami/deps/uv",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestShimScript(t *testing.T) {
	plain := ShimScript("/data/.ami/deps/uv/current/bin/uv", false)
	if !strings.HasPrefix(plain, "#!/bin/sh\n") || !strings.HasSuffix(plain, "exec '/data/.ami/deps/uv/current/bin/uv' \"$@\"\n") {
		t.Errorf("plain shim = %q", plain)
	}
	if strings.Contains(plain, "SSL_CERT_FILE") {
		t.Error("plain shim must not touch SSL_CERT_FILE")
	}
	agent := ShimScript("/data/.ami/deps/claude-code/current/bin/claude", true)
	for _, want := range []string{
		`if [ -z "${SSL_CERT_FILE:-}" ] && [ -f /opt/ami/toolkit/certs/ca-certificates.crt ]; then`,
		"export SSL_CERT_FILE=/opt/ami/toolkit/certs/ca-certificates.crt",
		`exec '/data/.ami/deps/claude-code/current/bin/claude' "$@"`,
	} {
		if !strings.Contains(agent, want) {
			t.Errorf("agent shim missing %q:\n%s", want, agent)
		}
	}
}

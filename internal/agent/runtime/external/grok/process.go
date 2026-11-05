package grok

import (
	"sync"

	acpclient "github.com/kindredzhang/ami/internal/agent/runtime/external/acp/client"
	"github.com/kindredzhang/ami/internal/agent/runtime/external/agentproc"
)

// grokProcess is what a `grok agent stdio` process reports through Grok's own
// notifications, kept beside the ACP core's state of it.
type grokProcess struct {
	conn *acpclient.Conn

	mu sync.Mutex
	// models is the catalog Grok last pushed; nil until it has.
	models map[string]any
	// toolStatus is Grok's latest report on the Ami tool gateway of each
	// session; statusWake closes, and is replaced, when a report arrives.
	toolStatus map[string]mcpStatus
	statusWake chan struct{}
}

func newGrokProcess(conn *acpclient.Conn) *grokProcess {
	p := &grokProcess{conn: conn, toolStatus: map[string]mcpStatus{}, statusWake: make(chan struct{})}
	conn.Subscribe(modelsUpdateMethod, p.updateModels)
	conn.Subscribe(mcpStatusMethod, p.updateToolStatus)
	return p
}

// processEnv is the whole environment of a grok process. apiKey is empty for
// a Grok account login, which Grok reads from its home.
func processEnv(home, apiKey string) []string {
	env := agentproc.BaseEnv(
		"GROK_HOME="+home,
		// The managed dependency owns the CLI version.
		"GROK_DISABLE_AUTOUPDATER=1",
		// Ami keeps the transcript: a stopped turn stays in Grok's history
		// as it stays in Ami's.
		"GROK_CANCEL_REWIND=0",
		// Ami shows none of these, and each costs a model call per turn.
		"GROK_TURN_SUMMARY=0",
		"GROK_SESSION_RECAP=0",
		// Ami owns memory. Grok's would carry across every conversation of
		// the Agent, which share its home.
		"GROK_MEMORY=0",
	)
	if apiKey != "" {
		env = append(env, "XAI_API_KEY="+apiKey)
	}
	return env
}

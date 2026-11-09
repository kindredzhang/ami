package tools

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"
)

// fakeExpertResolver serves expert definitions from memory.
type fakeExpertResolver struct {
	experts map[string]SpawnExpert
	listErr error
}

func (r *fakeExpertResolver) ListExperts(_ context.Context, _ string) ([]SpawnExpertSummary, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	out := make([]SpawnExpertSummary, 0, len(r.experts))
	for name, expert := range r.experts {
		out = append(out, SpawnExpertSummary{
			Name:        name,
			DisplayName: expert.DisplayName,
			Skills:      sortedKeys(expert.Skills),
			InheritsAll: !expert.SkillsScoped,
		})
	}
	return out, nil
}

func (r *fakeExpertResolver) ResolveExpert(_ context.Context, _, name string) (SpawnExpert, error) {
	expert, ok := r.experts[name]
	if !ok {
		return SpawnExpert{}, ErrExpertNotFound
	}
	return expert, nil
}

func sortedKeys(m map[string]SkillDetail) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// capturingSpawnAgent records the run config the provider hands to the runtime,
// which is the only place the expert's effect is observable end to end.
type capturingSpawnAgent struct {
	*fakeSpawnAgent
	lastConfig *SpawnRunConfig
}

func (a *capturingSpawnAgent) Generate(ctx context.Context, cfg SpawnRunConfig) (*SpawnResult, error) {
	a.lastConfig = &cfg
	return a.fakeSpawnAgent.Generate(ctx, cfg)
}

// GenerateWithWatchdog is the entry the provider actually uses for a run that
// may need a watchdog, so capturing only Generate would record nothing.
func (a *capturingSpawnAgent) GenerateWithWatchdog(ctx context.Context, cfg SpawnRunConfig, touch func()) (*SpawnResult, error) {
	a.lastConfig = &cfg
	return a.fakeSpawnAgent.GenerateWithWatchdog(ctx, cfg, touch)
}

func TestSpawnWithExpertScopesSkillsAndPrompt(t *testing.T) {
	t.Parallel()

	p, _, _, _ := newAgentControlProvider(t, &fakeSpawnAgent{})
	capture := &capturingSpawnAgent{fakeSpawnAgent: &fakeSpawnAgent{}}
	p.SetAgent(capture)
	p.SetExpertResolver(&fakeExpertResolver{experts: map[string]SpawnExpert{
		"analyst": {
			Name:          "analyst",
			DisplayName:   "Data Analyst",
			Skills:        map[string]SkillDetail{"alpha": {Description: "a", Content: "A"}},
			SkillsScoped:  true,
			SystemSection: "## Active expert: Data Analyst\n\nBe brief.",
		},
	}})
	p.SetSystemPromptFunc(func(string) string { return "base subagent prompt" })

	session := SessionContext{
		BotID:     "bot1",
		SessionID: "parent1",
		// The parent holds the bot's whole set; the expert must not inherit it.
		Skills: map[string]SkillDetail{
			"alpha": {Description: "a"},
			"beta":  {Description: "b"},
		},
	}

	if _, err := executeAgentTool(t, p, session, "spawn_agent", map[string]any{
		"task":   "count the rows",
		"expert": "analyst",
	}); err != nil {
		t.Fatalf("spawn_agent: %v", err)
	}

	if capture.lastConfig == nil {
		t.Fatal("runtime was never invoked")
	}
	cfg := *capture.lastConfig

	// Skill scoping is the core promise: exactly the expert's slice, not the
	// parent's full set.
	if len(cfg.Skills) != 1 {
		t.Fatalf("skills = %v, want only alpha", sortedKeys(cfg.Skills))
	}
	if _, ok := cfg.Skills["alpha"]; !ok {
		t.Errorf("skills = %v, want alpha", sortedKeys(cfg.Skills))
	}
	if _, ok := cfg.Skills["beta"]; ok {
		t.Error("beta came from the parent and must not reach an expert run")
	}

	// The skill provider must not reload the bot's catalog over this set.
	if !cfg.Identity.SkillsScoped {
		t.Error("SkillsScoped must be set or the skill provider widens the set")
	}
	if cfg.Identity.ExpertName != "analyst" {
		t.Errorf("ExpertName = %q, want analyst", cfg.Identity.ExpertName)
	}

	// The persona is appended to the standing subagent prompt, not substituted
	// for it: the run still needs the rules that prompt carries.
	if !strings.Contains(cfg.System, "base subagent prompt") {
		t.Errorf("expert run dropped the base prompt: %q", cfg.System)
	}
	if !strings.Contains(cfg.System, "## Active expert: Data Analyst") {
		t.Errorf("expert persona missing from prompt: %q", cfg.System)
	}
}

func TestSpawnWithInheritingExpertKeepsParentSkills(t *testing.T) {
	t.Parallel()

	// A persona-only expert scopes no skills, so it must run with the parent's
	// set. Getting this backwards would quietly strip a bot of every skill
	// whenever someone created a voice-only expert.
	p, _, _, _ := newAgentControlProvider(t, &fakeSpawnAgent{})
	capture := &capturingSpawnAgent{fakeSpawnAgent: &fakeSpawnAgent{}}
	p.SetAgent(capture)
	p.SetExpertResolver(&fakeExpertResolver{experts: map[string]SpawnExpert{
		"voice": {
			Name:          "voice",
			DisplayName:   "Voice",
			SystemSection: "## Active expert: Voice",
		},
	}})

	session := SessionContext{
		BotID:     "bot1",
		SessionID: "parent1",
		Skills:    map[string]SkillDetail{"alpha": {}, "beta": {}},
	}

	if _, err := executeAgentTool(t, p, session, "spawn_agent", map[string]any{
		"task":   "say hello",
		"expert": "voice",
	}); err != nil {
		t.Fatalf("spawn_agent: %v", err)
	}

	cfg := *capture.lastConfig
	if len(cfg.Skills) != 2 {
		t.Fatalf("skills = %v, want the parent's two", sortedKeys(cfg.Skills))
	}
	if cfg.Identity.SkillsScoped {
		t.Error("an inheriting expert must not claim its set is scoped")
	}
}

func TestSpawnUnknownExpertFailsBeforeCreatingASession(t *testing.T) {
	t.Parallel()

	p, _, sessions, _ := newAgentControlProvider(t, &fakeSpawnAgent{})
	p.SetExpertResolver(&fakeExpertResolver{experts: map[string]SpawnExpert{}})

	session := SessionContext{BotID: "bot1", SessionID: "parent1"}
	_, err := executeAgentTool(t, p, session, "spawn_agent", map[string]any{
		"task":   "do work",
		"expert": "ghost",
	})
	if err == nil {
		t.Fatal("want an error for an unknown expert")
	}
	// The message has to name the recovery step: without it the model guesses
	// another expert name instead of asking what exists.
	if !strings.Contains(err.Error(), "list_experts") {
		t.Errorf("error = %v, want it to point at list_experts", err)
	}
	if !strings.Contains(err.Error(), "ghost") {
		t.Errorf("error = %v, want it to name the expert that was asked for", err)
	}
	if len(sessions.sessions) != 0 {
		t.Errorf("created %d sessions, want none", len(sessions.sessions))
	}
}

func TestSpawnWithoutExpertIsUnaffected(t *testing.T) {
	t.Parallel()

	p, _, _, _ := newAgentControlProvider(t, &fakeSpawnAgent{})
	capture := &capturingSpawnAgent{fakeSpawnAgent: &fakeSpawnAgent{}}
	p.SetAgent(capture)
	p.SetExpertResolver(&fakeExpertResolver{experts: map[string]SpawnExpert{}})

	session := SessionContext{
		BotID:     "bot1",
		SessionID: "parent1",
		Skills:    map[string]SkillDetail{"alpha": {}},
	}
	if _, err := executeAgentTool(t, p, session, "spawn_agent", map[string]any{
		"task": "plain work",
	}); err != nil {
		t.Fatalf("spawn_agent: %v", err)
	}
	cfg := *capture.lastConfig
	if len(cfg.Skills) != 1 {
		t.Fatalf("skills = %v, want the parent's", sortedKeys(cfg.Skills))
	}
	if cfg.Identity.SkillsScoped || cfg.Identity.ExpertName != "" {
		t.Error("a plain spawn must not carry expert identity")
	}
}

func TestListExpertsToolReportsTheCatalog(t *testing.T) {
	t.Parallel()

	p, _, _, _ := newAgentControlProvider(t, &fakeSpawnAgent{})
	p.SetExpertResolver(&fakeExpertResolver{experts: map[string]SpawnExpert{
		"analyst": {Name: "analyst", DisplayName: "Data Analyst", SkillsScoped: true, Skills: map[string]SkillDetail{"alpha": {}}},
		"voice":   {Name: "voice"},
	}})

	session := SessionContext{BotID: "bot1", SessionID: "parent1"}
	out, err := executeAgentTool(t, p, session, "list_experts", map[string]any{})
	if err != nil {
		t.Fatalf("list_experts: %v", err)
	}
	payload, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("result = %T, want a map", out)
	}
	// The tool output has been through a JSON round trip, so the collection
	// arrives as generic maps rather than the typed summaries.
	experts, ok := payload["experts"].([]any)
	if !ok {
		t.Fatalf("experts = %T", payload["experts"])
	}
	if len(experts) != 2 {
		t.Fatalf("got %d experts, want 2", len(experts))
	}
	names := make([]string, 0, len(experts))
	for _, raw := range experts {
		row, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("expert row = %T", raw)
		}
		name, _ := row["name"].(string)
		names = append(names, name)
	}
	sort.Strings(names)
	if strings.Join(names, ",") != "analyst,voice" {
		t.Errorf("names = %v, want analyst,voice", names)
	}
}

func TestListExpertsWithoutResolverIsEmptyNotAnError(t *testing.T) {
	t.Parallel()

	// Deployments without expert support still expose the tool: a stable tool
	// surface matters more than hiding one that answers "none".
	p, _, _, _ := newAgentControlProvider(t, &fakeSpawnAgent{})
	session := SessionContext{BotID: "bot1", SessionID: "parent1"}
	out, err := executeAgentTool(t, p, session, "list_experts", map[string]any{})
	if err != nil {
		t.Fatalf("list_experts: %v", err)
	}
	payload := out.(map[string]any)
	// JSON numbers decode as float64 through the tool output round trip.
	if count, _ := payload["count"].(float64); count != 0 {
		t.Errorf("count = %v, want 0", payload["count"])
	}
}

func TestResolveExpertEmptyNameIsANoOp(t *testing.T) {
	t.Parallel()

	// Every plain spawn passes through the resolver; it must not error or touch
	// the resolver, or spawning an ordinary subagent would depend on the expert
	// subsystem being configured.
	p, _, _, _ := newAgentControlProvider(t, &fakeSpawnAgent{})
	p.SetExpertResolver(&fakeExpertResolver{listErr: errors.New("must not be called")})

	got, err := p.resolveExpert(context.Background(), SessionContext{BotID: "bot1"}, "  ")
	if err != nil {
		t.Fatalf("resolveExpert: %v", err)
	}
	if got.Name != "" {
		t.Errorf("got expert %q, want none", got.Name)
	}
}

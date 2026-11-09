package core

import (
	"errors"
	"testing"

	agenttools "github.com/kindredzhang/ami/internal/agent/tool"
	"github.com/kindredzhang/ami/internal/experts"
)

// These tests pin the two decisions the resolver makes that are easy to get
// wrong and expensive when wrong:
//
//  1. An expert that names skills gets exactly those, never the bot's full set.
//  2. An expert whose named skills all disappeared gets an empty set, not the
//     full set. Silently widening is the failure mode that would let an expert
//     run tools its author never granted it.

func TestNarrowExpertSkillsKeepsOnlyNamed(t *testing.T) {
	t.Parallel()

	available := map[string]agenttools.SkillDetail{
		"alpha": {Description: "a"},
		"beta":  {Description: "b"},
		"gamma": {Description: "c"},
	}

	scoped, missing := narrowExpertSkills(available, []string{"alpha", "gamma"})
	if len(scoped) != 2 {
		t.Fatalf("got %d skills, want 2: %v", len(scoped), keys(scoped))
	}
	if _, ok := scoped["beta"]; ok {
		t.Error("beta was not named and must not be included")
	}
	if len(missing) != 0 {
		t.Errorf("missing = %v, want none", missing)
	}
}

func TestNarrowExpertSkillsReportsMissingNames(t *testing.T) {
	t.Parallel()

	available := map[string]agenttools.SkillDetail{"alpha": {Description: "a"}}

	// A name the bot does not have is dropped rather than resolved, but it is
	// reported so the failure is diagnosable instead of looking like an
	// intentional empty expert.
	scoped, missing := narrowExpertSkills(available, []string{"alpha", "ghost"})
	if len(scoped) != 1 {
		t.Fatalf("got %v, want only alpha", keys(scoped))
	}
	if len(missing) != 1 || missing[0] != "ghost" {
		t.Errorf("missing = %v, want [ghost]", missing)
	}
}

func TestNarrowExpertSkillsAllMissingYieldsEmptyNotNil(t *testing.T) {
	t.Parallel()

	// Nil means "inherit everything" downstream. An expert scoped to skills that
	// all vanished must therefore produce a non-nil empty map.
	scoped, missing := narrowExpertSkills(map[string]agenttools.SkillDetail{"alpha": {}}, []string{"ghost"})
	if scoped == nil {
		t.Fatal("scoped must be non-nil so the run does not inherit the bot's skills")
	}
	if len(scoped) != 0 {
		t.Errorf("got %v, want empty", keys(scoped))
	}
	if len(missing) != 1 {
		t.Errorf("missing = %v, want one entry", missing)
	}
}

func TestNarrowExpertSkillsEmptyNamesYieldsEmptyNotNil(t *testing.T) {
	t.Parallel()

	// Callers only reach narrowing when the expert named skills, but if that
	// ever changes the result must stay non-nil: an empty named set is still a
	// deliberate scope, not permission to inherit.
	scoped, missing := narrowExpertSkills(map[string]agenttools.SkillDetail{"alpha": {}}, nil)
	if scoped == nil {
		t.Fatal("scoped must be non-nil")
	}
	if len(scoped) != 0 || len(missing) != 0 {
		t.Errorf("got %v / %v, want empty / empty", keys(scoped), missing)
	}
}

func TestExpertNotFoundIsDistinctFromInvalidName(t *testing.T) {
	t.Parallel()

	// The tool layer turns ErrExpertNotFound into "call list_experts" guidance
	// and passes everything else through. Conflating the two would make a
	// malformed name look like a missing expert, and the model would then retry
	// a name that can never work.
	if errors.Is(agenttools.ErrExpertNotFound, experts.ErrInvalidName) {
		t.Fatal("not-found must not be confused with an invalid name")
	}
	if !errors.Is(agenttools.ErrExpertNotFound, agenttools.ErrExpertNotFound) {
		t.Fatal("sentinel must match itself")
	}
}

func keys(m map[string]agenttools.SkillDetail) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

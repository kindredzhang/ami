package application

import (
	"encoding/json"
	"strings"
	"testing"
)

// instructionText decodes a message body the same way the model transport
// does: newTextContent stores a JSON string literal, not the raw text.
func instructionText(t *testing.T, msg *ModelMessage) string {
	t.Helper()
	var text string
	if err := json.Unmarshal(msg.Content, &text); err != nil {
		t.Fatalf("decode message content: %v", err)
	}
	return text
}

func TestAssignedExpertInstructionNamesTheSpawnTarget(t *testing.T) {
	msg := buildAssignedExpertInstruction(&AssignedExpert{
		Name:        "ops-analyst",
		DisplayName: "运营分析专家",
	})
	if msg == nil {
		t.Fatal("expected an instruction for a named expert")
	}
	text := instructionText(t, msg)
	for _, want := range []string{
		"运营分析专家",
		`expert="ops-analyst"`,
		"spawn_agent",
		"Do not perform the task yourself",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("instruction missing %q:\n%s", want, text)
		}
	}
}

func TestAssignedExpertInstructionFallsBackToName(t *testing.T) {
	msg := buildAssignedExpertInstruction(&AssignedExpert{Name: "ops-analyst"})
	if msg == nil {
		t.Fatal("expected an instruction for a named expert")
	}
	text := instructionText(t, msg)
	if !strings.Contains(text, `"ops-analyst"`) {
		t.Fatalf("instruction should label the expert by its id when no display name is set:\n%s", text)
	}
}

func TestAssignedExpertInstructionRefusesSilentPersonaFallback(t *testing.T) {
	// The whole point of going through spawn_agent is that a missing expert is
	// an error the user sees, not the main agent quietly answering as them.
	msg := buildAssignedExpertInstruction(&AssignedExpert{Name: "ops-analyst"})
	if msg == nil {
		t.Fatal("expected an instruction")
	}
	if !strings.Contains(instructionText(t, msg), "do not fall back to answering as that expert") {
		t.Fatalf("instruction must forbid impersonating a missing expert:\n%s", instructionText(t, msg))
	}
}

func TestAssignedExpertInstructionIsAbsentWithoutSelection(t *testing.T) {
	if msg := buildAssignedExpertInstruction(nil); msg != nil {
		t.Fatalf("no selection must produce no instruction, got %#v", msg)
	}
	if msg := buildAssignedExpertInstruction(&AssignedExpert{Name: "   "}); msg != nil {
		t.Fatalf("a blank name must produce no instruction, got %#v", msg)
	}
}

// The @ selection only works if the instruction the main agent receives names
// the same expert id the spawn tool will resolve against. A rename on one side
// without the other would produce a valid-looking instruction that fails at
// spawn time.
func TestAssignedExpertInstructionMatchesSpawnArgument(t *testing.T) {
	const expertName = "ops-analyst"
	msg := buildAssignedExpertInstruction(&AssignedExpert{Name: expertName, DisplayName: "运营分析专家"})
	if msg == nil {
		t.Fatal("expected an instruction")
	}
	text := instructionText(t, msg)
	if !strings.Contains(text, `expert="`+expertName+`"`) {
		t.Fatalf("instruction must pass the expert id as the spawn_agent expert argument:\n%s", text)
	}
}

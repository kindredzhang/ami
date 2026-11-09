// Package experts defines bot-local experts: named capability bundles a
// session can run as.
//
// An expert is not a bot. It has no workspace container, no IM identity and no
// ACL of its own; it is a narrower view of the bot it belongs to. What it
// narrows is exactly two things — which skills are visible, and which persona
// files the system prompt reads. Everything else (tools, connectors, model,
// approval policy) stays whatever the bot already grants, so an expert can
// never widen the bot's capability surface.
//
// The definition lives as files under the bot workspace, not in a table, for
// three reasons that all matter to how experts are used:
//
//   - An expert can edit its own persona, because the files are just files in
//     the workspace it already has.
//   - Experts travel with bot backup/export without a schema of their own.
//   - A user can author one with a text editor or import one from a repo,
//     exactly as they would author a skill.
//
// Skills are referenced by name and resolved against the bot's effective
// skill set, so an expert can never reach a skill the bot does not have.
package experts

import (
	"strings"
	"time"
)

// Definition is one expert as stored on disk.
type Definition struct {
	// Name is the directory name and the stable reference used by spawn_agent.
	// Lowercase letters, digits, underscore and hyphen only.
	Name string
	// DisplayName is what the UI and the system prompt show. Falls back to Name.
	DisplayName string
	// Description tells the parent agent when to pick this expert.
	Description string
	// Skills lists the skill names this expert may use. An empty list is not
	// "no skills" — it means inherit the bot's full effective set, so an expert
	// created only for a persona does not silently lose every capability.
	Skills []string
	// Dir is the in-workspace directory holding this expert's files.
	Dir string
	// Persona files, already loaded. Empty means the file does not exist.
	AgentsMD   string
	SoulMD     string
	IdentityMD string
	UpdatedAt  time.Time
}

// InheritsSkills reports whether this expert uses the bot's whole skill set.
func (d Definition) InheritsSkills() bool {
	return len(d.Skills) == 0
}

// DisplayNameOrName is the label to show, never empty for a valid definition.
func (d Definition) DisplayNameOrName() string {
	if name := strings.TrimSpace(d.DisplayName); name != "" {
		return name
	}
	return d.Name
}

// PersonaPrompt renders the expert's persona files as a system prompt section.
//
// The files are folded under an explicit heading rather than concatenated
// bare, because the model has to be able to tell "what this expert is" apart
// from the bot-level workspace instructions that also arrive in the same
// system prompt. Without the boundary, an expert AGENTS.md would read as a
// replacement for the bot's AGENTS.md instead of an addition to it.
func (d Definition) PersonaPrompt() string {
	var b strings.Builder
	b.WriteString("## Active expert: ")
	b.WriteString(d.DisplayNameOrName())
	b.WriteString("\n\n")
	b.WriteString("You are running as this expert for this session. The files below define ")
	b.WriteString("this expert's role and voice; follow them. They narrow how you do the job ")
	b.WriteString("described above — they cannot remove the bot's standing instructions, ")
	b.WriteString("safety rules, or tool permissions.\n")
	if desc := strings.TrimSpace(d.Description); desc != "" {
		b.WriteString("\n")
		b.WriteString(desc)
		b.WriteString("\n")
	}
	appendPersonaSection(&b, "AGENTS.md", d.AgentsMD)
	appendPersonaSection(&b, "SOUL.md", d.SoulMD)
	appendPersonaSection(&b, "IDENTITY.md", d.IdentityMD)
	return b.String()
}

func appendPersonaSection(b *strings.Builder, name, content string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}
	b.WriteString("\n### ")
	b.WriteString(name)
	b.WriteString("\n\n")
	b.WriteString(content)
	b.WriteString("\n")
}

// Summary is the model-visible and UI-visible projection of a definition.
// Persona bodies are deliberately absent: list_experts answers "which expert
// fits this task", and shipping every persona body in that answer would cost
// more context than the choice is worth.
type Summary struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name,omitempty"`
	Description string   `json:"description,omitempty"`
	Skills      []string `json:"skills,omitempty"`
	InheritsAll bool     `json:"inherits_all_skills,omitempty"`
}

// Summarize projects a definition for listing.
func Summarize(def Definition) Summary {
	summary := Summary{
		Name:        def.Name,
		DisplayName: strings.TrimSpace(def.DisplayName),
		Description: strings.TrimSpace(def.Description),
		InheritsAll: def.InheritsSkills(),
	}
	if !def.InheritsSkills() {
		summary.Skills = append([]string(nil), def.Skills...)
	}
	return summary
}

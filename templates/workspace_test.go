package templates

import (
	"io/fs"
	"testing"
)

// The workspace template tree mixes visible markdown files with hidden
// dotfile directories (.ami). A plain //go:embed pattern silently drops
// dotfiles, so assert the full expected layout survives embedding.
func TestWorkspaceFSContainsBootstrapFiles(t *testing.T) {
	t.Parallel()

	wantFiles := []string{
		"AGENTS.md",
		"MEMORY.md",
		"PROFILES.md",
		".ami/hooks.json",
		".ami/skills/.gitkeep",
		".ami/skills/skill-creator/SKILL.md",
		".ami/skills/hooks-setup/SKILL.md",
		".ami/skills/ami-help/SKILL.md",
	}
	fsys := WorkspaceFS()
	for _, name := range wantFiles {
		if _, err := fs.Stat(fsys, name); err != nil {
			t.Errorf("workspace template missing %q: %v", name, err)
		}
	}
}

func TestWorkspaceFSHasNoUnexpectedRoots(t *testing.T) {
	t.Parallel()

	entries, err := fs.ReadDir(WorkspaceFS(), ".")
	if err != nil {
		t.Fatalf("read workspace template root: %v", err)
	}
	want := map[string]bool{
		"AGENTS.md":   false,
		"MEMORY.md":   false,
		"PROFILES.md": false,
		".ami":        true,
	}
	for _, entry := range entries {
		isDir, ok := want[entry.Name()]
		if !ok {
			t.Errorf("unexpected root entry %q in workspace template", entry.Name())
			continue
		}
		if entry.IsDir() != isDir {
			t.Errorf("root entry %q: got isDir=%v, want %v", entry.Name(), entry.IsDir(), isDir)
		}
	}
	if len(entries) != len(want) {
		t.Errorf("workspace template root has %d entries, want %d", len(entries), len(want))
	}
}

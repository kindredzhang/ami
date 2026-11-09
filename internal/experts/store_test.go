package experts

import (
	"context"
	"errors"
	"io"
	"path"
	"strings"
	"testing"

	pb "github.com/kindredzhang/ami/internal/workspace/bridgepb"
)

// fakeClient is an in-memory workspace: enough of the bridge surface for the
// store to list and read, so the tests exercise real parsing rather than a
// mocked store.
type fakeClient struct {
	dirs  []string
	files map[string]string
}

func (c *fakeClient) ListDirAll(_ context.Context, dir string, _ bool) ([]*pb.FileEntry, error) {
	if _, ok := c.files[dir]; !ok && !c.hasDir(dir) {
		return nil, errors.New("no such directory")
	}
	entries := make([]*pb.FileEntry, 0, len(c.dirs))
	for _, d := range c.dirs {
		if path.Dir(d) != dir {
			continue
		}
		entries = append(entries, &pb.FileEntry{Path: d, IsDir: true})
	}
	return entries, nil
}

// hasDir reports whether a directory exists. A path that only appears as a
// parent of a listed directory still exists — the real bridge lists a populated
// root, it does not require the root itself to have been created explicitly.
func (c *fakeClient) hasDir(dir string) bool {
	dir = strings.TrimSuffix(dir, "/")
	for _, d := range c.dirs {
		if d == dir || strings.HasPrefix(d, dir+"/") {
			return true
		}
	}
	return false
}

func (c *fakeClient) ReadRaw(_ context.Context, filePath string) (io.ReadCloser, error) {
	content, ok := c.files[filePath]
	if !ok {
		return nil, errors.New("no such file")
	}
	return io.NopCloser(strings.NewReader(content)), nil
}

func newStoreFor(files map[string]string, dirs []string) *Store {
	return NewStore(&fakeClient{files: files, dirs: dirs}, nil)
}

func TestListReadsManifestAndPersona(t *testing.T) {
	t.Parallel()

	store := newStoreFor(map[string]string{
		"/data/experts/analyst/expert.json": `{"display_name":"Data Analyst","description":"Reads dashboards","skills":["b","a","a"]}`,
		"/data/experts/analyst/AGENTS.md":   "# Analyst\n\nBe brief.",
	}, []string{"/data/experts/analyst"})

	defs, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("got %d experts, want 1", len(defs))
	}
	def := defs[0]
	if def.Name != "analyst" {
		t.Errorf("Name = %q, want analyst", def.Name)
	}
	if def.DisplayName != "Data Analyst" {
		t.Errorf("DisplayName = %q", def.DisplayName)
	}
	// Duplicates and ordering are normalized so the prompt names a skill once.
	if got := strings.Join(def.Skills, ","); got != "a,b" {
		t.Errorf("Skills = %q, want a,b", got)
	}
	if def.InheritsSkills() {
		t.Error("expert naming skills must not report inheritance")
	}
	if !strings.Contains(def.AgentsMD, "Be brief") {
		t.Errorf("AgentsMD = %q", def.AgentsMD)
	}
}

func TestExpertWithOnlyPersonaInheritsSkills(t *testing.T) {
	t.Parallel()

	// A directory with markdown and no manifest is a valid expert. It must
	// inherit the bot's skills: an expert created only to change voice would
	// otherwise be unable to use any tool skill at all.
	store := newStoreFor(map[string]string{
		"/data/experts/voice/SOUL.md": "Warm and concise.",
	}, []string{"/data/experts/voice"})

	defs, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("got %d experts, want 1", len(defs))
	}
	if !defs[0].InheritsSkills() {
		t.Error("expert without a skills list must inherit the bot's skills")
	}
}

func TestListSkipsStrayAndBrokenDirectories(t *testing.T) {
	t.Parallel()

	// One unreadable expert must not take the whole listing down: the caller is
	// a model choosing an expert, and "none" is a worse answer than "the ones
	// that work".
	store := newStoreFor(map[string]string{
		"/data/experts/broken/expert.json": "{not json",
		"/data/experts/good/AGENTS.md":     "Fine.",
	}, []string{"/data/experts/broken", "/data/experts/good", "/data/experts/empty", "/data/experts/BadName"})

	defs, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(defs) != 1 || defs[0].Name != "good" {
		t.Fatalf("got %v, want only the readable expert", defs)
	}
}

func TestListOnMissingRootIsEmptyNotAnError(t *testing.T) {
	t.Parallel()

	store := newStoreFor(nil, nil)
	defs, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(defs) != 0 {
		t.Fatalf("got %d experts, want 0", len(defs))
	}
}

func TestGetRejectsUnsafeNames(t *testing.T) {
	t.Parallel()

	store := newStoreFor(nil, nil)
	for _, name := range []string{"", "Upper", "../escape", "with space", "a/b"} {
		if _, err := store.Get(context.Background(), name); !errors.Is(err, ErrInvalidName) {
			t.Errorf("Get(%q) error = %v, want ErrInvalidName", name, err)
		}
	}
}

func TestPersonaPromptFencesFilesUnderActiveExpert(t *testing.T) {
	t.Parallel()

	def := Definition{
		Name:        "analyst",
		DisplayName: "Data Analyst",
		Description: "Reads dashboards",
		AgentsMD:    "# Role\n\nBe brief.",
		SoulMD:      "Dry wit.",
	}
	prompt := def.PersonaPrompt()

	// The heading is load-bearing: the subagent prompt already carries the
	// bot's own instructions, and without an explicit boundary the expert's
	// AGENTS.md would read as a replacement for them.
	if !strings.Contains(prompt, "## Active expert: Data Analyst") {
		t.Fatalf("prompt missing expert heading:\n%s", prompt)
	}
	for _, want := range []string{"### AGENTS.md", "Be brief.", "### SOUL.md", "Dry wit.", "Reads dashboards"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "IDENTITY.md") {
		t.Error("prompt must not mention a file the expert does not have")
	}
}

func TestPersonaPromptKeepsHeadingWithoutFiles(t *testing.T) {
	t.Parallel()

	// A manifest-only expert still needs the heading: it tells the model which
	// expert it is acting as, which is the whole point of the section even when
	// there is no persona prose to carry it.
	if got := (Definition{Name: "x"}).PersonaPrompt(); !strings.Contains(got, "## Active expert") {
		t.Fatalf("want the heading even without files, got %q", got)
	}
}

// relativePathClient mirrors the real bridge, which reports a listed child as
// its path relative to the listed directory ("ops-analyst"), not as an absolute
// in-workspace path. The other fake returns absolute paths, so only this test
// covers the shape production actually sends.
type relativePathClient struct {
	files map[string]string
}

func (c *relativePathClient) ListDirAll(_ context.Context, dir string, _ bool) ([]*pb.FileEntry, error) {
	// One entry per child directory, keyed by the directory itself.
	seen := map[string]struct{}{}
	entries := make([]*pb.FileEntry, 0, 1)
	for filePath := range c.files {
		childDir := path.Dir(filePath)
		if path.Dir(childDir) != dir {
			continue
		}
		name := path.Base(childDir)
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		// Relative to dir, exactly as the bridge answers.
		entries = append(entries, &pb.FileEntry{Path: name, IsDir: true})
	}
	return entries, nil
}

func (c *relativePathClient) ReadRaw(_ context.Context, filePath string) (io.ReadCloser, error) {
	content, ok := c.files[filePath]
	if !ok {
		return nil, ErrNotFound
	}
	return io.NopCloser(strings.NewReader(content)), nil
}

func TestListLoadsExpertsFromRelativeEntryPaths(t *testing.T) {
	client := &relativePathClient{files: map[string]string{
		"/data/experts/ops-analyst/expert.json": `{"display_name":"运营分析专家"}`,
		"/data/experts/ops-analyst/AGENTS.md":   "先给结论，再给依据。",
	}}

	defs, err := NewStore(client, nil).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("got %d experts, want 1", len(defs))
	}
	if defs[0].Name != "ops-analyst" {
		t.Fatalf("name = %q, want ops-analyst", defs[0].Name)
	}
	if defs[0].DisplayName != "运营分析专家" {
		t.Fatalf("display name = %q", defs[0].DisplayName)
	}
	if !strings.Contains(defs[0].AgentsMD, "先给结论") {
		t.Fatalf("persona not loaded: %q", defs[0].AgentsMD)
	}
}

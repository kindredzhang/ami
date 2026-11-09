package experts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path"
	"regexp"
	"sort"
	"strings"

	pb "github.com/kindredzhang/ami/internal/workspace/bridgepb"
)

// RootDir is where experts live inside the bot workspace.
const RootDir = "/data/experts"

// manifestFile is the per-expert descriptor. Persona bodies stay in the
// markdown files next to it so an author can edit prose without touching YAML.
const manifestFile = "expert.json"

// personaFiles are read in this order when building the prompt section.
var personaFiles = []string{"AGENTS.md", "SOUL.md", "IDENTITY.md"}

// namePattern keeps expert names usable as directory names, URL segments and
// identifiers in the parent agent's tool call. It is deliberately stricter
// than what a filesystem allows: an expert name reaches the model as a token,
// and a name with a slash or a space is a name the model will mangle.
var namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

// ErrNotFound is returned when a requested expert has no directory.
var ErrNotFound = errors.New("expert not found")

// ErrInvalidName is returned for a name that cannot be an expert directory.
var ErrInvalidName = errors.New("invalid expert name")

// FileClient is the workspace file surface experts are read through. It is the
// same shape the skills package uses, so tests can drive both with one fake.
type FileClient interface {
	ListDirAll(ctx context.Context, path string, recursive bool) ([]*pb.FileEntry, error)
	ReadRaw(ctx context.Context, path string) (io.ReadCloser, error)
}

// manifest is the on-disk shape of expert.json. It is intentionally small and
// flat: anything richer belongs in the persona markdown, where the model reads
// it as prose rather than as configuration.
type manifest struct {
	DisplayName string   `json:"display_name,omitempty"`
	Description string   `json:"description,omitempty"`
	Skills      []string `json:"skills,omitempty"`
}

// Store reads experts out of one bot workspace.
type Store struct {
	client FileClient
	logger *slog.Logger
}

// NewStore builds a reader over the given workspace file client.
func NewStore(client FileClient, logger *slog.Logger) *Store {
	if logger == nil {
		logger = slog.Default()
	}
	return &Store{client: client, logger: logger.With(slog.String("component", "experts"))}
}

// ValidateName reports whether name can be an expert directory.
func ValidateName(name string) error {
	if !namePattern.MatchString(strings.TrimSpace(name)) {
		return fmt.Errorf("%w: %q", ErrInvalidName, name)
	}
	return nil
}

// List returns every expert in the workspace, sorted by name.
//
// A broken expert is logged and skipped rather than failing the whole listing:
// one malformed expert.json must not make every expert in the bot unavailable,
// and the caller (a listing for a model or a UI) has no way to act on the error
// beyond retrying.
func (s *Store) List(ctx context.Context) ([]Definition, error) {
	entries, err := s.client.ListDirAll(ctx, RootDir, false)
	if err != nil {
		// A bot that never created an expert has no directory. That is an empty
		// list, not a failure — the caller's next step (pick an expert) has the
		// same answer either way.
		return nil, nil
	}
	defs := make([]Definition, 0, len(entries))
	for _, entry := range entries {
		if !entry.GetIsDir() {
			continue
		}
		name := path.Base(strings.TrimSpace(entry.GetPath()))
		if ValidateName(name) != nil {
			continue
		}
		// The bridge reports entries relative to the listed directory, so a
		// child arrives as "ops-analyst" rather than "/data/experts/ops-analyst".
		// Loading always addresses the expert by its canonical path.
		def, loadErr := s.load(ctx, name, path.Join(RootDir, name))
		if loadErr != nil {
			if errors.Is(loadErr, ErrNotFound) {
				continue
			}
			s.logger.WarnContext(ctx, "skip unreadable expert",
				slog.String("expert", name),
				slog.Any("error", loadErr))
			continue
		}
		defs = append(defs, def)
	}
	sort.Slice(defs, func(i, j int) bool { return defs[i].Name < defs[j].Name })
	return defs, nil
}

// Get loads one expert by name.
func (s *Store) Get(ctx context.Context, name string) (Definition, error) {
	name = strings.TrimSpace(name)
	if err := ValidateName(name); err != nil {
		return Definition{}, err
	}
	return s.load(ctx, name, path.Join(RootDir, name))
}

// Summaries lists every expert in the model-facing projection.
func (s *Store) Summaries(ctx context.Context) ([]Summary, error) {
	defs, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Summary, 0, len(defs))
	for _, def := range defs {
		out = append(out, Summarize(def))
	}
	return out, nil
}

func (s *Store) load(ctx context.Context, name, dir string) (Definition, error) {
	def := Definition{Name: name, Dir: dir}

	raw, err := s.readText(ctx, path.Join(dir, manifestFile))
	switch {
	case err == nil:
		var m manifest
		if jsonErr := json.Unmarshal([]byte(raw), &m); jsonErr != nil {
			return Definition{}, fmt.Errorf("parse %s: %w", manifestFile, jsonErr)
		}
		def.DisplayName = m.DisplayName
		def.Description = m.Description
		def.Skills = normalizeSkillNames(m.Skills)
	case errors.Is(err, ErrNotFound):
		// A directory with only markdown is a valid expert: the manifest is
		// optional so a hand-written expert works without ceremony.
	default:
		return Definition{}, err
	}

	for _, file := range personaFiles {
		content, readErr := s.readText(ctx, path.Join(dir, file))
		if readErr != nil && !errors.Is(readErr, ErrNotFound) {
			return Definition{}, readErr
		}
		switch file {
		case "AGENTS.md":
			def.AgentsMD = content
		case "SOUL.md":
			def.SoulMD = content
		case "IDENTITY.md":
			def.IdentityMD = content
		}
	}

	// A directory with neither persona nor manifest is not an expert; it is
	// probably a stray directory under /data/experts.
	if def.AgentsMD == "" && def.SoulMD == "" && def.IdentityMD == "" &&
		def.DisplayName == "" && def.Description == "" && len(def.Skills) == 0 {
		return Definition{}, ErrNotFound
	}
	return def, nil
}

// readText reads a workspace file as text, mapping a missing file to
// ErrNotFound so callers can treat "absent" differently from "unreadable".
func (s *Store) readText(ctx context.Context, filePath string) (string, error) {
	rc, err := s.client.ReadRaw(ctx, filePath)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrNotFound, filePath)
	}
	defer func() { _ = rc.Close() }()
	raw, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", filePath, err)
	}
	return string(raw), nil
}

// normalizeSkillNames trims, de-duplicates and orders skill names so a
// manifest with the same skill listed twice does not produce a prompt that
// names it twice.
func normalizeSkillNames(names []string) []string {
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

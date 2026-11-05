// Command skillhub serves an internal workspace dependency registry.
//
// Ami's workspace dependency catalog speaks a small protocol: a paginated index
// of release descriptors, a release document per dependency revision, and a
// content-addressed download for each artifact. Pointing
// supermarket.catalog_base_url at this server replaces the public upstream with
// a source the operator controls — which is what an intranet deployment needs,
// because a dependency ships shell scripts, so whoever publishes it decides
// what runs on every developer machine.
//
// There is no database and no write path. Dependencies are directories on disk;
// publishing is "drop a directory in, restart". Every document is validated
// through the catalog loader's own decoders before it is served, so this server
// cannot emit something the client would reject.
//
// Layout under -data:
//
//	dependencies/<id>/dependency.yaml   manifest (authoritative)
//	dependencies/<id>/artifact.tar.gz   artifact; must contain the same manifest
//	dependencies/<id>/icon.svg          optional icon (SVG only)
//
// The manifest's `scripts` block names files inside the archive. Run
// `skillhub -check` to validate a data directory without starting a server.
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/kindredzhang/ami/internal/workspacedeps/catalog"
)

const (
	// registryID is pinned by the loader (catalog.OfficialRegistry); the path
	// is part of the protocol, so it is not configurable.
	registryID = catalog.OfficialRegistry
	// sourceDirName is the directory under -data that holds dependencies.
	sourceDirName = "dependencies"
	// maxPageLimit matches the ceiling the client requests and validates.
	maxPageLimit = 256
)

func main() {
	var (
		addr    = flag.String("addr", ":8090", "listen address")
		dataDir = flag.String("data", "./skillhub-data", "directory holding published dependencies")
		check   = flag.Bool("check", false, "validate the data directory and exit")
	)
	flag.Parse()

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	reg, err := loadRegistry(*dataDir)
	if err != nil {
		log.Error("registry failed to load", slog.String("data", *dataDir), slog.Any("error", err))
		os.Exit(1)
	}
	for _, d := range reg.deps {
		log.Info("dependency published",
			slog.String("id", d.id), slog.String("name", d.name),
			slog.String("version", d.version), slog.String("category", d.category),
			slog.String("revision", d.revision))
	}
	log.Info("registry loaded", slog.Int("dependencies", len(reg.deps)), slog.String("revision", reg.revision))

	if *check {
		fmt.Printf("ok: %d dependencies, revision %s\n", len(reg.deps), reg.revision)
		return
	}

	srv := &http.Server{
		Addr:              *addr,
		Handler:           reg.routes(log),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Info("skillhub listening", slog.String("addr", *addr))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server stopped", slog.Any("error", err))
		os.Exit(1)
	}
}

// published is one dependency that already passed the loader's validation.
type published struct {
	id       string
	name     string
	version  string
	category string
	// revision is the release document's sha256. The loader verifies metadata
	// against it, so it is derived from the bytes actually served.
	revision   string
	descriptor []byte // one element of the index "data" array
	release    []byte // the release document
	artifact   []byte // served by digest
	digest     string
	icon       []byte
	iconDigest string
}

type registry struct {
	deps     []*published
	byID     map[string]*published
	byDigest map[string]*published
	byIcon   map[string]*published
	// revision is the snapshot identity. The client refuses a paginated
	// catalog whose revision moves between pages.
	revision string
}

// loadRegistry reads and validates every dependency directory under dataDir.
//
// The validation path is the loader's own: DecodeRelease checks the release
// envelope against its revision, and FromArtifact re-derives the manifest and
// artifact digests from the archive. Anything that survives both is exactly
// what a client will accept.
func loadRegistry(dataDir string) (*registry, error) {
	root := filepath.Join(dataDir, sourceDirName)
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// An empty registry is a valid state: a site may point Ami here
			// before publishing anything. A stable revision stays valid.
			return &registry{
				byID:     map[string]*published{},
				byDigest: map[string]*published{},
				byIcon:   map[string]*published{},
				revision: snapshotRevision(nil),
			}, nil
		}
		return nil, fmt.Errorf("read %s: %w", root, err)
	}

	reg := &registry{
		byID:     map[string]*published{},
		byDigest: map[string]*published{},
		byIcon:   map[string]*published{},
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dep, err := loadDependency(filepath.Join(root, entry.Name()))
		if err != nil {
			// Name the directory: a half-finished publish is the usual cause,
			// and the operator needs to know which one to look at.
			return nil, fmt.Errorf("dependency %q: %w", entry.Name(), err)
		}
		if prev, dup := reg.byID[dep.id]; dup {
			return nil, fmt.Errorf("dependency %q is published twice (%s and %q share the id)",
				dep.id, prev.revision, entry.Name())
		}
		if prev, dup := reg.byDigest[dep.digest]; dup && prev.id != dep.id {
			return nil, fmt.Errorf("dependencies %q and %q share an artifact", prev.id, dep.id)
		}
		reg.byID[dep.id] = dep
		reg.byDigest[dep.digest] = dep
		if dep.iconDigest != "" {
			reg.byIcon[dep.iconDigest] = dep
		}
		reg.deps = append(reg.deps, dep)
	}

	sort.Slice(reg.deps, func(i, j int) bool { return reg.deps[i].id < reg.deps[j].id })
	reg.revision = snapshotRevision(reg.deps)
	return reg, nil
}

// loadDependency validates one directory and renders its documents.
func loadDependency(dir string) (*published, error) {
	manifestPath := filepath.Join(dir, catalog.ManifestFileName)
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", catalog.ManifestFileName, err)
	}
	archive, err := os.ReadFile(filepath.Join(dir, "artifact.tar.gz"))
	if err != nil {
		return nil, fmt.Errorf("read artifact.tar.gz: %w", err)
	}

	// The revision is the release document's digest, so the document is built
	// first and then handed back to the loader for verification. Deriving it
	// from anything else (the archive, the manifest) would fail DecodeRelease.
	// The icon bytes are read before the release is built: the document carries
	// the icon's digest and size, and a placeholder would be rejected by the
	// client's own validation.
	iconName, iconBytes, err := readIcon(dir, manifestBytes)
	if err != nil {
		return nil, err
	}
	release, err := buildRelease(manifestBytes, archive, iconBytes)
	if err != nil {
		return nil, err
	}
	revision := sha256Hex(release)

	// Verify end to end the way a client will, including that the manifest in
	// the release matches the one inside the archive.
	def, err := catalog.FromArtifact(release, revision, archive, "skillhub://"+dir)
	if err != nil {
		return nil, err
	}

	verified := def.Dependency()
	dep := &published{
		id:       verified.ID,
		name:     verified.Name,
		version:  verified.Version.Channel,
		category: string(verified.Category),
		revision: revision,
		release:  release,
		artifact: archive,
		digest:   sha256Hex(archive),
	}
	dep.descriptor, err = buildDescriptor(release, revision)
	if err != nil {
		return nil, err
	}

	if iconName != "" {
		dep.icon = iconBytes
		dep.iconDigest = sha256Hex(iconBytes)
	}
	return dep, nil
}

// releaseDoc mirrors the release envelope the loader decodes. Field order and
// the omission rules matter only through the revision hash, which is computed
// over whatever this function emits — the client compares bytes, not fields.
type releaseDoc struct {
	SchemaVersion  string          `json:"schema_version"`
	RegistryID     string          `json:"registry_id"`
	DependencyID   string          `json:"dependency_id"`
	Manifest       json.RawMessage `json:"manifest"`
	ManifestDigest string          `json:"manifest_digest"`
	Artifact       artifactRef     `json:"artifact"`
	Icon           *iconRef        `json:"icon,omitempty"`
}

// descriptorDoc is a release plus the revision it can be fetched at.
type descriptorDoc struct {
	releaseDoc
	Revision string `json:"revision"`
}

type artifactRef struct {
	Format           string `json:"format"`
	Digest           string `json:"digest"`
	Size             int64  `json:"size"`
	UncompressedSize int64  `json:"uncompressed_size"`
	ArchiveSize      int64  `json:"archive_size"`
	FileCount        int    `json:"file_count"`
	ContentType      string `json:"content_type"`
}

type iconRef struct {
	Digest      string `json:"digest"`
	Size        int    `json:"size"`
	ContentType string `json:"content_type"`
}

// buildRelease renders the release document for one dependency.
//
// The manifest is embedded verbatim (not re-encoded): FromArtifact compares it
// against the copy inside the archive by value, and a round-trip through a
// struct could reorder or drop fields that the archive preserves.
func buildRelease(manifestBytes, archive, iconBytes []byte) ([]byte, error) {
	manifest, err := decodeManifest(manifestBytes)
	if err != nil {
		return nil, err
	}
	// dependency.yaml is YAML, but the release envelope carries the manifest as
	// a JSON object — the loader decodes it with encoding/json and compares it
	// to the archive's YAML copy by value. Rendering that JSON here is the one
	// place the two representations meet.
	manifestJSON, err := manifestAsJSON(manifestBytes)
	if err != nil {
		return nil, err
	}

	stats, err := measureArchive(archive)
	if err != nil {
		return nil, err
	}

	doc := releaseDoc{
		SchemaVersion:  "1",
		RegistryID:     registryID,
		DependencyID:   manifest.ID,
		Manifest:       manifestJSON,
		ManifestDigest: manifestDigest(manifestBytes, archive),
		Artifact: artifactRef{
			Format: "ami_dependency_v1",
			Digest: sha256Hex(archive),
			Size:   int64(len(archive)),
			// ArchiveSize is the decompressed tar length, and UncompressedSize
			// the sum of file contents. They are different numbers: the tar
			// adds 512-byte headers and padding. The loader checks both.
			UncompressedSize: stats.contentBytes,
			ArchiveSize:      stats.tarBytes,
			FileCount:        stats.files,
			ContentType:      "application/gzip",
		},
	}

	if manifest.Icon != "" {
		doc.Icon = &iconRef{
			Digest:      sha256Hex(iconBytes),
			Size:        len(iconBytes),
			ContentType: "image/svg+xml",
		}
	}
	return marshalNoEscape(doc)
}

// buildDescriptor renders the index element: the release document plus the
// revision to fetch it at.
func buildDescriptor(release []byte, revision string) ([]byte, error) {
	var doc descriptorDoc
	if err := json.Unmarshal(release, &doc.releaseDoc); err != nil {
		return nil, err
	}
	doc.Revision = revision
	return marshalNoEscape(doc)
}

// manifestDigest reproduces the loader's dependency digest: the manifest plus
// every script file it references, hashed in name order.
//
// It is not a hash of dependency.yaml alone — the loader folds the scripts in so
// that editing an install script changes the identity. Computing it any other
// way makes every release fail "manifest identity mismatch".
func manifestDigest(manifestBytes, archive []byte) string {
	files := map[string][]byte{catalog.ManifestFileName: manifestBytes}
	// Only the manifest and the scripts it names are part of the identity.
	// Archive extras such as the icon are deliberately excluded: the loader
	// folds exactly dep.Scripts.configured() into the digest, so including the
	// icon here produces a digest no client can reproduce.
	for name, body := range archivedFiles(archive) {
		if name == catalog.ManifestFileName || strings.EqualFold(name, catalog.ManifestFileName) {
			continue
		}
		if isDeclaredScript(name, manifestBytes) {
			files[name] = body
		}
	}
	return catalog.DigestFiles(files)
}

// isDeclaredScript reports whether name is one of the script files the manifest
// configures.
func isDeclaredScript(name string, manifestBytes []byte) bool {
	scripts, err := decodeScripts(manifestBytes)
	if err != nil {
		return false
	}
	for _, candidate := range scripts {
		if candidate == name {
			return true
		}
	}
	return false
}

// decodeScripts lists the script files the manifest references, in manifest
// order.
func decodeScripts(manifestBytes []byte) ([]string, error) {
	var m struct {
		Scripts struct {
			Install     string `yaml:"install"`
			CheckUpdate string `yaml:"check_update"`
			Update      string `yaml:"update"`
			Remove      string `yaml:"remove"`
			Reinstall   string `yaml:"reinstall"`
			Version     string `yaml:"version"`
		} `yaml:"scripts"`
	}
	if err := yaml.Unmarshal(manifestBytes, &m); err != nil {
		return nil, err
	}
	var files []string
	for _, name := range []string{
		m.Scripts.Install, m.Scripts.CheckUpdate, m.Scripts.Update,
		m.Scripts.Remove, m.Scripts.Reinstall, m.Scripts.Version,
	} {
		if name != "" {
			files = append(files, name)
		}
	}
	return files, nil
}

// archivedFiles returns each regular file in the archive by name.
func archivedFiles(archive []byte) map[string][]byte {
	files := map[string][]byte{}
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return files
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err != nil {
			return files
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		body, err := io.ReadAll(tr)
		if err != nil {
			return files
		}
		files[hdr.Name] = body
	}
}

// manifestAsJSON converts the YAML manifest into the JSON object the release
// envelope carries. Numbers keep their YAML scalar form; the loader compares
// the decoded structures, not the text.
func manifestAsJSON(manifestBytes []byte) (json.RawMessage, error) {
	var raw any
	if err := yaml.Unmarshal(manifestBytes, &raw); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	encoded, err := marshalNoEscape(raw)
	if err != nil {
		return nil, fmt.Errorf("encode manifest: %w", err)
	}
	return json.RawMessage(encoded), nil
}

// readIcon returns the icon declared by the manifest, if any.
//
// The protocol accepts SVG icons only, and the client rejects any other type.
// Checking here keeps the error attached to the file an operator can fix.
func readIcon(dir string, manifestBytes []byte) (string, []byte, error) {
	manifest, err := decodeManifest(manifestBytes)
	if err != nil {
		return "", nil, err
	}
	if manifest.Icon == "" {
		return "", nil, nil
	}
	if !strings.EqualFold(filepath.Ext(manifest.Icon), ".svg") {
		return "", nil, fmt.Errorf("icon %q must be an SVG", manifest.Icon)
	}
	iconBytes, err := os.ReadFile(filepath.Join(dir, manifest.Icon))
	if err != nil {
		return "", nil, fmt.Errorf("read icon %q: %w", manifest.Icon, err)
	}
	return manifest.Icon, iconBytes, nil
}

// manifestView is the slice of the manifest this server needs to render a
// release. Everything else is passed through untouched as raw YAML.
//
// dependency.yaml is YAML, not JSON — the archive carries the same file and the
// catalog loader parses it as YAML, so it must be read the same way here.
type manifestView struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Category string `yaml:"category"`
	Icon     string `yaml:"icon"`
}

func decodeManifest(data []byte) (manifestView, error) {
	var m manifestView
	if err := yaml.Unmarshal(data, &m); err != nil {
		return m, fmt.Errorf("decode manifest: %w", err)
	}
	if strings.TrimSpace(m.ID) == "" {
		return m, errors.New("manifest has no id")
	}
	return m, nil
}

// archiveMetrics are the three size facts the protocol carries.
type archiveMetrics struct {
	// tarBytes is the decompressed tar stream length, headers and padding
	// included. The loader compares it against the whole decompressed body.
	tarBytes int64
	// contentBytes is the sum of regular file contents, excluding tar headers.
	contentBytes int64
	files        int
}

// measureArchive computes the archive metrics the release envelope reports.
// Getting tarBytes and contentBytes confused is the easiest way to publish a
// release that no client accepts, so both are derived here in one pass.
func measureArchive(archive []byte) (archiveMetrics, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return archiveMetrics{}, fmt.Errorf("artifact is not gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()

	// The decompressed length is the tar stream itself: read it once for size
	// and again through the tar reader for the per-entry facts.
	decompressed, err := io.ReadAll(io.LimitReader(gz, catalog.MaxArtifactBytes+1))
	if err != nil {
		return archiveMetrics{}, fmt.Errorf("read artifact: %w", err)
	}
	if int64(len(decompressed)) > catalog.MaxArtifactBytes {
		return archiveMetrics{}, errors.New("artifact exceeds the byte budget")
	}

	metrics := archiveMetrics{tarBytes: int64(len(decompressed))}
	tr := tar.NewReader(bytes.NewReader(decompressed))
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return archiveMetrics{}, fmt.Errorf("read artifact: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		metrics.files++
		metrics.contentBytes += hdr.Size
	}
	if metrics.files == 0 {
		return archiveMetrics{}, errors.New("artifact contains no files")
	}
	return metrics, nil
}

// snapshotRevision identifies the published set. It changes whenever any
// dependency changes, and is stable while nothing does.
func snapshotRevision(deps []*published) string {
	files := make(map[string][]byte, len(deps))
	for _, d := range deps {
		files[d.id] = []byte(registryID + "\x00" + d.revision + "\x00" + d.digest)
	}
	return strings.TrimPrefix(catalog.DigestFiles(files), "sha256:")
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func marshalNoEscape(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	// Encoder appends a newline; drop it so the revision hashes only the JSON.
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// routes wires the endpoints the catalog worker calls.
func (r *registry) routes(log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/dependencies", func(w http.ResponseWriter, req *http.Request) {
		if got := req.URL.Query().Get("registry"); got != "" && got != registryID {
			// Serving another registry's catalog would make the client cache a
			// snapshot that is not ours; 404 keeps the mismatch visible.
			http.NotFound(w, req)
			return
		}
		page, limit, err := pageParams(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		r.writeIndex(w, page, limit)
	})

	mux.HandleFunc("GET /api/registries/"+registryID+"/dependencies/{id}/releases/{revision}", func(w http.ResponseWriter, req *http.Request) {
		dep, ok := r.byID[req.PathValue("id")]
		if !ok || dep.revision != req.PathValue("revision") {
			// An unknown revision is gone, not unavailable: retrying cannot
			// help, and the client distinguishes the two.
			http.NotFound(w, req)
			return
		}
		writeJSON(w, dep.release)
	})

	mux.HandleFunc("GET /api/artifacts/dependency/{digest}", func(w http.ResponseWriter, req *http.Request) {
		dep, ok := r.byDigest[req.PathValue("digest")]
		if !ok {
			http.NotFound(w, req)
			return
		}
		writeBytes(w, "application/gzip", dep.artifact)
	})

	mux.HandleFunc("GET /api/artifacts/icon/{digest}", func(w http.ResponseWriter, req *http.Request) {
		dep, ok := r.byIcon[req.PathValue("digest")]
		if !ok {
			http.NotFound(w, req)
			return
		}
		writeBytes(w, "image/svg+xml", dep.icon)
	})

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, fmt.Appendf(nil, `{"status":"ok","dependencies":%d,"revision":%q}`, len(r.deps), r.revision))
	})

	return logRequests(mux, log)
}

// writeIndex renders one page. The client rejects a page whose data length is
// not exactly what total/limit/page imply, so the slice is computed from the
// same numbers that are reported.
func (r *registry) writeIndex(w http.ResponseWriter, page, limit int) {
	total := len(r.deps)
	start := min((page-1)*limit, total)
	end := min(start+limit, total)

	var b bytes.Buffer
	b.WriteString(`{"total":`)
	b.WriteString(strconv.Itoa(total))
	b.WriteString(`,"page":`)
	b.WriteString(strconv.Itoa(page))
	b.WriteString(`,"limit":`)
	b.WriteString(strconv.Itoa(limit))
	b.WriteString(`,"revision":"`)
	b.WriteString(r.revision)
	b.WriteString(`","data":[`)
	for i := start; i < end; i++ {
		if i > start {
			b.WriteByte(',')
		}
		b.Write(r.deps[i].descriptor)
	}
	b.WriteString(`]}`)
	writeJSON(w, b.Bytes())
}

func pageParams(req *http.Request) (int, int, error) {
	page, limit := 1, maxPageLimit
	q := req.URL.Query()
	if raw := q.Get("page"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			return 0, 0, fmt.Errorf("invalid page %q", raw)
		}
		page = v
	}
	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > maxPageLimit {
			return 0, 0, fmt.Errorf("invalid limit %q", raw)
		}
		limit = v
	}
	return page, limit, nil
}

func writeJSON(w http.ResponseWriter, body []byte) {
	writeBytes(w, "application/json", body)
}

func writeBytes(w http.ResponseWriter, contentType string, body []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	_, _ = w.Write(body)
}

func logRequests(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, req)
		if req.URL.Path == "/healthz" {
			return
		}
		log.Info("request",
			slog.String("method", req.Method), slog.String("path", req.URL.Path),
			slog.Int("status", rec.status), slog.Duration("latency", time.Since(start)))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

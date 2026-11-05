package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"github.com/kindredzhang/ami/internal/config"
	"github.com/kindredzhang/ami/internal/server"
)

// newDisabledMarketEcho builds an echo instance whose market is turned off, and
// whose upstream points at a server that fails the test if it is ever reached.
func newDisabledMarketEcho(t *testing.T) *echo.Echo {
	t.Helper()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("disabled market contacted upstream: %s", r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(upstream.Close)

	cfg := config.Config{}
	cfg.Supermarket.BaseURL = upstream.URL
	cfg.Supermarket.Disabled = true

	h := NewSupermarketHandler(slog.New(slog.DiscardHandler), cfg)
	e := echo.New()
	// The real server installs this handler; without it apperror codes are not
	// translated to their HTTP status and everything surfaces as a 500.
	e.HTTPErrorHandler = server.NewHTTPErrorHandler(slog.New(slog.DiscardHandler))
	h.Register(e)
	return e
}

// A disabled market must answer listings with an empty envelope. A failed
// request here would show users an error dialog for a source they cannot
// reach — the intranet case this switch exists for.
func TestDisabledMarketAnswersEmptyListings(t *testing.T) {
	e := newDisabledMarketEcho(t)

	for path, wantKey := range map[string]string{
		"/supermarket/skills":     "data",
		"/supermarket/apps":       "data",
		"/supermarket/categories": "data",
		"/supermarket/registries": "data",
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
			}
			var got map[string]json.RawMessage
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode body: %v (body: %s)", err, rec.Body.String())
			}
			raw, ok := got[wantKey]
			if !ok {
				t.Fatalf("response has no %q key: %s", wantKey, rec.Body.String())
			}
			var items []json.RawMessage
			if err := json.Unmarshal(raw, &items); err != nil {
				t.Fatalf("decode %q: %v", wantKey, err)
			}
			if len(items) != 0 {
				t.Fatalf("%q has %d items, want 0", wantKey, len(items))
			}
		})
	}
}

// A single-item lookup is not a listing: with the market off it must report the
// item as absent rather than inventing an empty success.
func TestDisabledMarketSingleLookupIsNotFound(t *testing.T) {
	e := newDisabledMarketEcho(t)

	req := httptest.NewRequest(http.MethodGet, "/supermarket/registries/ami/apps/foo", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

// An enabled market must still proxy — the switch must not change the default.
func TestEnabledMarketStillProxies(t *testing.T) {
	var gotPath string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total":1,"page":1,"limit":1,"data":[{"id":"kept"}]}`))
	}))
	t.Cleanup(upstream.Close)

	cfg := config.Config{}
	cfg.Supermarket.BaseURL = upstream.URL

	h := NewSupermarketHandler(slog.New(slog.DiscardHandler), cfg)
	e := echo.New()
	e.HTTPErrorHandler = server.NewHTTPErrorHandler(slog.New(slog.DiscardHandler))
	h.Register(e)

	req := httptest.NewRequest(http.MethodGet, "/supermarket/skills", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	if gotPath != "/api/skills" {
		t.Fatalf("upstream path = %q, want /api/skills", gotPath)
	}
	if body := rec.Body.String(); body != `{"total":1,"page":1,"limit":1,"data":[{"id":"kept"}]}` {
		t.Fatalf("body = %s, want the upstream payload", body)
	}
}

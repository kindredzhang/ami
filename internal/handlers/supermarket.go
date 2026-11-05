package handlers

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/kindredzhang/ami/internal/apperror"
	"github.com/kindredzhang/ami/internal/config"
	"github.com/kindredzhang/ami/internal/errs"
	supermarketclient "github.com/kindredzhang/ami/internal/supermarket"
)

// SupermarketHandler proxies the read-only Supermarket catalog. Installing
// Apps into bots is the AppsHandler's job.
type SupermarketHandler struct {
	upstream *supermarketclient.Client
	logger   *slog.Logger
	// disabled answers every listing with an empty result instead of proxying.
	// An intranet deployment points this at a curated internal source or turns
	// the market off entirely; either way it must never look broken.
	disabled bool
}

func NewSupermarketHandler(log *slog.Logger, cfg config.Config) *SupermarketHandler {
	return &SupermarketHandler{
		upstream: supermarketclient.NewClient(cfg.Supermarket.GetBaseURL(), nil),
		logger:   log.With(slog.String("handler", "supermarket")),
		disabled: cfg.Supermarket.Disabled,
	}
}

func (h *SupermarketHandler) Register(e *echo.Echo) {
	g := e.Group("/supermarket")
	g.GET("/skills", h.ListSkills)
	g.GET("/apps", h.ListApps)
	g.GET("/categories", h.ListCategories)
	g.GET("/registries", h.ListRegistries)
	g.GET("/registries/:registry_id/apps", h.ListRegistryApps)
	g.GET("/registries/:registry_id/apps/:app_id", h.GetRegistryApp)
	g.GET("/registries/:registry_id/apps/:app_id/releases/:revision", h.GetRegistryAppRelease)
	g.GET("/registries/:registry_id/apps/:app_id/skills/:skill_id", h.GetRegistrySkill)
	g.GET("/artifacts/icon/:digest", h.GetRegistrySkillIcon)
}

// emptyListings maps a listing endpoint to the empty envelope the market's
// protocol defines, so a disabled market renders "nothing published yet"
// rather than an error dialog. Paths absent from this table are real lookups
// (one app, one release) and stay 404s.
var emptyListings = map[string]string{
	"/api/skills":             `{"total":0,"page":1,"limit":0,"data":[]}`,
	"/api/apps":               `{"total":0,"page":1,"limit":0,"data":[]}`,
	"/api/categories":         `{"data":[]}`,
	"/api/registries":         `{"data":[]}`,
	"/api/dependency-catalog": `{"total":0,"page":1,"limit":0,"data":[]}`,
}

// proxy forwards a GET request to the supermarket and streams the JSON response back.
func (h *SupermarketHandler) proxy(c echo.Context, upstreamPath string) error {
	if h.disabled {
		if body, ok := emptyListings[upstreamPath]; ok {
			return c.JSONBlob(http.StatusOK, []byte(body))
		}
		return apperror.Wrap(apperror.CodeSupermarketNotFound,
			errs.New("supermarket is disabled", slog.String("path", upstreamPath)), nil)
	}
	requestPath := upstreamPath
	if qs := c.QueryString(); qs != "" {
		requestPath += "?" + qs
	}
	resp, err := h.upstream.Get(c.Request().Context(), requestPath, "application/json")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "supermarket unreachable").
			WithInternal(errs.WrapDependency(err, "proxy supermarket request", slog.String("path", upstreamPath)))
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return upstreamStatusError(resp.StatusCode, upstreamPath)
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(resp.StatusCode)
	_, _ = io.Copy(c.Response(), resp.Body)
	return nil
}

// upstreamStatusError answers a Supermarket response that is not a success
// without forwarding its body. A missing resource is unknown to the client;
// any other status is Supermarket's failure.
func upstreamStatusError(status int, upstreamPath string) error {
	attrs := []slog.Attr{slog.Int("upstream_status", status), slog.String("path", upstreamPath)}
	if status == http.StatusNotFound {
		return apperror.Wrap(apperror.CodeSupermarketNotFound, errs.New("supermarket resource not found", attrs...), nil)
	}
	return echo.NewHTTPError(http.StatusBadGateway, "supermarket request failed").
		WithInternal(errs.NewDependency("supermarket answered with an error status", attrs...))
}

// --- Supermarket upstream types (for swagger) ---

type SupermarketAuthor = supermarketclient.Author

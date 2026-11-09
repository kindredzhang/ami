package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/kindredzhang/ami/internal/bots"
	"github.com/kindredzhang/ami/internal/errs"
	"github.com/kindredzhang/ami/internal/experts"
)

type ExpertListItem struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name,omitempty"`
	Description string   `json:"description,omitempty"`
	Skills      []string `json:"skills,omitempty"`
	InheritsAll bool     `json:"inherits_all_skills,omitempty"`
}

type ExpertsResponse struct {
	Experts []ExpertListItem `json:"experts"`
}

// ListExperts godoc
// @Summary List the experts this bot can run as
// @Tags experts
// @Param bot_id path string true "Bot ID"
// @Success 200 {object} ExpertsResponse
// @Failure 400 {object} server.Problem
// @Failure 403 {object} server.Problem
// @Failure 500 {object} server.Problem
// @Router /bots/{bot_id}/experts [get].
//
// This backs the composer's @ picker. It reads the same workspace the run
// reads, so what the picker offers is exactly what spawn_agent can resolve.
func (h *ContainerdHandler) ListExperts(c echo.Context) error {
	botID, err := h.requireBotAccessWithPermission(c, bots.PermissionChat)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	store, err := h.ExpertStore(ctx, botID)
	if err != nil {
		return errs.Wrap(err, "expert store")
	}
	defs, err := store.List(ctx)
	if err != nil {
		return errs.Wrap(err, "list experts")
	}
	items := make([]ExpertListItem, 0, len(defs))
	for _, def := range defs {
		summary := experts.Summarize(def)
		items = append(items, ExpertListItem{
			Name:        summary.Name,
			DisplayName: summary.DisplayName,
			Description: summary.Description,
			Skills:      summary.Skills,
			InheritsAll: summary.InheritsAll,
		})
	}
	return c.JSON(http.StatusOK, ExpertsResponse{Experts: items})
}

// ExpertStore returns a reader over the given bot's experts.
//
// It resolves the workspace client the same way skill loading does, so an
// expert read sees the same filesystem the run does. Returning the store
// rather than a snapshot keeps the caller free to list or fetch one expert
// without paying for both.
func (h *ContainerdHandler) ExpertStore(ctx context.Context, botID string) (*experts.Store, error) {
	ctx, _, err := h.pinCurrentWorkspaceTarget(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("pin workspace target: %w", err)
	}
	client, err := h.getGRPCClient(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("workspace client: %w", err)
	}
	return experts.NewStore(client, h.logger), nil
}

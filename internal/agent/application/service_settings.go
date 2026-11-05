package application

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/kindredzhang/ami/internal/agent/runtime/native"
	"github.com/kindredzhang/ami/internal/db"
	"github.com/kindredzhang/ami/internal/settings"
)

func (s *Service) loadBotSettings(ctx context.Context, botID string) (settings.Settings, error) {
	if s.settingsService == nil {
		return settings.Settings{}, errors.New("settings service not configured")
	}
	return s.settingsService.GetBot(ctx, botID)
}

func (s *Service) loadBotRuntimeInfo(ctx context.Context, botID string) (native.BotInfo, bool) {
	info := native.BotInfo{ID: strings.TrimSpace(botID)}
	if s.queries == nil {
		return info, false
	}
	botUUID, err := db.ParseUUID(botID)
	if err != nil {
		return info, false
	}
	row, err := s.queries.GetBotByID(ctx, botUUID)
	if err != nil {
		s.logger.DebugContext(ctx, "failed to load bot metadata for loop detection",
			slog.String("bot_id", botID),
			slog.Any("error", err),
		)
		return info, false
	}
	info.Name = strings.TrimSpace(row.Name)
	if row.DisplayName.Valid {
		info.DisplayName = strings.TrimSpace(row.DisplayName.String)
	}
	if row.Timezone.Valid {
		info.Timezone = strings.TrimSpace(row.Timezone.String)
	}
	info.OwnerUserID = uuidToString(row.OwnerUserID)
	return info, parseLoopDetectionEnabledFromMetadata(row.Metadata)
}

func parseLoopDetectionEnabledFromMetadata(payload []byte) bool {
	if len(payload) == 0 {
		return false
	}
	var metadata map[string]any
	if err := json.Unmarshal(payload, &metadata); err != nil || metadata == nil {
		return false
	}
	features, ok := metadata["features"].(map[string]any)
	if !ok {
		return false
	}
	loopDetection, ok := features["loop_detection"].(map[string]any)
	if !ok {
		return false
	}
	enabled, ok := loopDetection["enabled"].(bool)
	if !ok {
		return false
	}
	return enabled
}

// loadUserMemory reads the expert owner's account-level memory from user
// metadata. It is intentionally best-effort: a failure here must never block a
// turn, so errors are logged and an empty string returned.
func (s *Service) loadUserMemory(ctx context.Context, ownerUserID string) string {
	owner := strings.TrimSpace(ownerUserID)
	if owner == "" || s.accountService == nil {
		return ""
	}
	account, err := s.accountService.Get(ctx, owner)
	if err != nil {
		s.logger.DebugContext(ctx, "failed to load account memory",
			slog.String("owner_user_id", owner),
			slog.Any("error", err))
		return ""
	}
	raw, ok := account.Metadata["user_memory"]
	if !ok {
		return ""
	}
	text, _ := raw.(string)
	return strings.TrimSpace(text)
}

// uuidToString renders a pgtype UUID as its canonical string, or "" when unset.
func uuidToString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}
	return uuid.UUID(value.Bytes).String()
}

-- name: GetBotOverlayConfig :one
SELECT
  overlay_enabled,
  overlay_provider,
  overlay_config
FROM bots
WHERE team_id = public.ami_current_team_id() AND id = $1;

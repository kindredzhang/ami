package qq

import (
	"log/slog"

	"github.com/kindredzhang/ami/internal/channel"
	"github.com/kindredzhang/ami/internal/channel/identities"
	"github.com/kindredzhang/ami/internal/channel/route"
	"github.com/kindredzhang/ami/internal/media"
)

func ProvideQQAdapter(log *slog.Logger, mediaService *media.Service, identityService *identities.Service, routeService *route.DBService) channel.Adapter {
	adapter := NewQQAdapter(log)
	adapter.SetAssetOpener(mediaService)
	adapter.SetChannelIdentityResolver(identityService)
	adapter.SetRouteResolver(routeService)
	return adapter
}

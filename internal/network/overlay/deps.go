package overlay

import (
	netctl "github.com/kindredzhang/ami/internal/network"
	"github.com/kindredzhang/ami/internal/network/overlay/internal/sidecar"
)

type ProviderDeps struct {
	SidecarRuntime sidecar.Runtime
	Runtime        netctl.RuntimeDescriptor
	StateRoot      string
}

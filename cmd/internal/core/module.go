package core

import (
	"log/slog"

	"go.uber.org/fx"

	"github.com/kindredzhang/ami/internal/accounts"
	"github.com/kindredzhang/ami/internal/acl"
	"github.com/kindredzhang/ami/internal/agent/context/compaction"
	userinput "github.com/kindredzhang/ami/internal/agent/decision/input"
	"github.com/kindredzhang/ami/internal/agentcredential"
	"github.com/kindredzhang/ami/internal/apps"
	audiopkg "github.com/kindredzhang/ami/internal/audio"
	"github.com/kindredzhang/ami/internal/boot"
	"github.com/kindredzhang/ami/internal/bots"
	"github.com/kindredzhang/ami/internal/channelaccess"
	"github.com/kindredzhang/ami/internal/chat/event"
	"github.com/kindredzhang/ami/internal/config"
	"github.com/kindredzhang/ami/internal/connectors"
	dbstore "github.com/kindredzhang/ami/internal/db/store"
	"github.com/kindredzhang/ami/internal/fetchproviders"
	"github.com/kindredzhang/ami/internal/mcp"
	memprovider "github.com/kindredzhang/ami/internal/memory/adapters"
	"github.com/kindredzhang/ami/internal/models"
	"github.com/kindredzhang/ami/internal/oauthclients"
	"github.com/kindredzhang/ami/internal/policy"
	"github.com/kindredzhang/ami/internal/providertemplates"
	"github.com/kindredzhang/ami/internal/schedule"
	"github.com/kindredzhang/ami/internal/searchproviders"
	"github.com/kindredzhang/ami/internal/supermarket"
	"github.com/kindredzhang/ami/internal/userruntime"
	videopkg "github.com/kindredzhang/ami/internal/video"
	"github.com/kindredzhang/ami/internal/workdir"
	"github.com/kindredzhang/ami/internal/workspace"
	"github.com/kindredzhang/ami/internal/workspacedeps"
)

// FoundationModule assembles process-neutral domain infrastructure shared by
// Server and Channel. It intentionally excludes Agent, workspace runtimes,
// schedulers, and provider bootstrap loops.
func FoundationModule() fx.Option {
	return fx.Options(
		fx.Provide(
			provideLogger,
			provideDBConn,
			providePostgresStore,
			provideDBQueries,
			provideAccountStore,
			bots.NewService,
			provideAccountService,
			acl.NewService,
			channelaccess.NewService,
			userinput.NewService,
			policy.NewService,
			oauthclients.NewRegistry,
			event.NewHub,
			provideSessionService,
			provideMessageService,
		),
		fx.Invoke(setupTelemetry),
		// A new human account gets a starter expert. Wired with Invoke because
		// accounts must not import bots (bots depends on accounts' concepts,
		// not the other way round).
		fx.Invoke(wireDefaultBotProvisioner),
	)
}

// wireDefaultBotProvisioner connects account creation to starter-expert
// provisioning without making the accounts package depend on bots.
func wireDefaultBotProvisioner(accountsService *accounts.Service, botsService *bots.Service) {
	accountsService.SetDefaultBotProvisioner(botsService)
}

// ServerModule assembles the Server-owned Agent and workspace runtime. It
// expects FoundationModule and the Channel catalog/runtime interfaces to be
// provided by the composing command.
func ServerModule() fx.Option {
	return fx.Options(
		fx.Provide(
			boot.ProvideRuntimeConfig,
			provideContainerService,
			provideOverlayProviderRegistry,
			provideNetworkService,
			provideNetworkController,
			provideSettingsService,
			provideBotAgentsService,
			provideToolApprovalService,
			providePGVectorStore,
			provideUserRuntimeStore,
			provideBotRemoteRuntimeBindingStore,
			provideBotWorkdirStore,
			provideUserRuntimeHub,
			userruntime.NewService,
			workspace.NewRemoteWorkspaceService,
			provideUserRuntimePipe,
			provideWikiStore,
			provideWorkspaceManager,
			provideBotWorkspaceService,
			workdir.NewService,
			provideBridgeProvider,
			provideMemoryLLM,
			memprovider.NewService,
			provideMemoryProviderRegistry,
			models.NewService,
			agentcredential.NewService,
			agentcredential.NewAuthorizationService,
			provideACPSessionPool,
			provideAgentProcesses,
			provideCodexDriver,
			provideClaudeCodeDriver,
			provideGrokDriver,
			provideRuntimeRegistry,
			provideExternalAgentLoginHandler,
			provideHooksService,
			provideProvidersService,
			provideModelCatalog,
			providertemplates.NewService,
			fetchproviders.NewService,
			searchproviders.NewService,
			mcp.NewConnectionService,
			connectors.NewService,
			connectors.NewSource,
			provideAppService,
			mcp.NewToolSessionContextStore,
			provideAudioRegistry,
			audiopkg.NewService,
			provideVideoRegistry,
			videopkg.NewService,
			provideAudioTempStore,
			provideMediaStorage,
			provideMediaService,
			provideSessionRunLedger,
			provideRuntimeFenceActivator,
			provideSessionRuntimeManager,
			provideAttentionService,
			provideDisplayService,
			provideWorkspaceDependencyCatalog,
			provideWorkspaceDependencyService,
			provideWorkspaceDependencyUpdateWorker,
			provideAgent,
			provideAgentService,
			provideTurnService,
			provideScheduleTriggerer,
			provideScheduleSessionCreator,
			schedule.NewService,
			compaction.NewService,
			provideContainerdHandler,
			provideBotBackupService,
			provideFederationGateway,
			provideFederationSource,
			provideACPToolSource,
			provideToolGatewayService,
			provideBackgroundManager,
			provideToolProviders,
			provideOAuthService,
		),
		fx.Invoke(
			injectToolProviders,
			injectBackgroundTaskEvents,
			injectACPToolProviders,
			injectBotConnectorLifecycle,
			injectBotWorkspaceIntents,
			configureMemoryProviderRegistry,
			injectScheduleBotAgents,
			injectScheduleEvents,
			startProviderTemplateSync,
			startScheduleService,
			startContainerReconciliation,
			startBotWorkspaceReconciler,
			startWorkspaceDependencyMaintenance,
			startBackgroundTaskCleanup,
			startAudioTempStoreCleanup,
		),
	)
}

func provideAppService(log *slog.Logger, cfg config.Config, queries dbstore.Queries, manager *workspace.Manager, workspaceDeps *workspacedeps.Service, connectorService *connectors.Service) *apps.Service {
	upstream := supermarket.NewClient(cfg.Supermarket.GetBaseURL(), nil)
	publisher := apps.NewSupermarketPublisher(supermarket.NewInstaller(upstream, manager, log))
	return apps.NewService(apps.Options{
		Store:        apps.NewPostgresStore(queries),
		Registry:     publisher,
		Skills:       publisher,
		Dependencies: workspaceDeps,
		Connectors:   connectorService,
		Logger:       log,
	})
}

// Module preserves the all-in-one composition API for tests and transitional
// callers. Production commands compose the two modules explicitly.
func Module() fx.Option {
	return fx.Options(FoundationModule(), ServerModule())
}

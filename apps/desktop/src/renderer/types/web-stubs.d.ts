// Type stubs for @ami/web subpath imports consumed by the renderer.
// We route typechecking through these stubs (via tsconfig `paths`) so vue-tsc
// does not recursively typecheck @ami/web's source tree. @ami/web
// owns its own types/CI. Vite ignores `paths` and resolves the real `exports`
// entries at bundle time, so runtime behavior is unchanged.

// Marker to make this file a module (not an ambient script). Required so
// that paths-mapped dynamic `import('@ami/web/...')` calls succeed.
// TS otherwise complains that the resolved file "is not a module".
export {}

declare module '@ami/web/router-guards/onboarding' {
  export function ensureOnboarding(): Promise<boolean>
  export function hasOnboardingInProgress(): boolean
}

declare module '@ami/web/routes' {
  import type { RouteRecordRaw } from 'vue-router'
  export function createAppRoutes(platform: 'web' | 'desktop'): RouteRecordRaw[]
}

declare module '@ami/web/router' {
  import type { Router } from 'vue-router'
  const router: Router
  export default router
}

declare module '@ami/web/i18n' {
  import type { I18n } from 'vue-i18n'
  import type { ComputedRef } from 'vue'
  const i18n: I18n
  export default i18n
  export function i18nRef(key: string): ComputedRef<string>
}

declare module '@ami/web/api-client' {
  export interface SetupApiClientOptions {
    baseUrl?: string
    fetch?: typeof fetch
    onUnauthorized?: () => void
  }
  export interface SdkUrlOptions {
    url: string
    path?: Record<string, unknown>
    query?: Record<string, unknown>
  }
  export function sdkAuthQuery(): { token?: string }
  export function sdkApiBaseUrl(): string
  export function sdkApiUrl(options: SdkUrlOptions): string
  export function sdkWebSocketUrl(options: SdkUrlOptions): string
  export function setupApiClient(options?: SetupApiClientOptions): void
}

declare module '@ami/web/lib/file-drop-guard' {
  export function installFileDropGuard(): void
}

declare module '@ami/web/lib/keyboard-commands' {
  export const appKeyboardCommands: {
    readonly closeCurrentWorkspaceTab: 'close-current-workspace-tab'
    readonly saveActiveFile: 'save-active-file'
    readonly toggleSidebar: 'toggle-sidebar'
    readonly openSettings: 'open-settings'
    readonly newChatSession: 'new-chat-session'
    readonly focusChatInput: 'focus-chat-input'
    readonly showSessions: 'show-sessions'
    readonly showFiles: 'show-files'
    readonly showSchedule: 'show-schedule'
    readonly showSupermarket: 'show-supermarket'
    readonly nextWorkspaceTab: 'next-workspace-tab'
    readonly previousWorkspaceTab: 'previous-workspace-tab'
    readonly splitWorkspaceRight: 'split-workspace-right'
    readonly splitWorkspaceBelow: 'split-workspace-below'
    readonly newTerminal: 'new-terminal'
    readonly newBrowser: 'new-browser'
    readonly closeMediaLightbox: 'close-media-lightbox'
    readonly mediaLightboxPrev: 'media-lightbox-prev'
    readonly mediaLightboxNext: 'media-lightbox-next'
  }
  export type AppKeyboardCommand =
    typeof appKeyboardCommands[keyof typeof appKeyboardCommands]
  export type KeyboardCommandHandler = () => boolean | void
  export type UnhandledKeyboardCommandCallback = (command: AppKeyboardCommand) => void
  export interface KeyboardCommandApi {
    onKeyboardCommand(cb: (command: AppKeyboardCommand) => void): (() => void) | void
  }
  export interface KeyboardCommandRegistry {
    register(command: AppKeyboardCommand, handler: KeyboardCommandHandler): () => void
    dispatch(command: AppKeyboardCommand): boolean
    connect(api: KeyboardCommandApi, onUnhandled?: UnhandledKeyboardCommandCallback): () => void
  }
  export function isAppKeyboardCommand(value: unknown): value is AppKeyboardCommand
  export function createKeyboardCommandRegistry(canDispatch?: (command: AppKeyboardCommand) => boolean): KeyboardCommandRegistry
}

declare module '@ami/web/lib/keyboard-context' {
  export function selectActiveKeyboardBindings<T extends { scope: string }>(bindings: T[], root?: Document): T[]
  import type { AppKeyboardCommand } from '@ami/web/lib/keyboard-commands'
  export function canDispatchKeyboardCommand(command: AppKeyboardCommand, route: { name?: unknown; path: string }, root?: Document, hasWorkspace?: boolean): boolean
}

declare module '@ami/web/lib/keyboard-bindings' {
  import type { AppKeyboardCommand } from '@ami/web/lib/keyboard-commands'
  export type DesktopDelivery = 'menu'
  export type BrowserBehavior = 'intercept' | 'passthrough'
  export type KeyboardScope = 'global' | 'workspace' | 'mediaLightbox'
  export interface KeyboardBinding {
    command: AppKeyboardCommand
    key: string
    mod?: boolean
    alt?: boolean
    shift?: boolean
    repeat?: boolean
    desktop?: DesktopDelivery
    browser: BrowserBehavior
    scope: KeyboardScope
    i18nKey: string
  }
  export const keyboardBindings: KeyboardBinding[]
  export const RESERVED_BROWSER_COMBOS: Set<string>
  export function toElectronAccelerator(binding: Pick<KeyboardBinding, 'key' | 'mod' | 'alt' | 'shift'>): string
  export function acceleratorForCommand(command: AppKeyboardCommand): string | undefined
  export function selectWebBindings(bindings: KeyboardBinding[]): KeyboardBinding[]
}

declare module '@ami/web/lib/browser-keyboard-shortcuts' {
  import type { AppKeyboardCommand, KeyboardCommandRegistry } from '@ami/web/lib/keyboard-commands'
  export interface BrowserKeyboardShortcutBinding {
    command: AppKeyboardCommand
    key: string
    mod?: boolean
    alt?: boolean
    shift?: boolean
    repeat?: boolean
  }
  export function handleBrowserKeyboardShortcut(
    event: {
      key: string
      metaKey: boolean
      ctrlKey: boolean
      altKey: boolean
      shiftKey: boolean
      preventDefault(): void
    },
    registry: Pick<KeyboardCommandRegistry, 'dispatch'>,
    bindings: BrowserKeyboardShortcutBinding[],
  ): boolean
  export function connectBrowserKeyboardShortcuts(
    registry: Pick<KeyboardCommandRegistry, 'dispatch'>,
    bindings: BrowserKeyboardShortcutBinding[],
    target?: unknown,
  ): () => void
  export function connectBrowserKeyboardShortcutsLive(
    registry: Pick<KeyboardCommandRegistry, 'dispatch'>,
    getBindings: () => BrowserKeyboardShortcutBinding[],
    target?: unknown,
  ): () => void
}

declare module '@ami/web/store/keyboard-shortcuts' {
  import type { KeyboardBinding } from '@ami/web/lib/keyboard-bindings'
  import type { AppKeyboardCommand } from '@ami/web/lib/keyboard-commands'
  export type ConflictKind = 'none' | 'same-scope' | 'cross-scope' | 'reserved' | 'invalid'
  export interface ConflictResult {
    kind: ConflictKind
    collidesWith?: AppKeyboardCommand
  }
  export function useKeyboardShortcutsStore(pinia?: unknown): {
    overrides: Record<string, string>
    effectiveBindings: KeyboardBinding[]
    isOverridden(command: AppKeyboardCommand): boolean
    detectConflict(command: AppKeyboardCommand, combo: unknown): ConflictResult
    setBinding(command: AppKeyboardCommand, combo: string): ConflictResult
    resetBinding(command: AppKeyboardCommand): void
    resetAll(): void
  }
}

declare module '@ami/web/composables/useKeyboardCommand' {
  import type { InjectionKey } from 'vue'
  import type {
    AppKeyboardCommand,
    KeyboardCommandHandler,
    KeyboardCommandRegistry,
  } from '@ami/web/lib/keyboard-commands'
  export const KEYBOARD_REGISTRY: InjectionKey<KeyboardCommandRegistry>
  export function useKeyboardCommand(command: AppKeyboardCommand, handler: KeyboardCommandHandler): void
}

declare module '@ami/web/pages/home/commands/workspace-tab-commands' {
  import type { AppKeyboardCommand, KeyboardCommandRegistry } from '@ami/web/lib/keyboard-commands'
  export interface WorkspaceTabCommandStore {
    activeId: string | null
    requestCloseTab(id: string): void
  }
  export function handleWorkspaceKeyboardCommand(
    command: AppKeyboardCommand,
    store: WorkspaceTabCommandStore,
  ): boolean
  export function registerWorkspaceTabCommands(
    registry: Pick<KeyboardCommandRegistry, 'register'>,
    store: WorkspaceTabCommandStore,
  ): () => void
}

declare module '@ami/web/store/settings' {
  export type ThemePreference = 'light' | 'dark' | 'system'
  export function useSettingsStore(pinia?: unknown): {
    theme: ThemePreference
  }
}

declare module '@ami/web/store/workspace-tabs' {
  export function useWorkspaceTabsStore(pinia?: unknown): {
    activeId: string | null
    requestCloseTab: (id: string) => void
  }
}

declare module '@ami/web/store/user' {
  export function useUserStore(): {
    userInfo: {
      role: string
    }
  }
}

declare module '@ami/web/store/chat-list' {
  // Desktop needs the bot snapshot for two things: re-pulling it when a
  // sibling window invalidates bot config (composer agent menu), and
  // mirroring the list into the native tray's Bots submenu.
  export function useChatStore(pinia?: unknown): {
    bots: Array<{
      id?: string
      name?: string
      display_name?: string
    }>
    refreshBots: () => Promise<void>
  }
}

declare module '@ami/web/store/capabilities' {
  export function useCapabilitiesStore(): {
    load: () => Promise<void>
  }
}

declare module '@ami/web/store/update' {
  export function useUpdateStore(): {
    checking: boolean
    checked: boolean
    hasUpdate: boolean
    latestVersion: string
    releaseBody: string
    releaseUrl: string
    check: () => Promise<boolean>
    checkAtStartup: () => Promise<void>
  }
}

declare module '@ami/web/composables/useDialogMutation' {
  export function useDialogMutation(): {
    run: <T>(action: () => Promise<T>, options?: { fallbackMessage?: string, onSuccess?: () => void }) => Promise<T | undefined>
  }
}

declare module '@ami/web/constants/acl-presets' {
  export const defaultAclPreset: string
  export const aclPresetOptions: Array<{ value: string, titleKey: string, descriptionKey?: string }>
}

declare module '@ami/web/utils/timezones' {
  export const emptyTimezoneValue: string
}

declare module '@ami/web/lib/desktop-shell' {
  import type { InjectionKey } from 'vue'
  export const DesktopShellKey: InjectionKey<boolean>
  export type DesktopRuntimeStatus = 'disabled' | 'connecting' | 'connected' | 'disconnected' | 'stopped' | 'error'
  export interface DesktopRuntimeState {
    enabled: boolean
    runtimeId?: string
    runtimeName?: string
    status: DesktopRuntimeStatus
    deviceName: string
    error?: string
  }
  export interface DesktopRuntimeBridge {
    runtimeState(): Promise<DesktopRuntimeState>
    configureRuntime(config: { runtimeId: string, name: string, key: string, teamId?: string } | null): Promise<DesktopRuntimeState>
    setRuntimePaused(paused: boolean): Promise<DesktopRuntimeState>
    onRuntimeStateChanged(listener: (state: DesktopRuntimeState) => void): () => void
  }
  export const DesktopRuntimeKey: InjectionKey<DesktopRuntimeBridge | undefined>
  export interface DesktopWindowBridge {
    isFullScreen(): Promise<boolean>
    onFullScreenChanged(listener: (fullScreen: boolean) => void): () => void
    setIgnoreMenuShortcuts?(ignore: boolean): Promise<void>
  }
  export const DesktopWindowKey: InjectionKey<DesktopWindowBridge | undefined>
  export type DesktopUpdateStatus =
    | 'idle'
    | 'checking'
    | 'up-to-date'
    | 'installing'
    | 'downloading'
    | 'downloaded'
    | 'error'
    | 'unavailable'
  export interface DesktopUpdateInfo {
    version: string
    platform: string
    enabled: boolean
  }
  export interface DesktopUpdateState {
    status: DesktopUpdateStatus
    autoUpdate: boolean
    currentVersion: string
    latestVersion: string | null
    progress: number | null
    error: string | null
    releaseNotes: string | null
  }
  export interface DesktopUpdateBridge {
    getInfo(): Promise<DesktopUpdateInfo>
    getState(): Promise<DesktopUpdateState>
    check(): Promise<DesktopUpdateState>
    setAutoUpdate(enabled: boolean): Promise<DesktopUpdateState>
    install(): Promise<DesktopUpdateState>
    onStateChanged(listener: (state: DesktopUpdateState) => void): () => void
  }
  export const DesktopUpdatesKey: InjectionKey<DesktopUpdateBridge | undefined>
}

declare module '@ami/web/lib/auth-session' {
  export type AuthSessionClearReason = 'login' | 'logout' | 'token-cleared' | 'unauthorized'
  export function notifyAuthSessionCleared(reason: AuthSessionClearReason): void
}

declare module '@ami/web/lib/query-cache-persistence' {
  import type { PiniaColadaPlugin, QueryCache, UseQueryEntryFilter } from '@pinia/colada'
  export const QUERY_CACHE_STORAGE_KEY: string
  export const persistableQueryFilter: UseQueryEntryFilter
  export function createQueryCachePersistencePlugin(options?: {
    key?: string
    storage?: Storage
    debounce?: number
  }): PiniaColadaPlugin
  export function whenQueryCacheRestored(): Promise<void>
  export function saveQueryCacheToDiskNow(queryCache: QueryCache, storage?: Storage): void
  export function removeQueryCacheFromDisk(storage?: Storage): void
  export function cancelPendingQueryCacheSave(): void
}

declare module '@ami/web/pages/login/transition' {
  export const LOGIN_ENTRY_ANIMATION_KEY: string
}

declare module '@ami/web/composables/useBackOr' {
  import type { ComputedRef } from 'vue'
  import type { RouteLocationRaw, Router } from 'vue-router'
  export function installBackHistory(router: Router): void
  export function useBackOr(fallback: RouteLocationRaw): () => void
  export function useBackAffordance(fallback: RouteLocationRaw): {
    onBack: () => void
    label: ComputedRef<string>
  }
}

declare module '@ami/web/style.css'

// Fallback for every Vue SFC reachable through the @ami/web/* wildcard
// export. The TS ambient-module `*` token matches multi-segment paths
// (slashes included), so this single declaration covers `pages/.../*.vue`,
// `components/.../*.vue`, `layout/.../*.vue`, etc.
declare module '@ami/web/*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<Record<string, never>, Record<string, never>, unknown>
  export default component
}

declare module '@ami/web/components/provider-icon/preload' {
  export function configureProviderIconLoader(loader: (url: string) => Promise<string>): void
}

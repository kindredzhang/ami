import {
  createRouter,
  createMemoryHistory,
  type RouteLocationNormalized,
  type RouteRecordRaw,
} from 'vue-router'
import { createAppRoutes } from '@ami/web/routes'
import { ensureOnboarding, hasOnboardingInProgress } from '@ami/web/router-guards/onboarding'
import { useUserStore } from '@ami/web/store/user'
import { installBackHistory } from '@ami/web/composables/useBackOr'

const routes: RouteRecordRaw[] = [
  {
    path: '/connect',
    name: 'ConnectServer',
    component: () => import('../connect/ConnectServer.vue'),
  },
  ...createAppRoutes('desktop'),
]

const router = createRouter({
  history: createMemoryHistory(),
  routes,
})

// Memory history keeps no readable back-stack, so back affordances rely on this
// afterEach-based tracker instead of history state. See useBackOr.
installBackHistory(router)

router.onError((error: Error) => {
  const isChunkLoadError =
    error.message.includes('Failed to fetch dynamically imported module') ||
    error.message.includes('Importing a module script failed') ||
    error.message.includes('error loading dynamically imported module')
  if (isChunkLoadError) {
    console.warn('[Router] Chunk load failed, reloading...', error.message)
    window.location.reload()
    return
  }
  throw error
})

router.beforeEach(async (to: RouteLocationNormalized) => {
  // Dev component wall: allow in dev builds, never reachable in prod.
  if (to.path.startsWith('/dev/')) {
    return import.meta.env.DEV ? true : { path: '/' }
  }

  if (to.path === '/connect') return true

  const token = localStorage.getItem('token')
  if (to.fullPath === '/login') {
    return token ? { path: '/' } : true
  }
  if (to.path.startsWith('/oauth/')) {
    return true
  }
  if (!token) {
    return { name: 'Login' }
  }
  if (to.meta.adminOnly) {
    const userStore = useUserStore()
    if (String(userStore.userInfo.role).toLowerCase() !== 'admin') {
      return { name: 'bots' }
    }
  }

  // Onboarding: redirect completed users away, let incomplete users through
  if (to.path === '/onboarding') {
    if (hasOnboardingInProgress()) return true
    const completed = await ensureOnboarding()
    return completed ? { path: '/' } : true
  }

  const completed = await ensureOnboarding()
  if (!completed) {
    return { path: '/onboarding' }
  }

  return true
})

// Dev convenience: reach the component wall from devtools without a URL bar
// (memory history). e.g. `window.__amiRouter.push('/dev/components')`.
if (import.meta.env.DEV) {
  ;(window as unknown as { __amiRouter?: typeof router }).__amiRouter = router
}

export default router

import { defineStore } from 'pinia'
import { ref } from 'vue'

// Update detection for the web build is disabled in this deployment: the
// platform is self-hosted on an intranet, where the upstream release feed is
// neither reachable nor a place we point users at. The desktop build has its
// own updater (main/updates.ts, fed by AMI_DESKTOP_UPDATE_BASE_URL).
//
// The store is kept — with its original shape — so the About page and the
// desktop renderer's ambient types keep working; nothing ever sets hasUpdate
// here, so no update affordance is ever drawn.
export const useUpdateStore = defineStore('update', () => {
  const checking = ref(false)
  const checked = ref(false)
  const hasUpdate = ref(false)
  const latestVersion = ref('')
  const releaseBody = ref('')
  const releaseUrl = ref('')

  async function check(): Promise<boolean> {
    checking.value = true
    try {
      checked.value = true
      return false
    } finally {
      checking.value = false
    }
  }

  async function checkAtStartup(): Promise<void> {
    // No-op: nothing to check against in a private deployment.
  }

  return { checking, checked, hasUpdate, latestVersion, releaseBody, releaseUrl, check, checkAtStartup }
})

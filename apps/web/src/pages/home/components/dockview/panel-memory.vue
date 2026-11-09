<template>
  <!-- Account-level memory. Scope, stated once and never re-litigated in the
       UI: this is ONE page of background that follows the signed-in user into
       every expert, not a per-expert notebook. It lives in a workspace tab
       because it is long-form writing — the user needs room to read what they
       wrote against the conversation they were just in. -->
  <div class="flex h-full min-h-0 flex-col overflow-hidden bg-background">
    <ScrollArea class="min-h-0 flex-1">
      <div class="mx-auto w-full max-w-2xl px-6 py-8">
        <header class="mb-5">
          <h1 class="text-title font-semibold text-foreground">
            {{ t('memory.title') }}
          </h1>
          <p class="mt-1.5 text-label leading-relaxed text-muted-foreground">
            {{ t('memory.description') }}
          </p>
        </header>

        <div
          v-if="loadError"
          class="mb-4"
        >
          <CalloutBanner
            variant="destructive"
            :title="loadError"
          />
        </div>

        <div class="space-y-3">
          <Skeleton
            v-if="loading"
            class="h-52 w-full rounded-lg"
          />
          <template v-else>
            <Textarea
              v-model="draft"
              :placeholder="t('memory.placeholder')"
              :rows="12"
              :maxlength="MAX_LENGTH"
              class="min-h-52 resize-y text-control leading-relaxed"
            />
            <div class="flex flex-wrap items-center justify-between gap-3">
              <span class="text-label text-muted-foreground">
                {{ t('memory.appliesTo') }}
              </span>
              <div class="flex items-center gap-3">
                <span class="text-label tabular-nums text-muted-foreground">
                  {{ t('memory.counter', { used: draft.length, max: MAX_LENGTH }) }}
                </span>
                <Button
                  size="sm"
                  :disabled="!dirty || saving"
                  :loading="saving"
                  @click="save"
                >
                  {{ t('memory.save') }}
                </Button>
              </div>
            </div>
          </template>
        </div>
      </div>
    </ScrollArea>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Button, CalloutBanner, ScrollArea, Skeleton, Textarea, toast } from '@felinic/ui'
import { getUsersMe, putUsersMe } from '@ami/sdk'
import { resolveApiErrorMessage } from '@/utils/api-error'

/** Mirrors the server-side cap in accounts (4000 runes, truncated server-side). */
const MAX_LENGTH = 4000

const { t } = useI18n()

const loading = ref(true)
const saving = ref(false)
const loadError = ref('')
const draft = ref('')
const saved = ref('')

const dirty = computed(() => draft.value !== saved.value)

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await getUsersMe({ throwOnError: true })
    const value = String(data.metadata?.user_memory ?? '')
    draft.value = value
    saved.value = value
  } catch (error) {
    loadError.value = resolveApiErrorMessage(error, t('memory.loadFailed'), { prefixFallback: true })
  } finally {
    loading.value = false
  }
}

async function save() {
  if (saving.value || !dirty.value) return
  saving.value = true
  try {
    const { data } = await putUsersMe({
      body: { metadata: { user_memory: draft.value } },
      throwOnError: true,
    })
    // Trust the server's echo: it truncates on rune boundaries, so the value we
    // show next must be the stored one, not the one we typed.
    const stored = String(data.metadata?.user_memory ?? '')
    draft.value = stored
    saved.value = stored
    toast.success(t('memory.saved'))
  } catch (error) {
    toast.error(resolveApiErrorMessage(error, t('memory.saveFailed'), { prefixFallback: true }))
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

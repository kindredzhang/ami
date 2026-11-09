<template>
  <div
    class="shrink-0 [-webkit-app-region:drag]"
    :class="compact ? 'px-4 pb-3 pt-4' : 'px-4 pb-3.5 pt-3'"
    data-native-sidebar-surface
    data-dongfeng-brand
  >
    <div class="flex min-w-0 items-center gap-3">
      <img
        v-if="logoVisible"
        src="@/assets/brand/dongfeng-logo.jpg"
        :alt="t('sidebar.brand.logoAlt')"
        class="size-8 shrink-0 rounded-full object-contain"
        width="32"
        height="32"
        @error="logoVisible = false"
      >
      <div class="min-w-0 flex-1">
        <div
          class="truncate text-title leading-tight font-semibold tracking-normal text-foreground"
          :title="t('sidebar.brand.organization')"
        >
          {{ t('sidebar.brand.organization') }}
        </div>
        <p
          class="mt-1 truncate text-label leading-tight font-medium tracking-[0.04em] text-accent-red-deep"
          :title="t('sidebar.brand.platform')"
        >
          {{ t('sidebar.brand.platform') }}
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

withDefaults(defineProps<{
  /** The mobile drawer has no native window controls, so it gets a little more air. */
  compact?: boolean
}>(), {
  compact: false,
})

const { t } = useI18n()
const logoVisible = ref(true)
</script>

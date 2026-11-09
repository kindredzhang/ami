<template>
  <span
    class="inline-flex items-center gap-1 shrink-0 rounded-sm bg-accent-red-soft px-1 py-px text-xs text-foreground"
    :title="displayName || name"
  >
    <UserRoundCheck class="size-2.5 shrink-0" />
    <span class="truncate max-w-40">{{ displayName || name }}</span>
  </span>
</template>

<script setup lang="ts">
// A compact identity token for "this output came from expert X".
//
// It exists so a reader can scan a transcript and tell expert work apart from
// the main agent's without opening anything: the name is the discriminator,
// and it stays a single token so it fits inline next to a status.
import { computed } from 'vue'
import { UserRoundCheck } from 'lucide-vue-next'

const props = defineProps<{
  name: string
  displayName?: string
}>()

const displayName = computed(() => (props.displayName ?? '').trim())
</script>

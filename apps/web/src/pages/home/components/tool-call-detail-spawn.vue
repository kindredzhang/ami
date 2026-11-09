<template>
  <div class="space-y-1.5">
    <div
      v-if="experts.length"
      class="space-y-1"
    >
      <div
        v-for="expert in experts"
        :key="expert.name"
        class="flex items-start gap-1.5 text-xs px-1 py-0.5"
      >
        <ExpertChip
          :name="expert.name"
          :display-name="expert.display_name"
        />
        <span
          v-if="expert.description"
          class="text-muted-foreground line-clamp-2 min-w-0"
        >{{ expert.description }}</span>
        <span
          v-else
          class="shrink-0 text-muted-foreground"
        >{{ expert.skills?.length ?? 0 }}</span>
      </div>
    </div>

    <div
      v-else-if="agents.length"
      class="space-y-1"
    >
      <button
        v-for="agent in agents"
        :key="agent.agent_id || agent.session_id"
        class="flex items-center gap-1.5 text-xs w-full text-left rounded-sm px-1 py-0.5 transition-colors"
        :class="agent.session_id ? 'cursor-pointer' : ''"
        @click="agent.session_id ? navigateToSession(agent.session_id) : undefined"
      >
        <CircleDot
          class="size-3 shrink-0"
          :class="statusClass(agent.status)"
        />
        <ExpertChip
          v-if="agent.expert_name"
          :name="agent.expert_name"
          :display-name="agent.expert_display_name"
        />
        <span
          v-else
          class="font-mono text-foreground shrink-0"
        >{{ agent.agent_id }}</span>
        <span class="text-muted-foreground truncate">{{ agent.status }}</span>
        <ExternalLink
          v-if="agent.session_id"
          class="size-3 text-muted-foreground/50 shrink-0 ml-auto"
        />
      </button>
    </div>

    <div
      v-else-if="result"
      class="space-y-1"
    >
      <button
        class="flex items-center gap-1.5 text-xs w-full text-left rounded-sm px-1 py-0.5 transition-colors"
        :class="result.session_id ? 'cursor-pointer' : ''"
        @click="result.session_id ? navigateToSession(result.session_id) : undefined"
      >
        <CircleDot
          class="size-3 shrink-0"
          :class="statusClass(result.status)"
        />
        <ExpertChip
          v-if="expertName"
          :name="expertName"
          :display-name="expertDisplayName"
        />
        <span
          v-else-if="result.agent_id"
          class="font-mono text-foreground shrink-0"
        >{{ result.agent_id }}</span>
        <span
          v-if="result.status"
          class="text-muted-foreground truncate"
        >{{ result.status }}</span>
        <ExternalLink
          v-if="result.session_id"
          class="size-3 text-muted-foreground/50 shrink-0 ml-auto"
        />
      </button>

      <div class="space-y-1 pt-1 border-t border-border/50">
        <p
          v-if="result.task_id"
          class="text-xs text-muted-foreground"
        >
          <span class="font-mono">{{ result.task_id }}</span>
        </p>
        <PreviewBox v-if="result.text">
          {{ result.text }}
        </PreviewBox>
        <p
          v-if="result.error"
          class="text-xs text-destructive"
        >
          {{ result.error }}
        </p>
      </div>
    </div>

    <EmptyRow v-else>
      {{ t('chat.tools.detail.noTasks') }}
    </EmptyRow>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { CircleDot, ExternalLink } from 'lucide-vue-next'
import { useI18n } from 'vue-i18n'
import { useChatStore } from '@/store/chat-list'
import type { ToolCallBlock } from '@/store/chat-list'
import { useWorkspaceTabsStore } from '@/store/workspace-tabs'
import EmptyRow from './tool-detail/empty-row.vue'
import ExpertChip from './tool-detail/expert-chip.vue'
import PreviewBox from './tool-detail/preview-box.vue'

interface ExpertRow {
  name: string
  display_name?: string
  description?: string
  skills?: string[]
}

interface AgentResult {
  agent_id?: string
  session_id?: string
  task_id?: string
  status?: string
  text?: string
  error?: string
  expert_name?: string
  expert_display_name?: string
}

const props = defineProps<{ block: ToolCallBlock }>()
const { t } = useI18n()
const chatStore = useChatStore()
const workspaceTabs = useWorkspaceTabsStore()

function resolveResult(): Record<string, unknown> | null {
  if (!props.block.result) return null
  const result = props.block.result as Record<string, unknown>
  return (result.structuredContent as Record<string, unknown>) ?? result
}

const result = computed<AgentResult | null>(() => {
  const r = resolveResult()
  if (!r) return null
  // Both listing shapes carry a collection where a single run carries its
  // fields; treating either as one run would render an empty status row.
  if (Array.isArray(r.agents) || Array.isArray(r.experts)) return null
  return r as AgentResult
})

const agents = computed<AgentResult[]>(() => {
  const r = resolveResult()
  if (!r || !Array.isArray(r.agents)) return []
  return r.agents as AgentResult[]
})

// list_experts answers a different question than spawn_agent, so it gets a
// different rendering: the reader is choosing who should do the work, which
// needs each expert's purpose, not a run status.
const experts = computed<ExpertRow[]>(() => {
  const r = resolveResult()
  if (!r || !Array.isArray(r.experts)) return []
  return (r.experts as ExpertRow[]).filter(expert => typeof expert?.name === 'string' && expert.name !== '')
})

function callArgs(): Record<string, unknown> {
  const raw = props.block.input
  return raw !== null && typeof raw === 'object' ? raw as Record<string, unknown> : {}
}

// The expert this spawn named, read from the call's arguments rather than the
// result: a running spawn has no result yet, and the reader still needs to
// know which expert is working.
const expertName = computed(() => {
  const raw = callArgs().expert
  return typeof raw === 'string' ? raw.trim() : ''
})

// The result carries the display name once the run resolved the expert; before
// that the raw name is all there is, and the chip falls back to it.
const expertDisplayName = computed(() => {
  const raw = result.value?.expert_display_name
  return typeof raw === 'string' ? raw.trim() : ''
})

function statusClass(status?: string) {
  switch (status) {
    case 'completed':
      return 'text-success'
    case 'failed':
    case 'killed':
      return 'text-destructive'
    case 'running':
      return 'text-primary'
    case 'queued':
      return 'text-muted-foreground'
    default:
      return 'text-muted-foreground/70'
  }
}

function navigateToSession(sessionId: string) {
  if (!sessionId || !chatStore.currentBotId) return
  // Open (or focus) the spawned session in the subagent region to the right of
  // the conversation; activation selects it. An expert run is titled with the
  // expert, so a workspace holding several side-by-side subagents stays
  // readable: the tab strip then says which expert each one is, not just which
  // task it was given.
  const label = expertDisplayName.value || expertName.value
  const rawTask = callArgs().task
  const task = typeof rawTask === 'string' ? rawTask.trim() : ''
  const title = label ? `${label}${task ? ` · ${task}` : ''}` : undefined
  workspaceTabs.openSubagentSession({ sessionId, title })
}
</script>

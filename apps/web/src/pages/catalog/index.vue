<template>
  <!-- 专家 → 技能. A read-only catalog: it answers "what do I get and what can
       it do", so the entries are static showcase data and nothing here touches
       the network.

       The page also carries the 应用选择 concept the product is heading
       towards: an agent picks ONE expert to execute a task, that expert brings
       its own skill set and system prompt, and an extra list of skills can be
       layered on top. Nothing is persisted yet — selection is local state that
       makes the model concrete, so the real picker can replace it in place. -->
  <div class="flex h-full min-h-0 flex-col overflow-hidden bg-background">
    <ScrollArea class="min-h-0 flex-1">
      <div class="mx-auto w-full max-w-[1180px] px-8 py-9 pb-28">
        <!-- Masthead. One sentence of framing, no slogan. -->
        <header class="mb-7">
          <h1 class="text-heading font-semibold text-foreground">
            {{ t('catalog.title') }}
          </h1>
          <p class="mt-1.5 max-w-[62ch] text-label leading-relaxed text-muted-foreground">
            {{ t('catalog.description') }}
          </p>
        </header>

        <!-- Experts -->
        <section class="mb-10">
          <div class="mb-3 flex items-baseline justify-between gap-4">
            <div class="flex items-baseline gap-2">
              <h2 class="text-control font-medium text-foreground">
                {{ t('catalog.experts.title') }}
              </h2>
              <span class="text-caption text-muted-foreground">
                {{ t('catalog.experts.hint') }}
              </span>
            </div>
            <span class="text-caption tabular-nums text-muted-foreground">
              {{ t('catalog.experts.count', { count: experts.length }) }}
            </span>
          </div>

          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-3">
            <button
              v-for="expert in experts"
              :key="expert.id"
              type="button"
              class="group flex w-full items-start gap-3 rounded-xl border-1 px-4 py-3.5 text-left outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring"
              :class="isExpertSelected(expert.id) ? SELECTED_CARD : IDLE_CARD"
              :aria-pressed="isExpertSelected(expert.id)"
              @click="toggleExpert(expert.id)"
            >
              <span
                class="flex size-9 shrink-0 items-center justify-center rounded-lg"
                :class="accentTileClass(expert.accent)"
              >
                <component
                  :is="iconFor(expert.icon)"
                  :stroke-width="1.75"
                  class="size-[18px]"
                />
              </span>

              <span class="min-w-0 flex-1">
                <span class="flex items-start justify-between gap-2">
                  <span
                    class="truncate text-control font-medium text-foreground"
                    :title="expert.name"
                  >
                    {{ expert.name }}
                  </span>
                  <Badge
                    v-if="expert.tags?.length"
                    variant="secondary"
                    class="shrink-0"
                  >
                    {{ expert.tags[0] }}
                  </Badge>
                </span>

                <span class="mt-0.5 block truncate text-caption text-muted-foreground">
                  {{ expert.summary }}
                </span>

                <span class="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-caption text-muted-foreground">
                  <span class="truncate">{{ expert.owner }}</span>
                  <span aria-hidden="true">·</span>
                  <span class="tabular-nums">{{ expert.createdAt }}</span>
                </span>

                <span class="mt-2.5 flex flex-wrap items-center gap-1.5">
                  <span class="inline-flex items-center gap-1 rounded-sm bg-muted-soft px-1.5 py-0.5 text-caption text-foreground">
                    <component
                      :is="modelIconFor(expert.model)"
                      class="size-3"
                    />
                    {{ expert.model }}
                  </span>
                  <span class="inline-flex items-center gap-1 rounded-sm bg-muted-soft px-1.5 py-0.5 text-caption text-muted-foreground">
                    <Wand2 class="size-3" />
                    {{ t('catalog.experts.skillCount', { count: expert.skills.length }) }}
                  </span>
                </span>
              </span>
            </button>
          </div>
        </section>

        <!-- Skills -->
        <section>
          <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
            <div class="flex items-baseline gap-2">
              <h2 class="text-control font-medium text-foreground">
                {{ t('catalog.skills.title') }}
              </h2>
              <span class="text-caption text-muted-foreground">
                {{ t('catalog.skills.hint') }}
              </span>
            </div>
            <SegmentedControl
              v-model="skillGroup"
              :items="groupOptions"
              :aria-label="t('catalog.skills.groupLabel')"
            />
          </div>

          <div class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-3">
            <button
              v-for="skill in visibleSkills"
              :key="skill.id"
              type="button"
              class="group flex w-full items-start gap-3 rounded-xl border-1 px-4 py-3.5 text-left outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring"
              :class="isSkillSelected(skill.id) ? SELECTED_CARD : IDLE_CARD"
              :aria-pressed="isSkillSelected(skill.id)"
              @click="toggleSkill(skill.id)"
            >
              <span
                class="flex size-9 shrink-0 items-center justify-center rounded-lg text-control"
                :class="accentTileClass(skill.accent)"
                aria-hidden="true"
              >
                {{ skill.icon }}
              </span>

              <span class="min-w-0 flex-1">
                <span class="flex items-start justify-between gap-2">
                  <span
                    class="truncate text-control font-medium text-foreground"
                    :title="skill.name"
                  >
                    {{ skill.name }}
                  </span>
                  <span class="shrink-0 text-caption tabular-nums text-muted-foreground">
                    {{ skill.version }}
                  </span>
                </span>

                <!-- Two lines then clamp: these descriptions are a sentence or
                     two, and a ragged card height makes the grid unreadable. -->
                <span class="mt-1 line-clamp-2 block text-caption leading-relaxed text-muted-foreground">
                  {{ skill.description }}
                </span>

                <span class="mt-2 flex items-center gap-1.5 text-caption text-muted-foreground">
                  <span>{{ t('catalog.publishedAt') }}</span>
                  <span class="tabular-nums">{{ skill.publishedAt }}</span>
                </span>
              </span>
            </button>
          </div>

          <p
            v-if="!visibleSkills.length"
            class="py-10 text-center text-label text-muted-foreground"
          >
            {{ t('catalog.skills.empty') }}
          </p>
        </section>
      </div>
    </ScrollArea>

    <!-- Selection dock. Pinned rather than in-flow: the selection stays visible
         while reading further down the catalog, which is the whole point of
         choosing one. It only exists once something is picked, so the page
         starts clean. -->
    <div
      v-if="hasSelection"
      class="shrink-0 border-t border-border-soft bg-background px-8 py-3"
    >
      <div class="mx-auto flex w-full max-w-[1180px] flex-wrap items-center gap-3">
        <div class="flex min-w-0 flex-1 flex-wrap items-center gap-2">
          <span class="text-caption text-muted-foreground">{{ t('catalog.selection.label') }}</span>

          <span
            v-if="selectedExpert"
            class="inline-flex max-w-[18rem] items-center gap-1.5 rounded-md bg-accent-red-soft px-2 py-1 text-caption text-accent-red-deep"
          >
            <Bot class="size-3.5 shrink-0" />
            <span class="truncate">{{ selectedExpert.name }}</span>
            <button
              type="button"
              class="shrink-0 cursor-pointer rounded-2xs text-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring"
              :aria-label="t('catalog.selection.clearExpert')"
              @click="toggleExpert(selectedExpert.id)"
            >
              <X class="size-3" />
            </button>
          </span>

          <span
            v-for="skillId in extraSkillIds"
            :key="skillId"
            class="inline-flex max-w-[16rem] items-center gap-1.5 rounded-md bg-muted-soft px-2 py-1 text-caption text-foreground"
          >
            <Wand2 class="size-3.5 shrink-0" />
            <span class="truncate">{{ skillName(skillId) }}</span>
            <button
              type="button"
              class="shrink-0 cursor-pointer rounded-2xs text-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring"
              :aria-label="t('catalog.selection.clearSkill')"
              @click="toggleSkill(skillId)"
            >
              <X class="size-3" />
            </button>
          </span>
        </div>

        <div class="flex shrink-0 items-center gap-2">
          <span class="text-caption text-muted-foreground">
            {{ t('catalog.selection.totalSkills', { count: effectiveSkillIds.length }) }}
          </span>
          <Button
            size="sm"
            variant="outline"
            @click="previewOpen = true"
          >
            {{ t('catalog.selection.preview') }}
          </Button>
          <Button
            size="sm"
            variant="ghost"
            @click="clearSelection"
          >
            {{ t('catalog.selection.clear') }}
          </Button>
        </div>
      </div>
    </div>

    <!-- Preview: what an agent would actually receive. Skill list is the union
         of the expert's own skills and the ones layered on top, which is the
         mechanic the real 应用选择 has to implement. -->
    <Dialog v-model:open="previewOpen">
      <DialogContent class="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{{ t('catalog.preview.title') }}</DialogTitle>
          <DialogDescription>
            {{ t('catalog.preview.description') }}
          </DialogDescription>
        </DialogHeader>

        <div class="mt-4 space-y-4">
          <div>
            <p class="mb-1.5 text-caption text-muted-foreground">
              {{ t('catalog.preview.expert') }}
            </p>
            <p class="text-control text-foreground">
              {{ selectedExpert?.name ?? t('catalog.preview.noExpert') }}
            </p>
          </div>

          <div>
            <p class="mb-1.5 text-caption text-muted-foreground">
              {{ t('catalog.preview.skills', { count: effectiveSkillIds.length }) }}
            </p>
            <div class="flex flex-wrap gap-1.5">
              <span
                v-for="skillId in effectiveSkillIds"
                :key="skillId"
                class="rounded-sm bg-muted-soft px-1.5 py-0.5 text-caption text-foreground"
              >
                {{ skillName(skillId) }}
              </span>
            </div>
          </div>

          <div>
            <p class="mb-1.5 text-caption text-muted-foreground">
              {{ t('catalog.preview.systemPrompt') }}
            </p>
            <p class="rounded-md bg-muted-soft px-3 py-2 text-caption leading-relaxed text-foreground">
              {{ selectedExpert?.systemPrompt ?? t('catalog.preview.noPrompt') }}
            </p>
          </div>
        </div>

        <DialogFooter class="mt-6">
          <DialogClose as-child>
            <Button variant="outline">
              {{ t('catalog.preview.close') }}
            </Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, type Component } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Badge,
  Button,
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  ScrollArea,
  SegmentedControl,
} from '@felinic/ui'
import {
  Bot,
  Cpu,
  ShieldCheck,
  Target,
  Wand2,
  X,
} from 'lucide-vue-next'
import {
  ClaudeColor,
  DeepseekColor,
  GeminiColor,
  Openai,
  QwenColor,
  ZhipuColor,
} from '@ami/icon'
import { catalogExperts, catalogSkills, type CatalogSkill } from './data'

const { t } = useI18n()

const experts = catalogExperts

const skillGroup = ref<CatalogSkill['group']>('builtin')

const groupOptions = computed(() => [
  { value: 'builtin', label: t('catalog.skills.groups.builtin') },
  { value: 'personal', label: t('catalog.skills.groups.personal') },
])

const visibleSkills = computed(() => catalogSkills.filter(skill => skill.group === skillGroup.value))

// --- 应用选择 -------------------------------------------------------------
// One expert, plus any number of skills layered on top. Both are local state:
// the page is a reference view, not a form that saves. Splitting the two lets
// the dock show "the expert's own skills" and "the ones you added" differently
// without the state knowing about each other.

const selectedExpertId = ref<string | null>(null)
/** Skill ids picked on top of the expert's own set. */
const extraSkillIds = ref<string[]>([])
const previewOpen = ref(false)

const selectedExpert = computed(() =>
  experts.find(expert => expert.id === selectedExpertId.value) ?? null,
)

/** The expert's own skills first, then the extras — no duplicates. */
const effectiveSkillIds = computed(() => {
  const ids = [...(selectedExpert.value?.skills ?? [])]
  for (const id of extraSkillIds.value) {
    if (!ids.includes(id)) ids.push(id)
  }
  return ids
})

const hasSelection = computed(() => selectedExpertId.value !== null || extraSkillIds.value.length > 0)

function isExpertSelected(id: string): boolean {
  return selectedExpertId.value === id
}

/** Clicking the selected expert again deselects it; picking another swaps. */
function toggleExpert(id: string) {
  selectedExpertId.value = selectedExpertId.value === id ? null : id
}

function isSkillSelected(id: string): boolean {
  return extraSkillIds.value.includes(id)
}

function toggleSkill(id: string) {
  const index = extraSkillIds.value.indexOf(id)
  if (index >= 0) extraSkillIds.value.splice(index, 1)
  else extraSkillIds.value.push(id)
}

function clearSelection() {
  selectedExpertId.value = null
  extraSkillIds.value = []
}

function skillName(id: string): string {
  return catalogSkills.find(skill => skill.id === id)?.name ?? id
}

// Card chrome. The selected state borrows the sidebar's active language — a
// red hairline plus the soft red wash — so "chosen" looks the same everywhere
// in the app. The idle state is the plain card edge.
// ui-allow-style: the catalog owns a local selectable-card system — the same
// posture as the sidebar rows and PersonaTile, whose hover chrome also lives
// with its owner rather than in a shared component.
const IDLE_CARD = 'border-border bg-card hover:bg-[color:var(--ui-hover)]' /* ui-allow-style */
const SELECTED_CARD = 'border-accent-red-border bg-accent-red-soft'

// --- icons ---------------------------------------------------------------

/** Lucide glyph per expert entry. Unknown keys fall back to a neutral bot mark. */
const EXPERT_ICONS: Record<string, Component> = {
  Bot,
  Cpu,
  ShieldCheck,
  Target,
}

function iconFor(key: string): Component {
  return EXPERT_ICONS[key] ?? Bot
}

/**
 * Model brand glyph. Match on a prefix so "openai/gpt-5", "OpenAI · o3" and
 * "deepseek-v4" all resolve, and fall back to a generic mark rather than an
 * empty slot when a name is unfamiliar.
 */
const MODEL_ICONS: Array<[RegExp, Component]> = [
  [/openai|gpt/i, Openai],
  [/claude|anthropic/i, ClaudeColor],
  [/deepseek/i, DeepseekColor],
  [/gemini|google/i, GeminiColor],
  [/qwen|通义/i, QwenColor],
  [/glm|zhipu|智谱/i, ZhipuColor],
]

function modelIconFor(model: string): Component {
  return MODEL_ICONS.find(([pattern]) => pattern.test(model))?.[1] ?? Bot
}

/**
 * Accent tiles. Each entry is a soft wash + matching ink so the glyph stays
 * legible in both schemes; the washes come from the shared soft tokens, never
 * a hand-written `/NN` alpha.
 */
const ACCENT_TILES: Record<string, string> = {
  purple: 'bg-accent-purple-soft text-accent-purple-deep',
  green: 'bg-accent-green-soft text-accent-green-deep',
  blue: 'bg-accent-blue-soft text-accent-blue-deep',
  orange: 'bg-accent-orange-soft text-accent-orange-deep',
  red: 'bg-accent-red-soft text-accent-red-deep',
  teal: 'bg-accent-teal-soft text-accent-teal-deep',
}

function accentTileClass(accent: string): string {
  return ACCENT_TILES[accent] ?? 'bg-muted-soft text-muted-foreground'
}
</script>

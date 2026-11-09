<!-- eslint-disable vue/no-mutating-props -->
<template>
  <!-- id is a scroll anchor: the Overview "choose a model" reminder navigates
       here with ?section=interaction, and bot-settings.vue scrolls to it. -->
  <SettingsSection
    id="settings-section-interaction"
    :title="$t('bots.settings.blocks.interaction')"
  >
    <!-- The default-agent picker is gone: this deployment has no external
         agents (ACP), so Ami is the only runtime and an empty
         default_bot_agent_id already means "Ami". Nothing to choose. -->
    <SettingsRow
      :label="$t('bots.settings.chatModel')"
      stack="sm"
    >
      <div class="w-full sm:w-56">
        <ModelSelect
          v-model="form.chat_model_id"
          v-model:reasoning-effort="form.reasoning_effort"
          :models="models"
          :providers="providers"
          model-type="chat"
          :placeholder="$t('bots.settings.chatModelPlaceholder')"
          show-reasoning
        />
      </div>
    </SettingsRow>

    <SettingsRow
      :label="$t('bots.settings.showToolCallsInIM')"
    >
      <Switch
        :model-value="form.show_tool_calls_in_im"
        @update:model-value="(val) => form.show_tool_calls_in_im = !!val"
      />
    </SettingsRow>
  </SettingsSection>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { SettingsRow, SettingsSection, Switch } from '@felinic/ui'
import ModelSelect from './model-select.vue'
import { reconcileStoredEffort } from './reasoning-effort'
import type { AcpprofilePublicProfile, BotagentsBotAgent, SettingsSettings, ModelsGetResponse, ProvidersGetResponse } from '@ami/sdk'

type InteractionSettingsForm = SettingsSettings & {
  chat_runtime: string
  chat_acp_agent_id: string
  chat_acp_project_path: string
  chat_acp_project_mode: string
  default_bot_agent_id: string
}

const props = defineProps<{
  form: InteractionSettingsForm
  models: ModelsGetResponse[]
  providers: ProvidersGetResponse[]
  botAgents: BotagentsBotAgent[]
  acpProfiles: AcpprofilePublicProfile[]
}>()

const chatModelReasoning = computed(() => {
  if (!props.form.chat_model_id) return undefined
  return props.models.find((m) => m.id === props.form.chat_model_id)?.reasoning
})

// Switching models can strand the stored tier on a model that does not offer it.
// A model with no thinking at all has nothing to migrate to, so it is left alone
// rather than cleared — the stored value is still meaningful if the user
// switches back.
watch(chatModelReasoning, (options) => {
  if (!options?.supported) return
  const current = props.form.reasoning_effort ?? ''
  const next = reconcileStoredEffort(current, options)
  if (next === current) return
  // eslint-disable-next-line vue/no-mutating-props
  props.form.reasoning_effort = next
}, { immediate: true })
</script>

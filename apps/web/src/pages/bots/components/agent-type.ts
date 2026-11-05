import { externalAgentDisplayName, normalizeAgentID } from '@/utils/external-agent'
import type { SegmentedItem } from '@felinic/ui'
import type { AcpprofilePublicProfile } from '@ami/sdk'
import { DIRECT_BOT_AGENT_RUNTIMES, isDirectBotAgentRuntime } from '@/utils/bot-agent'

// AMI_AGENT_VALUE is the built-in-agent segment's value in agent-type-pill.
// It is NOT an agent id — a reserved marker so the chooser can never collide
// with a normalized ACP profile id.
export const AMI_AGENT_VALUE = 'ami'

// agentTypeItems builds the pill's segments: the built-in Ami agent always
// first, then the direct external agent runtimes (they have no ACP profile),
// then one segment per ACP profile the server publishes.
export function agentTypeItems(profiles: AcpprofilePublicProfile[]): SegmentedItem<string>[] {
  const direct = DIRECT_BOT_AGENT_RUNTIMES.map(runtime => ({
    value: runtime,
    label: externalAgentDisplayName(runtime, runtime),
  }))
  const agents = profiles.flatMap((profile) => {
    const agentId = normalizeAgentID(profile.id)
    // Skip entries that fail normalization or would collide with the reserved
    // built-in segment value or a direct runtime name.
    if (!agentId || agentId === AMI_AGENT_VALUE || isDirectBotAgentRuntime(agentId)) return []
    return [{ value: agentId, label: profile.display_name || externalAgentDisplayName(agentId, agentId) }]
  })
  return [{ value: AMI_AGENT_VALUE, label: 'Ami' }, ...direct, ...agents]
}

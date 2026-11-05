import { describe, expect, it } from 'vitest'
import type { AcpprofilePublicProfile } from '@ami/sdk'
import { AMI_AGENT_VALUE, agentTypeItems } from './agent-type'

function profile(id: string, displayName = ''): AcpprofilePublicProfile {
  return { id, display_name: displayName } as AcpprofilePublicProfile
}

describe('agentTypeItems', () => {
  // A profile that normalizes to the reserved built-in value would replace the
  // Ami segment's identity, and one that normalizes to a direct runtime name
  // would duplicate that segment — both are skipped, never rendered.
  it('skips profiles that collide with the reserved Ami segment or a direct runtime', () => {
    const items = agentTypeItems([profile('Ami'), profile('Claude-Code'), profile('gemini')])
    expect(items.map(item => item.value)).toEqual([AMI_AGENT_VALUE, 'codex', 'claude-code', 'grok-build', 'gemini'])
  })
})

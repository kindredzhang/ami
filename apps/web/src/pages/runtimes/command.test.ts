import { describe, expect, it } from 'vitest'

import { buildRuntimeConnectCommand, detectRuntimeHostPlatform } from './command'

const key = `mrk_${'a'.repeat(64)}`
const teamId = '11111111-1111-4111-8111-111111111111'

describe('buildRuntimeConnectCommand', () => {
  it('includes the credential team ID required by hosted gateways', () => {
    expect(buildRuntimeConnectCommand('https://ami.example/api', {
      key,
      team_id: teamId,
    })).toBe(
      `npm install -g @ami/runtime@latest && ami-runtime enroll --server https://ami.example/api --key ${key} --team-id ${teamId} && ami-runtime service install && ami-runtime service start`,
    )
  })

  it('keeps credentials from older self-hosted servers usable', () => {
    expect(buildRuntimeConnectCommand('https://ami.example/api', { key }))
      .toBe(`npm install -g @ami/runtime@latest && ami-runtime enroll --server https://ami.example/api --key ${key} && ami-runtime service install && ami-runtime service start`)
  })

  it('enables plaintext WebSockets only for loopback development servers', () => {
    expect(buildRuntimeConnectCommand('http://127.0.0.1:18080', {
      key,
      team_id: teamId,
    })).toBe(
      `npm install -g @ami/runtime@latest && ami-runtime enroll --server http://127.0.0.1:18080 --key ${key} --team-id ${teamId} --insecure-localhost && ami-runtime service install && ami-runtime service start`,
    )
  })

  it('replaces saved enrollment only when rebinding is explicitly selected', () => {
    const credential = { key, team_id: teamId }
    const firstConnection = buildRuntimeConnectCommand('http://localhost:18083/api', credential)
    const rebind = buildRuntimeConnectCommand('http://localhost:18083/api', credential, 'replace')
    expect(firstConnection).not.toContain('--replace')
    expect(rebind).toBe(firstConnection.replace(
      '--insecure-localhost &&', '--insecure-localhost --replace &&',
    ))
  })

  it('runs the same chain through cmd on Windows', () => {
    const credential = { key, team_id: teamId }
    const unix = buildRuntimeConnectCommand('https://ami.example/api', credential)
    expect(buildRuntimeConnectCommand('https://ami.example/api', credential, 'connect', 'linux')).toBe(unix)
    expect(buildRuntimeConnectCommand('https://ami.example/api', credential, 'connect', 'windows')).toBe(`cmd /c "${unix}"`)
  })
})

describe('detectRuntimeHostPlatform', () => {
  it('guesses the platform from the browser and falls back to macOS', () => {
    expect(detectRuntimeHostPlatform('Mozilla/5.0 (Windows NT 10.0; Win64; x64)')).toBe('windows')
    expect(detectRuntimeHostPlatform('Mozilla/5.0 (X11; Linux x86_64)')).toBe('linux')
    expect(detectRuntimeHostPlatform('Mozilla/5.0 (Linux; Android 14)')).toBe('macos')
    expect(detectRuntimeHostPlatform('Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)')).toBe('macos')
  })
})

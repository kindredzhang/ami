import { describe, expect, it, vi } from 'vitest'

// The registry pulls in the i18n module, which reads localStorage at import
// time. Node has none, so it is stubbed before the import is evaluated.
vi.hoisted(() => {
  for (const name of ['localStorage', 'sessionStorage']) {
    Object.defineProperty(globalThis, name, {
      configurable: true,
      value: {
        getItem: () => null,
        setItem: () => {},
        removeItem: () => {},
        clear: () => {},
      },
    })
  }
})

const { getToolDisplay, getToolTitle, toolBucket, toolFragmentKind } =
  await import('./tool-call-registry')

// These pin the parts of the expert UI a reader actually depends on: an expert
// spawn must be distinguishable from a plain spawn, and a list_experts call
// must not be filed under "other" where it reads as noise.

// getToolDisplay takes a whole ToolCallBlock; this supplies the two fields the
// presentation reads so each test states only what it is about.
function blockFor(toolName: string, input: Record<string, unknown>) {
  return {
    toolName,
    input,
    result: undefined,
    running: false,
    tool_call_id: 'call_1',
    done: true,
  } as never
}

function displayFor(toolName: string, input: Record<string, unknown>) {
  return getToolDisplay(blockFor(toolName, input))
}

// The real translator resolves against the app's i18n bundle; a key-echoing
// stand-in keeps the test on the key selection this change is about.
function titleFor(toolName: string, input: Record<string, unknown>) {
  return getToolTitle(blockFor(toolName, input), key => key)
}

describe('expert tool presentation', () => {
  it('files list_experts with the agent tools', () => {
    // Filing it elsewhere would split the agent story across two summary groups
    // in the collapsed header.
    expect(toolBucket('list_experts')).toBe('agent')
    expect(toolFragmentKind('list_experts')).toBe('agents')
  })

  it('keeps spawn_agent in the agent bucket', () => {
    expect(toolBucket('spawn_agent')).toBe('agent')
    expect(toolFragmentKind('spawn_agent')).toBe('agents')
  })
})

describe('expert spawn presentation', () => {
  it('labels an expert spawn differently from a plain spawn', () => {
    const plain = titleFor('spawn_agent', { task: 'do work' }).action
    const withExpert = titleFor('spawn_agent', { task: 'do work', expert: 'analyst' }).action
    expect(withExpert).toBe('chat.tools.spawn_expert')
    expect(plain).toBe('chat.tools.spawn_agent')
  })

  it('shows the expert name as the target', () => {
    const display = displayFor('spawn_agent', {
      task: 'count rows',
      expert: 'analyst',
      id: 'agent_3',
    })
    // The expert is the stable half of the call; the task varies every time.
    expect(display.target).toBe('analyst')
  })

  it('falls back to the agent id when no expert is named', () => {
    const display = displayFor('spawn_agent', { task: 'count rows', id: 'agent_3' })
    expect(display.target).toBe('agent_3')
  })

  it('threads fork and background through for an expert spawn', () => {
    expect(displayFor('spawn_agent', { task: 't', expert: 'e', fork: true }).actionKey)
      .toBe('spawn_expert_fork')
    expect(displayFor('spawn_agent', { task: 't', expert: 'e', run_in_background: true }).actionKey)
      .toBe('spawn_expert_background')
    expect(displayFor('spawn_agent', { task: 't', expert: 'e', fork: true, run_in_background: true }).actionKey)
      .toBe('spawn_expert_fork_background')
  })

  it('keeps the plain spawn keys when no expert is named', () => {
    expect(displayFor('spawn_agent', { task: 't' }).actionKey).toBe('spawn_agent')
    expect(displayFor('spawn_agent', { task: 't', fork: true }).actionKey).toBe('spawn_agent_fork')
  })
})

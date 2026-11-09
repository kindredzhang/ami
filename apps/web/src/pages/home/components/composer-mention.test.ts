import { describe, expect, it } from 'vitest'
import { findMentionQuery, matchesMentionQuery, removeMentionToken } from './composer-mention'

describe('findMentionQuery', () => {
  it('opens on a bare @', () => {
    expect(findMentionQuery('@')).toEqual({ start: 0, query: '' })
  })

  it('opens after whitespace', () => {
    expect(findMentionQuery('分析一下 @')).toEqual({ start: 5, query: '' })
  })

  it('carries the typed query lowercased', () => {
    expect(findMentionQuery('@Ops')).toEqual({ start: 0, query: 'ops' })
  })

  // A user who typed "@@" — or retyped after dismissing the picker — must still
  // get a picker. Reporting nothing here makes the composer look broken.
  it('opens on a repeated @', () => {
    expect(findMentionQuery('@@')).toEqual({ start: 1, query: '' })
    expect(findMentionQuery('@@op')).toEqual({ start: 1, query: 'op' })
  })

  it('stays closed for an email-like token inside a word', () => {
    expect(findMentionQuery('mail user@host')).toBeNull()
  })

  it('stays closed when prose follows the mention', () => {
    expect(findMentionQuery('@ops 帮我看下')).toBeNull()
  })

  it('stays closed in a multi-line draft', () => {
    expect(findMentionQuery('first line\n@ops')).toBeNull()
  })

  it('stays closed without an @', () => {
    expect(findMentionQuery('')).toBeNull()
    expect(findMentionQuery('普通文本')).toBeNull()
  })
})

describe('removeMentionToken', () => {
  it('drops the token and the gap before it', () => {
    expect(removeMentionToken('分析一下 @op', 5)).toBe('分析一下')
  })

  it('leaves an empty draft when the token was the whole text', () => {
    expect(removeMentionToken('@', 0)).toBe('')
  })

  it('ignores an out-of-range start', () => {
    expect(removeMentionToken('abc', 99)).toBe('abc')
  })
})

describe('matchesMentionQuery', () => {
  const expert = {
    name: 'ops-analyst',
    display_name: '运营分析专家',
    description: '分析运营看板数据',
  }

  it('matches everything on an empty query', () => {
    expect(matchesMentionQuery(expert, '')).toBe(true)
  })

  it('matches the id, the display name and the description', () => {
    expect(matchesMentionQuery(expert, 'ops')).toBe(true)
    expect(matchesMentionQuery(expert, '运营')).toBe(true)
    expect(matchesMentionQuery(expert, '看板')).toBe(true)
  })

  it('rejects an unrelated query', () => {
    expect(matchesMentionQuery(expert, 'finance')).toBe(false)
  })
})

/**
 * The composer's @ token detection.
 *
 * Kept out of chat-pane.vue so the boundary rules can be tested directly: this
 * is the one place where a small regex difference decides whether the expert
 * picker opens at all, and a composer that silently stops responding to "@" is
 * indistinguishable from a missing feature.
 */
export interface MentionQuery {
  /** Index of the "@" that starts the token. */
  start: number
  /** Text after the "@", lowercased for matching. */
  query: string
}

/**
 * Returns the trailing @ token in `text`, or null when the caret is not in one.
 *
 * Rules:
 * - The token must be at the end of the draft: a completed mention followed by
 *   more prose is not a live picker.
 * - Multi-line drafts never open it; a newline means the user moved on.
 * - The character before the "@" must be whitespace, another "@", or nothing.
 *   That keeps `user@host` as plain text while still letting a user who typed
 *   "@@" (or retyped after dismissing the picker) reopen it.
 */
export function findMentionQuery(text: string): MentionQuery | null {
  if (!text || text.includes('\n')) return null
  const match = /@([^\s@]*)$/.exec(text)
  const token = match?.[1]
  if (token === undefined) return null
  const start = text.length - token.length - 1
  const before = text[start - 1] ?? ''
  if (before !== '' && !/[\s@]/.test(before)) return null
  return { start, query: token.toLowerCase() }
}

/**
 * Removes the @ token from the draft once an expert is chosen, trimming the
 * whitespace that ran up to it so the selection does not leave a trailing gap.
 */
export function removeMentionToken(text: string, start: number): string {
  if (start < 0 || start > text.length) return text
  return text.slice(0, start).replace(/\s+$/, '')
}

/** Filter for the picker: name, display name and description all match. */
export function matchesMentionQuery(
  expert: { name: string, display_name?: string, description?: string },
  query: string,
): boolean {
  if (!query) return true
  const haystack = `${expert.name} ${expert.display_name ?? ''} ${expert.description ?? ''}`.toLowerCase()
  return haystack.includes(query)
}

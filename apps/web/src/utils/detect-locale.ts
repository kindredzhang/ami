import type { Locale } from '@/i18n'

// This build is deployed inside Dongfeng Group, so Chinese is the default: a
// machine whose browser reports anything other than Japanese resolves to 中文
// instead of falling through to English. An explicit choice still wins — this
// only decides the first-run value (see getInitialLocale in i18n.ts).
export function detectLocale(): Locale {
  if (typeof navigator === 'undefined') return 'zh'
  const normalized = (navigator.language || '').toLowerCase()
  // Only Japanese is honoured; everything else starts in Chinese. A Dongfeng
  // intranet seat should not land in an English UI because its browser reports
  // en-US, and English/Japanese are one click away in 设置 → 外观.
  if (normalized.startsWith('ja')) return 'ja'
  return 'zh'
}

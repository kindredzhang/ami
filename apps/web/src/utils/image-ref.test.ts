import { describe, expect, it } from 'vitest'
import { shortenImageRef } from './image-ref'

describe('shortenImageRef', () => {
  it('returns empty string for missing values', () => {
    expect(shortenImageRef(undefined)).toBe('')
    expect(shortenImageRef(null)).toBe('')
    expect(shortenImageRef('')).toBe('')
  })

  it('strips docker hub library prefix', () => {
    expect(shortenImageRef('docker.io/library/nginx:latest')).toBe('nginx:latest')
  })

  it('strips docker hub registry prefix for namespaced images', () => {
    expect(shortenImageRef('docker.io/ami/ami:latest')).toBe('ami/ami:latest')
  })

  it('preserves non-docker-hub registries', () => {
    expect(shortenImageRef('ghcr.io/kindredzhang/ami:latest')).toBe('ghcr.io/kindredzhang/ami:latest')
  })
})

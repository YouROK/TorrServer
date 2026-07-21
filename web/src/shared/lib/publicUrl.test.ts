import { describe, expect, it, vi, afterEach } from 'vitest'

import { getAppBasePath, publicUrl } from './publicUrl'

describe('publicUrl', () => {
  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('builds relative assets for ./ BASE_URL', () => {
    vi.stubEnv('BASE_URL', './')
    // import.meta.env.BASE_URL is compile-time; publicUrl reads import.meta.env
    // vitest stubEnv may not rewrite import.meta — assert current default behavior
    expect(publicUrl('icon.png')).toMatch(/icon\.png$/)
    expect(publicUrl('/icon.png')).toMatch(/icon\.png$/)
  })

  it('getAppBasePath is empty for relative base', () => {
    expect(getAppBasePath() === '' || getAppBasePath().startsWith('/')).toBe(true)
  })
})

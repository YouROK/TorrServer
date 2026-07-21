import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { GStreamerRuntime } from 'shared/api/types'

describe('shouldUseGStreamerPlayer', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.stubGlobal('window', {
      location: { protocol: 'http:', hostname: 'localhost', port: '8090' },
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  const runtimeOn: GStreamerRuntime = {
    built_in: true,
    config: { TranscodeAVI: true },
  }
  const runtimeOffAvi: GStreamerRuntime = {
    built_in: true,
    config: { TranscodeAVI: false },
  }

  it('returns false when GStreamer is not built in', async () => {
    const { shouldUseGStreamerPlayer } = await import('./gstreamer')
    expect(shouldUseGStreamerPlayer('movie.mkv', { built_in: false })).toBe(false)
    expect(shouldUseGStreamerPlayer('movie.mkv', null)).toBe(false)
  })

  it('uses GStreamer for mkv/webm containers', async () => {
    const { shouldUseGStreamerPlayer } = await import('./gstreamer')
    expect(shouldUseGStreamerPlayer('/path/film.mkv', runtimeOn)).toBe(true)
    expect(shouldUseGStreamerPlayer('clip.mk3d', runtimeOn)).toBe(true)
    expect(shouldUseGStreamerPlayer('a.webm?x=1', runtimeOn)).toBe(true)
  })

  it('honors TranscodeAVI for avi files', async () => {
    const { shouldUseGStreamerPlayer } = await import('./gstreamer')
    expect(shouldUseGStreamerPlayer('old.avi', runtimeOn)).toBe(true)
    expect(shouldUseGStreamerPlayer('old.avi', runtimeOffAvi)).toBe(false)
  })

  it('skips common progressive formats', async () => {
    const { shouldUseGStreamerPlayer } = await import('./gstreamer')
    expect(shouldUseGStreamerPlayer('a.mp4', runtimeOn)).toBe(false)
    expect(shouldUseGStreamerPlayer('a.m4v', runtimeOn)).toBe(false)
  })
})

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

describe('externalLink', () => {
  beforeEach(() => {
    vi.resetModules()
    vi.stubGlobal('window', {
      location: {
        href: 'https://ui.example:8443/web/',
        protocol: 'https:',
        hostname: 'ui.example',
        port: '8443',
      },
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('resolves relative links against the page origin when base is missing', async () => {
    const { externalLink } = await import('./mediaBase')
    const url = externalLink('/stream/file.mkv?link=abc&play', null)
    expect(url.toString()).toBe('https://ui.example:8443/stream/file.mkv?link=abc&play')
  })

  it('swaps protocol, hostname, and port from the media base', async () => {
    const { externalLink } = await import('./mediaBase')
    const url = externalLink('/stream/file.mkv?link=abc&play', 'http://192.168.1.2:8090')
    expect(url.toString()).toBe('http://192.168.1.2:8090/stream/file.mkv?link=abc&play')
  })

  it('clears the port when the media base has none', async () => {
    const { externalLink } = await import('./mediaBase')
    const url = externalLink('https://ui.example:8443/stream/x', 'http://media.example')
    expect(url.toString()).toBe('http://media.example/stream/x')
  })

  it('keeps the original link when the media base is malformed', async () => {
    const { externalLink } = await import('./mediaBase')
    const url = externalLink('/stream/x', 'not a url')
    expect(url.toString()).toBe('https://ui.example:8443/stream/x')
  })
})

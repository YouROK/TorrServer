import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest'

describe('Hosts URL construction', () => {
  beforeEach(() => {
    vi.resetModules()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.unstubAllEnvs()
  })

  it('builds endpoint URLs from window.location when VITE_SERVER_HOST is unset', async () => {
    vi.stubEnv('VITE_SERVER_HOST', '')
    vi.stubGlobal('window', {
      location: { protocol: 'https:', hostname: 'ts.example', port: '8443' },
    })

    const hosts = await import('./hosts')
    expect(hosts.getTorrServerHost()).toBe('https://ts.example:8443')
    expect(hosts.torrentsHost()).toBe('https://ts.example:8443/torrents')
    expect(hosts.streamHost()).toBe('https://ts.example:8443/stream')
    expect(hosts.playlistTorrHost).toBe(hosts.streamHost)
    expect(hosts.playlistTorrHost()).toBe(hosts.streamHost())
  })
})

/** Shared fixtures for mocked Playwright e2e (no live TorrServer). */

export const HASH = '0123456789abcdef0123456789abcdef01234567'

export const fixtureTorrent = {
  hash: HASH,
  title: 'E2E Mock Torrent',
  name: 'e2e.mock.mkv',
  category: 'movie',
  poster: '',
  torrent_size: 1024 * 1024 * 100,
  loaded_size: 0,
  stat: 5,
  file_stats: [{ id: 1, path: 'e2e.mock.mkv', length: 1024 * 1024 * 100 }],
}

export const fixtureSettings = {
  CacheSize: 64 * 1024 * 1024,
  ReaderReadAHead: 95,
  PreloadCache: 50,
  EnableBonjour: false,
  EnableDLNA: false,
  EnableTorznabSearch: true,
  EnableRutorSearch: false,
  TrackersListURL: 'https://example.test/trackers.txt',
  DefaultTrackers: 'udp://audit.example:1337/announce',
  TorznabUrls: [
    {
      Host: 'http://127.0.0.1:9696/1',
      Key: 'mockkey',
      Name: 'mock-indexer',
      Categories: '2000',
      CatType: 'manual',
    },
  ],
  RetrackersMode: 1,
  TorrentDisconnectTimeout: 30,
  ConnectionsLimit: 25,
  ResponsiveMode: true,
  ShowFSActiveTorr: true,
  StoreSettingsInJson: true,
  TMDBSettings: { APIKey: '', APIURL: 'https://api.themoviedb.org/3' },
}

export const fixtureWaf = {
  whitelist: '',
  blacklist: '',
  referers: '',
  ip_enabled: false,
  referer_enabled: true,
  read_only: false,
  warnings: [],
}

export const fixtureRuntime = {
  dlna_enabled: false,
  bonjour_enabled: false,
  friendly_name: '',
  webdav_enabled: false,
  webdav_path: '/dav',
  fuse_enabled: false,
  fuse_path: '',
}

const json = body => ({
  status: 200,
  contentType: 'application/json',
  body: JSON.stringify(body),
})

/** Intercept TorrServer APIs so the embedded SPA can run without a backend. */
export async function installMocks(page, { requireAuth = false } = {}) {
  const waf = { ...fixtureWaf }

  await page.route('**/echo', route => route.fulfill({ status: 200, contentType: 'text/plain', body: 'MatriX.144' }))
  await page.route('**/gst/echo', route => route.fulfill({ status: 404, body: '' }))
  await page.route('**/gst/settings', route => route.fulfill(json({})))
  await page.route('**/storage/settings', route => route.fulfill(json({})))
  await page.route('**/viewed', route => route.fulfill(json([])))
  await page.route('**/runtime/status', route => route.fulfill(json(fixtureRuntime)))
  await page.route('**/cache', route => route.fulfill(json({ Hash: HASH, Pieces: {}, Readers: [] })))

  await page.route('**/waf', async route => {
    if (requireAuth && !route.request().headers().authorization) {
      await route.fulfill({ status: 401, body: '' })
      return
    }
    if (route.request().method() === 'POST') {
      const body = route.request().postDataJSON() || {}
      waf.whitelist = body.whitelist ?? waf.whitelist
      waf.blacklist = body.blacklist ?? waf.blacklist
      waf.referers = body.referers ?? waf.referers
    }
    await route.fulfill(json(waf))
  })

  await page.route('**/settings', async route => {
    if (requireAuth && !route.request().headers().authorization) {
      await route.fulfill({ status: 401, body: '' })
      return
    }
    await route.fulfill(json(fixtureSettings))
  })

  await page.route('**/torrents', async route => {
    if (route.request().method() !== 'POST') {
      await route.continue()
      return
    }
    if (requireAuth && !route.request().headers().authorization) {
      await route.fulfill({ status: 401, body: '' })
      return
    }
    let action = 'list'
    try {
      action = route.request().postDataJSON()?.action || 'list'
    } catch {
      /* ignore */
    }
    if (action === 'list') {
      await route.fulfill(json([fixtureTorrent]))
      return
    }
    await route.fulfill(json(fixtureTorrent))
  })

  await page.route('**/torznab/search**', async route => {
    await route.fulfill({
      status: 200,
      contentType: 'application/xml',
      body: `<?xml version="1.0"?><rss version="2.0"><channel><item><title>Mock Result</title><guid>${HASH}</guid></item></channel></rss>`,
    })
  })
  await page.route('**/torznab/caps**', async route => {
    await route.fulfill({
      status: 200,
      contentType: 'application/xml',
      body: `<?xml version="1.0"?><caps><categories><category id="2000" name="Movies"/></categories></caps>`,
    })
  })
}

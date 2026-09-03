/**
 * Live torrent / stream / ffp checks.
 *
 *   LIVE=1 TORRSERVER_URL=http://127.0.0.1:18091 TS_USER=ts TS_PASS=ts yarn test:e2e:live
 *
 * Optional: LIVE_TORRENT_HASH (default e5a5bdb8…), TORZNAB_QUERY.
 * Indexer hosts are taken from the running server settings, or TORZNAB_HOST + TORZNAB_KEY.
 */
import assert from 'node:assert/strict'
import { test } from 'node:test'

import { chromium } from 'playwright'

const BASE = process.env.TORRSERVER_URL || ''
const USER = process.env.TS_USER || 'ts'
const PASS = process.env.TS_PASS || 'ts'
const HASH = (process.env.LIVE_TORRENT_HASH || 'e5a5bdb8ff6152657a1e051024b618dd37d76957').toLowerCase()
const live = process.env.LIVE === '1' && BASE !== ''

function authHeader() {
  return 'Basic ' + Buffer.from(`${USER}:${PASS}`).toString('base64')
}

async function api(path, { method = 'GET', json, headers = {} } = {}) {
  const res = await fetch(BASE + path, {
    method,
    headers: {
      Authorization: authHeader(),
      Accept: 'application/json',
      ...(json ? { 'Content-Type': 'application/json' } : {}),
      ...headers,
    },
    body: json ? JSON.stringify(json) : undefined,
  })
  return res
}

function skipIfNotLive(t) {
  if (!live) t.skip('set LIVE=1 TORRSERVER_URL=http://127.0.0.1:PORT')
}

test('live echo is MatriX', { timeout: 15_000 }, async t => {
  skipIfNotLive(t)
  const res = await api('/echo')
  const body = await res.text()
  assert.equal(res.status, 200)
  assert.match(body, /MatriX/)
})

test('live torrent get/add + stream stat + play range + ffp', { timeout: 180_000 }, async t => {
  skipIfNotLive(t)

  let got = await api('/torrents', { method: 'POST', json: { action: 'get', hash: HASH } })
  if (got.status !== 200) {
    const add = await api('/torrents', {
      method: 'POST',
      json: { action: 'add', link: HASH, save_to_db: true, title: 'live-e2e' },
    })
    assert.ok(add.status === 200, `add torrent failed: ${add.status} ${await add.text()}`)
  }

  let torrent
  const deadline = Date.now() + 120_000
  while (Date.now() < deadline) {
    const statRes = await fetch(`${BASE}/stream?link=${HASH}&stat`, {
      headers: { Authorization: authHeader(), Accept: 'application/json' },
    })
    if (statRes.ok) {
      torrent = await statRes.json()
      const files = torrent.file_stats || torrent.FileStats || []
      if (files.length > 0 && torrent.stat !== 1) break
    }
    await new Promise(r => setTimeout(r, 2000))
  }
  assert.ok(torrent, 'no stream?stat payload')
  const files = torrent.file_stats || []
  assert.ok(files.length > 0, `no file_stats: ${JSON.stringify(torrent).slice(0, 400)}`)
  const fileId = files[0].id ?? files[0].Id ?? 1

  const play = await fetch(`${BASE}/stream?link=${HASH}&index=${fileId}&play`, {
    headers: { Authorization: authHeader(), Range: 'bytes=0-1023' },
  })
  const playStatus = play.status
  const buf = Buffer.from(await play.arrayBuffer())
  assert.ok([200, 206].includes(playStatus), `play status ${playStatus}`)
  assert.ok(buf.length > 0, 'empty play body')

  const ffp = await api(`/ffp/${HASH}/${fileId}`)
  const ffpStatus = ffp.status
  const ffpText = await ffp.text()
  if (ffpStatus === 200) {
    const probe = JSON.parse(ffpText)
    assert.ok(probe.format || probe.streams, `ffp json missing format/streams: ${ffpText.slice(0, 200)}`)
  } else {
    t.diagnostic(`ffp skipped/non-200: ${ffpStatus} ${ffpText.slice(0, 200)}`)
  }
})

test('live torznab search (server settings or TORZNAB_HOST)', { timeout: 60_000 }, async t => {
  skipIfNotLive(t)
  const setsRes = await api('/settings', { method: 'POST', json: { action: 'get' } })
  assert.equal(setsRes.status, 200)
  const sets = await setsRes.json()
  const fromEnv = process.env.TORZNAB_HOST
  const enabled = Boolean(fromEnv) || Boolean(sets.EnableTorznabSearch && (sets.TorznabUrls || []).length)
  if (!enabled) {
    t.skip('no Torznab indexers on server; set TORZNAB_HOST or enable in data/settings.json')
    return
  }
  const q = encodeURIComponent(process.env.TORZNAB_QUERY || 'matrix')
  const search = await api(`/torznab/search/?query=${q}`)
  const searchStatus = search.status
  const xml = await search.text()
  assert.ok(searchStatus === 200, `torznab search ${searchStatus} ${xml.slice(0, 200)}`)
  let parsed = null
  try {
    parsed = JSON.parse(xml)
  } catch {
    /* xml */
  }
  if (Array.isArray(parsed)) {
    t.diagnostic(`torznab results: ${parsed.length}`)
  } else {
    assert.match(xml, /rss|error|item|channel/i)
  }
})

test('live Playwright library card for hash', { timeout: 60_000 }, async t => {
  skipIfNotLive(t)
  const browser = await chromium.launch({ headless: true })
  const context = await browser.newContext({
    locale: 'en-US',
    viewport: { width: 1440, height: 900 },
    extraHTTPHeaders: { 'Accept-Language': 'en-US,en;q=0.9' },
  })
  const page = await context.newPage()
  try {
    await page.goto(BASE + '/', { waitUntil: 'domcontentloaded', timeout: 30_000 })
    await page.waitForSelector('input[name="username"], .torrent-card', { timeout: 20_000 })
    const userBox = page.locator('input[name="username"]')
    if (await userBox.isVisible().catch(() => false)) {
      await userBox.fill(USER)
      await page.locator('input[name="password"]').fill(PASS)
      await page.getByRole('button', { name: /sign in|войти/i }).click()
    }
    await page.waitForFunction(
      h =>
        [...document.querySelectorAll('.torrent-card')].some(
          el => (el.getAttribute('data-hash') || '').toLowerCase() === h,
        ),
      HASH,
      { timeout: 30_000 },
    )
  } finally {
    await context.close()
    await browser.close()
  }
})

#!/usr/bin/env node
/**
 * iPhone 16 Pro Max smoke screenshots for Details (Files/Stats/Swarm/Cache + ⋯ sheet).
 *
 * Starts against Vite (or embedded UI) and mocks TorrServer APIs so layout can be
 * inspected without a live torrent library:
 *
 *   yarn dev   # terminal A
 *   TORRSERVER_URL=http://127.0.0.1:5173 yarn screenshots:iphone
 *
 * Set LIVE=1 to skip fixtures and use a real TorrServer (must have ≥1 torrent).
 *
 * Requires: playwright + chromium.
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { chromium, devices } from 'playwright'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const outDir = path.resolve(__dirname, '../tmp/iphone16-promax')
const baseURL = process.env.TORRSERVER_URL || process.env.WEB_URL || 'http://127.0.0.1:5173'
const live = process.env.LIVE === '1'

const HASH = '0123456789abcdef0123456789abcdef01234567'

const fixtureTorrent = {
  hash: HASH,
  title:
    '[LAMPA] Дом дракона (3 сезон: 1-4 серии из 8) / House of the Dragon / 2026 / 4 х ДБ, 3 х ПМ, WEB-DLRip 2160p',
  name: 'House.of.the.Dragon.S03E01-E04.2160p.WEB-DL.x265',
  category: 'tv',
  poster: '',
  torrent_size: 37.09 * 1024 ** 3,
  loaded_size: 0,
  preloaded_bytes: 0,
  preload_size: 0,
  download_speed: 0,
  upload_speed: 0,
  active_peers: 1,
  total_peers: 51,
  connected_seeders: 0,
  pending_peers: 51,
  half_open_peers: 0,
  bytes_read: 95100,
  bytes_written: 955,
  bytes_read_data: 0,
  bytes_read_useful_data: 0,
  chunks_read: 0,
  chunks_read_useful: 0,
  chunks_read_wasted: 0,
  chunks_written: 0,
  stat: 3,
  file_stats: [
    { id: 1, path: 'House of the Dragon/S03E01.mkv', length: 9.02 * 1024 ** 3 },
    { id: 2, path: 'House of the Dragon/S03E02.mkv', length: 10.6 * 1024 ** 3 },
    { id: 3, path: 'House of the Dragon/S03E03.mkv', length: 8.9 * 1024 ** 3 },
    { id: 4, path: 'House of the Dragon/S03E04.mkv', length: 8.57 * 1024 ** 3 },
  ],
}

const fixtureCache = {
  Hash: HASH,
  Capacity: 256 * 1024 * 1024,
  Filled: 0,
  PiecesLength: 8 * 1024 * 1024,
  PiecesCount: 4748,
  Pieces: {},
  Readers: [],
}

const iPhone =
  devices['iPhone 16 Pro Max'] ??
  devices['iPhone 15 Pro Max'] ?? {
    userAgent:
      'Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Mobile/15E148 Safari/604.1',
    viewport: { width: 440, height: 956 },
    deviceScaleFactor: 3,
    isMobile: true,
    hasTouch: true,
    defaultBrowserType: 'chromium',
  }

fs.mkdirSync(outDir, { recursive: true })

async function shot(page, name) {
  const file = path.join(outDir, `${name}.png`)
  await page.screenshot({ path: file, fullPage: false })
  console.log('wrote', file)
}

async function installFixtures(page) {
  await page.route('**/echo', async route => {
    await route.fulfill({ status: 200, contentType: 'text/plain', body: '1' })
  })
  await page.route('**/settings', async route => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ CacheSize: 256 * 1024 * 1024 }),
    })
  })
  await page.route('**/gst/settings', async route => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
  })
  await page.route('**/storage/settings', async route => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
  })
  await page.route('**/viewed', async route => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: '[]' })
  })
  await page.route('**/cache', async route => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(fixtureCache),
    })
  })
  await page.route('**/torrents', async route => {
    if (route.request().method() !== 'POST') {
      await route.continue()
      return
    }
    let action = 'list'
    try {
      action = route.request().postDataJSON()?.action || 'list'
    } catch {
      /* ignore */
    }
    if (action === 'list') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([fixtureTorrent]),
      })
      return
    }
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(fixtureTorrent),
    })
  })
}

async function main() {
  const browser = await chromium.launch({ headless: true })
  const context = await browser.newContext({
    ...iPhone,
    locale: 'en-US',
  })
  const page = await context.newPage()

  if (!live) await installFixtures(page)

  await page.goto(baseURL, { waitUntil: 'domcontentloaded', timeout: 60_000 })
  await page.waitForSelector('.torrent-card', { timeout: 45_000 })

  await page.locator('.torrent-card [role="button"]').first().click()

  const details = page.locator('.ts-details-modal')
  await details.waitFor({ state: 'visible', timeout: 20_000 })
  // Wait out HeroUI enter animation so geometry asserts are stable.
  await page.waitForTimeout(400)

  const style = await details.evaluate(el => {
    const cs = getComputedStyle(el)
    const rect = el.getBoundingClientRect()
    return {
      borderRadius: cs.borderRadius,
      transform: cs.transform,
      width: cs.width,
      height: cs.height,
      inset: `${cs.top} ${cs.right} ${cs.bottom} ${cs.left}`,
      position: cs.position,
      rect: { x: rect.x, y: rect.y, w: rect.width, h: rect.height },
      viewport: { w: window.innerWidth, h: window.innerHeight },
    }
  })
  fs.writeFileSync(path.join(outDir, 'details-computed.json'), JSON.stringify(style, null, 2))
  console.log('details computed', style)

  const overflowX = Math.max(0, Math.abs(style.rect.x), style.rect.w - style.viewport.w)
  const overflowY = Math.max(0, Math.abs(style.rect.y), style.rect.h - style.viewport.h)
  if (overflowX > 2 || overflowY > 2 || (style.transform && style.transform !== 'none')) {
    throw new Error(
      `Details modal geometry not production-ready: overflowX=${overflowX.toFixed(1)} overflowY=${overflowY.toFixed(1)} transform=${style.transform}`,
    )
  }

  const tabIds = ['files', 'stats', 'swarm', 'cache']
  for (let i = 0; i < tabIds.length; i++) {
    await page.locator('.ts-details-tabs [role="tab"]').nth(i).click()
    await page.waitForTimeout(500)
    await shot(page, `details-${tabIds[i]}`)
  }

  // Cache snake must fill the panel — regression guard for the 1-row collapse.
  await page.locator('.ts-details-tabs [role="tab"]').nth(3).click()
  await page.waitForTimeout(600)
  const snake = await page.locator('.ts-details-cache-snake').evaluate(el => {
    const rect = el.getBoundingClientRect()
    const panel = el.closest('[role="tabpanel"]')?.getBoundingClientRect()
    return { snakeH: rect.height, panelH: panel?.height ?? 0 }
  })
  console.log('cache snake geometry', snake)
  if (snake.panelH > 120 && snake.snakeH < Math.min(120, snake.panelH * 0.35)) {
    throw new Error(
      `Cache snake too short for panel: snakeH=${snake.snakeH.toFixed(1)} panelH=${snake.panelH.toFixed(1)}`,
    )
  }

  await page.locator('.ts-details-tabs [role="tab"]').nth(1).click()
  await page.waitForTimeout(300)
  await page.locator('.ts-details-modal').getByRole('button', { name: /^Actions$/i }).click()
  await page.waitForTimeout(500)
  await shot(page, 'details-stats-actions-sheet')

  await browser.close()
  console.log(`screenshots:iphone ok → ${outDir}`)
}

main().catch(err => {
  console.error(err)
  process.exit(1)
})

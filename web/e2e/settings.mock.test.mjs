import assert from 'node:assert/strict'
import { after, before, test } from 'node:test'

import { chromium } from 'playwright'

import { HASH, installMocks } from './fixtures.mjs'
import { embedPagesRoot, serveEmbedDir } from './static-server.mjs'

let server
let origin
let browser

before(async () => {
  ;({ server, origin } = await serveEmbedDir(embedPagesRoot()))
  browser = await chromium.launch({ headless: true })
})

after(async () => {
  await browser?.close()
  await new Promise(resolve => server?.close(resolve))
})

async function newPage(opts) {
  const context = await browser.newContext({
    locale: 'en-US',
    viewport: { width: 1440, height: 900 },
    extraHTTPHeaders: { 'Accept-Language': 'en-US,en;q=0.9' },
  })
  const page = await context.newPage()
  await installMocks(page, opts)
  return { context, page }
}

test('mocked library shows torrent card and no Quick-VLC', { timeout: 45_000 }, async () => {
  const { context, page } = await newPage()
  try {
    await page.goto(origin + '/', { waitUntil: 'domcontentloaded' })
    await page.locator(`.torrent-card[data-hash="${HASH}"]`).waitFor({ timeout: 20_000 })
    const text = await page.locator('body').innerText()
    assert.equal(/quick\s*vlc/i.test(text), false)
  } finally {
    await context.close()
  }
})

test('mocked login form appears when torrents API returns 401', { timeout: 45_000 }, async () => {
  const { context, page } = await newPage({ requireAuth: true })
  try {
    await page.goto(origin + '/', { waitUntil: 'domcontentloaded' })
    await page.locator('input[name="username"]').waitFor({ timeout: 20_000 })
    await page.locator('input[name="username"]').fill('audit')
    await page.locator('input[name="password"]').fill('auditpass')
    await page.getByRole('button', { name: 'Sign in' }).click()
    await page.locator(`.torrent-card[data-hash="${HASH}"]`).waitFor({ timeout: 20_000 })
  } finally {
    await context.close()
  }
})

test('mocked Settings Access / Network / MCP / Torznab', { timeout: 60_000 }, async () => {
  const { context, page } = await newPage()
  try {
    await page.goto(origin + '/', { waitUntil: 'domcontentloaded' })
    await page.getByRole('button', { name: 'Settings' }).first().waitFor({ timeout: 20_000 })
    await page.getByRole('button', { name: 'Settings' }).first().click()
    await page.getByRole('heading', { name: 'Settings' }).waitFor()

    const tab = name => page.getByRole('tab', { name })
    const clickTab = async name => {
      await tab(name).first().click()
      await page.waitForTimeout(400)
    }

    await clickTab('Access')
    await page.getByLabel('IP whitelist').waitFor({ timeout: 15_000 })
    await page.getByLabel('IP blacklist').fill('203.0.113.8')
    await page.getByRole('button', { name: 'Save access lists' }).click()
    await page.getByText('Access lists saved and reloaded').waitFor({ timeout: 10_000 })

    await clickTab('Network')
    const trackersUrl = await page.getByLabel('Custom remote trackers list URL').inputValue()
    assert.equal(trackersUrl.includes('example.test'), true)

    await clickTab('Torznab')
    await page.getByText('mock-indexer').waitFor({ timeout: 10_000 })
    await page.getByText('2000').first().waitFor()

    await clickTab('Advanced')
    await page.getByText('MCP', { exact: true }).first().waitFor()
    await page.getByRole('button', { name: 'Copy MCP URL' }).waitFor()

    await clickTab('Application')
    await page.getByText('Prompt to open video in VLC').waitFor()
  } finally {
    await context.close()
  }
})

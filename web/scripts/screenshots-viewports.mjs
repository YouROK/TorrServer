#!/usr/bin/env node
/**
 * Viewport matrix screenshots: library chrome + Settings on phone / tablet / desktop.
 *
 * Uses the embedded SPA (after `yarn build`) unless TORRSERVER_URL is set:
 *
 *   yarn build
 *   yarn screenshots:viewports
 *
 *   TORRSERVER_URL=http://127.0.0.1:3000 yarn screenshots:viewports   # against yarn dev
 *
 * Requires: playwright + chromium.
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { chromium } from 'playwright'

import { installMocks } from '../e2e/fixtures.mjs'
import { embedPagesRoot, serveEmbedDir } from '../e2e/static-server.mjs'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const outDir = path.resolve(__dirname, '../tmp/viewports')
const liveUrl = process.env.TORRSERVER_URL || process.env.WEB_URL || ''

const VIEWPORTS = [
  { id: 'phone', width: 390, height: 844, phone: true },
  { id: 'tablet', width: 768, height: 1024, phone: false },
  { id: 'desktop', width: 1440, height: 900, phone: false },
]

fs.mkdirSync(outDir, { recursive: true })

async function shot(page, name) {
  const file = path.join(outDir, `${name}.png`)
  await page.screenshot({ path: file, fullPage: false })
  console.log('wrote', file)
}

async function openSettings(page, phone) {
  if (phone) {
    await page.locator('.ts-bottom-nav').getByRole('button', { name: 'More' }).click()
    await page.locator('.ts-sheet-modal').getByRole('button', { name: 'Settings' }).click()
  } else {
    await page.getByRole('button', { name: 'Settings' }).first().click()
  }
  await page.getByRole('heading', { name: 'Settings' }).waitFor({ timeout: 20_000 })
  await page.waitForTimeout(400)
}

async function main() {
  let server
  let origin = liveUrl
  if (!origin) {
    ; ({ server, origin } = await serveEmbedDir(embedPagesRoot()))
  }

  const browser = await chromium.launch({ headless: true })
  try {
    for (const vp of VIEWPORTS) {
      const context = await browser.newContext({
        locale: 'en-US',
        viewport: { width: vp.width, height: vp.height },
        colorScheme: 'dark',
        extraHTTPHeaders: { 'Accept-Language': 'en-US,en;q=0.9' },
      })
      const page = await context.newPage()
      if (!liveUrl) await installMocks(page)

      await page.goto(origin.replace(/\/$/, '') + '/', { waitUntil: 'domcontentloaded', timeout: 60_000 })
      await page.waitForSelector('.torrent-card', { timeout: 45_000 })
      await page.waitForTimeout(300)
      await shot(page, `${vp.id}-library`)

      await openSettings(page, vp.phone)
      await shot(page, `${vp.id}-settings`)
      await context.close()
    }
  } finally {
    await browser.close()
    if (server) await new Promise(resolve => server.close(resolve))
  }

  console.log(`screenshots:viewports ok → ${outDir}`)
}

main().catch(err => {
  console.error(err)
  process.exit(1)
})

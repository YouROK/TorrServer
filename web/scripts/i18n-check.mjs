#!/usr/bin/env node
/**
 * Locale parity check: every language must contain the same flattened key set as en.
 * Also fails on flat dotted top-level keys like "Search.Tracker" (use nested objects).
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const localesDir = path.join(root, 'src/locales')
const langs = ['en', 'ru', 'ua', 'bg', 'fr', 'ro', 'zh']

function flatten(obj, prefix = '', out = {}) {
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k
    if (v && typeof v === 'object' && !Array.isArray(v)) flatten(v, key, out)
    else out[key] = v
  }
  return out
}

let failed = false
const locales = {}
for (const lang of langs) {
  const file = path.join(localesDir, lang, 'translation.json')
  const raw = JSON.parse(fs.readFileSync(file, 'utf8'))
  const flatKeys = Object.keys(raw).filter(k => k.includes('.'))
  if (flatKeys.length) {
    console.error(`[${lang}] flat dotted top-level keys forbidden:`, flatKeys.slice(0, 10))
    failed = true
  }
  if (typeof raw.Search === 'string') {
    console.error(`[${lang}] "Search" must be a nested object; use nav.Search for the label`)
    failed = true
  }
  if (!raw.nav || typeof raw.nav.Search !== 'string') {
    console.error(`[${lang}] missing nav.Search string`)
    failed = true
  }
  locales[lang] = flatten(raw)
}

const enKeys = new Set(Object.keys(locales.en))
for (const lang of langs) {
  if (lang === 'en') continue
  const keys = new Set(Object.keys(locales[lang]))
  const missing = [...enKeys].filter(k => !keys.has(k))
  const extra = [...keys].filter(k => !enKeys.has(k))
  if (missing.length || extra.length) {
    failed = true
    console.error(`[${lang}] missing ${missing.length}, extra ${extra.length}`)
    if (missing.length) console.error('  missing:', missing.slice(0, 20).join(', '))
    if (extra.length) console.error('  extra:', extra.slice(0, 20).join(', '))
  }
}

if (failed) {
  console.error('i18n:check failed')
  process.exit(1)
}
console.log(`i18n:check ok — ${enKeys.size} keys × ${langs.length} locales`)

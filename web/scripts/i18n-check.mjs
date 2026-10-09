#!/usr/bin/env node
/**
 * Locale parity check: every language must contain the same flattened key set as en.
 * Also fails on flat dotted top-level keys like "Search.Tracker" (use nested objects).
 * Scans web/src for static t('…') / i18n.t('…') keys and fails if any are missing from en.
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const localesDir = path.join(root, 'src/locales')
const srcDir = path.join(root, 'src')
const langs = ['en', 'ru', 'ua', 'bg', 'fr', 'ro', 'zh']

/** Dynamic t(`prefix${…}`) / t('ThemePalette'+id) prefixes — not full keys. */
const DYNAMIC_PREFIXES = [
  'HTTPSSettings.Source.',
  'ThemePalette',
  'WAF.Lists.',
  'WAF.WarningCodes.',
]

function flatten(obj, prefix = '', out = {}) {
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k
    if (v && typeof v === 'object' && !Array.isArray(v)) flatten(v, key, out)
    else out[key] = v
  }
  return out
}

function walkTsFiles(dir, out = []) {
  for (const ent of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, ent.name)
    if (ent.isDirectory()) walkTsFiles(p, out)
    else if (ent.isFile() && /\.(ts|tsx)$/.test(ent.name)) out.push(p)
  }
  return out
}

function collectStaticTKeys(files) {
  const re = /(?:\bi18n\.t|\bt)\(\s*['"]([^'"]+)['"]/g
  const keys = new Set()
  for (const file of files) {
    const text = fs.readFileSync(file, 'utf8')
    for (const m of text.matchAll(re)) keys.add(m[1])
  }
  return keys
}

function isDynamicKey(key) {
  return DYNAMIC_PREFIXES.some(p => key === p || key.startsWith(p))
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

const usedKeys = collectStaticTKeys(walkTsFiles(srcDir))
const missingUsed = [...usedKeys].filter(k => !enKeys.has(k) && !isDynamicKey(k)).sort()
if (missingUsed.length) {
  failed = true
  console.error(`used t() keys missing from en: ${missingUsed.length}`)
  console.error('  ', missingUsed.slice(0, 30).join(', '))
}

if (failed) {
  console.error('i18n:check failed')
  process.exit(1)
}
console.log(`i18n:check ok — ${enKeys.size} keys × ${langs.length} locales; ${usedKeys.size} static t() keys covered`)

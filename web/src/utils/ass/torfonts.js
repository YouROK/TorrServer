import { readFontNames } from './sfnt'

// Fonts that lie in the torrent next to the video: the ASS file names fonts, the font files
// are found by the names inside them (file names do not matter) and handed to the renderer.

const FONT_FILE = /\.(ttf|otf|ttc|woff2?)$/i
const MAX_FILES = 300
const CONCURRENCY = 4

export const isFontFile = path => FONT_FILE.test(path || '')

export const normalizeFontName = name =>
  String(name || '')
    .replace(/^@/, '')
    .trim()
    .toLowerCase()

const log = (...args) => {
  // eslint-disable-next-line no-console
  console.info('[ASS fonts]', ...args)
}

// url -> Promise<Array|null>; a font file is read once per page load
const nameCache = new Map()

const indexFile = file => {
  if (!nameCache.has(file.url)) {
    nameCache.set(
      file.url,
      readFontNames(file.url).catch(error => {
        log(`cannot read ${file.path || file.url}: ${error.message}`)
        return null
      }),
    )
  }
  return nameCache.get(file.url)
}

// Runs the tasks with a limited number of simultaneous requests
const runPool = async (items, worker) => {
  const results = new Array(items.length)
  let next = 0
  const run = async () => {
    while (next < items.length) {
      const index = next
      next++
      // eslint-disable-next-line no-await-in-loop
      results[index] = await worker(items[index])
    }
  }
  await Promise.all(Array.from({ length: Math.min(CONCURRENCY, items.length) }, run))
  return results
}

// files: [{ path, url, source }] (source: 'external' or 'embedded'); wantedNames: font names used by the ASS
// file; preferSource: files of this source go first, and when the same font is in several files the first
// one is used. Returns [{ name, url, bold, italic, byFamily }] with name normalized (lower case).
export const resolveFonts = async (files, wantedNames, preferSource) => {
  const wanted = new Set((wantedNames || []).map(normalizeFontName).filter(Boolean))
  const ordered = [...(files || [])].sort((a, b) => (b.source === preferSource) - (a.source === preferSource))
  const candidates = ordered.filter(file => isFontFile(file.path)).slice(0, MAX_FILES)
  if (!wanted.size || !candidates.length) return []

  const indexed = await runPool(candidates, indexFile)
  const found = []
  const seen = new Set()
  const faces = new Set() // the same face from another file is skipped: the first file wins
  let unreadable = 0

  indexed.forEach((infos, i) => {
    if (!infos) {
      unreadable++
      return
    }
    infos.forEach(info => {
      const kinds = [
        ['family', info.families],
        ['full', info.fulls],
        ['postscript', info.postscript],
      ]
      kinds.forEach(([kind, names]) => {
        names.forEach(raw => {
          const name = normalizeFontName(raw)
          const key = `${candidates[i].url}|${name}`
          const face = `${name}|${kind === 'family'}|${info.bold}|${info.italic}`
          if (!wanted.has(name) || seen.has(key) || faces.has(face)) return
          seen.add(key)
          faces.add(face)
          found.push({
            name,
            url: candidates[i].url,
            bold: info.bold,
            italic: info.italic,
            byFamily: kind === 'family',
          })
        })
      })
    })
  })

  const matched = new Set(found.map(face => face.name))
  const missing = [...wanted].filter(name => !matched.has(name))
  const summary = `${candidates.length} font file(s), ${unreadable} unreadable`
  const usage = `used by subtitles: ${wanted.size}, found: ${matched.size}`
  log(`${summary}; ${usage}${missing.length ? `, missing: ${missing.join(', ')}` : ''}`)
  return found
}

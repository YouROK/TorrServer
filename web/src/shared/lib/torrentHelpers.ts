import axios from 'axios'
import { remote as parseTorrentRemote } from 'parse-torrent'
import ptt from 'parse-torrent-title'
import { tmdbSettingsHost } from 'shared/api/hosts'
import type { TMDBSettingsConfig } from 'shared/api/types'

type TMDBSettings = Required<Pick<TMDBSettingsConfig, 'APIKey' | 'APIURL' | 'ImageURL' | 'ImageURLRu'>> &
  TMDBSettingsConfig

let tmdbSettingsCache: TMDBSettings | null = null

export const clearTMDBCache = () => {
  tmdbSettingsCache = null
}

const defaultTMDBSettings = (): TMDBSettings => ({
  APIKey: import.meta.env.VITE_TMDB_API_KEY || '',
  APIURL: 'https://api.themoviedb.org/3',
  ImageURL: 'https://image.tmdb.org',
  ImageURLRu: 'https://imagetmdb.com',
})

const mergeTMDBSettings = (data?: TMDBSettingsConfig | null): TMDBSettings => ({
  ...defaultTMDBSettings(),
  ...data,
  APIKey: data?.APIKey || import.meta.env.VITE_TMDB_API_KEY || '',
})

/** Ensure absolute `https://` host and strip trailing slash. */
const normalizeUrl = (url: string | undefined, fallback: string): string => {
  const trimmed = (url || fallback).trim().replace(/\/$/, '')
  if (/^https?:\/\//i.test(trimmed)) return trimmed
  return `https://${trimmed.replace(/^\/\//, '')}`
}

/** Force `…/3/search/multi` even when settings store a truncated API base. */
const buildTmdbSearchUrl = (apiURL: string | undefined): string => {
  let base = normalizeUrl(apiURL, 'https://api.themoviedb.org')
  if (!base.includes('/3/search/multi')) {
    base = base.replace(/\/3.*$/, '').replace(/\/search.*$/, '')
    base = `${base}/3/search/multi`
  }
  return base
}

const getTMDBSettings = async (): Promise<TMDBSettings> => {
  if (tmdbSettingsCache) return tmdbSettingsCache
  try {
    const { data } = await axios.get(tmdbSettingsHost())
    tmdbSettingsCache = mergeTMDBSettings(data)
    return tmdbSettingsCache
  } catch {
    tmdbSettingsCache = defaultTMDBSettings()
    return tmdbSettingsCache
  }
}

interface TmdbSearchResult {
  poster_path?: string
  [key: string]: unknown
}

/** TMDB multi-search → absolute `w300` poster URLs (RU uses ImageURLRu host when set). */
export const getMoviePosters = async (movieName: string, language = 'en'): Promise<string[] | null> => {
  const settings = await getTMDBSettings()
  if (!settings.APIKey) return null

  const url = buildTmdbSearchUrl(settings.APIURL)
  const imgHost = normalizeUrl(
    language === 'ru' ? settings.ImageURLRu : settings.ImageURL,
    language === 'ru' ? 'https://imagetmdb.com' : 'https://image.tmdb.org',
  )

  return axios
    .get(url, {
      params: {
        api_key: settings.APIKey,
        language,
        include_image_language: `${language},null,en`,
        query: movieName,
      },
    })
    .then(({ data: { results } }: { data: { results: TmdbSearchResult[] } }) =>
      results.filter(el => el.poster_path).map(el => `${imgHost}/t/p/w300${el.poster_path}`),
    )
    .catch(() => null)
}

export const checkImageURL = async (url?: string | null): Promise<boolean> => {
  if (!url || !url.match(/.(\.jpg|\.jpeg|\.png|\.gif|\.svg||\.webp).*$/i)) return false
  return true
}

const magnetRegex = /^magnet:\?xt=urn:[a-z0-9].*/i
const hashRegex = /^\b[0-9a-f]{32}\b$|^\b[0-9a-f]{40}\b$|^\b[0-9a-f]{64}\b$/i
const torrentRegex = /^.*\.(torrent)$/i
const linkRegex = /^(http(s?)):\/\/.*/i
const torrsRegex = /^(?:web\+)?torrs:\/\/.*/i

/** Accepts infohash, magnet, `.torrent` path, http(s) torrent URL, or `torrs://` / `web+torrs://` deep link. */
export const checkTorrentSource = (source: string): boolean =>
  source.match(hashRegex) !== null ||
  source.match(magnetRegex) !== null ||
  source.match(torrentRegex) !== null ||
  source.match(linkRegex) !== null ||
  source.match(torrsRegex) !== null

const POSTER_SEARCH_MAX_LEN = 50
const POSTER_SEARCH_MAX_WORDS = 4

/**
 * Trim release-group noise so TMDB search hits the movie/show name
 * (cut at `[`/`(`/` / `, cap words/length, prefer parse-torrent-title).
 */
export const shortenTitleForPosterSearch = (
  fullTitle: string,
  opts: { maxWords?: number; maxLen?: number } = {},
): string => {
  const maxWords = opts.maxWords ?? POSTER_SEARCH_MAX_WORDS
  const maxLen = opts.maxLen ?? POSTER_SEARCH_MAX_LEN
  if (!fullTitle || typeof fullTitle !== 'string') return ''
  const trimmed = fullTitle.trim()
  if (!trimmed) return ''
  let base = trimmed
  for (const sep of [' [', ' (', ' / ']) {
    const i = base.indexOf(sep)
    if (i > 0) base = base.slice(0, i).trim()
  }
  try {
    const parsed = ptt.parse(base)
    if (parsed?.title && parsed.title.length <= maxLen + 15) base = parsed.title
  } catch {
    // ignore
  }
  const words = base.split(/\s+/).filter(Boolean)
  const byWords = words.slice(0, maxWords).join(' ')
  if (byWords.length <= maxLen) return byWords.trim()
  const cut = byWords.slice(0, maxLen)
  const lastSpace = cut.lastIndexOf(' ')
  const result = lastSpace > 0 ? cut.slice(0, lastSpace) : cut
  return result.trim() || trimmed.slice(0, maxLen).trim()
}

export interface ParseTorrentTitleResult {
  parsedTitle: string | null
  originalName: string | null
}

/**
 * Prefer the shorter of torrent-level vs first-file parse-torrent-title results
 * (release names are often noisier than the inner video filename).
 */
export const parseTorrentTitle = (
  parsingSource: string | File,
  callback: (result: ParseTorrentTitleResult) => void,
): void => {
  parseTorrentRemote(parsingSource, (err, parsed = {}) => {
    const { name, files } = parsed
    if (!name || err) return callback({ parsedTitle: null, originalName: null })

    const torrentName = ptt.parse(name).title
    const nameOfFileInsideTorrent = files ? ptt.parse(files[0].name || '').title : null

    let newTitle = torrentName
    if (nameOfFileInsideTorrent) {
      newTitle = torrentName.length < nameOfFileInsideTorrent.length ? torrentName : nameOfFileInsideTorrent
    }

    callback({ parsedTitle: newTitle, originalName: name })
  })
}

/** Resolve infoHash from a .torrent File (async wrapper around parse-torrent `remote`). */
export const parseTorrentInfoHash = (file: File): Promise<string | null> =>
  new Promise(resolve => {
    parseTorrentRemote(file, (err, parsed) => {
      if (err || !parsed?.infoHash) resolve(null)
      else resolve(String(parsed.infoHash))
    })
  })

/** Resolve infoHash from magnet / hash / torrent URL string. */
export const parseSourceInfoHash = (source: string): Promise<string | null> =>
  new Promise(resolve => {
    const trimmed = source.trim()
    if (!trimmed) {
      resolve(null)
      return
    }
    parseTorrentRemote(trimmed, (err, parsed) => {
      if (err || !parsed?.infoHash) resolve(null)
      else resolve(String(parsed.infoHash).toLowerCase())
    })
  })

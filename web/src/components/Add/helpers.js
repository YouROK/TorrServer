import axios from 'axios'
import parseTorrent from 'parse-torrent'
import ptt from 'parse-torrent-title'
import { tmdbSettingsHost } from 'utils/Hosts'

// Cache for TMDB settings to avoid repeated API calls
let tmdbSettingsCache = null

// Clear TMDB settings cache - call this when settings are updated
export const clearTMDBCache = () => {
  tmdbSettingsCache = null
}

// Fetch TMDB settings from backend
const getTMDBSettings = async () => {
  if (tmdbSettingsCache) {
    return tmdbSettingsCache
  }

  try {
    const { data } = await axios.get(tmdbSettingsHost())
    tmdbSettingsCache = data
    return data
  } catch (error) {
    return {
      APIKey: process.env.REACT_APP_TMDB_API_KEY || '',
      APIURL: 'https://api.themoviedb.org/3',
      ImageURL: 'https://image.tmdb.org',
      ImageURLRu: 'https://imagetmdb.com',
    }
  }
}

export const getMoviePosters = async (movieName, language = 'en') => {
  const settings = await getTMDBSettings()

  // If no API key is configured, return null
  if (!settings.APIKey) {
    return null
  }

  // Build API URL - automatically append /3/search/multi
  let apiURL = settings.APIURL.replace(/^https?:\/\//, '').replace(/\/$/, '')

  // If URL doesn't already contain the full path, add /3/search/multi
  if (!apiURL.includes('/3/search/multi')) {
    // Remove any partial paths that might exist
    apiURL = apiURL.replace(/\/3.*$/, '').replace(/\/search.*$/, '')
    apiURL = `${apiURL}/3/search/multi`
  }

  const url = `${window.location.protocol}//${apiURL}`

  // Build image URL - strip protocol and trailing slash
  const imgHost = `${window.location.protocol}//${
    language === 'ru'
      ? settings.ImageURLRu.replace(/^https?:\/\//, '').replace(/\/$/, '')
      : settings.ImageURL.replace(/^https?:\/\//, '').replace(/\/$/, '')
  }`

  return axios
    .get(url, {
      params: {
        api_key: settings.APIKey,
        language,
        include_image_language: `${language},null,en`,
        query: movieName,
      },
    })
    .then(({ data: { results } }) =>
      results.filter(el => el.poster_path).map(el => `${imgHost}/t/p/w300${el.poster_path}`),
    )
    .catch(() => null)
}

export const checkImageURL = async url => {
  if (!url || !url.match(/.(\.jpg|\.jpeg|\.png|\.gif|\.svg||\.webp).*$/i)) return false
  return true
}

const magnetRegex = /^magnet:\?xt=urn:[a-z0-9].*/i
export const hashRegex = /^\b[0-9a-f]{32}\b$|^\b[0-9a-f]{40}\b$|^\b[0-9a-f]{64}\b$/i
const torrentRegex = /^.*\.(torrent)$/i
const linkRegex = /^(http(s?)):\/\/.*/i
const torrsRegex = /^(torrs):\/\/.*/i

export const checkTorrentSource = source =>
  source.match(hashRegex) !== null ||
  source.match(magnetRegex) !== null ||
  source.match(torrentRegex) !== null ||
  source.match(linkRegex) !== null ||
  source.match(torrsRegex) !== null

/** Max length for TMDB/search API query; long torrent names exceed this. */
const POSTER_SEARCH_MAX_LEN = 50
/** Max words to use from title for poster search. */
const POSTER_SEARCH_MAX_WORDS = 4

/**
 * Shortens a long torrent title for poster search (TMDB).
 * Uses part before " [", " (", " / " and limits by words/length so the API gets a valid query.
 * @param {string} fullTitle - Raw torrent title
 * @param {{ maxWords?: number, maxLen?: number }} opts - Optional limits
 * @returns {string} Short title suitable for getMoviePosters()
 */
export const shortenTitleForPosterSearch = (fullTitle, opts = {}) => {
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
  } catch (_) {
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

export const parseTorrentTitle = (parsingSource, callback) => {
  parseTorrent.remote(parsingSource, (err, { name, files } = {}) => {
    if (!name || err) return callback({ parsedTitle: null, originalName: null })

    const torrentName = ptt.parse(name).title
    const nameOfFileInsideTorrent = files ? ptt.parse(files[0].name).title : null

    let newTitle = torrentName
    if (nameOfFileInsideTorrent) {
      // taking shorter title because in most cases it is more accurate
      newTitle = torrentName.length < nameOfFileInsideTorrent.length ? torrentName : nameOfFileInsideTorrent
    }

    callback({ parsedTitle: newTitle, originalName: name })
  })
}

import { getTorrServerHost, subsHost } from './Hosts'

// The server host may be relative (REACT_APP_SERVER_HOST=".") and a relative URL cannot be the base of new URL():
// resolve it against the document first, the same way fetch() resolves the other links to the server
const serverUrl = path => new URL(path, new URL(`${getTorrServerHost()}/`, document.baseURI))

// Link to the list of external subtitle tracks of a video file: GET /subs/{hash}/{id}
export const subsListUrl = (hash, fileId) => `${subsHost()}/${encodeURIComponent(hash)}/${fileId}`

// Resolves to [] on any error: subtitles are optional (old servers have no /subs)
export const fetchExternalSubtitles = async (url, signal) => {
  try {
    const response = await fetch(url, { signal })
    if (!response.ok) return []
    const list = await response.json()
    return Array.isArray(list) ? list : []
  } catch {
    return []
  }
}

// Absolute link to the subtitle file (the server converts it to WebVTT)
export const subtitleFileUrl = track => serverUrl(track.url).toString()

// File index (the "index" query parameter) of a stream link such as the one used for captions
export const linkFileIndex = link => {
  const match = /[?&]index=(\d+)/.exec(link || '')
  return match ? Number(match[1]) : null
}

// Name of an external track: its title, or the file name
export const trackLabel = track => track.title || track.name || 'Subtitle'

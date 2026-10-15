import { authFetch } from 'shared/api/authCredentials'
import { getTorrServerHost, subsHost } from 'shared/api/hosts'

export type SubTrack = {
  index: number
  name: string
  title: string
  lang: string
  format: string
  url: string
}

/**
 * Progressive-player caption URL via `GET /subs/{hash}/{id}` (server Match + WebVTT).
 * Empty string when none found or the request fails. HLS gets external subs from GST.
 */
export const resolveCaptionSrc = async (
  hash: string,
  fileId: number,
  preferLang?: string,
): Promise<string> => {
  try {
    const res = await authFetch(subsHost(hash, fileId), { cache: 'no-store' })
    if (!res.ok) return ''
    const tracks = (await res.json()) as SubTrack[]
    if (!Array.isArray(tracks) || tracks.length === 0) return ''
    const lang = preferLang?.slice(0, 2).toLowerCase()
    const pick = (lang && tracks.find(t => t.lang === lang)) || tracks[0]
    if (!pick?.url) return ''
    return pick.url.startsWith('http') ? pick.url : `${getTorrServerHost()}${pick.url}`
  } catch {
    return ''
  }
}

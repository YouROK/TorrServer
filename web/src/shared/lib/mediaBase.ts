import { useQuery } from '@tanstack/react-query'

import { authFetch, withAuthMediaUrl } from 'shared/api/authCredentials'
import { mediaBaseHost } from 'shared/api/hosts'

export const MEDIA_BASE_QUERY_KEY = 'external-media-base'

interface MediaBaseResponse {
  base?: string
}

/**
 * The server decides which scheme://host:port external players should use. It differs
 * from the page origin when the UI runs on TorrServer's self-signed HTTPS certificate,
 * which players reject. Not cached for long: certificates and settings change without
 * a restart.
 */
const loadMediaBase = async (): Promise<string | null> => {
  const response = await authFetch(mediaBaseHost())
  if (!response.ok) return null
  const data = (await response.json()) as MediaBaseResponse
  return data.base || null
}

export const useExternalMediaBase = (): string | null => {
  const { data } = useQuery<string | null, Error>({
    queryKey: [MEDIA_BASE_QUERY_KEY],
    queryFn: loadMediaBase,
    staleTime: 30 * 1000,
    gcTime: 60 * 1000,
    retry: 1,
    refetchOnWindowFocus: false,
  })

  return data || null
}

/**
 * Absolute URL of a media link for external players and copied links.
 * Scheme and host come from the media base. Basic userinfo is added when the host matches the page.
 * In-page playback must keep the original link (`withAuthMediaUrl` on the `<video>` src).
 */
export const externalLink = (link: string, base: string | null | undefined): URL => {
  const url = new URL(link, window.location.href)
  if (base) {
    try {
      const target = new URL(base)
      url.protocol = target.protocol
      url.hostname = target.hostname
      // set explicitly: assigning a host without a port would keep the old port
      url.port = target.port
    } catch {
      // malformed base: keep the original link
    }
  }
  return new URL(withAuthMediaUrl(url.toString()))
}

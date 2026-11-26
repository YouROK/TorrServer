import { useQuery } from 'react-query'

import { mediaBaseHost } from './Hosts'

export const MEDIA_BASE_QUERY_KEY = 'external-media-base'

// The server decides which scheme://host:port external players should use. It differs
// from the page origin when the UI runs on TorrServer's self-signed HTTPS certificate,
// which players reject. Not cached for long: certificates and settings change without
// a restart.
const loadMediaBase = async () => {
  const response = await fetch(mediaBaseHost())
  if (!response.ok) return null
  const { base } = await response.json()
  return base || null
}

export const useExternalMediaBase = () => {
  const { data } = useQuery(MEDIA_BASE_QUERY_KEY, loadMediaBase, {
    staleTime: 30 * 1000,
    cacheTime: 60 * 1000,
    retry: 1,
  })

  return data || null
}

// externalLink returns the absolute URL of a media link for external players and copied
// links. Only scheme and host are replaced; in-page playback must keep the original link.
export const externalLink = (link, base) => {
  const url = new URL(link, window.location.href)
  if (!base) return url
  try {
    const target = new URL(base)
    url.protocol = target.protocol
    url.hostname = target.hostname
    // set explicitly: assigning a host without a port would keep the old port
    url.port = target.port
  } catch {
    // malformed base: keep the original link
  }
  return url
}

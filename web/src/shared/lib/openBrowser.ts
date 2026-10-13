import { authFetch } from 'shared/api/authCredentials'

import { copyToClipboard } from './clipboard'
import { detectApplePlatform, detectStandaloneApp } from './platform'
import type { PlaylistButtonAction } from './posterPlay'

/** Desktop and mobile Safari: a new tab downloads the m3u attachment. */
export const browserLinkAttrs = {
  target: '_blank',
  rel: 'noopener noreferrer',
} as const

/**
 * iOS home-screen webview has no back control on the system file preview.
 * Handle the playlist in the page instead of navigating to it.
 */
export function savePlaylistInApp(ios: boolean, standalone: boolean): boolean {
  return ios && standalone
}

export function playlistFilename(pathname: string): string {
  let name = pathname.split('/').pop() || ''
  try {
    name = decodeURIComponent(name)
  } catch {
    // keep the raw segment
  }
  return /\.m3u8?$/i.test(name) ? name : 'playlist.m3u'
}

/** When the control is not itself an anchor. Call from the click turn. */
export function openInBrowser(url: string) {
  const a = document.createElement('a')
  a.href = url
  a.target = browserLinkAttrs.target
  a.rel = browserLinkAttrs.rel
  document.body.appendChild(a)
  a.click()
  a.remove()
}

/** Library “all playlists” and any other non-anchor playlist action. */
export function openPlaylistLink(url: string): Promise<'shared' | 'copied' | 'opened' | 'error'> {
  if (!savePlaylistInApp(detectApplePlatform().isIOS, detectStandaloneApp())) {
    openInBrowser(url)
    return Promise.resolve('opened')
  }
  return deliverPlaylist(url)
}

/**
 * Playlist button. `playerHref` is set only for an external player — download and copy pass null
 * so this does not import the player schemes (that module pulls in `window` at load).
 */
export function bindPlaylistButton(
  action: PlaylistButtonAction,
  url: string,
  playerHref: string | null,
  onResult: (kind: 'copied' | 'error') => void,
) {
  return (event: { preventDefault: () => void }) => {
    if (playerHref) {
      event.preventDefault()
      window.location.href = playerHref
      return
    }
    if (action === 'copyLink') {
      event.preventDefault()
      void copyToClipboard(url).then(
        () => onResult('copied'),
        () => onResult('error'),
      )
      return
    }
    bindPlaylistClick(url, onResult)(event)
  }
}

export function bindPlaylistClick(url: string, onResult: (kind: 'copied' | 'error') => void) {
  return (event: { preventDefault: () => void }) => {
    if (!savePlaylistInApp(detectApplePlatform().isIOS, detectStandaloneApp())) return
    event.preventDefault()
    void deliverPlaylist(url).then(kind => {
      if (kind === 'copied' || kind === 'error') onResult(kind)
    })
  }
}

async function deliverPlaylist(url: string): Promise<'shared' | 'copied' | 'error'> {
  if (typeof navigator.share === 'function') {
    try {
      const res = await authFetch(url)
      if (!res.ok) throw new Error(String(res.status))
      const file = new File([await res.blob()], playlistFilename(new URL(url, window.location.origin).pathname), {
        type: 'application/vnd.apple.mpegurl',
      })
      const nav = navigator as Navigator & { canShare?: (data: ShareData) => boolean }
      if (nav.canShare && !nav.canShare({ files: [file] })) throw new Error('share')
      await nav.share({ files: [file] })
      return 'shared'
    } catch (error) {
      if (error instanceof DOMException && error.name === 'AbortError') return 'shared'
    }
  }
  try {
    await copyToClipboard(url)
    return 'copied'
  } catch {
    return 'error'
  }
}

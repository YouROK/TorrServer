/** Lightweight now-playing signal for the ambient bar (player → Shell). */
export const NOW_PLAYING_EVENT = 'torrserver:now-playing'

export interface NowPlayingInfo {
  title: string
  hash?: string
}

export const setNowPlaying = (info: NowPlayingInfo | null): void => {
  try {
    window.dispatchEvent(new CustomEvent(NOW_PLAYING_EVENT, { detail: info }))
  } catch {
    // ignore
  }
}

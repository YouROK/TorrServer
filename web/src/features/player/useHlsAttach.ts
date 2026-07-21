import { useEffect, useRef, type RefObject } from 'react'
import Hls from 'hls.js'

import { getAuthorizationHeader, withAuthMediaUrl } from 'shared/api/authCredentials'

export interface SubtitleTrackInfo {
  id: number
  name?: string
  lang?: string
}

export interface UseHlsAttachOptions {
  enabled: boolean
  open: boolean
  videoRef: RefObject<HTMLVideoElement | null>
  /** Bumps when the video node attaches / detaches so the effect re-runs. */
  videoEpoch: number
  src: string
  playbackRate?: number
  onLoading?: (loading: boolean) => void
  onError?: () => void
  onNotSupported?: () => void
  onSubtitleTracks?: (tracks: SubtitleTrackInfo[]) => void
  onSubtitleTrack?: (id: number) => void
  onDuration?: (seconds: number) => void
}

const supportsNativeHls = (video: HTMLVideoElement): boolean =>
  Boolean(video.canPlayType('application/vnd.apple.mpegurl') || video.canPlayType('application/x-mpegURL'))

/**
 * Attach / detach hls.js (or native HLS) while the player modal is open.
 * Progressive playback should leave `enabled` false and set `video.src` instead.
 */
export function useHlsAttach({
  enabled,
  open,
  videoRef,
  videoEpoch,
  src,
  playbackRate = 1,
  onLoading,
  onError,
  onNotSupported,
  onSubtitleTracks,
  onSubtitleTrack,
  onDuration,
}: UseHlsAttachOptions): RefObject<Hls | null> {
  const hlsRef = useRef<Hls | null>(null)
  const onLoadingRef = useRef(onLoading)
  const onErrorRef = useRef(onError)
  const onNotSupportedRef = useRef(onNotSupported)
  const onSubtitleTracksRef = useRef(onSubtitleTracks)
  const onSubtitleTrackRef = useRef(onSubtitleTrack)
  const onDurationRef = useRef(onDuration)

  useEffect(() => {
    onLoadingRef.current = onLoading
    onErrorRef.current = onError
    onNotSupportedRef.current = onNotSupported
    onSubtitleTracksRef.current = onSubtitleTracks
    onSubtitleTrackRef.current = onSubtitleTrack
    onDurationRef.current = onDuration
  })

  useEffect(() => {
    const video = videoRef.current
    if (!enabled || !open || !video || !src) return undefined

    let hlsPlayer: Hls | null = null
    let usingNativeHls = false
    onLoadingRef.current?.(true)
    onSubtitleTracksRef.current?.([])

    const syncSubtitleTracks = () => {
      const tracks = (hlsPlayer?.subtitleTracks || []).map((track, id) => ({
        id,
        name: track.name,
        lang: track.lang,
      }))
      onSubtitleTracksRef.current?.(tracks)
    }

    if (Hls.isSupported()) {
      const authHeader = getAuthorizationHeader()
      hlsPlayer = new Hls({
        xhrSetup: xhr => {
          if (authHeader) xhr.setRequestHeader('Authorization', authHeader)
        },
      })
      hlsRef.current = hlsPlayer

      hlsPlayer.on(Hls.Events.MANIFEST_PARSED, () => {
        syncSubtitleTracks()
        onSubtitleTrackRef.current?.(hlsPlayer?.subtitleTrack ?? -1)
        video.playbackRate = playbackRate
        video.play().catch(() => {})
      })
      hlsPlayer.on(Hls.Events.LEVEL_LOADED, (_event, data) => {
        const total = data.details?.totalduration
        if (typeof total === 'number') onDurationRef.current?.(total)
      })
      hlsPlayer.on(Hls.Events.SUBTITLE_TRACKS_UPDATED, syncSubtitleTracks)
      hlsPlayer.on(Hls.Events.SUBTITLE_TRACK_SWITCH, (_event, data) => onSubtitleTrackRef.current?.(data.id))
      hlsPlayer.on(Hls.Events.ERROR, (_event, data) => {
        if (!data.fatal || !hlsPlayer) return
        if (data.type === Hls.ErrorTypes.NETWORK_ERROR) {
          hlsPlayer.startLoad()
        } else if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
          hlsPlayer.recoverMediaError()
        } else {
          hlsPlayer.stopLoad()
          onLoadingRef.current?.(false)
          onErrorRef.current?.()
        }
      })

      hlsPlayer.loadSource(src)
      hlsPlayer.attachMedia(video)
    } else if (supportsNativeHls(video)) {
      usingNativeHls = true
      video.src = withAuthMediaUrl(src)
      video.load()
      video.playbackRate = playbackRate
      video.play().catch(() => {})
    } else {
      onNotSupportedRef.current?.()
    }

    return () => {
      if (hlsRef.current === hlsPlayer) hlsRef.current = null
      hlsPlayer?.destroy()
      if (usingNativeHls) {
        video.pause()
        video.removeAttribute('src')
        video.load()
      }
      onSubtitleTracksRef.current?.([])
      onSubtitleTrackRef.current?.(-1)
    }
    // playbackRate applied on manifest parse; remount when source changes
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabled, open, videoRef, videoEpoch, src])

  return hlsRef
}

export function switchHlsSubtitleTrack(hls: Hls | null, trackId: number): void {
  if (!hls) return
  hls.subtitleTrack = trackId
}

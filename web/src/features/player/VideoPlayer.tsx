import { useCallback, useEffect, useRef, useState, type CSSProperties, type ReactNode } from 'react'
import axios from 'axios'
import Hls from 'hls.js'
import { Alert, Button, Modal, Popover, Spinner, useMediaQuery, useOverlayState } from '@heroui/react'
import { ArrowDown, Database, Maximize2, Minimize2, Music2, Play, Users, X } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { authFetch, withAuthMediaUrl } from 'shared/api/authCredentials'
import { useUpdateCache } from 'shared/cache/useUpdateCache'
import { useTorrentDetail } from 'shared/hooks/useTorrentDetail'
import { formatCacheFilledLabel, humanizeSpeed } from 'shared/lib/format'
import { gstreamerMasterUrl, gstreamerProbeUrl } from 'shared/lib/gstreamer'
import { setNowPlaying } from 'shared/lib/nowPlaying'
import { queryMax } from 'shared/theme/breakpoints'
import { iconBtn } from 'shared/ui/controlClasses'
import { PLAYER_DIALOG_EXPANDED, PLAYER_DIALOG_MOBILE, PLAYER_DIALOG_NORMAL } from 'shared/ui/dialogSizes'
import { iconMenu, iconPlayerCompact } from 'shared/ui/iconProps'
import { useModalOpen, useSyncModalOpen } from 'shared/ui/ModalOpenContext'

import { extractAudioTracks, formatAudioTrackDisplay, type ProbeTrack } from './audioTrackLabel'
import PlayerChrome from './PlayerChrome'
import { useHlsAttach, type SubtitleTrackInfo } from './useHlsAttach'
import { useTimecodePersist } from './useTimecodePersist'

export interface VideoPlayerProps {
  videoSrc: string
  downloadSrc?: string
  title?: string
  onNotSupported?: () => void
  hls?: boolean
  heartbeatSrc?: string
  showTrigger?: boolean
  inlineTrigger?: boolean
  inlineTriggerPrimary?: boolean
  initiallyOpen?: boolean
  onClose?: () => void
  captionSrc?: string
  hash?: string
  fileIndex?: number
  initialTimecode?: number
  trackTimecode?: boolean
  onViewedChange?: () => void
  audioTracks?: ProbeTrack[]
  audioIndex?: number
  onAudioIndexChange?: (index: number) => void
}

const HEARTBEAT_INTERVAL_MS = 30_000

function PlayerStat({ tip, children }: { tip: string; children: ReactNode }) {
  return (
    <span className='inline-flex min-w-0 max-w-full items-center gap-1' title={tip}>
      {children}
    </span>
  )
}

function nativeMimeType(url: string): string {
  switch (url.split('?')[0].split('.').pop()?.toLowerCase()) {
    case 'mp4':
      return 'video/mp4'
    case 'ogg':
    case 'ogv':
      return 'video/ogg'
    case 'webm':
      return 'video/webm'
    default:
      return ''
  }
}

const supportsNativeHls = (video: HTMLVideoElement): boolean =>
  Boolean(video.canPlayType('application/vnd.apple.mpegurl') || video.canPlayType('application/x-mpegURL'))

const supportsPiP = (): boolean =>
  typeof document !== 'undefined' && 'pictureInPictureEnabled' in document && Boolean(document.pictureInPictureEnabled)

/**
 * In-app stream player — Media Chrome OSD + hls.js / progressive TorrServer URLs.
 */
export default function VideoPlayer({
  videoSrc,
  downloadSrc = videoSrc,
  title,
  onNotSupported,
  hls = false,
  heartbeatSrc = '',
  showTrigger = true,
  inlineTrigger = false,
  inlineTriggerPrimary = false,
  initiallyOpen = false,
  onClose,
  captionSrc,
  hash,
  fileIndex,
  initialTimecode = 0,
  trackTimecode = false,
  onViewedChange,
  audioTracks = [],
  audioIndex = 0,
  onAudioIndexChange,
}: VideoPlayerProps) {
  const { t } = useTranslation()
  const isMobile = useMediaQuery(queryMax('dialog'))
  const { setImmersive } = useModalOpen()

  const videoRef = useRef<HTMLVideoElement | null>(null)
  const onNotSupportedRef = useRef(onNotSupported)
  const pendingSeekRef = useRef<number | null>(null)

  const [open, setOpen] = useState(initiallyOpen)
  const [videoEpoch, setVideoEpoch] = useState(0)
  const [loading, setLoading] = useState(true)
  const [mediaError, setMediaError] = useState(false)
  const [fullscreen, setFullscreen] = useState(false)
  const [expanded, setExpanded] = useState(false)
  const [subtitleTracks, setSubtitleTracks] = useState<SubtitleTrackInfo[]>([])
  const [audioMenuOpen, setAudioMenuOpen] = useState(false)
  const [activeAudioIndex, setActiveAudioIndex] = useState(audioIndex)
  const [playbackSrc, setPlaybackSrc] = useState(videoSrc)
  const [resolvedAudioTracks, setResolvedAudioTracks] = useState<ProbeTrack[]>(audioTracks)

  const liveHash = open ? hash : undefined
  const { data: liveTorrent } = useTorrentDetail(liveHash)
  const cache = useUpdateCache(hash, { enabled: open && Boolean(hash), fast: false })

  const shouldPersistTimecode = Boolean(hash && fileIndex != null && trackTimecode)
  const canSwitchAudio = hls && Boolean(hash) && fileIndex != null && resolvedAudioTracks.length > 1
  const showPip = supportsPiP()
  const showCaptionsButton = hls ? subtitleTracks.length > 0 : Boolean(captionSrc)

  const dialogStyle: CSSProperties | undefined = isMobile
    ? PLAYER_DIALOG_MOBILE
    : expanded
      ? PLAYER_DIALOG_EXPANDED
      : PLAYER_DIALOG_NORMAL

  const cinemaMaxHeight = isMobile ? undefined : expanded ? 'min(86dvh, calc(100dvh - 4rem))' : 'min(72dvh, 40rem)'

  const attachVideoNode = useCallback((node: HTMLVideoElement | null) => {
    videoRef.current = node
    setVideoEpoch(epoch => epoch + 1)
  }, [])

  useEffect(() => {
    onNotSupportedRef.current = onNotSupported
  }, [onNotSupported])

  // Reset media session when the launcher hands a new source / audio track.
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- remount playback session for new src/audio props
    setPlaybackSrc(videoSrc)
    setActiveAudioIndex(audioIndex)
    setResolvedAudioTracks(audioTracks)
    setMediaError(false)
    setLoading(true)
  }, [videoSrc, audioIndex, audioTracks])

  useHlsAttach({
    enabled: hls,
    open,
    videoRef,
    videoEpoch,
    src: playbackSrc,
    onLoading: setLoading,
    onError: () => setMediaError(true),
    onNotSupported: () => onNotSupportedRef.current?.(),
    onSubtitleTracks: setSubtitleTracks,
  })

  const { flushTimecode, onTimeUpdate, applyResumeIfNeeded } = useTimecodePersist({
    enabled: shouldPersistTimecode,
    hash,
    fileIndex,
    title,
    initialTimecode,
    videoRef,
    onViewedChange,
  })

  const openPlayer = useCallback(() => setOpen(true), [])

  const closePlayer = useCallback(() => {
    flushTimecode()
    setOpen(false)
    setMediaError(false)
    setAudioMenuOpen(false)
    setExpanded(false)
    onClose?.()
  }, [flushTimecode, onClose])

  const overlayState = useOverlayState({
    isOpen: open,
    onOpenChange: next => {
      if (next) setOpen(true)
      else closePlayer()
    },
  })

  useSyncModalOpen(open)

  useEffect(() => {
    if (!isMobile || !open) {
      setImmersive(false)
      return undefined
    }
    setImmersive(true)
    return () => setImmersive(false)
  }, [isMobile, open, setImmersive])

  useEffect(() => {
    if (!open) {
      setNowPlaying(null)
      return
    }
    // Suppress ambient Now Playing pill while the cinema modal is open (title is in OSD).
    setNowPlaying(null)
    return () => setNowPlaying(null)
  }, [open])

  useEffect(() => {
    const probe = document.createElement('video')
    const supported = hls
      ? Hls.isSupported() || supportsNativeHls(probe)
      : Boolean(probe.canPlayType(nativeMimeType(playbackSrc)))
    if (!supported) onNotSupportedRef.current?.()
  }, [hls, playbackSrc])

  // Progressive: set src on the element when not using hls.js
  useEffect(() => {
    const videoElement = videoRef.current
    if (!open || hls || !videoElement) return undefined
    setLoading(true)
    videoElement.src = withAuthMediaUrl(playbackSrc)
    videoElement.load()
    videoElement.play().catch(() => {})
    return () => {
      videoElement.pause()
      videoElement.removeAttribute('src')
      videoElement.load()
    }
  }, [open, hls, videoEpoch, playbackSrc])

  useEffect(() => {
    if (!open || !heartbeatSrc) return undefined
    const timer = window.setInterval(() => {
      authFetch(heartbeatSrc, { cache: 'no-store' }).catch(() => {})
    }, HEARTBEAT_INTERVAL_MS)
    return () => window.clearInterval(timer)
  }, [heartbeatSrc, open])

  useEffect(() => {
    const onFullscreenChange = () => setFullscreen(Boolean(document.fullscreenElement))
    document.addEventListener('fullscreenchange', onFullscreenChange)
    return () => document.removeEventListener('fullscreenchange', onFullscreenChange)
  }, [])

  // Restore position after GST audio remount
  useEffect(() => {
    const videoElement = videoRef.current
    if (!videoElement || pendingSeekRef.current == null) return undefined
    const target = pendingSeekRef.current
    const apply = () => {
      if (pendingSeekRef.current == null) return
      videoElement.currentTime = target
      pendingSeekRef.current = null
      videoElement.play().catch(() => {})
    }
    videoElement.addEventListener('loadedmetadata', apply, { once: true })
    return () => videoElement.removeEventListener('loadedmetadata', apply)
  }, [videoEpoch, playbackSrc])

  useEffect(() => {
    if (!open || !hls || !hash || fileIndex == null) return undefined
    if (audioTracks.length) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- adopt probe tracks passed from launcher
      setResolvedAudioTracks(audioTracks)
      return undefined
    }
    let cancelled = false
    void axios
      .get(gstreamerProbeUrl(hash, fileIndex))
      .then(({ data }) => {
        if (cancelled) return
        setResolvedAudioTracks(extractAudioTracks(data))
      })
      .catch(() => {})
    return () => {
      cancelled = true
    }
  }, [open, hls, hash, fileIndex, audioTracks])

  const switchAudioTrack = (index: number) => {
    if (!hash || fileIndex == null) return
    const video = videoRef.current
    if (video) pendingSeekRef.current = video.currentTime
    setActiveAudioIndex(index)
    setPlaybackSrc(gstreamerMasterUrl(hash, fileIndex, index))
    setAudioMenuOpen(false)
    onAudioIndexChange?.(index)
  }

  const chromeIcon = iconPlayerCompact

  const downloadSpeed = liveTorrent?.download_speed ?? 0
  const peersActive = liveTorrent?.active_peers
  const peersTotal = liveTorrent?.total_peers ?? 0
  const seeders = liveTorrent?.connected_seeders ?? 0
  const peersLabel = peersActive == null ? '—' : `${peersActive}/${peersTotal}`
  const seedersLabel = peersActive == null ? '—' : `↑${seeders}`
  const cacheFilledLabel = formatCacheFilledLabel(cache.Filled, cache.Capacity) ?? '—'

  const torrentStatsRow =
    hash != null ? (
      <div
        className='mt-0.5 flex min-w-0 flex-wrap items-center gap-x-2.5 gap-y-1 text-[11px] leading-none text-white/80 drop-shadow sm:text-xs'
        aria-label={`${t('DownloadSpeed')}, ${t('Peers')}, ${t('CacheFilled')}`}
      >
        <PlayerStat tip={t('DownloadSpeed')}>
          <ArrowDown className='size-3 shrink-0' strokeWidth={2.25} aria-hidden />
          <span className='tabular-nums'>{humanizeSpeed(downloadSpeed)}</span>
        </PlayerStat>
        <PlayerStat tip={t('Peers')}>
          <Users className='size-3 shrink-0' strokeWidth={2} aria-hidden />
          <span className='tabular-nums'>{peersLabel}</span>
          <span className='tabular-nums text-white/70'>{seedersLabel}</span>
        </PlayerStat>
        <PlayerStat tip={t('CacheFilled')}>
          <Database className='size-3 shrink-0' strokeWidth={2} aria-hidden />
          <span className='min-w-0 truncate tabular-nums'>{cacheFilledLabel}</span>
        </PlayerStat>
      </div>
    ) : null

  const audioExtra = canSwitchAudio ? (
    <Popover isOpen={audioMenuOpen} onOpenChange={setAudioMenuOpen}>
      <Popover.Trigger>
        <Button
          isIconOnly
          variant='ghost'
          className={`${iconBtn} text-white hover-fine:bg-white/15`}
          aria-label={t('SelectAudioTrack')}
        >
          <Music2 {...chromeIcon} aria-hidden />
        </Button>
      </Popover.Trigger>
      <Popover.Content
        placement='bottom end'
        className='max-h-72 w-72 overflow-y-auto border border-white/10 bg-neutral-950/95 p-1 backdrop-blur-xl'
      >
        <p className='px-2 py-1.5 text-[11px] font-semibold tracking-wide text-muted uppercase'>
          {t('SelectAudioTrack')}
        </p>
        {resolvedAudioTracks.map((track, index) => {
          const { title: trackTitle, meta } = formatAudioTrackDisplay(track, index)
          const selected = index === activeAudioIndex
          return (
            <Button
              key={index}
              variant={selected ? 'secondary' : 'ghost'}
              className='h-auto w-full justify-start gap-2 py-2'
              onPress={() => switchAudioTrack(index)}
            >
              <span className='min-w-0 flex-1 text-left'>
                <span className='block truncate text-sm font-medium'>{trackTitle}</span>
                {meta ? <span className='mt-0.5 block truncate text-xs text-muted'>{meta}</span> : null}
              </span>
            </Button>
          )
        })}
      </Popover.Content>
    </Popover>
  ) : null

  const expandExtra = !isMobile ? (
    <Button
      isIconOnly
      variant='ghost'
      className={`${iconBtn} text-white hover-fine:bg-white/15`}
      aria-label={expanded ? t('ExitFullscreen') : t('ExpandPlayer')}
      onPress={() => setExpanded(v => !v)}
    >
      {expanded ? <Minimize2 {...chromeIcon} aria-hidden /> : <Maximize2 {...chromeIcon} aria-hidden />}
    </Button>
  ) : null

  return (
    <>
      {showTrigger &&
        (inlineTrigger ? (
          <Button
            variant={inlineTriggerPrimary ? 'primary' : 'secondary'}
            size='sm'
            onPress={openPlayer}
            className={inlineTriggerPrimary ? 'min-h-10 shrink-0 px-3' : 'min-h-10 min-w-[72px] max-w-full flex-1'}
          >
            {inlineTriggerPrimary ? <Play {...iconMenu} fill='currentColor' aria-hidden /> : null}
            {t('Play')}
          </Button>
        ) : (
          <Button variant='secondary' onPress={openPlayer}>
            <Play {...iconMenu} aria-hidden />
            {t('Play')}
          </Button>
        ))}

      <Modal state={overlayState}>
        <Modal.Backdrop isDismissable={!fullscreen} className='bg-black/75 backdrop-blur-sm'>
          <Modal.Container
            size={isMobile ? 'full' : 'lg'}
            scroll='inside'
            className={isMobile ? 'ts-player-modal-full h-dvh p-0' : 'p-4 sm:p-5'}
          >
            <Modal.Dialog
              className={
                isMobile
                  ? 'h-dvh max-h-dvh overflow-hidden rounded-none border-0 bg-black text-white shadow-none'
                  : 'overflow-hidden border border-white/10 bg-[#0a0e0c] text-white shadow-2xl shadow-black/50'
              }
              style={dialogStyle}
            >
              <Modal.Body className={isMobile ? 'flex h-full min-h-0 flex-col gap-0 p-0' : 'gap-0 p-0'}>
                {mediaError ? (
                  <div className='space-y-3 p-4'>
                    <Alert status='danger'>{t('PlaybackError')}</Alert>
                    <Button
                      variant='secondary'
                      onPress={() => window.open(downloadSrc, '_blank', 'noopener,noreferrer')}
                    >
                      {t('OpenLink')}
                    </Button>
                  </div>
                ) : (
                  <PlayerChrome
                    isMobile={isMobile}
                    showPip={showPip}
                    showCaptionsButton={showCaptionsButton}
                    className={isMobile ? 'ts-player--mobile' : 'ts-player--desktop'}
                    style={
                      isMobile
                        ? { width: '100%', height: '100%' }
                        : { width: '100%', maxHeight: cinemaMaxHeight, aspectRatio: '16 / 9' }
                    }
                    topChrome={
                      <div className='flex items-start gap-2 px-3 pb-10 pt-[max(0.75rem,env(safe-area-inset-top))] pl-[max(0.75rem,env(safe-area-inset-left))] pr-[max(0.75rem,env(safe-area-inset-right))]'>
                        <div className='min-w-0 flex-1'>
                          <p
                            className='truncate text-sm font-semibold text-white drop-shadow'
                            title={title || t('Play')}
                          >
                            {title || t('Play')}
                          </p>
                          {torrentStatsRow}
                        </div>
                        {audioExtra}
                        {expandExtra}
                        <Button
                          isIconOnly
                          variant='ghost'
                          className={`${iconBtn} text-white hover-fine:bg-white/15`}
                          aria-label={t('Close')}
                          onPress={closePlayer}
                        >
                          <X {...chromeIcon} aria-hidden />
                        </Button>
                      </div>
                    }
                    extraControls={null}
                  >
                    <video
                      key={playbackSrc}
                      slot='media'
                      ref={attachVideoNode}
                      autoPlay
                      playsInline
                      crossOrigin='anonymous'
                      className='ts-player-video'
                      onTimeUpdate={() => {
                        onTimeUpdate()
                        if (loading) setLoading(false)
                      }}
                      onLoadedMetadata={() => {
                        setLoading(false)
                        setMediaError(false)
                        applyResumeIfNeeded()
                      }}
                      onWaiting={() => setLoading(true)}
                      onCanPlay={() => {
                        setLoading(false)
                        setMediaError(false)
                      }}
                      onPlaying={() => setLoading(false)}
                      onPause={flushTimecode}
                      onError={() => {
                        setLoading(false)
                        setMediaError(true)
                      }}
                    >
                      {!hls && captionSrc ? (
                        <track kind='captions' src={captionSrc} srcLang='und' label='Captions' default />
                      ) : null}
                    </video>
                  </PlayerChrome>
                )}
                {loading && !mediaError ? (
                  <div className='pointer-events-none absolute inset-0 z-10 grid place-items-center bg-black/55'>
                    <div className='flex flex-col items-center gap-3 px-4 text-center'>
                      <Spinner size='lg' color='current' className='text-accent' />
                      <p className='text-sm font-medium text-white/90'>{t('Buffering')}</p>
                    </div>
                  </div>
                ) : null}
              </Modal.Body>
            </Modal.Dialog>
          </Modal.Container>
        </Modal.Backdrop>
      </Modal>
    </>
  )
}

export type { SubtitleTrackInfo }

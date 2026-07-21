import { lazy, Suspense, useEffect, useMemo, useRef, useState } from 'react'
import axios from 'axios'
import { Button, Modal, Spinner, useOverlayState } from '@heroui/react'
import { Music2 } from 'lucide-react'
import { iconMenu } from 'shared/ui/iconProps'
import ptt from 'parse-torrent-title'
import { useTranslation } from 'react-i18next'
import { useQueryClient } from '@tanstack/react-query'
import type { BTSets, PlayableFile, TorrentFileStat } from 'shared/api/types'
import { streamHost } from 'shared/api/hosts'
import { getSettings, SETTINGS_QUERY_KEY } from 'shared/api/settings'
import { getTorrent } from 'shared/api/torrents'
import { listViewedEntries, VIEWED_QUERY_KEY } from 'shared/api/viewed'
import { humanizeSize } from 'shared/lib/format'
import {
  gstreamerHeartbeatUrl,
  gstreamerMasterUrl,
  gstreamerProbeUrl,
  shouldUseGStreamerPlayer,
  useGStreamerRuntime,
} from 'shared/lib/gstreamer'
import { useSyncModalOpen } from 'shared/ui/ModalOpenContext'
import { useOptionalAppToast } from 'shared/ui/Toast'

import { extractAudioTracks, formatAudioTrackDisplay, type ProbeTrack } from './audioTrackLabel'
import { waitForPlayableFiles } from './waitForPlayableFiles'
import { toPlayableFile } from 'shared/torrent/toPlayableFile'

/** Lazy: keeps hls.js out of the initial bundle — only fetched once a file is actually played. */
const VideoPlayer = lazy(() => import('features/player/VideoPlayer'))

interface ActivePlayer {
  src: string
  downloadSrc: string
  title: string
  hls: boolean
  heartbeatSrc?: string
  captionSrc?: string
  hash: string
  fileIndex: number
  initialTimecode: number
  trackTimecode: boolean
  audioTracks: ProbeTrack[]
  audioIndex: number
}

const fileBaseName = (path: string): string => path.split('\\').pop()?.split('/').pop() || path

const fileStreamUrl = (hash: string, file: Pick<PlayableFile, 'id' | 'path'>): string =>
  `${streamHost()}/${encodeURIComponent(fileBaseName(file.path))}?link=${hash}&index=${file.id}&play`

/**
 * Sidecar caption URL: same basename as the video plus `.srt` / `.vtt` elsewhere in the torrent.
 * Empty string when none found.
 */
export const findCaptionSrc = (file: PlayableFile, allFiles: TorrentFileStat[], hash: string): string => {
  const base = file.path.replace(/\.[^/.]+$/, '')
  const caption = allFiles.find(candidate => {
    const path = candidate.path ?? candidate.Path ?? ''
    const id = candidate.id ?? candidate.Id ?? -1
    return id !== file.id && path.startsWith(base) && /\.(srt|vtt)$/i.test(path)
  })
  if (!caption) return ''
  const captionFile = toPlayableFile(caption)
  return fileStreamUrl(hash, captionFile)
}

/** Parses a file's episode code ("S01E03") and title from its path, falling back to the raw name. */
const fileEpisodeInfo = (path: string, index: number): { code: string; title: string } => {
  const name = fileBaseName(path)
  const parsed = ptt.parse(name)
  const season = Number(parsed.season)
  const episode = Number(parsed.episode)
  const code = `${season ? `S${String(season).padStart(2, '0')}` : ''}${
    episode ? `E${String(episode).padStart(2, '0')}` : ''
  }`
  const title = parsed.title || name.replace(/\.[^/.]+$/, '')
  return { code: code || `#${index + 1}`, title }
}

export interface UsePlayLauncherArgs {
  hash: string
  displayName: string
  /** Playable files already known from a loaded torrent detail — falls back to fetching the torrent if empty. */
  knownPlayableFiles: PlayableFile[]
  /** Full file list (incl. sidecar captions) when already loaded in details. */
  knownAllFiles?: TorrentFileStat[]
  /** Persisted torrent.data for IN_DB fallback while live file_stats are empty. */
  torrentData?: string
  onViewedChange?: () => void
  onPlayerNotSupported?: (fileId: number) => void
  /** When set, auto-start this file once playable files are known (Continue Watching). */
  autoPlayFileId?: number
  /** Prefer this resume position over server viewed history (local Continue Watching). */
  autoPlayTimecode?: number
}

/**
 * Shared "Play" flow: resolves the file to play (prompting when there are several), probes audio
 * tracks for GStreamer-backed files, and opens the video player. Shared between the grid card's
 * quick-play button, the details dialog, and per-file rows.
 */
export function usePlayLauncher({
  hash,
  displayName,
  knownPlayableFiles,
  knownAllFiles,
  torrentData,
  onViewedChange,
  onPlayerNotSupported,
  autoPlayFileId,
  autoPlayTimecode,
}: UsePlayLauncherArgs) {
  const { t } = useTranslation()
  const toast = useOptionalAppToast()
  const queryClient = useQueryClient()
  const gstRuntime = useGStreamerRuntime()
  /** Per-file audio-track probe results, cached for the lifetime of this hook instance. */
  const audioProbeCache = useRef<Record<number, ProbeTrack[]>>({})
  const allTorrentFilesRef = useRef<TorrentFileStat[]>(knownAllFiles?.length ? knownAllFiles : [])
  const abortRef = useRef<AbortController | null>(null)
  const pendingTimecodeRef = useRef<number | undefined>(undefined)

  const [playableFiles, setPlayableFiles] = useState<PlayableFile[]>([])
  const [audioTracks, setAudioTracks] = useState<ProbeTrack[]>([])
  const [pendingAudioFile, setPendingAudioFile] = useState<PlayableFile | null>(null)
  const [isResolving, setIsResolving] = useState(false)
  const [resolvingFileId, setResolvingFileId] = useState<number | null>(null)
  const [activePlayer, setActivePlayer] = useState<ActivePlayer | null>(null)
  /** When set, file-picker selection hands off here instead of opening the built-in player. */
  const fileHandoffRef = useRef<((file: PlayableFile) => void) | null>(null)

  const filePickerState = useOverlayState()
  const audioPickerState = useOverlayState()

  useSyncModalOpen(filePickerState.isOpen || audioPickerState.isOpen)

  useEffect(() => {
    if (!filePickerState.isOpen) fileHandoffRef.current = null
  }, [filePickerState.isOpen])

  useEffect(() => {
    if (knownAllFiles?.length) allTorrentFilesRef.current = knownAllFiles
  }, [knownAllFiles])

  useEffect(() => {
    return () => {
      abortRef.current?.abort()
    }
  }, [])

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- clear resolving spinner when audio picker closes
    if (!audioPickerState.isOpen) setResolvingFileId(null)
  }, [audioPickerState.isOpen])

  const notifyViewedChange = () => {
    void queryClient.invalidateQueries({ queryKey: VIEWED_QUERY_KEY(hash) })
    onViewedChange?.()
  }

  const ensureTrackTimecode = async (): Promise<boolean> => {
    const cached = queryClient.getQueryData<BTSets>(SETTINGS_QUERY_KEY)
    if (cached) return Boolean(cached.TrackTimecode)
    try {
      const settings = await queryClient.ensureQueryData({
        queryKey: SETTINGS_QUERY_KEY,
        queryFn: ({ signal }) => getSettings(signal),
      })
      return Boolean(settings.TrackTimecode)
    } catch {
      return false
    }
  }

  const ensureAllTorrentFiles = async (): Promise<TorrentFileStat[]> => {
    if (allTorrentFilesRef.current.length) return allTorrentFilesRef.current
    const detail = await getTorrent(hash)
    allTorrentFilesRef.current = detail.file_stats || []
    return allTorrentFilesRef.current
  }

  const resolveInitialTimecode = async (
    fileIndex: number,
    trackTimecode: boolean,
    override?: number,
  ): Promise<number> => {
    if (override != null && override > 0) return override
    if (!trackTimecode) return 0
    const entries = await listViewedEntries(hash)
    return entries.find(entry => entry.file_index === fileIndex)?.timecode ?? 0
  }

  const openPlayer = async (
    file: PlayableFile,
    audioIndex = 0,
    tracksForPlayer: ProbeTrack[] = [],
    timecodeOverride?: number,
  ) => {
    const trackTimecode = await ensureTrackTimecode()
    const useHls = shouldUseGStreamerPlayer(file.path, gstRuntime)
    const directStream = fileStreamUrl(hash, file)
    const allFiles = await ensureAllTorrentFiles()
    const captionSrc = findCaptionSrc(file, allFiles, hash)
    const override = timecodeOverride ?? pendingTimecodeRef.current
    pendingTimecodeRef.current = undefined
    const initialTimecode = await resolveInitialTimecode(file.id, trackTimecode, override)
    const audioTracksForPlayer =
      tracksForPlayer.length > 0 ? tracksForPlayer : useHls ? audioProbeCache.current[file.id] || [] : []

    setActivePlayer({
      src: useHls ? gstreamerMasterUrl(hash, file.id, audioIndex) : directStream,
      downloadSrc: directStream,
      title: fileBaseName(file.path) || displayName,
      hls: useHls,
      heartbeatSrc: useHls ? gstreamerHeartbeatUrl(hash) : undefined,
      captionSrc: captionSrc || undefined,
      hash,
      fileIndex: file.id,
      initialTimecode,
      trackTimecode,
      audioTracks: audioTracksForPlayer,
      audioIndex,
    })
    filePickerState.close()
    audioPickerState.close()
    setPendingAudioFile(null)
    setResolvingFileId(null)
  }

  const resolveAndPlay = async (file: PlayableFile, timecodeOverride?: number) => {
    setResolvingFileId(file.id)
    setIsResolving(true)
    let openedAudioPicker = false
    pendingTimecodeRef.current = timecodeOverride
    try {
      const useHls = shouldUseGStreamerPlayer(file.path, gstRuntime)
      if (!useHls) {
        await openPlayer(file, 0, [], timecodeOverride)
        return
      }

      const openAudioPicker = (tracks: ProbeTrack[]) => {
        setPendingAudioFile(file)
        setAudioTracks(tracks)
        audioPickerState.open()
        openedAudioPicker = true
      }

      const cachedTracks = audioProbeCache.current[file.id]
      if (cachedTracks !== undefined) {
        if (cachedTracks.length <= 1) {
          await openPlayer(file, 0, cachedTracks, timecodeOverride)
          return
        }
        openAudioPicker(cachedTracks)
        return
      }

      try {
        const { data } = await axios.get(gstreamerProbeUrl(hash, file.id))
        const tracks = extractAudioTracks(data)
        audioProbeCache.current[file.id] = tracks
        if (tracks.length <= 1) {
          await openPlayer(file, 0, tracks, timecodeOverride)
          return
        }
        openAudioPicker(tracks)
      } catch {
        await openPlayer(file, 0, [], timecodeOverride)
      }
    } finally {
      setIsResolving(false)
      if (!openedAudioPicker) setResolvingFileId(null)
    }
  }

  const autoPlayDoneRef = useRef(false)
  useEffect(() => {
    if (autoPlayFileId == null || autoPlayDoneRef.current) return
    const file =
      knownPlayableFiles.find(item => item.id === autoPlayFileId) ||
      playableFiles.find(item => item.id === autoPlayFileId)
    if (!file) return
    autoPlayDoneRef.current = true
    // eslint-disable-next-line react-hooks/set-state-in-effect -- one-shot Continue Watching autoplay
    void resolveAndPlay(file, autoPlayTimecode)
    // eslint-disable-next-line react-hooks/exhaustive-deps -- one-shot resume when the target file appears
  }, [autoPlayFileId, autoPlayTimecode, knownPlayableFiles, playableFiles])

  const pickFileFromList = (file: PlayableFile) => {
    const handoff = fileHandoffRef.current
    fileHandoffRef.current = null
    if (handoff) {
      filePickerState.close()
      handoff(file)
      return
    }
    void resolveAndPlay(file)
  }

  /**
   * Resolve playable file(s): one file → `onFile` immediately; several → file picker then `onFile`.
   * Used by poster Play for copy-link / external-player handoff without opening VideoPlayer.
   */
  const resolvePlayableFile = async (onFile: (file: PlayableFile) => void) => {
    abortRef.current?.abort()
    const controller = new AbortController()
    abortRef.current = controller
    fileHandoffRef.current = null

    setIsResolving(true)
    setResolvingFileId(null)
    try {
      let candidates = knownPlayableFiles
      if (!candidates.length) {
        const { playable, allFiles } = await waitForPlayableFiles(hash, {
          metadataData: torrentData,
          signal: controller.signal,
        })
        allTorrentFilesRef.current = allFiles
        candidates = playable
      } else if (!allTorrentFilesRef.current.length) {
        await ensureAllTorrentFiles()
      }

      if (!candidates.length) {
        toast?.showToast({ message: t('NoPlayableFiles'), severity: 'info' })
        return
      }
      if (candidates.length === 1) {
        onFile(candidates[0])
        return
      }
      fileHandoffRef.current = onFile
      setPlayableFiles(candidates)
      filePickerState.open()
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') return
      toast?.showToast({ message: t('Error'), severity: 'error' })
    } finally {
      setIsResolving(false)
    }
  }

  const handlePlay = async () => {
    fileHandoffRef.current = null
    await resolvePlayableFile(file => {
      void resolveAndPlay(file)
    })
  }

  const playFile = (file: PlayableFile) => {
    fileHandoffRef.current = null
    void resolveAndPlay(file)
  }

  const playerModals = useMemo(
    () => (
      <>
        <Modal state={filePickerState}>
          <Modal.Backdrop isDismissable>
            <Modal.Container size='md'>
              <Modal.Dialog aria-label={t('Play')} className='sm:min-w-[28rem]'>
                <Modal.Header>
                  <Modal.Heading>{t('Play')}</Modal.Heading>
                </Modal.Header>
                <Modal.Body className='flex max-h-[60vh] flex-col gap-1.5 overflow-y-auto'>
                  {playableFiles.map((file, index) => {
                    const { code, title } = fileEpisodeInfo(file.path, index)
                    const pending = isResolving && resolvingFileId === file.id
                    return (
                      <Button
                        key={file.id}
                        variant='ghost'
                        className='h-auto w-full flex-col items-start justify-start gap-0.5 px-3 py-2.5'
                        isPending={pending}
                        onPress={() => pickFileFromList(file)}
                      >
                        {({ isPending }) => (
                          <>
                            <span className='flex w-full items-center gap-2 text-sm font-medium'>
                              {isPending ? <Spinner size='sm' color='current' /> : null}
                              <span className='shrink-0 rounded bg-accent/15 px-1.5 py-0.5 text-xs font-semibold text-accent'>
                                {code}
                              </span>
                              <span className='truncate'>{title}</span>
                            </span>
                            <span className='text-xs text-muted'>{humanizeSize(file.length)}</span>
                          </>
                        )}
                      </Button>
                    )
                  })}
                </Modal.Body>
              </Modal.Dialog>
            </Modal.Container>
          </Modal.Backdrop>
        </Modal>

        <Modal state={audioPickerState}>
          <Modal.Backdrop isDismissable>
            <Modal.Container size='sm'>
              <Modal.Dialog aria-label={t('SelectAudioTrack')}>
                <Modal.Header>
                  <Modal.Heading>{t('SelectAudioTrack')}</Modal.Heading>
                </Modal.Header>
                <Modal.Body className='flex max-h-[60vh] flex-col gap-1.5 overflow-y-auto'>
                  {audioTracks.map((track, index) => {
                    const { title, meta } = formatAudioTrackDisplay(track, index)
                    return (
                      <Button
                        key={index}
                        variant='ghost'
                        className='h-auto w-full justify-start gap-3 px-3 py-2.5'
                        onPress={() => pendingAudioFile && void openPlayer(pendingAudioFile, index, audioTracks)}
                      >
                        <span className='flex size-9 shrink-0 items-center justify-center rounded-lg bg-accent/15 text-accent'>
                          <Music2 {...iconMenu} aria-hidden />
                        </span>
                        <span className='min-w-0 flex-1 text-left'>
                          <span className='block truncate text-sm font-semibold text-foreground'>{title}</span>
                          {meta ? <span className='mt-0.5 block truncate text-xs text-muted'>{meta}</span> : null}
                        </span>
                      </Button>
                    )
                  })}
                </Modal.Body>
              </Modal.Dialog>
            </Modal.Container>
          </Modal.Backdrop>
        </Modal>

        {activePlayer ? (
          <Suspense
            fallback={
              <div className='fixed inset-0 z-[100] grid place-items-center bg-black/55'>
                <div className='flex flex-col items-center gap-3'>
                  <Spinner size='lg' color='current' className='text-accent' />
                  <p className='text-sm font-medium text-white/90'>{t('Buffering')}</p>
                </div>
              </div>
            }
          >
            <VideoPlayer
              initiallyOpen
              showTrigger={false}
              title={activePlayer.title}
              videoSrc={activePlayer.src}
              downloadSrc={activePlayer.downloadSrc}
              hls={activePlayer.hls}
              heartbeatSrc={activePlayer.heartbeatSrc}
              captionSrc={activePlayer.captionSrc}
              hash={activePlayer.hash}
              fileIndex={activePlayer.fileIndex}
              initialTimecode={activePlayer.initialTimecode}
              trackTimecode={activePlayer.trackTimecode}
              audioTracks={activePlayer.audioTracks}
              audioIndex={activePlayer.audioIndex}
              onAudioIndexChange={index => {
                setActivePlayer(current =>
                  current
                    ? {
                        ...current,
                        audioIndex: index,
                        src: gstreamerMasterUrl(current.hash, current.fileIndex, index),
                      }
                    : null,
                )
              }}
              onViewedChange={notifyViewedChange}
              onNotSupported={() => onPlayerNotSupported?.(activePlayer.fileIndex)}
              onClose={() => setActivePlayer(null)}
            />
          </Suspense>
        ) : null}
      </>
    ),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [
      filePickerState,
      audioPickerState,
      playableFiles,
      audioTracks,
      pendingAudioFile,
      activePlayer,
      isResolving,
      resolvingFileId,
      t,
    ],
  )

  return {
    handlePlay: () => void handlePlay(),
    resolvePlayableFile: (onFile: (file: PlayableFile) => void) => void resolvePlayableFile(onFile),
    playFile,
    isResolving,
    resolvingFileId,
    playerModals,
  }
}

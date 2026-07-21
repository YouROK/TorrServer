import { CheckCircle2 } from 'lucide-react'
import { memo, useCallback, useMemo, useState, type ReactNode } from 'react'
import ptt from 'parse-torrent-title'
import { useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { streamHost } from 'shared/api/hosts'
import { authFetch } from 'shared/api/authCredentials'
import type { PlayableFile, TorrentFileStat } from 'shared/api/types'
import { remViewedFile, VIEWED_QUERY_KEY } from 'shared/api/viewed'
import { shouldUseGStreamerPlayer, useGStreamerRuntime } from 'shared/lib/gstreamer'
import { useExternalPlayers } from 'shared/lib/externalPlayers'
import { humanizeSize } from 'shared/lib/format'
import { useOptionalAppToast } from 'shared/ui/Toast'
import { usePlayLauncher } from 'features/player/usePlayLauncher'

import FileRowActions from './FileRowActions'
import MediaInfoDialog from './MediaInfoDialog'

/**
 * Global `parse-torrent-title` handler extensions for Russian-language release
 * naming conventions. These mutate the shared `ptt` singleton, so they must be
 * registered exactly once at module scope (this file owns torrent file parsing).
 */
ptt.addHandler('episode', /(\d{1,4})[- |. ]серия|серия[- |. ](\d{1,4})/i, { type: 'integer' })
ptt.addHandler('season', /sezon[- |. ](\d{1,3})|(\d{1,3})[- |. ]sezon/i, { type: 'integer' })
ptt.addHandler('season', /сезон[- |. ](\d{1,3})|(\d{1,3})[- |. ]сезон/i, { type: 'integer' })

export interface FilesDataGridProps {
  playableFileList?: PlayableFile[]
  viewedFileList?: number[]
  selectedSeason?: number
  seasonAmount?: number[] | null
  hash: string
  displayName?: string
  torrentData?: string
  allFileStats?: TorrentFileStat[]
  onViewedChange?: () => void
}

interface FileRow {
  id: number
  name: string
  season?: number
  episode?: number
  resolution?: string
  size: number
  viewed: boolean
  path: string
  link: string
  playerKey: string
  fullLink: string
  playable: PlayableFile
}

function episodeBadge(episode?: number): string | null {
  if (episode == null) return null
  return `E${String(episode).padStart(2, '0')}`
}

/**
 * Compact episode/file card used in the details Files list.
 * Mobile: single horizontal row (badge + title | Play + …) — no stacked full-width Play.
 * Desktop (`sm+`): title row and action strip sit side-by-side with roomier padding.
 */
function EpisodeRow({
  row,
  actions,
  onUnmarkViewed,
}: {
  row: FileRow
  actions: ReactNode
  onUnmarkViewed?: () => void
}) {
  const { t } = useTranslation()
  const badge = episodeBadge(row.episode)
  const title = row.episode != null ? row.name.replace(/^E\d+\s*[·.-]\s*/i, '').trim() || row.name : row.name

  return (
    <div
      className={`flex items-center gap-2 rounded-lg border border-border bg-surface px-2.5 py-2 sm:gap-3 sm:rounded-xl sm:px-3 sm:py-2 ${
        row.viewed ? 'opacity-80' : 'border-l-[3px] border-l-accent'
      }`}
    >
      <div className='flex min-w-0 flex-1 items-center gap-2 sm:gap-3'>
        {badge ? (
          <span className='inline-flex h-8 min-w-9 shrink-0 items-center justify-center rounded-md bg-accent-soft px-1.5 text-xs font-bold tabular-nums text-accent sm:h-8 sm:min-w-9 sm:rounded-lg sm:px-2 sm:text-sm'>
            {badge}
          </span>
        ) : null}
        <div className='min-w-0 flex-1'>
          <div className='flex min-w-0 flex-wrap items-center gap-x-2 gap-y-0.5'>
            <p className='min-w-0 truncate text-sm font-semibold text-foreground' title={row.path}>
              {title}
            </p>
            {row.viewed ? (
              <button
                type='button'
                className='inline-flex shrink-0 items-center gap-1 rounded-md text-xs text-accent hover-fine:underline'
                onClick={onUnmarkViewed}
              >
                <CheckCircle2 size={14} strokeWidth={1.75} aria-hidden />
                {t('Viewed')}
                <span className='text-muted'>· {t('Clear')}</span>
              </button>
            ) : null}
          </div>
          <p className='mt-0.5 truncate text-[11px] leading-tight text-muted sm:text-xs'>
            {[row.season != null ? `${t('Season')} ${row.season}` : null, row.resolution, humanizeSize(row.size)]
              .filter(Boolean)
              .join(' · ')}
          </p>
        </div>
      </div>
      <div className='shrink-0'>{actions}</div>
    </div>
  )
}

/**
 * Playable-file list for a torrent: parse seasons/episodes, wire PlayLauncher,
 * and render one {@link EpisodeRow} per file (plus MediaInfo dialog).
 */
const FilesDataGrid = memo(
  ({
    playableFileList,
    viewedFileList,
    selectedSeason,
    seasonAmount,
    hash,
    displayName,
    torrentData,
    allFileStats = [],
    onViewedChange,
  }: FilesDataGridProps) => {
    const { t } = useTranslation()
    const toast = useOptionalAppToast()
    const queryClient = useQueryClient()
    const [unsupportedPlayerKeys, setUnsupportedPlayerKeys] = useState<Record<string, boolean>>({})
    const [mediaInfo, setMediaInfo] = useState<{ fileId: number; fileName: string } | null>(null)
    const gstRuntime = useGStreamerRuntime()
    const { buildExternalPlayers, shouldShowOpenLink } = useExternalPlayers()

    const knownPlayableFiles = useMemo(() => playableFileList || [], [playableFileList])
    const onPlayerNotSupported = useCallback(
      (fileId: number) => {
        const useGst = knownPlayableFiles.find(file => file.id === fileId)
        const key = `${fileId}:${useGst && shouldUseGStreamerPlayer(useGst.path, gstRuntime) ? 'gst' : 'stream'}`
        setUnsupportedPlayerKeys(current => ({ ...current, [key]: true }))
      },
      [knownPlayableFiles, gstRuntime],
    )

    const { playFile, isResolving, resolvingFileId, playerModals } = usePlayLauncher({
      hash,
      displayName: displayName || hash,
      knownPlayableFiles,
      knownAllFiles: allFileStats,
      torrentData,
      onViewedChange,
      onPlayerNotSupported,
    })

    const notifyViewedChange = () => {
      void queryClient.invalidateQueries({ queryKey: VIEWED_QUERY_KEY(hash) })
      onViewedChange?.()
    }

    const unmarkViewed = async (fileId: number) => {
      try {
        await remViewedFile(hash, fileId)
        notifyViewedChange()
      } catch {
        toast?.showToast({ message: t('Error'), severity: 'error' })
      }
    }

    const preloadBuffer = (fileId: number) => void authFetch(`${streamHost()}?link=${hash}&index=${fileId}&preload`)

    const buildFileLink = (path: string, id: number) => {
      const fileName = path.split('\\').pop()!.split('/').pop()!
      return `${streamHost()}/${encodeURIComponent(fileName)}?link=${hash}&index=${id}&play`
    }

    const fileHasEpisodeText = !!playableFileList?.find(({ path }) => ptt.parse(path).episode)
    const shouldDisplayFullFileName = (playableFileList?.length ?? 0) > 1 && !fileHasEpisodeText

    const filteredFiles = useMemo(() => {
      if (!playableFileList?.length) return []
      return playableFileList.filter(({ path }) => {
        const { season } = ptt.parse(path)
        return season == null || season === selectedSeason || !seasonAmount?.length
      })
    }, [playableFileList, selectedSeason, seasonAmount])

    const rows = useMemo<FileRow[]>(
      () =>
        filteredFiles.map(file => {
          const parsed = ptt.parse(file.path)
          const link = buildFileLink(file.path, file.id)
          const useGStreamer = shouldUseGStreamerPlayer(file.path, gstRuntime)
          const playerKey = `${file.id}:${useGStreamer ? 'gst' : 'stream'}`
          const fullLink = new URL(link, window.location.href).toString()
          const fileName = file.path.split('/').pop() || file.path
          const episodeLabel =
            parsed.episode != null
              ? `E${parsed.episode}${parsed.title ? ` · ${parsed.title}` : ''}`
              : shouldDisplayFullFileName
                ? file.path
                : parsed.title || fileName
          return {
            id: file.id,
            name: episodeLabel,
            season: parsed.season,
            episode: parsed.episode,
            resolution: parsed.resolution,
            size: file.length,
            viewed: viewedFileList?.includes(file.id) ?? false,
            path: file.path,
            link,
            playerKey,
            fullLink,
            playable: file,
          }
        }),
      // eslint-disable-next-line react-hooks/exhaustive-deps -- link builders close over hash/gstRuntime
      [filteredFiles, viewedFileList, shouldDisplayFullFileName, hash, gstRuntime],
    )

    if (!playableFileList?.length) {
      return <p className='py-6 text-center text-sm text-muted'>{t('NoPlayableFiles')}</p>
    }

    return (
      <div className='space-y-1.5'>
        {rows.map(row => (
          <EpisodeRow
            key={row.id}
            row={row}
            onUnmarkViewed={row.viewed ? () => void unmarkViewed(row.id) : undefined}
            actions={
              <FileRowActions
                preloadLabel={t('Preload')}
                onPreload={() => preloadBuffer(row.id)}
                playerSupported={!unsupportedPlayerKeys[row.playerKey]}
                onPlay={() => playFile(row.playable)}
                isPlayPending={isResolving && resolvingFileId === row.id}
                openLinkHref={row.link}
                showOpenLink={shouldShowOpenLink}
                copyText={row.fullLink}
                externalPlayers={buildExternalPlayers(row.fullLink)}
                onProbeMedia={() => setMediaInfo({ fileId: row.id, fileName: row.name })}
              />
            }
          />
        ))}

        {playerModals}

        {mediaInfo ? (
          <MediaInfoDialog
            open
            hash={hash}
            fileId={mediaInfo.fileId}
            fileName={mediaInfo.fileName}
            onClose={() => setMediaInfo(null)}
          />
        ) : null}
      </div>
    )
  },
  (prev, next) =>
    prev.hash === next.hash &&
    prev.selectedSeason === next.selectedSeason &&
    prev.playableFileList === next.playableFileList &&
    prev.viewedFileList === next.viewedFileList &&
    prev.seasonAmount === next.seasonAmount &&
    prev.allFileStats === next.allFileStats &&
    prev.displayName === next.displayName &&
    prev.torrentData === next.torrentData &&
    prev.onViewedChange === next.onViewedChange,
)

export default FilesDataGrid

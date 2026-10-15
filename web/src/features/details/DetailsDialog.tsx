import { Button, Checkbox, Modal, Spinner, Tabs, ToggleButton, ToggleButtonGroup, useMediaQuery } from '@heroui/react'
import { closeButtonVariants } from '@heroui/styles'
import { ImagePlus, Pencil, X } from 'lucide-react'
import { useEffect, useMemo, useState, useCallback } from 'react'
import ptt from 'parse-torrent-title'
import { useTranslation } from 'react-i18next'
import type { TorrentStat } from 'shared/api/types'
import { listViewedFiles } from 'shared/api/viewed'
import { useUpdateCache } from 'shared/cache/useUpdateCache'
import { useTorrentDetail } from 'shared/hooks/useTorrentDetail'
import { useLocalBoolPref } from 'shared/hooks/useLocalPref'
import {
  bufferAheadBytes,
  bufferFillPercent,
  formatBufferAheadLabel,
  formatBufferFilledLabel,
  formatCacheFilledLabel,
  getPeerString,
  humanizeSize,
  humanizeSpeed,
  removeRedundantCharacters,
  resolveBufferTargetBytes,
} from 'shared/lib/format'
import { useSettingsQuery } from 'shared/hooks/useSettingsQuery'
import { filesFromMetadata } from 'shared/torrent/fileMetadata'
import { isFilePlayable } from 'shared/torrent/playable'
import { CLOSED, GETTING_INFO, IN_DB, PRELOAD, WORKING } from 'shared/torrent/states'
import { MEDIA_SHORT_VIEWPORT, queryMax } from 'shared/theme/breakpoints'
import AppDialog from 'shared/ui/AppDialog'
import { DIALOG_DETAILS } from 'shared/ui/dialogSizes'
import { iconChrome, iconEmpty } from 'shared/ui/iconProps'
import { toPlayableFile } from 'shared/torrent/toPlayableFile'

import FileBrowser from './FileBrowser'
import CacheMapDialog from './CacheMapDialog'
import EditPosterDialog from './EditPosterDialog'
import SwarmStatsPanel, { TorrentProgressStrip } from './SwarmStatsPanel'
import SpeedCharts from './SpeedCharts'
import TorrentActions from './TorrentActions'
import TorrentCache from './TorrentCache'
import { usePlayLauncher } from 'features/player/usePlayLauncher'

export interface DetailsDialogProps {
  torrent: TorrentStat
  onClose: () => void
  onEdit?: (torrent: TorrentStat) => void
  /** Continue Watching: start this file after details open. */
  autoPlayFileId?: number
  autoPlayTimecode?: number
}

type DetailsTab = 'files' | 'stats' | 'cache'

function StatWidget({
  label,
  value,
  dense = false,
  compact = false,
  tight = false,
  hint,
}: {
  label: string
  value: string
  dense?: boolean
  /** Phone hero — allow 2-line labels so Russian strings aren't clipped to "Скорость загр..". */
  compact?: boolean
  /** Phone density pass — single-line labels, no reserved 2-line min-height. */
  tight?: boolean
  /** Longer semantics tooltip (Cache vs Buffer). */
  hint?: string
}) {
  const shown = value || '—'
  const box = dense ? (tight ? 'w-full px-1.5 py-0.5' : 'w-full px-1.5 py-1') : 'w-full px-2.5 py-2 sm:px-3'
  return (
    <div
      className={`min-w-0 rounded-md border border-border bg-surface-secondary text-center ${box}`}
      title={hint || undefined}
    >
      <span
        className={`block leading-tight text-muted ${
          tight
            ? 'truncate text-[11px]'
            : compact
              ? 'line-clamp-2 min-h-[1.75rem] text-[10px]'
              : dense
                ? 'truncate text-[11px]'
                : 'truncate text-[11px] sm:text-xs'
        }`}
        title={hint || label}
      >
        {label}
      </span>
      {/* Single-line value — long labels must not wrap and grow the hero. */}
      <span
        className={`block truncate font-bold tabular-nums text-foreground ${
          dense ? 'mt-0.5 h-3.5 text-[11px] leading-3.5' : 'mt-1 h-5 text-sm leading-5 sm:h-6 sm:text-base sm:leading-6'
        }`}
        title={shown}
      >
        {shown}
      </span>
    </div>
  )
}

/**
 * Derives a short display title from torrent `name` / `title` via parse-torrent-title,
 * dropping redundant year/resolution tokens already present in the primary string.
 */
function buildDisplayTitle(name: string | undefined, title: string | undefined): string {
  const parts: Array<string | number> = []
  const parsedName = name ? ptt.parse(name) : null

  if (title !== name) {
    parts.push(removeRedundantCharacters(title || ''))
  } else if (parsedName?.title) {
    parts.push(removeRedundantCharacters(parsedName.title))
  }

  if (parsedName?.year && !String(parts[0] || '').includes(String(parsedName.year))) {
    parts.push(parsedName.year)
  }
  if (parsedName?.resolution && !String(parts[0] || '').includes(String(parsedName.resolution))) {
    parts.push(parsedName.resolution)
  }

  const combined = parts.join('. ')
  const needsTrailingDot = combined.endsWith('.') && combined[combined.length - 2] === '.'
  return needsTrailingDot ? `${combined}.` : combined
}

function TitleRow({ title, subtitle, compact }: { title: string; subtitle: string | null; compact: boolean }) {
  const [expanded, setExpanded] = useState(false)
  /** Drop subtitle when it already appears inside the primary title (common after ptt parse). */
  const showSubtitle = !!subtitle && !title.toLocaleLowerCase().includes(subtitle.toLocaleLowerCase())

  const titleClass = `min-w-0 flex-1 font-bold text-foreground ${
    compact
      ? expanded
        ? 'text-[15px] leading-snug'
        : 'line-clamp-6 text-[15px] leading-snug'
      : 'line-clamp-2 text-lg leading-snug'
  }`

  return (
    <div className='min-w-0'>
      <div className='flex items-start gap-1'>
        {compact ? (
          <button
            type='button'
            className='min-w-0 flex-1 rounded-sm text-left outline-none ring-accent focus-visible:ring-2'
            aria-expanded={expanded}
            onClick={() => setExpanded(v => !v)}
          >
            <h2 className={titleClass} title={title}>
              {title}
            </h2>
          </button>
        ) : (
          <h2 className={titleClass} title={title}>
            {title}
          </h2>
        )}
      </div>
      {showSubtitle ? (
        <p className={`mt-0.5 truncate text-muted ${compact ? 'text-xs' : 'text-sm'}`} title={subtitle!}>
          {subtitle}
        </p>
      ) : null}
    </div>
  )
}

/** Full-detail sheet for a torrent: slim hero, files browser, live stats, and cache "snake" map. */
export default function DetailsDialog({
  torrent: initialTorrent,
  onClose,
  onEdit,
  autoPlayFileId,
  autoPlayTimecode,
}: DetailsDialogProps) {
  const { t } = useTranslation()
  /** Equal-width Files/Stats/Cache segments — only needed below the phone breakpoint. */
  const isMobile = useMediaQuery(queryMax('mobile'))
  const isShortViewport = useMediaQuery(MEDIA_SHORT_VIEWPORT)
  /** Compact chrome only on phone / short viewports — not force-fullscreen on a wide monitor. */
  const useCompactDetails = isMobile || isShortViewport

  const hash = initialTorrent.hash
  const { data: liveTorrent } = useTorrentDetail(hash, initialTorrent)
  const torrent = liveTorrent ?? initialTorrent

  const [activeTab, setActiveTab] = useState<DetailsTab>('files')
  const [viewedFileList, setViewedFileList] = useState<number[] | undefined>()
  const [seasonList, setSeasonList] = useState<number[] | null>(null)
  const [selectedSeason, setSelectedSeason] = useState<number | undefined>()
  const [isSnakeDebugMode, setIsSnakeDebugMode] = useLocalBoolPref('isSnakeDebugMode')
  const [cacheMapOpen, setCacheMapOpen] = useState(false)
  const [posterEditOpen, setPosterEditOpen] = useState(false)

  const {
    poster,
    title,
    name,
    stat,
    download_speed: downloadSpeed,
    upload_speed: uploadSpeed,
    torrent_size: torrentSize,
    file_stats: fileStats,
    data,
  } = torrent

  const playableFileList = useMemo(() => {
    const files = fileStats?.length ? fileStats.map(toPlayableFile) : filesFromMetadata(data)
    return files.filter(({ path }) => isFilePlayable(path))
  }, [fileStats, data])

  const resolvedTab = activeTab

  const cache = useUpdateCache(hash, {
    // Fast snake on Cache tab or while the large map dialog is open.
    fast: resolvedTab === 'cache' || cacheMapOpen,
  })
  const { data: btSettings } = useSettingsQuery()
  const preloadCachePercent = btSettings?.PreloadCache ?? 50
  const bufferTarget = resolveBufferTargetBytes(cache.Capacity, preloadCachePercent)
  // Streaming → playable ahead (null only when no active reader). Idle → preload progress.
  const bufferAhead = bufferAheadBytes(cache)
  const isStreamingBuffer = bufferAhead != null
  const bufferValue = isStreamingBuffer ? bufferAhead : (cache.Filled ?? 0)
  const bufferHint = isStreamingBuffer
    ? bufferAhead === 0
      ? t('BufferAheadEmptyHint')
      : t('BufferAheadHint')
    : t('BufferHint')
  const bufferTitle = isStreamingBuffer ? t('BufferAhead') : t('Buffer')
  const bufferLabel = isStreamingBuffer
    ? (formatBufferAheadLabel(bufferAhead) ?? '—')
    : (formatBufferFilledLabel(bufferValue, bufferTarget, { percent: 'always' }) ?? '—')
  const bufferPct = bufferFillPercent(bufferValue, bufferTarget)
  const cacheOccupiedPct = bufferFillPercent(cache.Filled, cache.Capacity)

  const seasonsFingerprint = useMemo(() => {
    const seasons: number[] = []
    playableFileList.forEach(({ path }) => {
      const season = ptt.parse(path).season
      if (season && !seasons.includes(season)) seasons.push(season)
    })
    seasons.sort((a, b) => a - b)
    return seasons.join(',')
  }, [playableFileList])

  useEffect(() => {
    const seasons = seasonsFingerprint
      ? seasonsFingerprint
          .split(',')
          .map(Number)
          .filter(n => Number.isFinite(n) && n > 0)
      : []
    // eslint-disable-next-line react-hooks/set-state-in-effect -- keep season list/selection valid when file list changes
    setSeasonList(seasons)
    setSelectedSeason(prev => {
      if (seasons.length === 0) return undefined
      if (prev != null && seasons.includes(prev)) return prev
      return seasons[0]
    })
  }, [seasonsFingerprint])

  useEffect(() => {
    let cancelled = false
    void listViewedFiles(hash)
      .then(list => {
        if (!cancelled) setViewedFileList(list)
      })
      .catch(() => {
        if (!cancelled) setViewedFileList(undefined)
      })
    return () => {
      cancelled = true
    }
  }, [hash])

  const refreshViewed = useCallback(async () => {
    try {
      setViewedFileList(await listViewedFiles(hash))
    } catch {
      setViewedFileList(undefined)
    }
  }, [hash])

  const statusLabel = (value?: number) => {
    const labels: Record<number, string> = {
      [GETTING_INFO]: t('TorrentGettingInfo'),
      [PRELOAD]: t('TorrentPreload'),
      [WORKING]: t('TorrentWorking'),
      [CLOSED]: t('TorrentClosed'),
      [IN_DB]: t('TorrentInDb'),
    }
    return value != null ? labels[value] || String(value) : '—'
  }

  const displayTitle = buildDisplayTitle(name, title) || title || name || hash
  const subtitle = name && title !== name ? ptt.parse(name).title || name : null
  // IN_DB torrents still carry a persisted file list in `data`. With no list yet (In DB or
  // getting info) keep one pending files view — don't claim the torrent has nothing to play.
  const filesUnknown = playableFileList.length === 0 && (stat === GETTING_INFO || stat === IN_DB)
  const isLoadingMetadata = stat === GETTING_INFO && filesUnknown
  const hasMultipleSeasons = (seasonList?.length ?? 0) > 1

  const cacheFilledValue = formatCacheFilledLabel(cache.Filled, cache.Capacity) ?? '—'

  // Continue Watching must not wait for the Stats tab — Files is the default surface.
  const { playerModals: autoPlayModals } = usePlayLauncher({
    hash,
    displayName: name || title || 'file',
    knownPlayableFiles: playableFileList,
    onViewedChange: refreshViewed,
    autoPlayFileId,
    autoPlayTimecode,
  })

  /** Six hero tiles, 3×2. Pieces live on Stats; category stays off the hero. */
  const heroStats = (
    <>
      <StatWidget
        dense={useCompactDetails}
        compact={useCompactDetails}
        tight={useCompactDetails}
        label={t('DownloadSpeed')}
        value={humanizeSpeed(downloadSpeed)}
      />
      <StatWidget
        dense={useCompactDetails}
        compact={useCompactDetails}
        tight={useCompactDetails}
        label={t('UploadSpeed')}
        value={humanizeSpeed(uploadSpeed)}
      />
      <StatWidget
        dense={useCompactDetails}
        compact={useCompactDetails}
        tight={useCompactDetails}
        label={t('Peers')}
        value={getPeerString(torrent) || '—'}
      />
      <StatWidget
        dense={useCompactDetails}
        compact={useCompactDetails}
        tight={useCompactDetails}
        label={t('Size')}
        value={humanizeSize(torrentSize)}
      />
      <StatWidget
        dense={useCompactDetails}
        compact={useCompactDetails}
        tight={useCompactDetails}
        label={t('CacheFilled')}
        value={cacheFilledValue}
        hint={t('CacheHint')}
      />
      <StatWidget
        dense={useCompactDetails}
        compact={useCompactDetails}
        tight={useCompactDetails}
        label={t('Status')}
        value={statusLabel(stat)}
      />
    </>
  )

  const torrentActions = (
    <TorrentActions
      hash={hash}
      torrsHash={torrent.torrs_hash}
      name={name}
      title={title}
      playableFileList={playableFileList}
      viewedFileList={viewedFileList}
      setViewedFileList={setViewedFileList}
      onViewedChange={refreshViewed}
      onDropped={onClose}
      onDeleted={onClose}
      compact={useCompactDetails}
    />
  )

  const editControl =
    onEdit != null ? (
      <button
        type='button'
        className={closeButtonVariants()}
        aria-label={t('EditTorrent')}
        onClick={() => onEdit(torrent)}
      >
        <Pencil {...iconChrome} aria-hidden />
      </button>
    ) : null

  const tabClass = isMobile ? 'min-h-11 w-auto flex-1 basis-0' : 'min-h-9 w-auto shrink-0'

  return (
    <>
      <AppDialog
        open
        onClose={onClose}
        size='lg'
        dialogClassName='ts-details-modal flex flex-col overflow-hidden'
        dialogStyle={DIALOG_DETAILS}
      >
        {/* Zero-height chrome: CloseTrigger is absolute; heading is screen-reader only.
            Fullscreen top safe-area lives on `.ts-details-hero` (index.css) — do not pad
            this header or Tailwind `p-0` would fight the global full-dialog header rule.
            On compact, edit sits in the top-right action cluster beside Close. */}
        <Modal.Header className='relative h-0 shrink-0 overflow-visible border-0 p-0'>
          <Modal.Heading className='sr-only'>{t('TorrentDetails')}</Modal.Heading>
          {editControl ? <div className='ts-details-edit-trigger absolute z-20'>{editControl}</div> : null}
          <Modal.CloseTrigger aria-label={t('Close')} className='shrink-0'>
            <X {...iconChrome} aria-hidden />
          </Modal.CloseTrigger>
        </Modal.Header>

        <Modal.Body className='flex min-h-0 flex-1 flex-col gap-2 overflow-hidden pt-0 sm:gap-3'>
          {useCompactDetails ? (
            /* Phone: poster beside the title, then one even 4×2 grid. */
            <div className='ts-details-hero ts-details-hero-flow shrink-0 space-y-2 bg-gradient-to-br from-accent-soft to-accent-soft/40'>
              <div className='flex items-start gap-2'>
                <button
                  type='button'
                  onClick={() => setPosterEditOpen(true)}
                  aria-label={t('AddDialog.AddPosterLinkInput')}
                  title={t('AddDialog.AddPosterLinkInput')}
                  className='grid aspect-[2/3] w-16 shrink-0 place-items-center overflow-hidden rounded-md bg-surface-secondary outline-none ring-accent transition-shadow focus-visible:ring-2'
                >
                  {poster ? (
                    <img
                      src={poster}
                      alt=''
                      className='h-full w-full object-cover'
                      onError={event => {
                        event.currentTarget.style.display = 'none'
                      }}
                    />
                  ) : (
                    <ImagePlus size={16} strokeWidth={1.75} className='text-muted' aria-hidden />
                  )}
                </button>
                <div className='min-w-0 flex-1 pr-[6.75rem]'>
                  <TitleRow title={displayTitle} subtitle={subtitle} compact />
                </div>
              </div>
              <div className='grid grid-cols-3 gap-1'>{heroStats}</div>
            </div>
          ) : (
            /*
                  Desktop: poster spans the title and one 3×2 metric grid.
                  Button clearance is only on the title row — the grid stays full width.
                */
            <div className='ts-details-hero grid shrink-0 grid-cols-[auto_minmax(0,1fr)] items-stretch gap-x-3 gap-y-2 rounded-xl bg-gradient-to-br from-accent-soft to-accent-soft/40 p-2.5'>
              <button
                type='button'
                onClick={() => setPosterEditOpen(true)}
                aria-label={t('AddDialog.AddPosterLinkInput')}
                title={t('AddDialog.AddPosterLinkInput')}
                className='group relative col-start-1 row-span-2 row-start-1 grid h-full min-h-0 w-36 shrink-0 self-stretch place-items-center overflow-hidden rounded-lg bg-surface-secondary outline-none ring-accent transition-shadow focus-visible:ring-2'
              >
                {poster ? (
                  <img
                    src={poster}
                    alt=''
                    className='h-full w-full object-cover'
                    onError={event => {
                      event.currentTarget.style.display = 'none'
                    }}
                  />
                ) : (
                  <ImagePlus {...iconEmpty} className='text-muted' aria-hidden />
                )}
                <span className='pointer-events-none absolute inset-0 grid place-items-center bg-black/0 px-1.5 text-center text-[10px] font-medium text-white opacity-0 transition-opacity group-hover:bg-black/45 group-hover:opacity-100 group-focus-visible:bg-black/45 group-focus-visible:opacity-100'>
                  {t('AddDialog.AddPosterLinkInput')}
                </span>
              </button>

              <div className='col-start-2 row-start-1 min-h-12 min-w-0 pr-[6.75rem]'>
                <TitleRow title={displayTitle} subtitle={subtitle} compact={false} />
              </div>

              <div className='col-start-2 row-start-2 grid min-w-0 w-full grid-cols-3 gap-1.5'>{heroStats}</div>
            </div>
          )}

          {/*
                HeroUI primary tabs default to `w-full` per tab; without `w-auto` + Indicator,
                only the first label was visible (others scrolled off-screen). Secondary variant
                gives a clear underline for the active section on both phone and desktop.
              */}
          <Tabs.Root
            variant='secondary'
            selectedKey={resolvedTab}
            onSelectionChange={key => setActiveTab(String(key) as DetailsTab)}
            className='ts-details-tabs flex min-h-0 flex-1 flex-col overflow-hidden'
          >
            <Tabs.ListContainer className='w-full max-w-full shrink-0'>
              <Tabs.List aria-label={t('TorrentDetails')} className={isMobile ? 'w-full min-w-full' : undefined}>
                <Tabs.Tab id='files' className={tabClass} aria-label={t('TorrentContent')}>
                  {t('TorrentFiles')}
                  <Tabs.Indicator />
                </Tabs.Tab>
                <Tabs.Tab id='stats' className={tabClass}>
                  {t('Stats')}
                  <Tabs.Indicator />
                </Tabs.Tab>
                <Tabs.Tab id='cache' className={tabClass}>
                  {t('Cache')}
                  <Tabs.Indicator />
                </Tabs.Tab>
              </Tabs.List>
            </Tabs.ListContainer>

            <Tabs.Panel id='files' className='flex min-h-0 flex-1 flex-col overflow-hidden pt-2 sm:pt-3'>
              <div className='min-h-0 flex-1 overflow-y-auto overscroll-contain'>
                {hasMultipleSeasons ? (
                  <div className='mb-2 min-h-11 sm:mb-3'>
                    <ToggleButtonGroup
                      selectionMode='single'
                      selectedKeys={selectedSeason != null ? [String(selectedSeason)] : []}
                      onSelectionChange={keys => {
                        const value = [...keys][0]
                        if (value != null) setSelectedSeason(Number(value))
                      }}
                      className='flex flex-wrap gap-2'
                    >
                      {seasonList!.map(season => (
                        <ToggleButton key={season} id={String(season)}>
                          {t('Season')} {season}
                        </ToggleButton>
                      ))}
                    </ToggleButtonGroup>
                  </div>
                ) : null}

                {filesUnknown ? (
                  <div
                    className='flex flex-col items-center justify-center gap-2 rounded-xl border border-border bg-surface px-4 py-8 text-center'
                    aria-busy={isLoadingMetadata || undefined}
                    aria-label={isLoadingMetadata ? t('TorrentGettingInfo') : t('TorrentInDb')}
                  >
                    {isLoadingMetadata ? <Spinner size='sm' /> : null}
                    <p className='text-sm text-muted'>
                      {isLoadingMetadata ? t('TorrentGettingInfo') : t('TorrentInDb')}
                    </p>
                  </div>
                ) : (
                  <FileBrowser
                    hash={hash}
                    playableFileList={playableFileList || []}
                    viewedFileList={viewedFileList}
                    selectedSeason={selectedSeason}
                    seasonAmount={seasonList}
                    allFileStats={fileStats}
                    onViewedChange={refreshViewed}
                  />
                )}
              </div>
              <div className='shrink-0 border-t border-border bg-surface pt-1 pb-[max(0.75rem,env(safe-area-inset-bottom,0px))]'>
                {torrentActions}
              </div>
            </Tabs.Panel>

            <Tabs.Panel id='stats' className='flex min-h-0 flex-1 flex-col overflow-hidden pt-3'>
              {useCompactDetails ? (
                <div className='min-h-0 flex-1 space-y-3 overflow-y-auto overscroll-contain pb-[max(0.75rem,env(safe-area-inset-bottom,0px))]'>
                  <TorrentProgressStrip torrent={torrent} cache={cache} preloadCachePercent={preloadCachePercent} />
                  <SpeedCharts downloadSpeed={downloadSpeed} uploadSpeed={uploadSpeed} compact />
                  <SwarmStatsPanel
                    torrent={torrent}
                    columns={2}
                    cacheReaders={cache.Readers?.length ?? 0}
                    cache={cache}
                  />
                </div>
              ) : (
                <div className='flex min-h-0 flex-1 flex-col gap-3 pb-[max(0.75rem,env(safe-area-inset-bottom,0px))]'>
                  <TorrentProgressStrip torrent={torrent} cache={cache} preloadCachePercent={preloadCachePercent} />
                  <div className='grid min-h-0 flex-1 grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)] items-stretch gap-3'>
                    <SpeedCharts downloadSpeed={downloadSpeed} uploadSpeed={uploadSpeed} compact fill />
                    <SwarmStatsPanel
                      torrent={torrent}
                      className='h-full min-h-0'
                      cacheReaders={cache.Readers?.length ?? 0}
                      cache={cache}
                    />
                  </div>
                </div>
              )}
            </Tabs.Panel>

            <Tabs.Panel
              id='cache'
              className='flex min-h-0 flex-1 flex-col gap-2 overflow-y-auto overscroll-contain pt-3 pb-[max(0.75rem,env(safe-area-inset-bottom,0px))] sm:gap-4 sm:pt-4'
            >
              <div className='flex shrink-0 flex-col gap-2'>
                <div className='flex items-center justify-between gap-2'>
                  {useCompactDetails ? null : (
                    <p className='text-sm font-semibold text-muted' title={t('CacheHint')}>
                      {t('Cache')}
                    </p>
                  )}
                  <div
                    className={`flex flex-wrap items-center gap-2 ${useCompactDetails ? 'w-full justify-between' : ''}`}
                  >
                    <Checkbox isSelected={isSnakeDebugMode} onChange={setIsSnakeDebugMode}>
                      <Checkbox.Content>
                        <Checkbox.Control>
                          <Checkbox.Indicator />
                        </Checkbox.Control>
                        {t('SnakeDebug')}
                      </Checkbox.Content>
                    </Checkbox>
                    <Button size='sm' variant='secondary' className='min-h-11' onPress={() => setCacheMapOpen(true)}>
                      {t('DetailedCacheView.button')}
                    </Button>
                  </div>
                </div>

                <div className='space-y-2.5 rounded-xl border border-border bg-surface-secondary p-2.5'>
                  <div title={bufferHint}>
                    <div className='mb-1 flex items-baseline justify-between gap-2 text-xs'>
                      <span className='truncate text-muted'>{bufferTitle}</span>
                      <span className='shrink-0 font-bold tabular-nums text-foreground'>{bufferLabel}</span>
                    </div>
                    <div className='h-2 overflow-hidden rounded-full bg-surface'>
                      <div
                        className='h-full rounded-full bg-accent transition-[width] duration-300'
                        style={{ width: `${bufferPct}%` }}
                      />
                    </div>
                  </div>
                  <div title={t('CacheHint')}>
                    <div className='mb-1 flex items-baseline justify-between gap-2 text-xs'>
                      <span className='truncate text-muted'>{t('CacheOccupied')}</span>
                      <span className='shrink-0 font-bold tabular-nums text-foreground'>{cacheFilledValue}</span>
                    </div>
                    <div className='h-2 overflow-hidden rounded-full bg-surface'>
                      <div
                        className='h-full rounded-full bg-accent transition-[width] duration-300'
                        style={{ width: `${cacheOccupiedPct}%` }}
                      />
                    </div>
                  </div>
                  <div className='flex flex-wrap gap-x-3 gap-y-1 text-[11px] text-muted'>
                    <span>
                      {t('CacheReaders')}:{' '}
                      <span className='font-semibold tabular-nums text-foreground'>{cache.Readers?.length ?? 0}</span>
                    </span>
                    <span>
                      {t('PiecesCount')}:{' '}
                      <span className='font-semibold tabular-nums text-foreground'>
                        {cache.PiecesCount != null ? String(cache.PiecesCount) : '—'}
                      </span>
                    </span>
                    <span>
                      {t('PiecesLength')}:{' '}
                      <span className='font-semibold tabular-nums text-foreground'>
                        {cache.PiecesLength != null ? humanizeSize(cache.PiecesLength) : '—'}
                      </span>
                    </span>
                  </div>
                </div>
              </div>

              {/* flex-1 height chain required: detailed snake sizes rows via ResizeObserver height.
                  min-h keeps at least ~4 rows when the 3×3 hero has already taken the sheet. */}
              <div className='flex min-h-[12rem] min-w-0 flex-1 flex-col'>
                <TorrentCache cache={cache} mode='detailed' isSnakeDebugMode={isSnakeDebugMode} hash={hash} />
              </div>
            </Tabs.Panel>
          </Tabs.Root>
        </Modal.Body>
      </AppDialog>

      <CacheMapDialog
        open={cacheMapOpen}
        onClose={() => setCacheMapOpen(false)}
        cache={cache}
        torrent={torrent}
        isSnakeDebugMode={isSnakeDebugMode}
        onSnakeDebugModeChange={setIsSnakeDebugMode}
        hash={hash}
      />
      <EditPosterDialog torrent={torrent} open={posterEditOpen} onClose={() => setPosterEditOpen(false)} />
      {autoPlayModals}
    </>
  )
}

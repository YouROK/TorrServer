import { Button, Checkbox, Modal, Tabs, ToggleButton, ToggleButtonGroup, useMediaQuery } from '@heroui/react'
import { ImagePlus, Pencil, X } from 'lucide-react'
import { useEffect, useMemo, useState, useCallback, type ReactNode } from 'react'
import ptt from 'parse-torrent-title'
import { useTranslation } from 'react-i18next'
import type { TorrentStat } from 'shared/api/types'
import { listViewedFiles } from 'shared/api/viewed'
import { useUpdateCache } from 'shared/cache/useUpdateCache'
import { useTorrentDetail } from 'shared/hooks/useTorrentDetail'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
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
import { MEDIA_SHORT_VIEWPORT, queryMax, queryMin } from 'shared/theme/breakpoints'
import AppDialog from 'shared/ui/AppDialog'
import { iconBtn } from 'shared/ui/controlClasses'
import { DIALOG_DETAILS } from 'shared/ui/dialogSizes'
import { iconChrome, iconEmpty } from 'shared/ui/iconProps'
import { toPlayableFile } from 'shared/torrent/toPlayableFile'

import FileBrowser from './FileBrowser'
import CacheMapDialog from './CacheMapDialog'
import EditPosterDialog from './EditPosterDialog'
import MetricRows from './MetricRows'
import SwarmStatsPanel from './SwarmStatsPanel'
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

type DetailsTab = 'files' | 'stats' | 'swarm' | 'cache'

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
  /** Longer semantics tooltip (Cache vs Buffer vs Loaded). */
  hint?: string
}) {
  const shown = value || '—'
  return (
    <div
      className={`min-w-0 rounded-md border border-border bg-surface-secondary text-center ${
        dense ? (tight ? 'w-full px-1.5 py-0.5' : 'w-full px-1.5 py-1') : 'px-2.5 py-2 sm:min-w-[104px] sm:px-3'
      }`}
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

function TitleRow({
  title,
  subtitle,
  compact,
  editControl,
}: {
  title: string
  subtitle: string | null
  compact: boolean
  editControl?: ReactNode
}) {
  const { t } = useTranslation()
  const [expanded, setExpanded] = useState(false)
  /** Drop subtitle when it already appears inside the primary title (common after ptt parse). */
  const showSubtitle = !!subtitle && !title.toLocaleLowerCase().includes(subtitle.toLocaleLowerCase())

  const titleClass = `min-w-0 flex-1 font-bold text-foreground ${
    compact
      ? expanded
        ? 'text-[15px] leading-snug'
        : 'line-clamp-3 text-[15px] leading-snug'
      : 'line-clamp-2 text-lg leading-snug'
  }`

  return (
    <div className='min-w-0 flex-1'>
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
        {editControl}
      </div>
      {showSubtitle ? (
        <p className={`mt-0.5 truncate text-muted ${compact ? 'text-xs' : 'text-sm'}`} title={subtitle!}>
          {subtitle}
        </p>
      ) : null}
      {compact ? (
        <button type='button' className='mt-0.5 text-xs font-medium text-accent' onClick={() => setExpanded(v => !v)}>
          {expanded ? t('ShowLess') : t('ShowMore')}
        </button>
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
  const isFullScreen = useDialogFullScreen()
  /** Equal-width Files/Stats/Cache segments — only needed below the phone breakpoint. */
  const isMobile = useMediaQuery(queryMax('mobile'))
  const isShortViewport = useMediaQuery(MEDIA_SHORT_VIEWPORT)
  /** Wider phones (≥420px): hero can fit 6 chips including Cache + Status. */
  const isPhoneWide = useMediaQuery(queryMin('phone'))
  /** Edge-to-edge sheet — AppDialog applies the same useDialogFullScreen policy. */
  const sheetFull = isFullScreen
  /** Compact chrome only on phone / short viewports — not force-fullscreen on a wide monitor. */
  const useCompactDetails = isMobile || isShortViewport
  const showHeroSix = useCompactDetails && (isMobile ? isPhoneWide : true)

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
    category,
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
  // IN_DB torrents still carry a persisted file list in `data` — only block the UI while the
  // live torrent is actively resolving metadata and we have nothing to show yet.
  const isLoadingMetadata = stat === GETTING_INFO && playableFileList.length === 0
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

  /** Live pulse in the hero — primary on all sizes; secondary only on desktop hero. */
  const primaryStats = (
    <>
      <StatWidget
        dense
        compact={useCompactDetails}
        tight={useCompactDetails && !showHeroSix}
        label={t('DownloadSpeed')}
        value={humanizeSpeed(downloadSpeed)}
      />
      <StatWidget
        dense
        compact={useCompactDetails}
        tight={useCompactDetails && !showHeroSix}
        label={t('UploadSpeed')}
        value={humanizeSpeed(uploadSpeed)}
      />
      <StatWidget
        dense
        compact={useCompactDetails}
        tight={useCompactDetails && !showHeroSix}
        label={t('Peers')}
        value={getPeerString(torrent) || '—'}
      />
      <StatWidget
        dense
        compact={useCompactDetails}
        tight={useCompactDetails && !showHeroSix}
        label={t('Size')}
        value={humanizeSize(torrentSize)}
      />
    </>
  )

  const cacheStatusStats = (
    <>
      <StatWidget
        dense
        compact={useCompactDetails}
        tight={useCompactDetails && !showHeroSix}
        label={t('CacheFilled')}
        value={cacheFilledValue}
        hint={t('CacheHint')}
      />
      <StatWidget
        dense
        compact={useCompactDetails}
        tight={useCompactDetails && !showHeroSix}
        label={t('Status')}
        value={statusLabel(stat)}
      />
    </>
  )

  const secondaryStats = (
    <>
      {cacheStatusStats}
      <StatWidget
        dense
        compact={useCompactDetails}
        tight={useCompactDetails}
        label={t('Category')}
        value={category || '—'}
      />
      <StatWidget
        dense
        compact={useCompactDetails}
        tight={useCompactDetails}
        label={t('PiecesCount')}
        value={cache.PiecesCount != null ? String(cache.PiecesCount) : '—'}
      />
      <StatWidget
        dense
        compact={useCompactDetails}
        tight={useCompactDetails}
        label={t('PiecesLength')}
        value={cache.PiecesLength != null ? humanizeSize(cache.PiecesLength) : '—'}
      />
    </>
  )

  /** Stats-tab dense rows (mobile) — omit Cache/Status when already in the 6-chip hero. */
  const secondaryMetricItems = [
    ...(showHeroSix
      ? []
      : [
          { label: t('CacheFilled'), value: cacheFilledValue, hint: t('CacheHint') },
          { label: t('Status'), value: statusLabel(stat) },
        ]),
    { label: t('Category'), value: category || '—' },
    { label: t('PiecesCount'), value: cache.PiecesCount != null ? String(cache.PiecesCount) : '—' },
    {
      label: t('PiecesLength'),
      value: cache.PiecesLength != null ? humanizeSize(cache.PiecesLength) : '—',
    },
  ]

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

  const heroStats = (
    <>
      {primaryStats}
      {secondaryStats}
    </>
  )

  const editControl =
    onEdit != null ? (
      <Button
        isIconOnly
        variant='ghost'
        className={`${iconBtn} shrink-0`}
        aria-label={t('EditTorrent')}
        onPress={() => onEdit(torrent)}
      >
        <Pencil {...iconChrome} aria-hidden />
      </Button>
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
          {useCompactDetails && editControl ? (
            <div className='ts-details-edit-trigger absolute z-20'>{editControl}</div>
          ) : null}
          <Modal.CloseTrigger aria-label={t('Close')} className='shrink-0'>
            <X {...iconChrome} aria-hidden />
          </Modal.CloseTrigger>
        </Modal.Header>

        <Modal.Body
          className={`flex min-h-0 flex-1 flex-col gap-2 overflow-hidden sm:gap-3${sheetFull ? ' pt-0' : ' pt-1'}`}
        >
          {useCompactDetails ? (
            /* Phone: flush hero on narrow CSS; rounded card only on wide desktop sheet.
                   Fullscreen overrides radius via `.modal__dialog--full .ts-details-hero`. */
            <div className='ts-details-hero shrink-0 space-y-2 rounded-xl bg-gradient-to-br from-accent-soft to-accent-soft/40 p-2 pr-[6.75rem]'>
              <div className='flex items-start gap-2'>
                <button
                  type='button'
                  onClick={() => setPosterEditOpen(true)}
                  aria-label={t('AddDialog.AddPosterLinkInput')}
                  title={t('AddDialog.AddPosterLinkInput')}
                  className='group relative grid h-12 min-h-11 w-11 min-w-11 shrink-0 place-items-center overflow-hidden rounded-md bg-surface-secondary outline-none ring-accent transition-shadow focus-visible:ring-2'
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
                <TitleRow title={displayTitle} subtitle={subtitle} compact />
              </div>
              <div className={`grid w-full gap-1 ${showHeroSix ? 'grid-cols-3' : 'grid-cols-2'}`}>
                {primaryStats}
                {showHeroSix ? cacheStatusStats : null}
              </div>
            </div>
          ) : (
            /*
                  Desktop: poster spans both rows; title on top-right; metrics sit under the title
                  (raised beside the poster) and stretch across the remaining width.
                  Fullscreen tablet: same flush hero + safe-area as phone (`.ts-details-hero`).
                */
            <div className='ts-details-hero grid shrink-0 grid-cols-[auto_minmax(0,1fr)] items-start gap-x-3 gap-y-1.5 rounded-xl bg-gradient-to-br from-accent-soft to-accent-soft/40 p-2.5 pr-12'>
              <button
                type='button'
                onClick={() => setPosterEditOpen(true)}
                aria-label={t('AddDialog.AddPosterLinkInput')}
                title={t('AddDialog.AddPosterLinkInput')}
                className='group relative col-start-1 row-span-2 row-start-1 grid aspect-[2/3] w-[72px] shrink-0 self-start place-items-center overflow-hidden rounded-lg bg-surface-secondary outline-none ring-accent transition-shadow focus-visible:ring-2'
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

              <div className='col-start-2 row-start-1 min-w-0'>
                <TitleRow title={displayTitle} subtitle={subtitle} compact={false} editControl={editControl} />
              </div>

              <div className='col-start-2 row-start-2 grid min-w-0 w-full grid-cols-9 gap-1'>{heroStats}</div>
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
                <Tabs.Tab id='swarm' className={tabClass}>
                  {t('SwarmStats')}
                  <Tabs.Indicator />
                </Tabs.Tab>
                <Tabs.Tab id='cache' className={tabClass}>
                  {t('Cache')}
                  <Tabs.Indicator />
                </Tabs.Tab>
              </Tabs.List>
            </Tabs.ListContainer>

            <Tabs.Panel id='files' className='min-h-0 flex-1 overflow-y-auto overscroll-contain pt-2 sm:pt-3'>
              <div className='pb-[max(0.75rem,env(safe-area-inset-bottom,0px))]'>
                {/* Reserve chip-row height while metadata loads or when multi-season — list won't jump. */}
                {hasMultipleSeasons || isLoadingMetadata ? (
                  <div className='mb-2 min-h-11 sm:mb-3'>
                    {hasMultipleSeasons ? (
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
                    ) : null}
                  </div>
                ) : null}

                {isLoadingMetadata ? (
                  <div className='space-y-2' aria-busy aria-label={t('TorrentGettingInfo')}>
                    {Array.from({ length: 4 }, (_, i) => (
                      <div
                        key={i}
                        className='h-[4.5rem] animate-pulse rounded-xl border border-border bg-surface sm:h-[3.25rem]'
                      />
                    ))}
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
            </Tabs.Panel>

            <Tabs.Panel
              id='stats'
              className={`flex min-h-0 flex-1 flex-col pt-3 ${
                useCompactDetails ? 'overflow-hidden' : 'gap-3 overflow-y-auto overscroll-contain'
              }`}
            >
              {useCompactDetails ? (
                <>
                  {/* Content-start: do not flex-stretch a void between cards and CTAs. */}
                  <div className='min-h-0 flex-1 space-y-3 overflow-y-auto overscroll-contain pb-2'>
                    {secondaryMetricItems.length > 0 ? (
                      <MetricRows title={t('Details')} items={secondaryMetricItems} columns={1} />
                    ) : null}
                    <SpeedCharts downloadSpeed={downloadSpeed} uploadSpeed={uploadSpeed} compact />
                    <SwarmStatsPanel
                      torrent={torrent}
                      variant='summary'
                      columns={2}
                      showTitle={false}
                      cache={cache}
                      preloadCachePercent={preloadCachePercent}
                    />
                  </div>
                  <div className='shrink-0 border-t border-border bg-surface/95 pt-2 backdrop-blur-sm'>
                    {torrentActions}
                  </div>
                </>
              ) : (
                <>
                  <div className='grid min-h-[14rem] shrink-0 grid-cols-[minmax(0,1.4fr)_minmax(0,1fr)] items-stretch gap-3'>
                    <SpeedCharts downloadSpeed={downloadSpeed} uploadSpeed={uploadSpeed} compact fill />
                    <SwarmStatsPanel
                      torrent={torrent}
                      variant='summary'
                      stretch
                      cache={cache}
                      preloadCachePercent={preloadCachePercent}
                    />
                  </div>
                  <div className='shrink-0'>{torrentActions}</div>
                </>
              )}
            </Tabs.Panel>

            <Tabs.Panel id='swarm' className='flex min-h-0 flex-1 flex-col overflow-y-auto overscroll-contain pt-3'>
              <SwarmStatsPanel
                torrent={torrent}
                variant='full'
                className={
                  useCompactDetails
                    ? 'w-full shrink-0 pb-[max(0.75rem,env(safe-area-inset-bottom,0px))]'
                    : 'min-h-0 w-full flex-1 pb-[max(0.75rem,env(safe-area-inset-bottom,0px))]'
                }
                cacheReaders={cache.Readers?.length ?? 0}
                cache={cache}
                preloadCachePercent={preloadCachePercent}
              />
            </Tabs.Panel>

            <Tabs.Panel id='cache' className='flex min-h-0 flex-1 flex-col gap-2 overflow-hidden pt-3 sm:gap-4 sm:pt-4'>
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

              {/* flex-1 height chain required: detailed snake sizes rows via ResizeObserver height. */}
              <div className='flex min-h-0 min-w-0 flex-1 flex-col'>
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
        isSnakeDebugMode={isSnakeDebugMode}
        onSnakeDebugModeChange={setIsSnakeDebugMode}
        hash={hash}
      />
      <EditPosterDialog torrent={torrent} open={posterEditOpen} onClose={() => setPosterEditOpen(false)} />
      {autoPlayModals}
    </>
  )
}

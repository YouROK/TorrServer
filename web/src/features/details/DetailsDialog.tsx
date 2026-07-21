import {
  Button,
  Checkbox,
  Modal,
  Tabs,
  ToggleButton,
  ToggleButtonGroup,
  useMediaQuery,
  useOverlayState,
} from '@heroui/react'
import { ImagePlus, Pencil, ChevronRight, X } from 'lucide-react'
import { useEffect, useMemo, useState, useCallback } from 'react'
import ptt from 'parse-torrent-title'
import { useTranslation } from 'react-i18next'
import type { TorrentStat } from 'shared/api/types'
import { listViewedFiles } from 'shared/api/viewed'
import { useUpdateCache } from 'shared/cache/useUpdateCache'
import { useTorrentDetail } from 'shared/hooks/useTorrentDetail'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import { useLocalBoolPref } from 'shared/hooks/useLocalPref'
import {
  getPeerString,
  formatCacheFilledLabel,
  humanizeSize,
  humanizeSpeed,
  removeRedundantCharacters,
} from 'shared/lib/format'
import { filesFromMetadata } from 'shared/torrent/fileMetadata'
import { isFilePlayable } from 'shared/torrent/playable'
import { CLOSED, GETTING_INFO, IN_DB, PRELOAD, WORKING } from 'shared/torrent/states'
import { MEDIA_SHORT_VIEWPORT, queryMax } from 'shared/theme/breakpoints'
import { useSyncModalOpen } from 'shared/ui/ModalOpenContext'
import { iconBtn } from 'shared/ui/controlClasses'
import { DIALOG_DETAILS, DIALOG_FULLSCREEN } from 'shared/ui/dialogSizes'
import { iconChrome, iconEmpty } from 'shared/ui/iconProps'
import { toPlayableFile } from 'shared/torrent/toPlayableFile'

import FileBrowser from './FileBrowser'
import CacheMapDialog from './CacheMapDialog'
import EditPosterDialog from './EditPosterDialog'
import CacheHeatSparkline from './CacheHeatSparkline'
import SwarmStatsPanel from './SwarmStatsPanel'
import SpeedCharts from './SpeedCharts'
import TorrentActions from './TorrentActions'
import TorrentCache from './TorrentCache'

export interface DetailsDialogProps {
  torrent: TorrentStat
  onClose: () => void
  onEdit?: (torrent: TorrentStat) => void
  /** Continue Watching: start this file after details open. */
  autoPlayFileId?: number
  autoPlayTimecode?: number
}

type DetailsTab = 'overview' | 'files' | 'cache'

function StatWidget({ label, value, dense = false }: { label: string; value: string; dense?: boolean }) {
  const shown = value || '—'
  return (
    <div
      className={`min-w-0 rounded-lg border border-border bg-surface-secondary text-center ${
        dense ? 'px-2 py-1.5' : 'px-2.5 py-2 sm:min-w-[104px] sm:px-3'
      }`}
    >
      <span
        className={`block truncate leading-tight text-muted ${dense ? 'text-[10px]' : 'text-[11px] sm:text-xs'}`}
        title={label}
      >
        {label}
      </span>
      {/* Single-line value — long CacheFilled must not wrap and grow the hero. */}
      <span
        className={`mt-0.5 block truncate font-bold tabular-nums text-foreground ${
          dense ? 'h-4 text-xs leading-4' : 'mt-1 h-5 text-sm leading-5 sm:h-6 sm:text-base sm:leading-6'
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

/** Full-detail sheet for a torrent: hero header, live speed chart, files browser and cache "snake" map. */
export default function DetailsDialog({
  torrent: initialTorrent,
  onClose,
  onEdit,
  autoPlayFileId,
  autoPlayTimecode,
}: DetailsDialogProps) {
  const { t } = useTranslation()
  const isFullScreen = useDialogFullScreen()
  /** Equal-width Overview/Files/Cache segments — only needed below the phone breakpoint. */
  const isMobile = useMediaQuery(queryMax('mobile'))
  const isShortViewport = useMediaQuery(MEDIA_SHORT_VIEWPORT)
  /** Phone-compact hero/actions — not tied to fullscreen surface (iPad landscape stays wide). */
  const useCompactDetails = isMobile || isShortViewport
  useSyncModalOpen(true)

  const overlayState = useOverlayState({
    isOpen: true,
    onOpenChange: open => {
      if (!open) onClose()
    },
  })

  const hash = initialTorrent.hash
  const { data: liveTorrent } = useTorrentDetail(hash, initialTorrent)
  const torrent = liveTorrent ?? initialTorrent

  const [activeTab, setActiveTab] = useState<DetailsTab>('overview')
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

  // Stay on Overview until the user (or Play → onShowFiles) switches tabs.
  // Do not auto-jump to Files when metadata arrives — that felt like the sheet "teleporting".
  const resolvedTab = activeTab

  const cache = useUpdateCache(hash, {
    // Fast snake on Cache tab or while the large map dialog is open.
    fast: resolvedTab === 'cache' || cacheMapOpen,
  })

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

  const primaryStats = (
    <>
      <StatWidget dense={!useCompactDetails} label={t('DownloadSpeed')} value={humanizeSpeed(downloadSpeed)} />
      <StatWidget dense={!useCompactDetails} label={t('UploadSpeed')} value={humanizeSpeed(uploadSpeed)} />
      <StatWidget dense={!useCompactDetails} label={t('Peers')} value={getPeerString(torrent) || '—'} />
      <StatWidget dense={!useCompactDetails} label={t('Size')} value={humanizeSize(torrentSize)} />
    </>
  )

  const secondaryStats = (
    <>
      <StatWidget dense label={t('Status')} value={statusLabel(stat)} />
      <StatWidget dense label={t('Category')} value={category || '—'} />
      <StatWidget dense label={t('PiecesCount')} value={cache.PiecesCount != null ? String(cache.PiecesCount) : '—'} />
      <StatWidget
        dense
        label={t('PiecesLength')}
        value={cache.PiecesLength != null ? humanizeSize(cache.PiecesLength) : '—'}
      />
      <StatWidget dense label={t('CacheFilled')} value={cacheFilledValue} />
    </>
  )

  return (
    <Modal.Root state={overlayState}>
      <Modal.Backdrop>
        <Modal.Container size={isFullScreen ? 'full' : 'lg'} scroll='inside'>
          {/* Inline style: HeroUI's size ceiling + our collapse-prevention floor (index.css) live in CSS
              layers, so a plain width utility can lose to them regardless of specificity — see AppDialog. */}
          <Modal.Dialog
            className='flex flex-col overflow-hidden'
            style={isFullScreen ? DIALOG_FULLSCREEN : DIALOG_DETAILS}
          >
            <Modal.Header className='relative flex shrink-0 flex-nowrap items-center gap-1 pr-12 sm:gap-2'>
              <Modal.Heading className='min-w-0 flex-1 truncate'>{t('TorrentDetails')}</Modal.Heading>
              {onEdit ? (
                <Button
                  isIconOnly
                  variant='ghost'
                  className={`${iconBtn} shrink-0`}
                  aria-label={t('EditTorrent')}
                  onPress={() => onEdit(torrent)}
                >
                  <Pencil {...iconChrome} aria-hidden />
                </Button>
              ) : null}
              <Modal.CloseTrigger aria-label={t('Close')} className='shrink-0'>
                <X {...iconChrome} aria-hidden />
              </Modal.CloseTrigger>
            </Modal.Header>

            <Modal.Body className='flex min-h-0 flex-1 flex-col gap-3 overflow-hidden'>
              {/* Compact hero on phone/short viewport; wide hero on tablet/desktop (incl. fullscreen iPad). */}
              <div
                className={`shrink-0 rounded-xl bg-gradient-to-br from-accent-soft to-accent-soft/40 ${
                  useCompactDetails ? 'space-y-3 p-3' : 'flex flex-col gap-3 p-3 sm:flex-row sm:items-start'
                }`}
              >
                {useCompactDetails ? (
                  <>
                    <div className='flex items-start gap-3'>
                      <button
                        type='button'
                        onClick={() => setPosterEditOpen(true)}
                        aria-label={t('AddDialog.AddPosterLinkInput')}
                        title={t('AddDialog.AddPosterLinkInput')}
                        className='group relative grid h-16 w-[42px] shrink-0 place-items-center overflow-hidden rounded-md bg-surface-secondary outline-none ring-accent transition-shadow focus-visible:ring-2'
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
                          <ImagePlus size={18} strokeWidth={1.75} className='text-muted' aria-hidden />
                        )}
                        <span className='pointer-events-none absolute inset-0 grid place-items-center bg-black/0 text-[9px] font-medium text-white opacity-0 transition-opacity group-focus-visible:bg-black/45 group-focus-visible:opacity-100'>
                          {t('AddDialog.AddPosterLinkInput')}
                        </span>
                      </button>
                      <div className='min-w-0 flex-1'>
                        <h2 className='line-clamp-2 text-base font-bold leading-snug text-foreground'>
                          {displayTitle}
                        </h2>
                        {subtitle ? (
                          <p className='mt-0.5 line-clamp-1 text-xs text-muted' title={subtitle}>
                            {subtitle}
                          </p>
                        ) : null}
                      </div>
                    </div>
                    <div className='grid grid-cols-2 gap-1.5'>{primaryStats}</div>
                  </>
                ) : (
                  <>
                    {/* Always reserve poster column so late poster URL doesn't reflow the stats grid. */}
                    <button
                      type='button'
                      onClick={() => setPosterEditOpen(true)}
                      aria-label={t('AddDialog.AddPosterLinkInput')}
                      title={t('AddDialog.AddPosterLinkInput')}
                      className='group relative mx-auto grid aspect-[2/3] w-full max-w-[96px] shrink-0 place-items-center overflow-hidden rounded-lg bg-surface-secondary outline-none ring-accent transition-shadow focus-visible:ring-2 sm:mx-0'
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
                      <span className='pointer-events-none absolute inset-0 grid place-items-center bg-black/0 px-2 text-center text-xs font-medium text-white opacity-0 transition-opacity group-hover:bg-black/45 group-hover:opacity-100 group-focus-visible:bg-black/45 group-focus-visible:opacity-100'>
                        {t('AddDialog.AddPosterLinkInput')}
                      </span>
                    </button>

                    <div className='min-w-0 flex-1'>
                      <h2 className='mb-0.5 break-words text-lg font-bold text-foreground'>{displayTitle}</h2>
                      {/* Reserve subtitle line so title-only torrents don't collapse hero height. */}
                      <p
                        className={`mb-1.5 truncate text-sm ${subtitle ? 'text-muted' : 'invisible'}`}
                        aria-hidden={!subtitle}
                      >
                        {subtitle || '\u00a0'}
                      </p>

                      {/* 9 slots → 2 rows on xl (5+4). min-h locks row stack so value updates never grow hero. */}
                      <div className='grid grid-cols-2 gap-1.5 sm:grid-cols-3 sm:min-h-[9.5rem] lg:grid-cols-4 lg:min-h-[6.5rem] xl:grid-cols-5 xl:min-h-[6.5rem]'>
                        {primaryStats}
                        {secondaryStats}
                      </div>
                    </div>
                  </>
                )}
              </div>

              {/*
                HeroUI primary tabs default to `w-full` per tab; without `w-auto` + Indicator,
                only the first label was visible (others scrolled off-screen). Secondary variant
                gives a clear underline for the active section on both phone and desktop.
              */}
              <Tabs.Root
                variant='secondary'
                selectedKey={resolvedTab}
                onSelectionChange={key => setActiveTab(String(key) as DetailsTab)}
                className='flex min-h-0 flex-1 flex-col overflow-hidden'
              >
                <Tabs.ListContainer className='w-full max-w-full shrink-0'>
                  <Tabs.List aria-label={t('TorrentDetails')} className={isMobile ? 'w-full min-w-full' : undefined}>
                    <Tabs.Tab
                      id='overview'
                      className={isMobile ? 'min-h-11 w-auto flex-1 basis-0' : 'min-h-9 w-auto shrink-0'}
                    >
                      {t('Overview')}
                      <Tabs.Indicator />
                    </Tabs.Tab>
                    <Tabs.Tab
                      id='files'
                      className={isMobile ? 'min-h-11 w-auto flex-1 basis-0' : 'min-h-9 w-auto shrink-0'}
                      aria-label={t('TorrentContent')}
                    >
                      {t('TorrentFiles')}
                      <Tabs.Indicator />
                    </Tabs.Tab>
                    <Tabs.Tab
                      id='cache'
                      className={isMobile ? 'min-h-11 w-auto flex-1 basis-0' : 'min-h-9 w-auto shrink-0'}
                    >
                      {t('Cache')}
                      <Tabs.Indicator />
                    </Tabs.Tab>
                  </Tabs.List>
                </Tabs.ListContainer>

                <Tabs.Panel id='overview' className='min-h-0 flex-1 space-y-3 overflow-y-auto overscroll-contain pt-3'>
                  {useCompactDetails ? (
                    <div className='grid grid-cols-2 gap-1.5 sm:grid-cols-3'>{secondaryStats}</div>
                  ) : null}

                  <SpeedCharts downloadSpeed={downloadSpeed} uploadSpeed={uploadSpeed} compact />

                  <SwarmStatsPanel torrent={torrent} />

                  <CacheHeatSparkline filled={cache.Filled} capacity={cache.Capacity} />

                  {useCompactDetails ? (
                    <div className='flex gap-2'>
                      <Button
                        size='sm'
                        variant='secondary'
                        className='min-h-11 flex-1 gap-2'
                        onPress={() => setActiveTab('cache')}
                      >
                        {t('Cache')}
                        <ChevronRight size={16} strokeWidth={1.75} aria-hidden />
                      </Button>
                      <Button
                        size='sm'
                        variant='ghost'
                        className='min-h-11 shrink-0'
                        onPress={() => setCacheMapOpen(true)}
                        aria-label={t('DetailedCacheView.button')}
                      >
                        {t('Cache')}
                      </Button>
                    </div>
                  ) : (
                    <div className='rounded-xl border border-border bg-surface-secondary p-2.5'>
                      <div className='mb-2 flex items-center justify-end gap-2'>
                        <Button size='sm' variant='ghost' onPress={() => setCacheMapOpen(true)}>
                          {t('DetailedCacheView.button')}
                        </Button>
                      </div>
                      <TorrentCache cache={cache} mode='mini' />
                    </div>
                  )}

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
                    onShowFiles={() => setActiveTab('files')}
                    autoPlayFileId={autoPlayFileId}
                    autoPlayTimecode={autoPlayTimecode}
                    compact={useCompactDetails}
                  />
                </Tabs.Panel>

                <Tabs.Panel id='files' className='min-h-0 flex-1 overflow-y-auto overscroll-contain pt-3 sm:pt-4'>
                  <div className='sm:rounded-xl sm:bg-surface-secondary sm:p-4'>
                    {/* Reserve chip-row height while metadata loads or when multi-season — list won't jump. */}
                    {hasMultipleSeasons || isLoadingMetadata ? (
                      <div className='mb-3 min-h-11 sm:mb-4'>
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
                  id='cache'
                  className='flex min-h-0 flex-1 flex-col gap-2 overflow-hidden pt-3 sm:gap-4 sm:pt-4'
                >
                  <div className='flex shrink-0 items-center justify-between gap-2'>
                    {useCompactDetails ? null : <p className='text-sm font-semibold text-muted'>{t('Cache')}</p>}
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

                  <div className='flex min-h-0 min-w-0 flex-1 flex-col'>
                    <TorrentCache cache={cache} mode='detailed' isSnakeDebugMode={isSnakeDebugMode} />
                  </div>
                </Tabs.Panel>
              </Tabs.Root>
            </Modal.Body>
          </Modal.Dialog>
        </Modal.Container>
      </Modal.Backdrop>

      <CacheMapDialog
        open={cacheMapOpen}
        onClose={() => setCacheMapOpen(false)}
        cache={cache}
        isSnakeDebugMode={isSnakeDebugMode}
        onSnakeDebugModeChange={setIsSnakeDebugMode}
      />
      <EditPosterDialog torrent={torrent} open={posterEditOpen} onClose={() => setPosterEditOpen(false)} />
    </Modal.Root>
  )
}

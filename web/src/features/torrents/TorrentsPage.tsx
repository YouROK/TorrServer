import { lazy, Suspense, useEffect, useMemo, useRef, useState } from 'react'
import { useGSAP } from '@gsap/react'
import gsap from 'gsap'
import { Button, Dropdown, Input, Label, Spinner, TextField, useMediaQuery } from '@heroui/react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Check,
  CheckSquare,
  ChevronDown,
  CloudOff,
  FolderPlus,
  History,
  LayoutGrid,
  Search,
  SearchX,
  Trash2,
  X,
  XSquare,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { LibraryHeaderPortal } from 'app/LibraryHeaderSlot'
import type { TorrentStat, ViewedFileEntry } from 'shared/api/types'
import { filteredPlaylistAllUrl } from 'shared/api/extras'
import { dropTorrent, markTorrentsDroppedInList, removeTorrent, TORRENTS_QUERY_KEY } from 'shared/api/torrents'
import { useLocalJsonPref } from 'shared/hooks/useLocalPref'
import { useTorrentsQuery } from 'shared/hooks/useTorrentsQuery'
import { ALL_VIEWED_QUERY_KEY, clearViewedFiles, listAllViewedEntries, VIEWED_QUERY_KEY } from 'shared/api/viewed'
import { buildContinueWatchShelf } from 'shared/lib/serverContinueWatching'
import { removeContinueWatching, type ContinueWatchEntry } from 'shared/lib/continueWatching'
import {
  DOWNLOAD_ALL_PLAYLISTS_EVENT,
  OPEN_EXPORT_LIBRARY_EVENT,
  OPEN_IMPORT_LIBRARY_EVENT,
} from 'shared/lib/libraryEvents'
import { openPlaylistLink } from 'shared/lib/openBrowser'
import { queryMax } from 'shared/theme/breakpoints'
import { iconBtn } from 'shared/ui/controlClasses'
import { iconEmpty, iconEmptyLg, iconMenu, iconNav } from 'shared/ui/iconProps'
import DialogErrorBoundary from 'shared/ui/DialogErrorBoundary'
import { useOptionalAppToast } from 'shared/ui/Toast'
import { logoutBasicAuth } from 'shared/api/authCredentials'

import TorrentCard from './TorrentCard'

const DetailsDialog = lazy(() => import('features/details/DetailsDialog'))
const EditTorrentDialog = lazy(() => import('features/add/EditTorrentDialog'))
const ExportLibraryDialog = lazy(() => import('./ExportLibraryDialog'))
const ImportLibraryDialog = lazy(() => import('./ImportLibraryDialog'))

const lazyDialogFallback = (
  <div className='grid place-items-center p-4'>
    <Spinner size='sm' />
  </div>
)

export interface TorrentsPageProps {
  sortABC: boolean
  sortCategory: string
  onAdd?: () => void
  onClearCategory?: () => void
}

/** ~140–180px poster tiles by viewport: denser on phone, roomier on desktop. */
const POSTER_GRID_CLASS = 'torrent-poster-grid grid min-h-full gap-4 p-4 pb-8 sm:gap-4 sm:p-4 md:gap-5 md:p-6'
const SKELETON_TILE_COUNT = 12

function sortTorrents(torrents: TorrentStat[], sortABC: boolean, sortCategory: string): TorrentStat[] {
  const inCategory = torrents.filter(torrent => {
    if (sortCategory === 'all') return true
    if (sortCategory === '') return !torrent.category
    return torrent.category === sortCategory
  })

  if (sortABC) {
    return [...inCategory].sort((a, b) => (a.title || '').localeCompare(b.title || '') || a.hash.localeCompare(b.hash))
  }

  return [...inCategory].sort((a, b) => {
    const recencyDiff = (b.timestamp || 0) - (a.timestamp || 0)
    return recencyDiff !== 0 ? recencyDiff : a.hash.localeCompare(b.hash)
  })
}

function prefersReducedMotion(): boolean {
  return typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

/**
 * Poster library grid: filter/sort, multi-select, Continue Watching, details/edit sheets.
 */
export default function TorrentsPage({ sortABC, sortCategory, onAdd, onClearCategory }: TorrentsPageProps) {
  const { t } = useTranslation()
  const isMobile = useMediaQuery(queryMax('mobile'))
  const toast = useOptionalAppToast()
  const queryClient = useQueryClient()
  const gridRef = useRef<HTMLDivElement>(null)
  const [detailsTorrent, setDetailsTorrent] = useState<TorrentStat | null>(null)
  const [resumePlay, setResumePlay] = useState<{ fileIndex: number; timecode: number } | null>(null)
  const [editingTorrent, setEditingTorrent] = useState<TorrentStat | null>(null)
  const [libraryQuery, setLibraryQuery] = useLocalJsonPref('libraryQuery', '')
  const [selectionMode, setSelectionMode] = useState(false)
  const [selectedHashes, setSelectedHashes] = useState<Set<string>>(() => new Set())
  const [continueTick, setContinueTick] = useState(0)

  const { data: torrents, isLoading, isError, error, refetch } = useTorrentsQuery()
  const { data: serverViewed } = useQuery({
    queryKey: ALL_VIEWED_QUERY_KEY,
    queryFn: listAllViewedEntries,
    enabled: Boolean(torrents?.length),
    staleTime: 15_000,
  })

  const [smartCollection, setSmartCollection] = useLocalJsonPref<'all' | 'active' | 'noPoster' | 'unwatched'>(
    'smartCollection',
    'all',
  )
  const [exportOpen, setExportOpen] = useState(false)
  const [importOpen, setImportOpen] = useState(false)

  useEffect(() => {
    const onImport = () => setImportOpen(true)
    const onExport = () => setExportOpen(true)
    const onPlaylists = () => {
      void openPlaylistLink(filteredPlaylistAllUrl(sortCategory, libraryQuery)).then(kind => {
        if (kind === 'copied' || kind === 'error') {
          toast?.showToast({
            message: t(kind === 'copied' ? 'Copied' : 'Error'),
            severity: kind === 'copied' ? 'success' : 'error',
          })
        }
      })
    }
    window.addEventListener(OPEN_IMPORT_LIBRARY_EVENT, onImport)
    window.addEventListener(OPEN_EXPORT_LIBRARY_EVENT, onExport)
    window.addEventListener(DOWNLOAD_ALL_PLAYLISTS_EVENT, onPlaylists)
    return () => {
      window.removeEventListener(OPEN_IMPORT_LIBRARY_EVENT, onImport)
      window.removeEventListener(OPEN_EXPORT_LIBRARY_EVENT, onExport)
      window.removeEventListener(DOWNLOAD_ALL_PLAYLISTS_EVENT, onPlaylists)
    }
  }, [sortCategory, libraryQuery, t, toast])

  const visibleTorrents = useMemo(() => {
    const sorted = torrents ? sortTorrents(torrents, sortABC, sortCategory) : []
    const q = libraryQuery.trim().toLowerCase()
    let list = sorted
    if (q) {
      list = list.filter(torrent => {
        const hay =
          `${torrent.title || ''} ${torrent.name || ''} ${torrent.category || ''} ${torrent.hash}`.toLowerCase()
        return hay.includes(q)
      })
    }
    if (smartCollection === 'active') {
      list = list.filter(tr => (tr.download_speed || 0) > 0 || (tr.upload_speed || 0) > 0 || (tr.active_peers || 0) > 0)
    } else if (smartCollection === 'noPoster') {
      list = list.filter(tr => !tr.poster)
    } else if (smartCollection === 'unwatched') {
      const viewedHashes = new Set((serverViewed || []).map(v => v.hash).filter(Boolean) as string[])
      list = list.filter(tr => !viewedHashes.has(tr.hash))
    }
    return list
  }, [torrents, sortABC, sortCategory, libraryQuery, smartCollection, serverViewed])

  const continueEntries = useMemo(
    () => buildContinueWatchShelf(torrents, serverViewed, continueTick),
    [torrents, serverViewed, continueTick],
  )

  const resumeByHash = useMemo(() => {
    const map = new Map<string, ContinueWatchEntry>()
    for (const entry of continueEntries) {
      if (!map.has(entry.hash)) map.set(entry.hash, entry)
    }
    return map
  }, [continueEntries])

  const collectionOptions = [
    ['all', t('All')],
    ['active', t('SmartCollectionActive')],
    ['noPoster', t('SmartCollectionNoPoster')],
    ['unwatched', t('SmartCollectionUnwatched')],
  ] as const
  const collectionLabel = collectionOptions.find(([id]) => id === smartCollection)?.[1] ?? t('All')

  const visibleHashKey = useMemo(() => visibleTorrents.map(torrent => torrent.hash).join(','), [visibleTorrents])
  const prevHashesRef = useRef<Set<string>>(new Set())
  const prevViewRef = useRef({ sortABC, sortCategory, libraryQuery })

  useGSAP(
    () => {
      if (prefersReducedMotion()) return
      const root = gridRef.current
      if (!root) return

      const currentHashes = visibleHashKey ? visibleHashKey.split(',') : []
      const prevHashes = prevHashesRef.current
      const viewChanged =
        prevViewRef.current.sortABC !== sortABC ||
        prevViewRef.current.sortCategory !== sortCategory ||
        prevViewRef.current.libraryQuery !== libraryQuery

      prevViewRef.current = { sortABC, sortCategory, libraryQuery }

      let tiles: Element[]
      if (viewChanged || prevHashes.size === 0) {
        tiles = Array.from(root.querySelectorAll('.torrent-card'))
      } else {
        const newHashes = currentHashes.filter(hash => hash && !prevHashes.has(hash))
        tiles = newHashes
          .map(hash => root.querySelector(`.torrent-card[data-hash="${CSS.escape(hash)}"]`))
          .filter((node): node is Element => node != null)
      }

      prevHashesRef.current = new Set(currentHashes)

      if (!tiles.length) return
      gsap.from(tiles, {
        opacity: 0,
        y: 18,
        duration: 0.45,
        stagger: 0.035,
        ease: 'power2.out',
        clearProps: 'opacity,transform',
      })
    },
    { scope: gridRef, dependencies: [visibleHashKey, sortABC, sortCategory, libraryQuery] },
  )

  const toggleSelect = (hash: string) => {
    setSelectedHashes(prev => {
      const next = new Set(prev)
      if (next.has(hash)) next.delete(hash)
      else next.add(hash)
      return next
    })
  }

  const exitSelection = () => {
    setSelectionMode(false)
    setSelectedHashes(new Set())
  }

  const runBulk = async (action: 'drop' | 'delete') => {
    const hashes = [...selectedHashes]
    if (!hashes.length) return
    const mutate = action === 'drop' ? dropTorrent : removeTorrent
    const previous = queryClient.getQueryData<TorrentStat[]>(TORRENTS_QUERY_KEY)
    const remove = new Set(hashes)
    if (action === 'drop') {
      markTorrentsDroppedInList(queryClient, hashes)
    } else {
      queryClient.setQueryData<TorrentStat[]>(TORRENTS_QUERY_KEY, prev => prev?.filter(item => !remove.has(item.hash)))
    }
    exitSelection()
    try {
      await Promise.all(hashes.map(hash => mutate(hash)))
      void queryClient.invalidateQueries({ queryKey: TORRENTS_QUERY_KEY })
      toast?.showToast({
        message: action === 'drop' ? t('DropTorrent') : t('Delete'),
        severity: 'success',
      })
    } catch {
      if (previous) queryClient.setQueryData(TORRENTS_QUERY_KEY, previous)
      toast?.showToast({ message: t('Error'), severity: 'error' })
    }
  }

  const dismissContinue = (hash: string) => {
    removeContinueWatching(hash)
    setContinueTick(v => v + 1)
    queryClient.setQueryData<ViewedFileEntry[]>(ALL_VIEWED_QUERY_KEY, old =>
      (old || []).filter(row => row.hash !== hash),
    )
    void clearViewedFiles(hash)
      .then(() =>
        Promise.all([
          queryClient.invalidateQueries({ queryKey: ALL_VIEWED_QUERY_KEY }),
          queryClient.invalidateQueries({ queryKey: VIEWED_QUERY_KEY(hash) }),
        ]),
      )
      .catch(() => {
        void queryClient.invalidateQueries({ queryKey: ALL_VIEWED_QUERY_KEY })
        toast?.showToast({ message: t('Error'), severity: 'error' })
      })
  }

  const headerBtn = `${iconBtn} text-app-header-foreground hover-fine:bg-white/10`

  const headerChrome = (
    <LibraryHeaderPortal>
      <TextField
        aria-label={t('FilterLibrary')}
        value={libraryQuery}
        onChange={setLibraryQuery}
        className='min-w-0 flex-1 overflow-hidden'
      >
        <Label className='sr-only'>{t('FilterLibrary')}</Label>
        <Input
          placeholder={t('FilterLibrary')}
          className='min-h-9 min-w-0 w-full border-transparent bg-white/15 text-sm text-app-header-foreground shadow-none outline-none placeholder:text-app-header-foreground/55'
        />
      </TextField>
      {libraryQuery ? (
        <Button
          isIconOnly
          variant='ghost'
          className={headerBtn}
          aria-label={t('Clear')}
          onPress={() => setLibraryQuery('')}
        >
          <X {...iconNav} aria-hidden />
        </Button>
      ) : null}
      {selectionMode ? (
        <Button isIconOnly variant='ghost' className={headerBtn} aria-label={t('Cancel')} onPress={exitSelection}>
          <XSquare {...iconNav} aria-hidden />
        </Button>
      ) : (
        <Button
          isIconOnly
          variant='ghost'
          className={headerBtn}
          aria-label={t('Select')}
          onPress={() => setSelectionMode(true)}
        >
          <CheckSquare {...iconNav} aria-hidden />
        </Button>
      )}
      <Dropdown>
        <Dropdown.Trigger>
          <Button
            variant='ghost'
            isIconOnly={isMobile}
            aria-label={collectionLabel}
            className={
              isMobile
                ? headerBtn
                : 'min-h-11 max-w-[7.5rem] shrink-0 gap-1 px-2 text-xs font-medium text-app-header-foreground hover-fine:bg-white/10 sm:max-w-[11rem] sm:text-sm'
            }
          >
            {isMobile ? (
              <span className='inline-flex size-full items-center justify-center [&>svg]:m-0 [&>svg]:block'>
                <LayoutGrid {...iconNav} aria-hidden />
              </span>
            ) : (
              <>
                <span className='truncate'>{collectionLabel}</span>
                <ChevronDown size={14} strokeWidth={1.75} aria-hidden />
              </>
            )}
          </Button>
        </Dropdown.Trigger>
        <Dropdown.Popover placement='bottom end' className='min-w-[12rem]'>
          <Dropdown.Menu aria-label={t('FilterLibrary')}>
            {collectionOptions.map(([id, label]) => (
              <Dropdown.Item
                key={id}
                textValue={label}
                onPress={() => setSmartCollection(id)}
                className='min-h-11 gap-2'
              >
                <span className='min-w-0 flex-1 truncate text-sm'>{label}</span>
                {smartCollection === id ? (
                  <Check {...iconMenu} className='shrink-0 text-accent' aria-hidden />
                ) : (
                  <span className='size-4 shrink-0' aria-hidden />
                )}
              </Dropdown.Item>
            ))}
          </Dropdown.Menu>
        </Dropdown.Popover>
      </Dropdown>
      {continueEntries.length > 0 && !selectionMode ? (
        <Dropdown>
          <Dropdown.Trigger>
            <Button variant='ghost' isIconOnly className={`${headerBtn} relative`} aria-label={t('ContinueWatching')}>
              <span className='inline-flex size-full items-center justify-center [&>svg]:m-0 [&>svg]:block'>
                <History {...iconNav} aria-hidden />
              </span>
              <span className='absolute top-0.5 right-0.5 grid min-w-4 place-items-center rounded-full bg-white/90 px-1 text-[10px] leading-4 font-semibold text-black'>
                {continueEntries.length}
              </span>
            </Button>
          </Dropdown.Trigger>
          <Dropdown.Popover
            placement='bottom end'
            className='max-h-[min(70dvh,24rem)] min-w-[16rem] overflow-y-auto p-1'
          >
            {continueEntries.map(entry => {
              const torrent = torrents?.find(item => item.hash === entry.hash)
              if (!torrent) return null
              return (
                <div
                  key={`${entry.hash}:${entry.fileIndex}`}
                  className='flex min-w-0 items-center gap-1 rounded-lg pr-1 hover-fine:bg-surface-secondary'
                >
                  <button
                    type='button'
                    className='min-w-0 flex-1 px-3 py-2 text-left'
                    onClick={() => {
                      setResumePlay({ fileIndex: entry.fileIndex, timecode: entry.timecode })
                      setDetailsTorrent(torrent)
                    }}
                  >
                    <p className='truncate text-sm font-semibold'>{entry.title}</p>
                    <p className='truncate text-xs text-muted'>{entry.fileName}</p>
                  </button>
                  <button
                    type='button'
                    className='shrink-0 rounded-md p-1.5 text-muted hover-fine:bg-surface-tertiary hover-fine:text-foreground'
                    aria-label={t('Clear')}
                    onClick={() => dismissContinue(entry.hash)}
                  >
                    <X {...iconMenu} aria-hidden />
                  </button>
                </div>
              )
            })}
          </Dropdown.Popover>
        </Dropdown>
      ) : null}
    </LibraryHeaderPortal>
  )

  const bulkBar =
    selectionMode && selectedHashes.size > 0 ? (
      <div className='flex flex-wrap items-center gap-2 border-b border-border/60 bg-background px-3 py-2 sm:px-4'>
        <Button size='sm' variant='outline' className='min-h-11' onPress={() => void runBulk('drop')}>
          {t('DropTorrent')} ({selectedHashes.size})
        </Button>
        <Button size='sm' variant='danger' className='min-h-11' onPress={() => void runBulk('delete')}>
          <Trash2 {...iconMenu} aria-hidden />
          {t('Delete')} ({selectedHashes.size})
        </Button>
      </div>
    ) : null

  if (isLoading) {
    return (
      <>
        {headerChrome}
        <div className={POSTER_GRID_CLASS}>
          {Array.from({ length: SKELETON_TILE_COUNT }, (_, index) => (
            <div key={index} className='flex flex-col gap-2'>
              <div className='aspect-[2/3] w-full animate-pulse rounded-2xl bg-surface-secondary' />
              <div className='h-3.5 w-4/5 animate-pulse rounded-full bg-surface-secondary' />
            </div>
          ))}
        </div>
      </>
    )
  }

  if (isError) {
    const status = (error as { response?: { status?: number } } | null)?.response?.status
    const unauthorized = status === 401 || status === 403
    return (
      <>
        {headerChrome}
        <div className='grid min-h-[60vh] place-items-center p-6 text-center'>
          <div className='flex flex-col items-center gap-3'>
            <CloudOff {...iconEmptyLg} className='text-muted' />
            <p className='text-lg font-semibold text-foreground'>
              {unauthorized ? t('AuthRequiredTitle') : t('NoServerConnection')}
            </p>
            {unauthorized ? <p className='max-w-md text-sm text-muted'>{t('AuthRequiredHint')}</p> : null}
            <div className='flex flex-wrap items-center justify-center gap-2'>
              {unauthorized ? (
                <Button variant='primary' onPress={() => logoutBasicAuth()}>
                  {t('SignIn')}
                </Button>
              ) : null}
              <Button variant={unauthorized ? 'secondary' : 'primary'} onPress={() => void refetch()}>
                {t('Retry')}
              </Button>
            </div>
          </div>
        </div>
      </>
    )
  }

  if (!torrents?.length) {
    return (
      <>
        {headerChrome}
        <div className='grid min-h-[60vh] place-items-center p-6'>
          <div className='flex min-h-[190px] w-full max-w-[280px] flex-col items-center justify-center gap-3 rounded-3xl border border-dashed border-border bg-surface p-8 text-center'>
            <FolderPlus size={40} strokeWidth={1.25} className='text-accent' aria-hidden />
            <p className='text-lg font-semibold text-foreground'>{t('NoTorrentsAdded')}</p>
            <Button variant='primary' onPress={onAdd} isDisabled={!onAdd}>
              {t('AddFirstTorrent')}
            </Button>
          </div>
        </div>
        <Suspense fallback={lazyDialogFallback}>
          <DialogErrorBoundary onClose={() => setExportOpen(false)}>
            <ExportLibraryDialog open={exportOpen} onClose={() => setExportOpen(false)} torrents={[]} />
          </DialogErrorBoundary>
          <DialogErrorBoundary onClose={() => setImportOpen(false)}>
            <ImportLibraryDialog open={importOpen} onClose={() => setImportOpen(false)} />
          </DialogErrorBoundary>
        </Suspense>
      </>
    )
  }

  return (
    <>
      {headerChrome}
      {bulkBar}

      {!visibleTorrents.length ? (
        <div className='grid min-h-[40vh] place-items-center p-6 text-center'>
          <div className='flex flex-col items-center gap-3 text-muted'>
            <SearchX {...iconEmpty} />
            <p>{libraryQuery.trim() ? t('NoSearchResults') : t('NoTorrentsInCategory')}</p>
            {libraryQuery.trim() ? (
              <Button variant='secondary' onPress={() => setLibraryQuery('')}>
                <Search {...iconMenu} aria-hidden />
                {t('Clear')}
              </Button>
            ) : onClearCategory ? (
              <Button variant='secondary' onPress={onClearCategory}>
                {t('ShowAllTorrents')}
              </Button>
            ) : null}
          </div>
        </div>
      ) : (
        <div ref={gridRef} className={POSTER_GRID_CLASS}>
          {visibleTorrents.map(torrent => {
            const resumeEntry = resumeByHash.get(torrent.hash)
            return (
              <TorrentCard
                key={torrent.hash}
                torrent={torrent}
                onSelect={torrent => {
                  setResumePlay(null)
                  setDetailsTorrent(torrent)
                }}
                onEdit={setEditingTorrent}
                selectionMode={selectionMode}
                selected={selectedHashes.has(torrent.hash)}
                onToggleSelect={toggleSelect}
                resumeEntry={resumeEntry}
                onResume={
                  resumeEntry
                    ? () => {
                        setResumePlay({ fileIndex: resumeEntry.fileIndex, timecode: resumeEntry.timecode })
                        setDetailsTorrent(torrent)
                      }
                    : undefined
                }
              />
            )
          })}
        </div>
      )}

      {detailsTorrent ? (
        <Suspense fallback={lazyDialogFallback}>
          <DialogErrorBoundary onClose={() => setDetailsTorrent(null)}>
            <DetailsDialog
              torrent={detailsTorrent}
              autoPlayFileId={resumePlay?.fileIndex}
              autoPlayTimecode={resumePlay?.timecode}
              onClose={() => {
                setDetailsTorrent(null)
                setResumePlay(null)
              }}
              onEdit={(torrent: TorrentStat) => {
                setDetailsTorrent(null)
                setResumePlay(null)
                setEditingTorrent(torrent)
              }}
            />
          </DialogErrorBoundary>
        </Suspense>
      ) : null}

      <Suspense fallback={lazyDialogFallback}>
        <EditTorrentDialog
          torrent={editingTorrent}
          open={Boolean(editingTorrent)}
          onClose={() => setEditingTorrent(null)}
        />
      </Suspense>

      <Suspense fallback={lazyDialogFallback}>
        <DialogErrorBoundary onClose={() => setExportOpen(false)}>
          <ExportLibraryDialog open={exportOpen} onClose={() => setExportOpen(false)} torrents={torrents ?? []} />
        </DialogErrorBoundary>
        <DialogErrorBoundary onClose={() => setImportOpen(false)}>
          <ImportLibraryDialog open={importOpen} onClose={() => setImportOpen(false)} />
        </DialogErrorBoundary>
      </Suspense>
    </>
  )
}

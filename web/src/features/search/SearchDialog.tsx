import {
  Button,
  Input,
  Label,
  ListBox,
  Modal,
  Select,
  Spinner,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  useMediaQuery,
} from '@heroui/react'
import { AlertCircle, ArrowDown, ArrowUp, SearchX } from 'lucide-react'
import { useEffect, useMemo, useRef, useState, type KeyboardEvent } from 'react'
import axios, { isAxiosError } from 'axios'
import { useTranslation } from 'react-i18next'
import { useQueryClient } from '@tanstack/react-query'

import type { SearchResultItem, TorznabCapsCategory, TorznabUrl } from 'shared/api/types'
import { fetchTorznabCaps, searchRutor, searchTorznab } from 'shared/api/search'
import { getSettings } from 'shared/api/settings'
import { addTorrent, TORRENTS_QUERY_KEY } from 'shared/api/torrents'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import { queryMax } from 'shared/theme/breakpoints'
import { formatSizeToClassicUnits, parseSizeToBytes } from 'shared/lib/format'
import { getMoviePosters, shortenTitleForPosterSearch } from 'shared/lib/torrentHelpers'
import AppDialog from 'shared/ui/AppDialog'
import { iconEmpty } from 'shared/ui/iconProps'
import { DIALOG_SEARCH } from 'shared/ui/dialogSizes'
import { useOptionalAppToast } from 'shared/ui/Toast'

import SearchResultsGrid from './SearchResultsGrid'
import { beginSearchRequest, isCurrentSearch } from './searchRequest'
import { flattenTorznabCategories, staticTorznabCategoryOptions, type TorznabCategoryOption } from './torznabCategories'

export interface SearchDialogProps {
  open: boolean
  onClose: () => void
}

type SortField = 'size' | 'seeds' | 'peers'
type SortDirection = 'asc' | 'desc'
type TrackerSelection = number | 'rutor'

const resultDedupeKey = (item: SearchResultItem): string =>
  (item.Hash || item.Magnet || item.Link || item.Title || '').toLowerCase()

const mergeSearchResults = (...lists: SearchResultItem[][]): SearchResultItem[] => {
  const seen = new Set<string>()
  const out: SearchResultItem[] = []
  for (const list of lists) {
    for (const item of list || []) {
      const key = resultDedupeKey(item)
      if (key) {
        if (seen.has(key)) continue
        seen.add(key)
      }
      out.push(item)
    }
  }
  return out
}

const axiosErrorMessage = (err: unknown, fallback: string): string => {
  if (isAxiosError(err)) {
    const data = err.response?.data as { error?: string } | undefined
    if (data?.error) return String(data.error)
    if (err.message) return err.message
  }
  if (err instanceof Error && err.message) return err.message
  return fallback
}

/**
 * Torznab + RuTor search sheet. Aborts in-flight requests on query change;
 * optional poster enrichment via TMDB when adding a hit.
 */
export default function SearchDialog({ open, onClose }: SearchDialogProps) {
  const { t, i18n } = useTranslation()
  const queryClient = useQueryClient()
  const toast = useOptionalAppToast()
  const isMobile = useMediaQuery(queryMax('mobile'))
  const isFullScreenBreakpoint = useDialogFullScreen()

  const [query, setQuery] = useState('')
  const [results, setResults] = useState<SearchResultItem[]>([])
  const [loading, setLoading] = useState(false)
  const [searched, setSearched] = useState(false)
  const [adding, setAdding] = useState(false)
  const [addingKey, setAddingKey] = useState<string | null>(null)
  const [trackers, setTrackers] = useState<TorznabUrl[]>([])
  const [enableRutor, setEnableRutor] = useState(false)
  const [enableTorznab, setEnableTorznab] = useState(false)
  const [selectedTracker, setSelectedTracker] = useState<TrackerSelection>(-1)
  const [sortField, setSortField] = useState<SortField>('seeds')
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc')
  const [settingsLoaded, setSettingsLoaded] = useState(false)
  const [category, setCategory] = useState('')
  const [pageOffset, setPageOffset] = useState(0)
  const [pageLimit, setPageLimit] = useState(100)
  const [hasMore, setHasMore] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const [capsCategories, setCapsCategories] = useState<TorznabCapsCategory[] | null>(null)
  const [capsLoading, setCapsLoading] = useState(false)
  const [searchError, setSearchError] = useState<string | null>(null)

  const userSortedRef = useRef(false)
  const searchAbortRef = useRef<AbortController | null>(null)
  const searchGenRef = useRef(0)

  const hasTorznab = enableTorznab && trackers.length > 0
  const hasRutor = enableRutor
  const hasAnySource = hasTorznab || hasRutor
  const showAllTrackers = hasTorznab
  const showCategorySelect = hasTorznab && selectedTracker !== 'rutor'
  const showLoadMore =
    !loading && hasMore && typeof selectedTracker === 'number' && selectedTracker >= 0 && results.length > 0

  const categoryOptions: TorznabCategoryOption[] = useMemo(() => {
    const allLabel = t('Torznab.AllCategories')
    if (typeof selectedTracker === 'number' && selectedTracker >= 0 && capsCategories?.length) {
      return flattenTorznabCategories(capsCategories, allLabel)
    }
    return staticTorznabCategoryOptions(allLabel)
  }, [capsCategories, selectedTracker, t])

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- reset category paging when tracker changes
    setCategory('')
    setPageOffset(0)
    setHasMore(false)
  }, [selectedTracker])

  useEffect(() => {
    if (!open || typeof selectedTracker !== 'number' || selectedTracker < 0) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- clear caps when search closed / All Trackers
      setCapsCategories(null)
      setCapsLoading(false)
      return
    }

    const ac = new AbortController()
    setCapsLoading(true)
    fetchTorznabCaps(selectedTracker, ac.signal)
      .then(caps => {
        if (ac.signal.aborted) return
        setCapsCategories(caps.categories || [])
        const nextLimit = caps.limits?.default || caps.limits?.max || 100
        setPageLimit(nextLimit > 0 ? nextLimit : 100)
      })
      .catch(() => {
        if (ac.signal.aborted) return
        setCapsCategories(null)
      })
      .finally(() => {
        if (!ac.signal.aborted) setCapsLoading(false)
      })

    return () => ac.abort()
  }, [open, selectedTracker])

  useEffect(() => {
    if (!open) return
    const ac = new AbortController()
    // eslint-disable-next-line react-hooks/set-state-in-effect -- load search settings when dialog opens
    setSettingsLoaded(false)
    getSettings(ac.signal)
      .then(data => {
        const urls = data.TorznabUrls || []
        const torznabOn = Boolean(data.EnableTorznabSearch)
        const rutorOn = Boolean(data.EnableRutorSearch)
        setTrackers(urls)
        setEnableTorznab(torznabOn)
        setEnableRutor(rutorOn)
        if (torznabOn && urls.length > 0) setSelectedTracker(-1)
        else if (rutorOn) setSelectedTracker('rutor')
      })
      .catch(() => {
        if (ac.signal.aborted) return
        setTrackers([])
        setEnableTorznab(false)
        setEnableRutor(false)
      })
      .finally(() => {
        if (!ac.signal.aborted) setSettingsLoaded(true)
      })
    return () => ac.abort()
  }, [open])

  useEffect(
    () => () => {
      searchAbortRef.current?.abort()
    },
    [],
  )

  const runSearch = async (opts?: { append?: boolean; categoryOverride?: string; offsetOverride?: number }) => {
    const q = query.trim()
    if (!q || !hasAnySource) return

    const append = opts?.append ?? false
    const searchCategory = opts?.categoryOverride ?? category
    const searchOffset = opts?.offsetOverride ?? 0
    const { ac, gen } = beginSearchRequest(searchAbortRef, searchGenRef)

    if (append) setLoadingMore(true)
    else {
      setLoading(true)
      setSearched(true)
      setSearchError(null)
      setResults([])
      setPageOffset(0)
      setHasMore(false)
    }

    try {
      if (selectedTracker === -1) {
        const requests: Promise<SearchResultItem[]>[] = []
        const torznabOpts = searchCategory ? { cat: searchCategory, signal: ac.signal } : { signal: ac.signal }
        if (hasTorznab) requests.push(searchTorznab(q, torznabOpts))
        if (hasRutor) requests.push(searchRutor(q, ac.signal))
        if (requests.length === 0) {
          if (isCurrentSearch(gen, searchGenRef.current, ac.signal)) {
            toast?.showToast({ message: t('Torznab.NoSearchSources'), severity: 'error' })
          }
          return
        }
        const settled = await Promise.allSettled(requests)
        if (!isCurrentSearch(gen, searchGenRef.current, ac.signal)) return

        const lists: SearchResultItem[][] = []
        let firstError: unknown
        for (const result of settled) {
          if (result.status === 'fulfilled') {
            lists.push(result.value)
          } else if (!firstError && !axios.isCancel(result.reason) && result.reason?.code !== 'ERR_CANCELED') {
            firstError = result.reason
          }
        }
        if (lists.length === 0) {
          throw firstError || new Error(t('Torznab.SearchFailed'))
        }
        const merged = mergeSearchResults(...lists)
        setResults(merged)
        setHasMore(false)
        if (!userSortedRef.current && merged.length > 0) {
          setSortField('seeds')
          setSortDirection('desc')
        }
        return
      }

      let next: SearchResultItem[] = []
      if (selectedTracker === 'rutor') {
        next = await searchRutor(q, ac.signal)
      } else if (typeof selectedTracker === 'number') {
        next = await searchTorznab(q, {
          index: selectedTracker,
          cat: searchCategory || undefined,
          offset: searchOffset,
          limit: pageLimit,
          signal: ac.signal,
        })
      }
      if (!isCurrentSearch(gen, searchGenRef.current, ac.signal)) return

      if (append) {
        setResults(prev => mergeSearchResults(prev, next))
        setPageOffset(searchOffset)
      } else {
        setResults(next)
        setPageOffset(0)
        if (!userSortedRef.current && next.length > 0) {
          setSortField('seeds')
          setSortDirection('desc')
        }
      }
      setHasMore(typeof selectedTracker === 'number' && selectedTracker >= 0 && next.length >= pageLimit)
    } catch (err) {
      if (axios.isCancel(err) || (isAxiosError(err) && err.code === 'ERR_CANCELED')) return
      if (!isCurrentSearch(gen, searchGenRef.current, ac.signal)) return
      const message = axiosErrorMessage(err, t('Torznab.SearchFailed'))
      if (!append) {
        setResults([])
        setSearchError(message)
      }
      toast?.showToast({ message, severity: 'error' })
    } finally {
      if (gen === searchGenRef.current) {
        if (append) setLoadingMore(false)
        else setLoading(false)
      }
    }
  }

  const handleSearch = () => void runSearch()

  const handleLoadMore = () => {
    const nextOffset = pageOffset + pageLimit
    void runSearch({ append: true, offsetOverride: nextOffset })
  }

  const handleCategoryChange = (value: string) => {
    setCategory(value)
    setPageOffset(0)
    setHasMore(false)
    if (searched && query.trim()) void runSearch({ categoryOverride: value })
  }

  const handleAdd = async (item: SearchResultItem) => {
    const link = item.Magnet || item.Link || item.Hash
    if (!link) return

    const key = resultDedupeKey(item)
    setAdding(true)
    setAddingKey(key)
    try {
      let poster = item.Poster || ''
      if (!poster && item.Title) {
        const posterQuery = shortenTitleForPosterSearch(item.Title) || item.Title
        const urls = await getMoviePosters(posterQuery, i18n.language?.startsWith('ru') ? 'ru' : 'en')
        poster = urls?.[0] || ''
      }
      await addTorrent({ link, title: item.Title, poster })
      await queryClient.invalidateQueries({ queryKey: TORRENTS_QUERY_KEY })
      toast?.showToast({ message: t('Search.TorrentAdded'), severity: 'success' })
    } catch {
      toast?.showToast({ message: t('Error'), severity: 'error' })
    } finally {
      setAdding(false)
      setAddingKey(null)
    }
  }

  const handleSortChip = (field: SortField) => {
    userSortedRef.current = true
    if (sortField === field) {
      setSortDirection(prev => (prev === 'asc' ? 'desc' : 'asc'))
      return
    }
    setSortField(field)
    setSortDirection('desc')
  }

  const sortedResults = useMemo(() => {
    if (results.length === 0) return results
    return [...results].sort((a, b) => {
      let aVal: number
      let bVal: number
      switch (sortField) {
        case 'size':
          aVal = parseSizeToBytes(a.Size || '0')
          bVal = parseSizeToBytes(b.Size || '0')
          break
        case 'seeds':
          aVal = a.Seed || 0
          bVal = b.Seed || 0
          break
        case 'peers':
          aVal = a.Peer || 0
          bVal = b.Peer || 0
          break
        default:
          return 0
      }
      if (aVal === bVal) return 0
      return sortDirection === 'asc' ? (aVal < bVal ? -1 : 1) : aVal > bVal ? -1 : 1
    })
  }, [results, sortField, sortDirection])

  const emptyMessage = (() => {
    if (searchError) return null
    if (!settingsLoaded) return null
    if (!hasAnySource) {
      if (enableTorznab && trackers.length === 0) return t('Torznab.NoIndexersConfigured')
      return t('Torznab.NoSearchSources')
    }
    if (loading) return null
    if (searched && results.length === 0) return t('Search.NoResults')
    return null
  })()

  const sortFields: { field: SortField; label: string }[] = [
    { field: 'size', label: t('Torznab.SortBySize') },
    { field: 'seeds', label: t('Torznab.SortBySeeds') },
    { field: 'peers', label: t('Torznab.SortByPeers') },
  ]

  const footerButtonClassName = isMobile ? 'min-h-11 px-4' : undefined
  const trackerKey = selectedTracker === 'rutor' ? 'rutor' : String(selectedTracker)

  return (
    <AppDialog
      open={open}
      onClose={onClose}
      size='lg'
      fullScreen={isFullScreenBreakpoint}
      dialogClassName='flex flex-col overflow-hidden'
      dialogStyle={isFullScreenBreakpoint ? undefined : DIALOG_SEARCH}
    >
      <Modal.Header className='shrink-0'>
        <Modal.Heading>{t('Torznab.SearchTorrents')}</Modal.Heading>
        <Modal.CloseTrigger aria-label={t('Close')} />
      </Modal.Header>
      <Modal.Body className='flex min-h-0 flex-1 flex-col overflow-hidden'>
        <div className='sticky top-0 z-10 shrink-0 -mx-1 border-b border-border/60 bg-surface px-1 pb-3 pt-1'>
          <div className='flex flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-end'>
            <Select
              selectedKey={trackerKey}
              onSelectionChange={key => {
                const value = String(key)
                setSelectedTracker(value === 'rutor' ? 'rutor' : Number(value))
              }}
              isDisabled={!hasAnySource}
              className='sm:min-w-[180px]'
            >
              <Label>{t('Search.Tracker')}</Label>
              <Select.Trigger>
                <Select.Value />
                <Select.Indicator />
              </Select.Trigger>
              <Select.Popover>
                <ListBox>
                  {showAllTrackers ? <ListBox.Item id='-1'>{t('Search.AllTrackers')}</ListBox.Item> : null}
                  {hasRutor ? <ListBox.Item id='rutor'>{t('Search.Rutor')}</ListBox.Item> : null}
                  {hasTorznab
                    ? trackers.map((tracker, index) => (
                        <ListBox.Item key={`${tracker.Host}-${tracker.Key}`} id={String(index)}>
                          {tracker.Name || tracker.Host}
                        </ListBox.Item>
                      ))
                    : null}
                </ListBox>
              </Select.Popover>
            </Select>

            {showCategorySelect ? (
              <Select
                selectedKey={category}
                onSelectionChange={key => handleCategoryChange(String(key))}
                isDisabled={!hasAnySource || capsLoading}
                className='sm:min-w-[180px]'
              >
                <Label>{t('Torznab.Category')}</Label>
                <Select.Trigger>
                  <Select.Value />
                  <Select.Indicator />
                </Select.Trigger>
                <Select.Popover>
                  <ListBox>
                    {categoryOptions.map(option => (
                      <ListBox.Item key={option.id || 'all'} id={option.id}>
                        <span style={option.indent ? { paddingLeft: `${option.indent * 1}rem` } : undefined}>
                          {option.name}
                        </span>
                      </ListBox.Item>
                    ))}
                  </ListBox>
                </Select.Popover>
              </Select>
            ) : null}

            <TextField
              value={query}
              onChange={setQuery}
              isDisabled={!hasAnySource}
              className='flex-1'
              onKeyDown={(e: KeyboardEvent<HTMLInputElement>) => {
                if (e.key === 'Enter') void handleSearch()
              }}
            >
              <Label>{t('Search.QueryLabel')}</Label>
              <Input placeholder={t('Torznab.SearchMoviesShows')} />
            </TextField>

            <Button
              variant='primary'
              onPress={() => void handleSearch()}
              isDisabled={loading || !hasAnySource || !query.trim()}
              className={footerButtonClassName}
            >
              {loading ? <Spinner size='sm' color='current' /> : t('nav.Search')}
            </Button>
          </div>
        </div>

        <div className='min-h-0 flex-1 overflow-y-auto overscroll-contain pt-3'>
          {searched && sortedResults.length > 0 ? (
            <div className='mb-3 flex flex-wrap items-center gap-2'>
              <p className='text-sm text-muted'>{t('Torznab.ResultsCount', { count: sortedResults.length })}</p>
              <ToggleButtonGroup selectionMode='single' selectedKeys={[sortField]} className='flex flex-wrap gap-1'>
                {sortFields.map(({ field, label }) => {
                  const active = sortField === field
                  return (
                    <ToggleButton key={field} id={field} onPress={() => handleSortChip(field)}>
                      {label}
                      {active ? (
                        sortDirection === 'asc' ? (
                          <ArrowUp className='ml-1' size={14} strokeWidth={1.75} aria-hidden />
                        ) : (
                          <ArrowDown className='ml-1' size={14} strokeWidth={1.75} aria-hidden />
                        )
                      ) : null}
                    </ToggleButton>
                  )
                })}
              </ToggleButtonGroup>
            </div>
          ) : null}

          <div className='min-h-[200px] sm:min-h-[280px]'>
            {loading ? (
              <div className='grid place-items-center py-16'>
                <Spinner size='lg' />
              </div>
            ) : null}

            {!loading && searchError ? (
              <div className='flex flex-col items-center gap-3 py-16 text-center'>
                <AlertCircle {...iconEmpty} className='text-danger' aria-hidden />
                <p className='text-muted'>{searchError}</p>
                <Button variant='primary' onPress={() => void runSearch()} className={footerButtonClassName}>
                  {t('Retry')}
                </Button>
              </div>
            ) : null}

            {!loading && emptyMessage ? (
              <div className='flex flex-col items-center gap-2 py-16 text-center text-muted'>
                <SearchX {...iconEmpty} aria-hidden />
                <p>{emptyMessage}</p>
              </div>
            ) : null}

            {!loading && sortedResults.length > 0 ? (
              <>
                <SearchResultsGrid
                  results={sortedResults}
                  adding={adding}
                  addingKey={addingKey}
                  resultDedupeKey={resultDedupeKey}
                  formatSize={item => formatSizeToClassicUnits(parseSizeToBytes(item.Size || '0'))}
                  onAdd={item => void handleAdd(item)}
                />
                {showLoadMore ? (
                  <div className='mt-4 flex justify-center'>
                    <Button
                      variant='secondary'
                      onPress={() => handleLoadMore()}
                      isDisabled={loadingMore}
                      className={footerButtonClassName}
                    >
                      {loadingMore ? <Spinner size='sm' color='current' /> : t('Torznab.LoadMore')}
                    </Button>
                  </div>
                ) : null}
              </>
            ) : null}
          </div>
        </div>
      </Modal.Body>
    </AppDialog>
  )
}

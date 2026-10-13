import {
  bufferAheadBytes,
  bufferFillPercent,
  formatBufferAheadLabel,
  formatBufferFilledLabel,
  humanizeSize,
  resolveBufferTargetBytes,
} from 'shared/lib/format'
import type { TorrentCache, TorrentStat } from 'shared/api/types'
import { useTranslation } from 'react-i18next'

import MetricRows, { type MetricRowItem } from './MetricRows'

export interface SwarmStatsPanelProps {
  torrent: TorrentStat
  className?: string
  columns?: 1 | 2
  /** When false, omit outer frame (parent already provides a panel). */
  framed?: boolean
  cacheReaders?: number | null
  /** Live `/cache` snapshot — Preload meter uses Filled vs Preload target; Cache line uses Filled vs Capacity. */
  cache?: TorrentCache | null
  /** Settings PreloadCache % for buffer target. Defaults to 50. */
  preloadCachePercent?: number | null
}

function formatDuration(seconds?: number): string | null {
  if (seconds == null || !Number.isFinite(seconds) || seconds <= 0) return null
  const total = Math.round(seconds)
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  const s = total % 60
  if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  return `${m}:${String(s).padStart(2, '0')}`
}

function hasPositive(n?: number | null): boolean {
  return n != null && Number.isFinite(n) && n > 0
}

function ProgressMeter({
  label,
  valueLabel,
  ratio,
  compact = false,
  title,
}: {
  label: string
  valueLabel: string
  ratio: number
  compact?: boolean
  title?: string
}) {
  const width = Number.isFinite(ratio) ? Math.min(100, Math.max(0, ratio)) : 0
  return (
    <div className='min-w-0' title={title}>
      <div className={`flex items-baseline justify-between gap-2 text-xs ${compact ? 'mb-0.5' : 'mb-1'}`}>
        <span className='truncate text-muted'>{label}</span>
        <span className='shrink-0 font-bold tabular-nums text-foreground'>{valueLabel}</span>
      </div>
      <div className={`overflow-hidden rounded-full bg-surface ${compact ? 'h-2' : 'h-2.5'}`}>
        <div className='h-full rounded-full bg-accent transition-[width] duration-300' style={{ width: `${width}%` }} />
      </div>
    </div>
  )
}

function useTorrentProgress(torrent: TorrentStat, cache?: TorrentCache | null, preloadCachePercent?: number | null) {
  const { t } = useTranslation()
  const loaded = torrent.loaded_size ?? 0
  const totalSize = torrent.torrent_size ?? 0
  const loadedPct = totalSize > 0 ? Math.min(100, Math.round((loaded / totalSize) * 100)) : 0
  const loadedRatio = totalSize > 0 ? loaded / totalSize : 0
  const loadedLabel =
    totalSize > 0 ? `${humanizeSize(loaded)} / ${humanizeSize(totalSize)} · ${loadedPct}%` : humanizeSize(loaded)

  const bufferTarget = resolveBufferTargetBytes(cache?.Capacity, preloadCachePercent)
  const bufferAhead = bufferAheadBytes(cache)
  const isStreamingBuffer = bufferAhead != null
  const bufferFilled = isStreamingBuffer ? bufferAhead : (cache?.Filled ?? 0)
  const preloadHint = isStreamingBuffer
    ? bufferAhead === 0
      ? t('BufferAheadEmptyHint')
      : t('BufferAheadHint')
    : t('BufferHint')
  const preloadTitle = isStreamingBuffer ? t('BufferAhead') : t('Buffer')
  const preloadPct = bufferFillPercent(bufferFilled, bufferTarget)
  const preloadLabel = isStreamingBuffer
    ? (formatBufferAheadLabel(bufferAhead) ?? '—')
    : (formatBufferFilledLabel(bufferFilled, bufferTarget, { percent: 'always' }) ??
      (bufferFilled > 0 ? humanizeSize(bufferFilled) : '—'))

  return {
    showLoaded: totalSize > 0 || loaded > 0,
    loadedLabel,
    loadedRatio,
    preloadTitle,
    preloadLabel,
    preloadPct,
    preloadHint,
  }
}

/** Full-width Loaded + Buffer meters above the Stats chart/swarm split. */
export function TorrentProgressStrip({
  torrent,
  cache,
  preloadCachePercent,
  className = '',
}: {
  torrent: TorrentStat
  cache?: TorrentCache | null
  preloadCachePercent?: number | null
  className?: string
}) {
  const { t } = useTranslation()
  const progress = useTorrentProgress(torrent, cache, preloadCachePercent)

  return (
    <div
      className={`grid shrink-0 grid-cols-1 gap-3 rounded-xl border border-border bg-surface-secondary p-2.5 sm:grid-cols-2 ${className}`.trim()}
    >
      {progress.showLoaded ? (
        <ProgressMeter label={t('Loaded')} valueLabel={progress.loadedLabel} ratio={progress.loadedRatio} />
      ) : (
        <ProgressMeter label={t('Loaded')} valueLabel='—' ratio={0} />
      )}
      <ProgressMeter
        label={progress.preloadTitle}
        valueLabel={progress.preloadLabel}
        ratio={progress.preloadPct}
        title={progress.preloadHint}
      />
    </div>
  )
}

/** Count chips always; mix bar is connected peers only so Pending cannot paint the whole bar. */
function PeerPanel({
  active,
  seeders,
  pending,
  halfOpen,
  labels,
}: {
  active: number
  seeders: number
  pending: number
  halfOpen: number
  labels: { active: string; seeders: string; pending: string; halfOpen: string }
}) {
  const chips = [
    { key: 'active', n: active, className: 'bg-accent', label: labels.active },
    { key: 'seeders', n: seeders, className: 'bg-accent/70', label: labels.seeders },
    { key: 'pending', n: pending, className: 'bg-warning', label: labels.pending },
    { key: 'half', n: halfOpen, className: 'bg-foreground/25', label: labels.halfOpen },
  ]
  const barParts = chips.filter(p => p.key !== 'pending')
  const barTotal = barParts.reduce((sum, p) => sum + Math.max(0, p.n), 0)

  return (
    <div className='min-w-0 space-y-1.5'>
      <div className='grid grid-cols-2 gap-1.5'>
        {chips.map(p => (
          <div
            key={p.key}
            className='flex min-w-0 items-center gap-1.5 rounded-lg bg-surface px-2 py-1.5'
            title={`${p.label}: ${p.n}`}
          >
            <span className={`size-1.5 shrink-0 rounded-full ${p.className}`} aria-hidden />
            <span className='min-w-0 truncate text-[10px] text-muted'>{p.label}</span>
            <span className='ml-auto shrink-0 text-xs font-bold tabular-nums text-foreground'>{p.n}</span>
          </div>
        ))}
      </div>
      <div className='flex h-2 overflow-hidden rounded-full bg-surface'>
        {barTotal > 0
          ? barParts.map(p =>
              p.n > 0 ? (
                <div
                  key={p.key}
                  className={`h-full ${p.className}`}
                  style={{ width: `${(p.n / barTotal) * 100}%` }}
                  title={`${p.label}: ${p.n}`}
                />
              ) : null,
            )
          : null}
      </div>
    </div>
  )
}

/** Transfer + peer composition for the Stats side card. Progress lives in {@link TorrentProgressStrip}. */
export default function SwarmStatsPanel({
  torrent,
  className = '',
  columns: columnsProp,
  framed = true,
  cacheReaders,
  cache,
}: SwarmStatsPanelProps) {
  const { t } = useTranslation()
  const columns = columnsProp ?? 2
  const durationLabel = formatDuration(torrent.duration_seconds)

  const transferItems: MetricRowItem[] = [
    { label: t('BytesRead'), value: humanizeSize(torrent.bytes_read) },
    { label: t('BytesWritten'), value: humanizeSize(torrent.bytes_written) },
    { label: t('UsefulRead'), value: humanizeSize(torrent.bytes_read_useful_data) },
  ]

  const detailItems: MetricRowItem[] = [
    ...(hasPositive(torrent.bytes_read_data)
      ? [{ label: t('BytesReadData'), value: humanizeSize(torrent.bytes_read_data) }]
      : []),
    ...(hasPositive(torrent.chunks_read) ? [{ label: t('ChunksRead'), value: String(torrent.chunks_read) }] : []),
    ...(hasPositive(torrent.chunks_written)
      ? [{ label: t('ChunksWritten'), value: String(torrent.chunks_written) }]
      : []),
    ...(hasPositive(torrent.chunks_read_wasted)
      ? [{ label: t('ChunksWasted'), value: String(torrent.chunks_read_wasted) }]
      : []),
    ...(hasPositive(torrent.chunks_read_useful)
      ? [{ label: t('ChunksUseful'), value: String(torrent.chunks_read_useful) }]
      : []),
    ...(hasPositive(torrent.pieces_dirtied_good)
      ? [{ label: t('PiecesDirtiedGood'), value: String(torrent.pieces_dirtied_good) }]
      : []),
    ...(hasPositive(torrent.pieces_dirtied_bad)
      ? [{ label: t('PiecesDirtiedBad'), value: String(torrent.pieces_dirtied_bad) }]
      : []),
    ...(cache?.PiecesCount != null ? [{ label: t('PiecesCount'), value: String(cache.PiecesCount) }] : []),
    ...(cache?.PiecesLength != null ? [{ label: t('PiecesLength'), value: humanizeSize(cache.PiecesLength) }] : []),
    ...(cacheReaders != null ? [{ label: t('CacheReaders'), value: String(cacheReaders) }] : []),
    ...(torrent.bit_rate ? [{ label: t('BitRate'), value: torrent.bit_rate }] : []),
    ...(durationLabel ? [{ label: t('FfpDuration'), value: durationLabel }] : []),
  ]

  const inner = (
    <>
      <div className='shrink-0 space-y-2'>
        <MetricRows framed={false} items={transferItems} columns={columns} dense />
        {detailItems.length > 0 ? <MetricRows framed={false} items={detailItems} columns={columns} dense /> : null}
      </div>

      <div className='mt-auto shrink-0 space-y-2 border-t border-border pt-2'>
        <PeerPanel
          active={torrent.active_peers ?? 0}
          seeders={torrent.connected_seeders ?? 0}
          pending={torrent.pending_peers ?? 0}
          halfOpen={torrent.half_open_peers ?? 0}
          labels={{
            active: t('ActivePeers'),
            seeders: t('Seeders'),
            pending: t('PendingPeers'),
            halfOpen: t('HalfOpenPeers'),
          }}
        />
      </div>
    </>
  )

  if (!framed) {
    return <div className={`flex min-h-0 flex-col ${className}`.trim() || undefined}>{inner}</div>
  }

  return (
    <div
      className={`flex min-h-0 flex-col rounded-xl border border-border bg-surface-secondary p-2.5 ${className}`.trim()}
    >
      {inner}
    </div>
  )
}

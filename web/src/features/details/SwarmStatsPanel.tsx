import { humanizeSize } from 'shared/lib/format'
import type { TorrentStat } from 'shared/api/types'
import { useTranslation } from 'react-i18next'

import MetricRows, { type MetricRowItem } from './MetricRows'

export interface SwarmStatsPanelProps {
  torrent: TorrentStat
  className?: string
  columns?: 1 | 2
  /** When false, omit outer frame (parent already provides a panel). */
  framed?: boolean
  /** Show uppercase Swarm title inside framed summary (desktop). Hide on mobile Stats. */
  showTitle?: boolean
  /**
   * Match SpeedCharts height on desktop Stats (min-h + mt-auto meters).
   * Never enable on mobile stack — causes an empty void.
   */
  stretch?: boolean
  /**
   * `summary` — Stats side card: transfer IO + Loaded|Preload (hero owns Peers/Cache).
   * `full` — Swarm tab: PeerMixBar + chunks + Loaded|Preload (no Peers·Seeds / Cache echo).
   */
  variant?: 'summary' | 'full'
  cacheReaders?: number | null
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

function pct(part: number, whole: number): number {
  if (whole <= 0) return 0
  return Math.min(100, Math.max(0, (part / whole) * 100))
}

function ProgressMeter({
  label,
  valueLabel,
  ratio,
  compact = false,
}: {
  label: string
  valueLabel: string
  ratio: number
  compact?: boolean
}) {
  const width = Number.isFinite(ratio) ? Math.min(100, Math.max(0, ratio)) : 0
  return (
    <div className='min-w-0'>
      <div className={`flex items-baseline justify-between gap-2 text-xs ${compact ? 'mb-0.5' : 'mb-1'}`}>
        <span className='truncate text-muted'>{label}</span>
        <span className='shrink-0 font-bold tabular-nums text-foreground'>{valueLabel}</span>
      </div>
      <div className={`overflow-hidden rounded-full bg-surface ${compact ? 'h-1.5' : 'h-2'}`}>
        <div className='h-full rounded-full bg-accent transition-[width] duration-300' style={{ width: `${width}%` }} />
      </div>
    </div>
  )
}

function PeerMixBar({
  active,
  seeders,
  pending,
  halfOpen,
  labels,
  compact = false,
}: {
  active: number
  seeders: number
  pending: number
  halfOpen: number
  labels: { active: string; seeders: string; pending: string; halfOpen: string }
  compact?: boolean
}) {
  const parts = [
    { key: 'active', n: active, className: 'bg-accent', label: labels.active },
    { key: 'seeders', n: seeders, className: 'bg-accent/70', label: labels.seeders },
    { key: 'pending', n: pending, className: 'bg-warning', label: labels.pending },
    { key: 'half', n: halfOpen, className: 'bg-foreground/25', label: labels.halfOpen },
  ]
  const total = parts.reduce((sum, p) => sum + Math.max(0, p.n), 0)

  return (
    <div className='min-w-0'>
      <div className={`flex flex-wrap gap-x-3 gap-y-1 text-[10px] text-muted ${compact ? 'mb-1' : 'mb-1.5'}`}>
        {parts.map(p => (
          <span key={p.key} className='inline-flex items-center gap-1'>
            <span className={`size-1.5 rounded-full ${p.className}`} aria-hidden />
            {p.label} <span className='font-semibold tabular-nums text-foreground'>{p.n}</span>
          </span>
        ))}
      </div>
      <div className={`flex overflow-hidden rounded-full bg-surface ${compact ? 'h-2' : 'h-2.5'}`}>
        {total > 0
          ? parts.map(p =>
              p.n > 0 ? (
                <div
                  key={p.key}
                  className={`h-full ${p.className}`}
                  style={{ width: `${(p.n / total) * 100}%` }}
                  title={`${p.label}: ${p.n}`}
                />
              ) : null,
            )
          : null}
      </div>
    </div>
  )
}

function LoadedPreloadMeters({
  loadedLabel,
  loadedPct,
  preloadLabel,
  preloadPct,
  loadedTitle,
  preloadTitle,
  compact = false,
}: {
  loadedLabel: string
  loadedPct: number
  preloadLabel: string
  preloadPct: number
  loadedTitle: string
  preloadTitle: string
  compact?: boolean
}) {
  return (
    <div className={`grid grid-cols-2 ${compact ? 'gap-1.5' : 'gap-2'}`}>
      <ProgressMeter label={loadedTitle} valueLabel={loadedLabel} ratio={loadedPct} compact={compact} />
      <ProgressMeter label={preloadTitle} valueLabel={preloadLabel} ratio={preloadPct} compact={compact} />
    </div>
  )
}

/** Swarm metrics — Stats transfer teaser or Swarm-tab peer detail (hero owns Peers/Cache). */
export default function SwarmStatsPanel({
  torrent,
  className = '',
  columns: columnsProp,
  framed = true,
  showTitle = true,
  stretch = false,
  variant = 'summary',
  cacheReaders,
}: SwarmStatsPanelProps) {
  const { t } = useTranslation()
  const isFull = variant === 'full'
  const columns = columnsProp ?? 2
  const doStretch = stretch && !isFull

  const pendingValue = torrent.pending_peers != null ? String(torrent.pending_peers) : '—'

  const loaded = torrent.loaded_size ?? 0
  const totalSize = torrent.torrent_size ?? 0
  const loadedPct = pct(loaded, totalSize)
  const preloadDone = torrent.preloaded_bytes ?? 0
  const preloadNeed = torrent.preload_size ?? 0
  const preloadPct = preloadNeed > 0 ? pct(preloadDone, preloadNeed) : 0
  const durationLabel = formatDuration(torrent.duration_seconds)
  const loadedLabel = totalSize > 0 ? `${humanizeSize(loaded)} · ${Math.round(loadedPct)}%` : humanizeSize(loaded)
  const preloadLabel =
    preloadNeed > 0 || preloadDone > 0
      ? `${humanizeSize(preloadDone)} / ${humanizeSize(preloadNeed || undefined)}`
      : '—'

  /** Stats: Half-open + transfer IO. Preload is meter-only; Peers/Cache stay in hero. */
  const summaryItems: MetricRowItem[] = [
    { label: t('HalfOpenPeers'), value: torrent.half_open_peers != null ? String(torrent.half_open_peers) : '—' },
    { label: t('BytesRead'), value: torrent.bytes_read != null ? humanizeSize(torrent.bytes_read) : '—' },
    { label: t('BytesWritten'), value: torrent.bytes_written != null ? humanizeSize(torrent.bytes_written) : '—' },
    {
      label: t('UsefulRead'),
      value: torrent.bytes_read_useful_data != null ? humanizeSize(torrent.bytes_read_useful_data) : '—',
    },
    {
      label: t('ChunksWasted'),
      value: torrent.chunks_read_wasted != null ? String(torrent.chunks_read_wasted) : '—',
    },
  ]

  /**
   * Swarm tab: PeerMixBar owns Active/Seeders/Pending/Half-open.
   * Skip Total peers (hero) and Half-open row (mix bar).
   */
  const fullItems: MetricRowItem[] = [
    { label: t('BytesRead'), value: torrent.bytes_read != null ? humanizeSize(torrent.bytes_read) : '—' },
    { label: t('BytesWritten'), value: torrent.bytes_written != null ? humanizeSize(torrent.bytes_written) : '—' },
    {
      label: t('UsefulRead'),
      value: torrent.bytes_read_useful_data != null ? humanizeSize(torrent.bytes_read_useful_data) : '—',
    },
    {
      label: t('BytesReadData'),
      value: torrent.bytes_read_data != null ? humanizeSize(torrent.bytes_read_data) : '—',
    },
    { label: t('ChunksRead'), value: torrent.chunks_read != null ? String(torrent.chunks_read) : '—' },
    { label: t('ChunksWritten'), value: torrent.chunks_written != null ? String(torrent.chunks_written) : '—' },
    {
      label: t('ChunksWasted'),
      value: torrent.chunks_read_wasted != null ? String(torrent.chunks_read_wasted) : '—',
    },
    {
      label: t('ChunksUseful'),
      value: torrent.chunks_read_useful != null ? String(torrent.chunks_read_useful) : '—',
    },
    ...(cacheReaders != null ? [{ label: t('CacheReaders'), value: String(cacheReaders) }] : []),
    ...(torrent.bit_rate ? [{ label: t('BitRate'), value: torrent.bit_rate }] : []),
    ...(durationLabel ? [{ label: t('FfpDuration'), value: durationLabel }] : []),
  ]

  const pendingHeader = isFull ? (
    <div className='mb-1.5 flex shrink-0 items-center gap-2'>
      <span className='size-2.5 rounded-full bg-warning' aria-hidden />
      <span className='text-xs text-muted'>{t('PendingPeers')}</span>
      <span className='text-sm font-bold tabular-nums text-foreground'>{pendingValue}</span>
    </div>
  ) : null

  const progressMeters = (
    <LoadedPreloadMeters
      loadedTitle={t('ServerStatusLoaded')}
      loadedLabel={loadedLabel}
      loadedPct={loadedPct}
      preloadTitle={t('Preloaded')}
      preloadLabel={preloadLabel}
      preloadPct={preloadPct}
      compact={isFull}
    />
  )

  /** Desktop stretch: pin meters to bottom of chart-height card. Mobile: flow under rows. */
  const summaryVisuals = !isFull ? (
    <div className={`border-t border-border pt-2.5 ${doStretch ? 'mt-auto' : 'mt-2'}`}>{progressMeters}</div>
  ) : null

  const fullVisuals = isFull ? (
    <div className='mt-2 space-y-1.5 border-t border-border pt-2'>
      <PeerMixBar
        active={torrent.active_peers ?? 0}
        seeders={torrent.connected_seeders ?? 0}
        pending={torrent.pending_peers ?? 0}
        halfOpen={torrent.half_open_peers ?? 0}
        labels={{
          active: t('ActivePeers'),
          seeders: t('ConnectedSeeders'),
          pending: t('PendingPeers'),
          halfOpen: t('HalfOpenPeers'),
        }}
        compact
      />
      {progressMeters}
    </div>
  ) : null

  const inner = (
    <>
      {framed && !isFull && showTitle ? (
        <p className='mb-1.5 shrink-0 text-xs font-semibold tracking-wide text-muted uppercase'>{t('SwarmStats')}</p>
      ) : null}
      {pendingHeader}
      <MetricRows framed={false} items={isFull ? fullItems : summaryItems} columns={columns} dense={isFull} />
      {summaryVisuals}
      {fullVisuals}
    </>
  )

  if (!framed) {
    return <div className={className || undefined}>{inner}</div>
  }

  return (
    <div
      className={`flex flex-col rounded-xl border border-border bg-surface-secondary p-2.5 ${
        doStretch ? 'h-full min-h-[14rem]' : ''
      } ${className}`.trim()}
    >
      {inner}
    </div>
  )
}

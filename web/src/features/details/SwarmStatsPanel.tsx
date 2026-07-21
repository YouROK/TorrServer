import { humanizeSize } from 'shared/lib/format'
import type { TorrentStat } from 'shared/api/types'
import { useTranslation } from 'react-i18next'

export interface SwarmStatsPanelProps {
  torrent: TorrentStat
}

/** Compact power-user transfer / peer counters already present on TorrentStat. */
export default function SwarmStatsPanel({ torrent }: SwarmStatsPanelProps) {
  const { t } = useTranslation()

  const rows: { label: string; value: string }[] = [
    { label: t('PendingPeers'), value: torrent.pending_peers != null ? String(torrent.pending_peers) : '—' },
    { label: t('HalfOpenPeers'), value: torrent.half_open_peers != null ? String(torrent.half_open_peers) : '—' },
    {
      label: t('Preloaded'),
      value:
        torrent.preloaded_bytes != null || torrent.preload_size != null
          ? `${humanizeSize(torrent.preloaded_bytes)} / ${humanizeSize(torrent.preload_size)}`
          : '—',
    },
    { label: t('BytesRead'), value: torrent.bytes_read != null ? humanizeSize(torrent.bytes_read) : '—' },
    { label: t('BytesWritten'), value: torrent.bytes_written != null ? humanizeSize(torrent.bytes_written) : '—' },
  ]

  return (
    <div className='rounded-xl border border-border bg-surface-secondary p-2.5'>
      <p className='mb-2 text-xs font-semibold tracking-wide text-muted uppercase'>{t('SwarmStats')}</p>
      <div className='grid grid-cols-2 gap-1.5 sm:grid-cols-3 lg:grid-cols-5'>
        {rows.map(row => (
          <div key={row.label} className='min-w-0 rounded-lg border border-border bg-surface px-2 py-1.5 text-center'>
            <span className='block truncate text-[10px] text-muted' title={row.label}>
              {row.label}
            </span>
            <span className='mt-0.5 block truncate text-xs font-bold tabular-nums text-foreground' title={row.value}>
              {row.value}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

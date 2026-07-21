import { useEffect, useId, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { humanizeSpeed } from 'shared/lib/format'

const HISTORY_LENGTH = 30
const CHART_WIDTH = 320
const CHART_HEIGHT = 96

interface SpeedChartsProps {
  downloadSpeed?: number | null
  uploadSpeed?: number | null
  /** Hide speed legend + shorter chart — used when hero already shows live speeds. */
  compact?: boolean
}

function buildAreaPath(values: number[], max: number): { line: string; area: string } {
  if (values.length === 0) return { line: '', area: '' }
  const step = values.length > 1 ? CHART_WIDTH / (values.length - 1) : 0
  const points = values.map((value, index) => {
    const x = index * step
    const y = CHART_HEIGHT - (Math.max(0, value) / max) * CHART_HEIGHT
    return [x, y] as const
  })
  const line = points.map(([x, y]) => `${x},${y}`).join(' ')
  const area = `0,${CHART_HEIGHT} ${line} ${CHART_WIDTH},${CHART_HEIGHT}`
  return { line, area }
}

/** Live download/upload sparkline for the torrent overview tab. */
export default function SpeedCharts({ downloadSpeed, uploadSpeed, compact = false }: SpeedChartsProps) {
  const { t } = useTranslation()
  const [downloadHistory, setDownloadHistory] = useState<number[]>(() => Array(HISTORY_LENGTH).fill(0))
  const [uploadHistory, setUploadHistory] = useState<number[]>(() => Array(HISTORY_LENGTH).fill(0))
  const gradientId = useId()

  useEffect(() => {
    const dl = Math.max(0, downloadSpeed ?? 0)
    const ul = Math.max(0, uploadSpeed ?? 0)
    // eslint-disable-next-line react-hooks/set-state-in-effect -- append speed samples into rolling history
    setDownloadHistory(prev => [...prev.slice(1), dl])
    setUploadHistory(prev => [...prev.slice(1), ul])
  }, [downloadSpeed, uploadSpeed])

  const peak = useMemo(() => Math.max(...downloadHistory, ...uploadHistory, 1), [downloadHistory, uploadHistory])
  const downloadPath = useMemo(() => buildAreaPath(downloadHistory, peak), [downloadHistory, peak])
  const uploadPath = useMemo(() => buildAreaPath(uploadHistory, peak), [uploadHistory, peak])

  return (
    <div className={`w-full min-w-0 rounded-xl border border-border bg-surface-secondary ${compact ? 'p-2.5' : 'p-4'}`}>
      {compact ? null : (
        <div className='mb-3 flex flex-wrap items-center gap-4'>
          <div className='flex items-center gap-2'>
            <span className='size-2.5 rounded-full bg-accent' aria-hidden />
            <span className='text-xs text-muted'>{t('DownloadSpeed')}</span>
            <span className='text-sm font-bold tabular-nums text-foreground'>{humanizeSpeed(downloadSpeed)}</span>
          </div>
          <div className='flex items-center gap-2'>
            <span className='size-2.5 rounded-full bg-warning' aria-hidden />
            <span className='text-xs text-muted'>{t('UploadSpeed')}</span>
            <span className='text-sm font-bold tabular-nums text-foreground'>{humanizeSpeed(uploadSpeed)}</span>
          </div>
        </div>
      )}

      <svg
        viewBox={`0 0 ${CHART_WIDTH} ${CHART_HEIGHT}`}
        className={`w-full ${compact ? 'h-16' : 'h-[110px]'}`}
        preserveAspectRatio='none'
        aria-hidden
      >
        <defs>
          <linearGradient id={gradientId} x1='0' y1='0' x2='0' y2='1'>
            <stop offset='0%' stopColor='currentColor' stopOpacity='0.35' className='text-accent' />
            <stop offset='100%' stopColor='currentColor' stopOpacity='0' className='text-accent' />
          </linearGradient>
        </defs>
        <polygon points={downloadPath.area} fill={`url(#${gradientId})`} />
        <polyline points={downloadPath.line} className='fill-none stroke-accent' strokeWidth='2' />
        <polyline
          points={uploadPath.line}
          className='fill-none stroke-warning'
          strokeWidth='1.5'
          strokeDasharray='4 3'
        />
      </svg>
    </div>
  )
}

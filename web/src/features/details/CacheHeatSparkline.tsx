import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

const HISTORY = 24

export interface CacheHeatSparklineProps {
  filled?: number | null
  capacity?: number | null
}

/** Session ring-buffer of cache fill ratio — tiny sparkline under Overview charts. */
export default function CacheHeatSparkline({ filled, capacity }: CacheHeatSparklineProps) {
  const { t } = useTranslation()
  const [history, setHistory] = useState<number[]>(() => Array(HISTORY).fill(0))

  const ratio = capacity && capacity > 0 ? Math.min(1, Math.max(0, (filled ?? 0) / capacity)) : 0

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- rolling sample of live cache fill
    setHistory(prev => [...prev.slice(1), ratio])
  }, [ratio])

  const path = useMemo(() => {
    const w = 120
    const h = 28
    const step = w / (HISTORY - 1)
    return history.map((v, i) => `${i === 0 ? 'M' : 'L'} ${i * step},${h - v * h}`).join(' ')
  }, [history])

  return (
    <div className='rounded-xl border border-border bg-surface-secondary p-2.5'>
      <div className='mb-1.5 flex items-center justify-between gap-2'>
        <p className='text-xs font-semibold tracking-wide text-muted uppercase'>{t('CacheHeat')}</p>
        <span className='text-xs tabular-nums text-foreground'>{Math.round(ratio * 100)}%</span>
      </div>
      <svg viewBox='0 0 120 28' className='h-8 w-full text-accent' preserveAspectRatio='none' aria-hidden>
        <path d={path} fill='none' stroke='currentColor' strokeWidth='2' />
      </svg>
    </div>
  )
}

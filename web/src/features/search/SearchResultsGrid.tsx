import { Button, Spinner, useMediaQuery } from '@heroui/react'
import { Users } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import type { SearchResultItem } from 'shared/api/types'
import { queryMax } from 'shared/theme/breakpoints'

interface SearchResultsGridProps {
  results: SearchResultItem[]
  adding: boolean
  addingKey: string | null
  resultDedupeKey: (item: SearchResultItem) => string
  formatSize: (item: SearchResultItem) => string
  onAdd: (item: SearchResultItem) => void
}

/**
 * Dense list/table of Torznab/Rutor results — no Film/poster placeholders (covers aren't
 * available until a torrent is added). Fixed table layout avoids horizontal scroll.
 */
export default function SearchResultsGrid({
  results,
  adding,
  addingKey,
  resultDedupeKey,
  formatSize,
  onAdd,
}: SearchResultsGridProps) {
  const { t } = useTranslation()
  const isCompact = useMediaQuery(queryMax('mobile'))

  if (isCompact) {
    return (
      <div className='space-y-2'>
        {results.map((item, index) => {
          const key = resultDedupeKey(item) || `${item.Title || 'item'}-${index}`
          const busy = adding && addingKey === key

          return (
            <div
              key={key}
              className={`flex gap-3 rounded-xl border border-border bg-surface p-3 ${busy ? 'border-accent/40 bg-accent-soft/30' : ''}`}
            >
              <div className='min-w-0 flex-1'>
                <div className='mb-1 flex flex-wrap items-center gap-1.5'>
                  {item.Tracker ? (
                    <span className='truncate rounded-full bg-surface-tertiary px-2 py-0.5 text-[10px] font-medium text-muted'>
                      {item.Tracker}
                    </span>
                  ) : null}
                </div>
                <p className='mb-1.5 break-words text-sm font-medium leading-snug text-foreground' title={item.Title}>
                  {item.Title || '—'}
                </p>
                <div className='flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted'>
                  <span>{formatSize(item)}</span>
                  {item.Seed != null || item.Peer != null ? (
                    <span className='inline-flex items-center gap-1'>
                      <Users className='size-3.5' strokeWidth={1.75} aria-hidden />
                      {item.Seed ?? 0}·{item.Peer ?? 0}
                    </span>
                  ) : null}
                </div>
              </div>
              <Button
                size='sm'
                variant='primary'
                isDisabled={adding}
                onPress={() => onAdd(item)}
                className='min-h-11 shrink-0 self-center px-3'
              >
                {busy ? <Spinner size='sm' color='current' /> : t('Add')}
              </Button>
            </div>
          )
        })}
      </div>
    )
  }

  return (
    <div className={`w-full overflow-x-hidden rounded-xl border border-border ${adding ? 'opacity-90' : ''}`}>
      <table className='w-full table-fixed border-collapse text-sm'>
        <thead className='bg-surface-tertiary text-left text-xs uppercase tracking-wide text-muted'>
          <tr>
            <th className='px-3 py-2.5'>{t('Name')}</th>
            <th className='w-28 px-3 py-2.5'>{t('Size')}</th>
            <th className='w-24 px-3 py-2.5'>{t('Search.Seeders')}</th>
            <th className='w-24 px-3 py-2.5'>{t('Search.Peers')}</th>
            <th className='w-28 px-3 py-2.5' />
          </tr>
        </thead>
        <tbody>
          {results.map((item, index) => {
            const key = resultDedupeKey(item) || `${item.Title || 'item'}-${index}`
            const busy = adding && addingKey === key

            return (
              <tr
                key={key}
                className={`border-t border-border ${busy ? 'bg-accent-soft/40' : 'hover-fine:bg-surface-secondary/80'}`}
              >
                <td className='px-3 py-2.5 align-top'>
                  <div className='flex flex-col gap-1'>
                    {item.Tracker ? (
                      <span className='w-fit max-w-full truncate rounded-full bg-surface-tertiary px-2 py-0.5 text-[10px] font-medium text-muted'>
                        {item.Tracker}
                      </span>
                    ) : null}
                    <span className='line-clamp-3 break-words leading-snug text-foreground' title={item.Title}>
                      {item.Title || '—'}
                    </span>
                  </div>
                </td>
                <td className='px-3 py-2.5 whitespace-nowrap text-muted'>{formatSize(item)}</td>
                <td className='px-3 py-2.5 tabular-nums text-muted'>{item.Seed ?? '—'}</td>
                <td className='px-3 py-2.5 tabular-nums text-muted'>{item.Peer ?? '—'}</td>
                <td className='px-3 py-2.5'>
                  <Button
                    size='sm'
                    variant='primary'
                    isDisabled={adding}
                    onPress={() => onAdd(item)}
                    className='min-h-10'
                  >
                    {busy ? <Spinner size='sm' color='current' /> : t('Add')}
                  </Button>
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}

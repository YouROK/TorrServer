import { Check } from 'lucide-react'

export interface PosterPickerProps {
  poster: string
  posterOptions: string[]
  onSelect: (url: string) => void
  disabled?: boolean
  /** scroll — horizontal strip (Add); wrap — grid (Edit). */
  layout?: 'scroll' | 'wrap'
}

/** TMDB poster thumbnail strip shared by Add and Edit torrent dialogs. */
export default function PosterPicker({
  poster,
  posterOptions,
  onSelect,
  disabled,
  layout = 'scroll',
}: PosterPickerProps) {
  if (!posterOptions.length) return null

  const isScroll = layout === 'scroll'

  return (
    <div className={isScroll ? 'mb-3 flex gap-2 overflow-x-auto pb-1' : 'flex flex-wrap gap-2'}>
      {posterOptions.map(url => {
        const selected = poster === url
        return (
          <button
            key={url}
            type='button'
            onClick={() => onSelect(url)}
            aria-pressed={selected}
            disabled={disabled}
            className={`relative shrink-0 overflow-hidden rounded-lg border-2 transition-colors ${
              isScroll ? 'h-24 w-16' : 'h-[132px] w-[88px]'
            } ${selected ? 'border-accent' : 'border-border hover:border-accent/50'}`}
          >
            <img src={url} alt='' className='h-full w-full object-cover' />
            {selected ? (
              <span className='absolute right-1 top-1 grid size-4 place-items-center rounded-full bg-accent text-accent-foreground'>
                <Check size={12} strokeWidth={1.75} aria-hidden />
              </span>
            ) : null}
          </button>
        )
      })}
    </div>
  )
}

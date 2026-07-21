import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { NOW_PLAYING_EVENT, type NowPlayingInfo } from 'shared/lib/nowPlaying'

/** Thin ambient strip when the in-app player is open (desktop chrome). */
export default function NowPlayingBar() {
  const { t } = useTranslation()
  const [info, setInfo] = useState<NowPlayingInfo | null>(null)

  useEffect(() => {
    const onPlay = (event: Event) => {
      setInfo((event as CustomEvent<NowPlayingInfo | null>).detail)
    }
    window.addEventListener(NOW_PLAYING_EVENT, onPlay as EventListener)
    return () => window.removeEventListener(NOW_PLAYING_EVENT, onPlay as EventListener)
  }, [])

  if (!info) return null

  return (
    <div className='pointer-events-none fixed inset-x-0 bottom-0 z-[60] flex justify-center p-3 pb-[calc(0.75rem+env(safe-area-inset-bottom,0px))] sm:pb-3'>
      <div className='pointer-events-auto max-w-lg truncate rounded-full border border-border bg-surface/95 px-4 py-2 text-sm shadow-lg backdrop-blur'>
        <span className='mr-2 text-xs font-semibold tracking-wide text-muted uppercase'>{t('NowPlaying')}</span>
        <span className='font-medium text-foreground'>{info.title}</span>
      </div>
    </div>
  )
}

import { Button, Modal, Spinner } from '@heroui/react'
import { AudioLines, Captions, Clapperboard, FileVideo, Layers } from 'lucide-react'
import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'

import { fetchFfp, type FfpProbeResult, type FfpStream } from 'shared/api/extras'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import AppDialog from 'shared/ui/AppDialog'
import { DIALOG_SHEET_M } from 'shared/ui/dialogSizes'

import {
  formatFfpBitrate,
  formatFfpBytes,
  formatFfpDuration,
  fpsFromRate,
  groupStreams,
  streamTag,
} from './mediaInfoFormat'

export interface MediaInfoDialogProps {
  open: boolean
  onClose: () => void
  hash: string
  fileId: number
  fileName: string
}

function Spec({ label, value }: { label: string; value?: string | number | null }) {
  if (value == null || value === '') return null
  return (
    <div className='min-w-0 rounded-lg bg-surface-secondary/80 px-2.5 py-1.5'>
      <p className='text-[10px] font-semibold uppercase tracking-wide text-muted'>{label}</p>
      <p className='mt-0.5 truncate text-sm font-medium tabular-nums text-foreground' title={String(value)}>
        {value}
      </p>
    </div>
  )
}

function StreamCard({
  index,
  stream,
  kindLabel,
  icon,
}: {
  index: number
  stream: FfpStream
  kindLabel: string
  icon: ReactNode
}) {
  const { t } = useTranslation()
  const lang = streamTag(stream, 'language')
  const title = streamTag(stream, 'title')
  const codec = stream.codec_long_name || stream.codec_name || '—'
  const profile = stream.profile

  const specs: Array<{ label: string; value?: string | number | null }> = []

  if (stream.codec_type === 'video') {
    if (stream.width && stream.height) {
      const ar =
        stream.display_aspect_ratio && stream.display_aspect_ratio !== '0:0' ? ` (${stream.display_aspect_ratio})` : ''
      specs.push({ label: t('Resolution'), value: `${stream.width}×${stream.height}${ar}` })
    }
    if (stream.pix_fmt) specs.push({ label: t('FfpPixel'), value: stream.pix_fmt })
    const fps = fpsFromRate(stream.r_frame_rate)
    if (fps) specs.push({ label: t('FfpFps'), value: fps })
    const br = formatFfpBitrate(stream.bit_rate)
    if (br) specs.push({ label: t('FfpBitrate'), value: br })
    if (stream.color_space || stream.color_transfer || stream.color_primaries) {
      specs.push({
        label: t('FfpColor'),
        value: [stream.color_space, stream.color_transfer, stream.color_primaries].filter(Boolean).join(' / '),
      })
    }
  } else if (stream.codec_type === 'audio') {
    if (stream.sample_rate) specs.push({ label: t('FfpSampleRate'), value: `${stream.sample_rate} Hz` })
    if (stream.channels || stream.channel_layout) {
      specs.push({
        label: t('FfpChannels'),
        value: stream.channel_layout || `${stream.channels} ch`,
      })
    }
    const br = formatFfpBitrate(stream.bit_rate)
    if (br) specs.push({ label: t('FfpBitrate'), value: br })
  } else if (stream.codec_type === 'subtitle') {
    if (stream.codec_name) specs.push({ label: t('FfpCodec'), value: stream.codec_name })
  }

  return (
    <article className='rounded-xl border border-border bg-surface p-3 shadow-sm'>
      <div className='mb-2 flex flex-wrap items-center gap-2'>
        <span className='grid size-8 shrink-0 place-items-center rounded-lg bg-accent-soft text-accent'>{icon}</span>
        <div className='min-w-0 flex-1'>
          <p className='text-sm font-semibold text-foreground'>
            #{index + 1} {kindLabel}
            {lang ? <span className='ml-1.5 text-xs font-medium uppercase text-muted'>{lang}</span> : null}
          </p>
          <p className='truncate text-xs text-muted' title={codec}>
            {codec}
            {profile ? ` · ${profile}` : ''}
          </p>
        </div>
      </div>
      {title ? <p className='mb-2 truncate text-sm text-foreground/90'>{title}</p> : null}
      {specs.length > 0 ? (
        <div className='grid grid-cols-2 gap-1.5 sm:grid-cols-3'>
          {specs.map(item => (
            <Spec key={item.label} label={item.label} value={item.value} />
          ))}
        </div>
      ) : null}
    </article>
  )
}

/** Nested sheet with full ffprobe streams — replaces the one-line toast. */
export default function MediaInfoDialog({ open, onClose, hash, fileId, fileName }: MediaInfoDialogProps) {
  const { t } = useTranslation()
  const isFullScreenBreakpoint = useDialogFullScreen()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [data, setData] = useState<FfpProbeResult | null>(null)

  useEffect(() => {
    if (!open) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- reset MediaInfo when dialog closes
      setData(null)
      setError(null)
      setLoading(false)
      return
    }
    const ac = new AbortController()
    setLoading(true)
    setError(null)
    setData(null)
    void fetchFfp(hash, fileId, ac.signal)
      .then(result => {
        if (!ac.signal.aborted) setData(result)
      })
      .catch(err => {
        if ((err as Error).name === 'CanceledError' || (err as Error).name === 'AbortError') return
        if (!ac.signal.aborted) setError((err as Error).message || t('Error'))
      })
      .finally(() => {
        if (!ac.signal.aborted) setLoading(false)
      })
    return () => ac.abort()
  }, [open, hash, fileId, t])

  const groups = useMemo(() => groupStreams(data), [data])
  const duration = formatFfpDuration(data?.format?.duration)
  const size = formatFfpBytes(data?.format?.size)
  const bitrate = formatFfpBitrate(data?.format?.bit_rate)
  const container = data?.format?.format_long_name || data?.format?.format_name

  return (
    <AppDialog
      open={open}
      onClose={onClose}
      size='md'
      fullScreen={isFullScreenBreakpoint}
      dialogClassName='flex flex-col overflow-hidden'
      dialogStyle={isFullScreenBreakpoint ? undefined : DIALOG_SHEET_M}
    >
      <Modal.Header className='shrink-0'>
        <Modal.Heading className='min-w-0 truncate'>{t('MediaInfo')}</Modal.Heading>
        <Modal.CloseTrigger aria-label={t('Close')} />
      </Modal.Header>
      <Modal.Body className='min-h-0 flex-1 gap-4 overflow-y-auto overscroll-contain'>
        <div className='rounded-xl border border-border bg-gradient-to-br from-accent-soft/50 to-surface-secondary p-3'>
          <p className='mb-2 flex items-center gap-2 text-sm font-semibold text-foreground'>
            <FileVideo className='text-accent' size={16} strokeWidth={1.75} aria-hidden />
            <span className='min-w-0 truncate' title={fileName}>
              {fileName}
            </span>
          </p>
          <div className='grid grid-cols-2 gap-1.5 sm:grid-cols-4'>
            <Spec label={t('FfpContainer')} value={container} />
            <Spec label={t('FfpDuration')} value={duration} />
            <Spec label={t('Size')} value={size} />
            <Spec label={t('FfpBitrate')} value={bitrate} />
          </div>
        </div>

        {loading ? (
          <div className='grid place-items-center py-12'>
            <Spinner size='lg' />
          </div>
        ) : null}

        {error ? <p className='rounded-lg bg-danger/10 px-3 py-2 text-sm text-danger'>{error}</p> : null}

        {!loading && !error && data ? (
          <div className='space-y-4'>
            {groups.video.length > 0 ? (
              <section className='space-y-2'>
                <h3 className='flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wide text-muted'>
                  <Clapperboard size={14} strokeWidth={1.75} aria-hidden />
                  {t('FfpVideo')} ({groups.video.length})
                </h3>
                {groups.video.map((stream, i) => (
                  <StreamCard
                    key={`v-${i}`}
                    index={i}
                    stream={stream}
                    kindLabel={t('FfpVideo')}
                    icon={<Clapperboard size={16} strokeWidth={1.75} aria-hidden />}
                  />
                ))}
              </section>
            ) : null}

            {groups.audio.length > 0 ? (
              <section className='space-y-2'>
                <h3 className='flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wide text-muted'>
                  <AudioLines size={14} strokeWidth={1.75} aria-hidden />
                  {t('FfpAudio')} ({groups.audio.length})
                </h3>
                {groups.audio.map((stream, i) => (
                  <StreamCard
                    key={`a-${i}`}
                    index={i}
                    stream={stream}
                    kindLabel={t('FfpAudio')}
                    icon={<AudioLines size={16} strokeWidth={1.75} aria-hidden />}
                  />
                ))}
              </section>
            ) : null}

            {groups.subtitle.length > 0 ? (
              <section className='space-y-2'>
                <h3 className='flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wide text-muted'>
                  <Captions size={14} strokeWidth={1.75} aria-hidden />
                  {t('FfpSubtitle')} ({groups.subtitle.length})
                </h3>
                {groups.subtitle.map((stream, i) => (
                  <StreamCard
                    key={`s-${i}`}
                    index={i}
                    stream={stream}
                    kindLabel={t('FfpSubtitle')}
                    icon={<Captions size={16} strokeWidth={1.75} aria-hidden />}
                  />
                ))}
              </section>
            ) : null}

            {groups.other.length > 0 ? (
              <section className='space-y-2'>
                <h3 className='flex items-center gap-1.5 text-xs font-semibold uppercase tracking-wide text-muted'>
                  <Layers size={14} strokeWidth={1.75} aria-hidden />
                  {t('FfpOther')} ({groups.other.length})
                </h3>
                {groups.other.map((stream, i) => (
                  <StreamCard
                    key={`o-${i}`}
                    index={i}
                    stream={stream}
                    kindLabel={stream.codec_type || t('FfpOther')}
                    icon={<Layers size={16} strokeWidth={1.75} aria-hidden />}
                  />
                ))}
              </section>
            ) : null}

            {!groups.video.length && !groups.audio.length && !groups.subtitle.length && !groups.other.length ? (
              <p className='py-6 text-center text-sm text-muted'>{t('NoData')}</p>
            ) : null}
          </div>
        ) : null}
      </Modal.Body>
      <Modal.Footer className='shrink-0'>
        <Button variant='secondary' onPress={onClose} autoFocus>
          {t('Close')}
        </Button>
      </Modal.Footer>
    </AppDialog>
  )
}

import { useTranslation } from 'react-i18next'
import { humanizeSize, humanizeTime } from 'utils/Utils'

import { ReadoutField, ReadoutNote, ReadoutTitle, ReadoutValue, ReadoutWrapper } from './style'

// The reader someone is watching: the server flags those that have streamed long enough and
// shown something, and among them the furthest into its session wins. Preloads and probes
// open readers too, but they are short-lived and never flagged.
export const activePlayback = playback =>
  (playback || []).reduce((best, entry) => {
    if (!best) return entry
    if (!!entry.viewing !== !!best.viewing) return entry.viewing ? entry : best
    return (entry.session_seconds || 0) > (best.session_seconds || 0) ? entry : best
  }, null)

// What the position is worth knowing by: where the picture is, and how much was assumed to
// sit in the player ahead of it. Without a duration only the offset is meaningful.
export const playbackPosition = entry =>
  entry?.duration > 0 ? humanizeTime(entry.timecode) : humanizeSize(entry?.position)

// Whether the buffer was measured from the stream or taken from the settings.
export const bufferMark = (entry, t) => (entry.buffer_measured ? t('BufferMeasuredMark') : t('BufferFallbackMark'))

export default function PlaybackReadout({ playback }) {
  const { t } = useTranslation()
  const entry = activePlayback(playback)
  if (!entry) return null

  // The buffer is measured in film time, and the size in bytes is that same answer converted
  // back through the file's own byte positions — so the two always agree. Seconds are the
  // more useful of the pair: they are how far the mark sits behind the picture.
  const bufferNote = entry.buffer_seconds > 0 && humanizeTime(entry.buffer_seconds)

  return (
    <ReadoutWrapper>
      <ReadoutField>
        <ReadoutTitle>{t('OnScreen')}</ReadoutTitle>
        <ReadoutValue>{playbackPosition(entry)}</ReadoutValue>
        <ReadoutNote>
          {/* Where the time came from: the container's own timestamps, or the average
              bitrate when the file carries none. */}
          {entry.source && `${entry.source === 'estimate' ? t('TimeEstimated') : entry.source} · `}
          {humanizeSize(entry.position)}
        </ReadoutNote>
      </ReadoutField>

      <ReadoutField>
        <ReadoutTitle>{t('ClientBuffer')}</ReadoutTitle>
        <ReadoutValue>{humanizeSize(entry.buffer)}</ReadoutValue>
        <ReadoutNote>
          {bufferMark(entry, t)}
          {bufferNote && ` · ${bufferNote}`}
        </ReadoutNote>
      </ReadoutField>

      <ReadoutField>
        <ReadoutTitle>{t('ReadHead')}</ReadoutTitle>
        <ReadoutValue>{humanizeSize(entry.head)}</ReadoutValue>
        <ReadoutNote>{humanizeTime(entry.session_seconds)}</ReadoutNote>
      </ReadoutField>
    </ReadoutWrapper>
  )
}

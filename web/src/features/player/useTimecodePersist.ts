import { useCallback, useEffect, useRef, type RefObject } from 'react'
import { setViewedFile } from 'shared/api/viewed'
import { rememberContinueWatching } from 'shared/lib/continueWatching'

const TIMECODE_SAVE_INTERVAL_MS = 5_000
const TIMECODE_RESUME_MARGIN_SEC = 5

export interface UseTimecodePersistOptions {
  enabled: boolean
  hash?: string
  fileIndex?: number
  title?: string
  initialTimecode?: number
  videoRef: RefObject<HTMLVideoElement | null>
  onViewedChange?: () => void
}

/** Periodic + pause/close persistence to /viewed and local Continue Watching. */
export function useTimecodePersist({
  enabled,
  hash,
  fileIndex,
  title,
  initialTimecode = 0,
  videoRef,
  onViewedChange,
}: UseTimecodePersistOptions) {
  const lastSaveRef = useRef(0)
  const resumeAppliedRef = useRef(false)
  const onViewedChangeRef = useRef(onViewedChange)

  useEffect(() => {
    onViewedChangeRef.current = onViewedChange
  }, [onViewedChange])

  useEffect(() => {
    resumeAppliedRef.current = false
  }, [hash, fileIndex, initialTimecode])

  const saveTimecode = useCallback(
    async (time: number) => {
      if (!enabled || !hash || fileIndex == null) return
      try {
        await setViewedFile(hash, fileIndex, time)
        rememberContinueWatching({
          hash,
          fileIndex,
          title: title || hash,
          fileName: title || String(fileIndex),
          timecode: time,
        })
        onViewedChangeRef.current?.()
      } catch {
        // ignore persistence errors — playback should continue
      }
    },
    [enabled, hash, fileIndex, title],
  )

  const flushTimecode = useCallback(() => {
    const el = videoRef.current
    if (!el || !enabled) return
    void saveTimecode(el.currentTime)
  }, [videoRef, enabled, saveTimecode])

  const onTimeUpdate = useCallback(() => {
    const el = videoRef.current
    if (!el || !enabled) return
    const now = Date.now()
    if (now - lastSaveRef.current < TIMECODE_SAVE_INTERVAL_MS) return
    lastSaveRef.current = now
    void saveTimecode(el.currentTime)
  }, [videoRef, enabled, saveTimecode])

  const applyResumeIfNeeded = useCallback(() => {
    const el = videoRef.current
    if (!el || resumeAppliedRef.current) return
    if (!(initialTimecode > TIMECODE_RESUME_MARGIN_SEC)) {
      resumeAppliedRef.current = true
      return
    }
    const duration = el.duration
    if (Number.isFinite(duration) && duration > 0 && initialTimecode >= duration - TIMECODE_RESUME_MARGIN_SEC) {
      resumeAppliedRef.current = true
      return
    }

    el.currentTime = initialTimecode
    resumeAppliedRef.current = true
  }, [videoRef, initialTimecode])

  return { flushTimecode, onTimeUpdate, applyResumeIfNeeded, saveTimecode }
}

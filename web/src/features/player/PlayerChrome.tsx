import type { CSSProperties, ReactNode } from 'react'
import {
  MediaCaptionsButton,
  MediaControlBar,
  MediaController,
  MediaFullscreenButton,
  MediaMuteButton,
  MediaPipButton,
  MediaPlaybackRateButton,
  MediaPlayButton,
  MediaSeekBackwardButton,
  MediaSeekForwardButton,
  MediaTimeDisplay,
  MediaTimeRange,
  MediaVolumeRange,
} from 'media-chrome/react'

export interface PlayerChromeProps {
  isMobile: boolean
  showPip: boolean
  showCaptionsButton: boolean
  topChrome: ReactNode
  /** Extra controls rendered inside the bottom bar (prefer media-* or simple buttons). */
  extraControls?: ReactNode
  children: ReactNode
  className?: string
  style?: CSSProperties
}

/** Disable Media Chrome native tooltips — they clash with HeroUI modal focus and our theme. */
const noTip = { noTooltip: true } as const

/**
 * Brand Media Chrome shell — TorrServer dark cinema OSD.
 * `children` must include the `<video slot="media" />` (and optional tracks).
 * Loading/buffering UI lives in VideoPlayer (HeroUI Spinner) — no MediaLoadingIndicator.
 */
export default function PlayerChrome({
  isMobile,
  showPip,
  showCaptionsButton,
  topChrome,
  extraControls,
  children,
  className,
  style,
}: PlayerChromeProps) {
  return (
    <MediaController className={`ts-player ${className || ''}`.trim()} autohide='2' style={style}>
      {children}
      <div slot='top-chrome' className='ts-player-top'>
        {topChrome}
      </div>
      <MediaControlBar className='ts-player-bar'>
        <MediaPlayButton {...noTip} className={isMobile ? 'ts-player-play-lg' : undefined} />
        <MediaSeekBackwardButton {...noTip} seekOffset={10} />
        <MediaSeekForwardButton {...noTip} seekOffset={10} />
        <MediaTimeDisplay showDuration />
        <MediaTimeRange />
        <MediaMuteButton {...noTip} />
        {!isMobile ? <MediaVolumeRange /> : null}
        {extraControls}
        {showCaptionsButton ? <MediaCaptionsButton {...noTip} /> : null}
        <MediaPlaybackRateButton {...noTip} rates={[0.75, 1, 1.25, 1.5, 2]} />
        {showPip ? <MediaPipButton {...noTip} /> : null}
        <MediaFullscreenButton {...noTip} />
      </MediaControlBar>
    </MediaController>
  )
}

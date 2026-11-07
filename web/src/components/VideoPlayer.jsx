import {
  Box,
  CircularProgress,
  DialogContent,
  DialogTitle,
  IconButton,
  Menu,
  MenuItem,
  Slider,
  Tooltip,
  Typography,
  useMediaQuery,
} from '@material-ui/core'
import { makeStyles, withStyles } from '@material-ui/core/styles'
import CloseIcon from '@material-ui/icons/Close'
import Forward10Icon from '@material-ui/icons/Forward10'
import FullscreenIcon from '@material-ui/icons/Fullscreen'
import FullscreenExitIcon from '@material-ui/icons/FullscreenExit'
import GetAppIcon from '@material-ui/icons/GetApp'
import PauseIcon from '@material-ui/icons/Pause'
import PictureInPictureIcon from '@material-ui/icons/PictureInPicture'
import PlayArrowIcon from '@material-ui/icons/PlayArrow'
import Replay10Icon from '@material-ui/icons/Replay10'
import SpeedIcon from '@material-ui/icons/Speed'
import SubtitlesIcon from '@material-ui/icons/Subtitles'
import VolumeOffIcon from '@material-ui/icons/VolumeOff'
import VolumeUpIcon from '@material-ui/icons/VolumeUp'
import Hls from 'hls.js'
import { useCallback, useEffect, useRef, useState } from 'react'
import { StyledDialog } from 'style/CustomMaterialUiStyles'
import { useTranslation } from 'react-i18next'
import { fetchExternalSubtitles, linkFileIndex, subtitleFileUrl, trackLabel } from 'utils/Subs'

import { StyledButton } from './TorrentCard/style'

function getMimeType(url) {
  const ext = url.split('?')[0].split('.').pop().toLowerCase()
  switch (ext) {
    case 'mp4':
      return 'video/mp4'
    case 'ogg':
    case 'ogv':
      return 'video/ogg'
    case 'webm':
      return 'video/webm'
    default:
      return ''
  }
}

// The controls hide after this time without movement of the mouse, touches and keys
const IDLE_MS = 3000

// MUI renders menus and tooltips in a layer of <body>, but in fullscreen the browser shows only the fullscreen
// element and its descendants, so they stay behind the video: in fullscreen they must be created inside it
const portalContainer = () => document.fullscreenElement || document.body

const PlayerTooltip = props => <Tooltip PopperProps={{ container: portalContainer }} {...props} />

const canPlayNativeHls = video =>
  Boolean(video.canPlayType('application/vnd.apple.mpegurl') || video.canPlayType('application/x-mpegURL'))

const PrettoSlider = withStyles(theme => ({
  root: {
    color: '#00a572',
    height: 6,
    [theme?.breakpoints?.down?.('sm')]: {
      height: 0,
    },
  },
  thumb: {
    height: 18,
    width: 18,
    backgroundColor: '#fff',
    border: '2px solid currentColor',
    marginTop: -6,
    marginLeft: -12,
    [theme?.breakpoints?.down?.('sm')]: {
      height: 15,
      width: 15,
      marginTop: -5,
      marginLeft: -7,
    },
  },
  track: {
    height: 6,
    borderRadius: 4,
    [theme?.breakpoints?.down?.('sm')]: {
      height: 5,
    },
  },
  rail: {
    height: 6,
    borderRadius: 4,
    [theme?.breakpoints?.down?.('sm')]: {
      height: 6,
    },
  },
}))(Slider)

const useStyles = makeStyles(theme => ({
  dialogPaper: {
    backgroundColor: '#fff',
    borderRadius: theme.spacing(1),
  },
  header: {
    backgroundColor: '#00a572',
    color: '#fff',
    padding: theme.spacing(1, 2),
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  videoWrapper: {
    position: 'relative',
    width: '100%',
    backgroundColor: '#000',
    overflow: 'hidden',
    '&:hover $controls, &:hover $centralControl, &:hover $skipButton': {
      opacity: 1,
    },
  },
  videoWrapperFull: {
    height: '100%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    '& video': { width: '100%', height: '100%', objectFit: 'contain' },
  },
  // No activity: the controls, the buttons over the picture and the cursor are hidden.
  // Defined after videoWrapper, so it wins over its :hover rules.
  videoWrapperIdle: {
    cursor: 'none',
    '& video': { cursor: 'none' },
    '& $controls, &:hover $controls, & $centralControl, &:hover $centralControl, & $skipButton, &:hover $skipButton': {
      opacity: 0,
      pointerEvents: 'none',
    },
  },
  video: {
    width: '100%',
    display: 'block',
    cursor: 'pointer',
    [theme.breakpoints.down('sm')]: {
      height: '94.5vh',
      width: '100vw',
      objectFit: 'contain',
    },
  },
  loadingOverlay: {
    position: 'absolute',
    top: 0,
    left: 0,
    width: '100%',
    height: '100%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: 'rgba(0,0,0,0.6)',
    zIndex: 4,
  },
  centralControl: {
    position: 'absolute',
    top: '50%',
    left: '50%',
    transform: 'translate(-50%, -50%)',
    borderRadius: '50%',
    padding: theme.spacing(1),
    backgroundColor: 'rgba(0,0,0,0.5)',
    opacity: 0,
    transition: 'opacity 200ms',
    zIndex: 3,
    color: '#fff',
    pointerEvents: 'none',
    animation: '$pulse 0.6s ease-out',
  },
  skipButton: {
    position: 'absolute',
    top: '50%',
    transform: 'translateY(-50%)',
    padding: theme.spacing(1),
    backgroundColor: 'rgba(0,0,0,0.4)',
    color: '#fff',
    opacity: 0,
    transition: 'opacity 200ms',
    zIndex: 3,
    '&:hover': { backgroundColor: 'rgba(0,0,0,0.6)' },
  },
  leftSkip: { left: theme.spacing(2) },
  rightSkip: { right: theme.spacing(2) },
  controls: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    width: '100%',
    background: 'linear-gradient(to top, rgba(0,0,0,0.8), transparent)',
    padding: theme.spacing(0, 3, 2, 3),
    transition: 'opacity 200ms',
    opacity: 0,
    display: 'flex',
    flexDirection: 'column',
    gap: theme.spacing(0.5),
    zIndex: 3,
    pointerEvents: 'auto',
    [theme.breakpoints.down('sm')]: {
      opacity: 1,
      padding: theme.spacing(0, 1, 2, 1),
      gap: theme.spacing(0),
      background: 'linear-gradient(to top, rgba(0,0,0,0.95), transparent)',
    },
  },
  timeRow: {
    color: '#fff',
    paddingLeft: theme.spacing(2),
    [theme.breakpoints.down('sm')]: {
      paddingLeft: theme.spacing(1),
      fontSize: 9,
    },
  },
  slider: {
    color: '#00e68a',
    '& .MuiSlider-thumb': { backgroundColor: '#00e68a' },
    '& .MuiSlider-track': { borderRadius: 2 },
  },
  controlRow: {
    display: 'flex',
    alignItems: 'center',
  },
  iconButton: {
    color: '#fff',
    padding: 12,
    '&:hover': { backgroundColor: 'rgba(255,255,255,0.1)' },
    [theme.breakpoints.down('sm')]: {
      padding: 10,
    },
  },
  speedMenu: { minWidth: 100 },
  subtitleMenu: {
    '& .MuiPaper-root': {
      minWidth: 180,
      maxWidth: 320,
    },
  },
  '@keyframes pulse': {
    '0%': { transform: 'translate(-50%, -50%) scale(0.5)', opacity: 0 },
    '50%': { transform: 'translate(-50%, -50%) scale(1)', opacity: 1 },
    '100%': { transform: 'translate(-50%, -50%) scale(1.3)', opacity: 0 },
  },
}))

// Helper function to format seconds to HH:MM:SS
const formatTime = seconds => {
  if (!isFinite(seconds)) return '00:00:00'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = Math.floor(seconds % 60)
  const hh = h.toString().padStart(2, '0')
  const mm = m.toString().padStart(2, '0')
  const ss = s.toString().padStart(2, '0')
  return `${hh}:${mm}:${ss}`
}

const subtitleLabel = track => {
  const name = track.name || track.lang || 'Subtitle'
  return track.lang && track.lang.toLowerCase() !== name.toLowerCase() ? `${name} (${track.lang})` : name
}

// External HLS subtitle renditions (see server/gstreamer/subtitles_external.go) are numbered
// 1000 + position in the /subs list; returns that position or null for embedded tracks
const externalHlsPosition = track => {
  const match = /\/subs\/(\d+)\.m3u8/.exec(track.url || '')
  return match && Number(match[1]) >= 1000 ? Number(match[1]) - 1000 : null
}

// Entries of the subtitle menu. Labels start with [int] for tracks inside the video file and
// with [ext] for external files.
//  kind 'hls'  - a track of the HLS playlist (embedded tracks and the WebVTT copy of external files)
//  kind 'text' - an external file shown as WebVTT (direct playback only)
const buildSubtitleMenu = (hlsMode, subtitleTracks, externalTracks) => {
  const items = []
  const hlsExternal = new Map()
  subtitleTracks.forEach((track, index) => {
    const position = externalHlsPosition(track)
    if (position === null) {
      items.push({ key: `hls:${index}`, label: `[int] ${subtitleLabel(track)}`, kind: 'hls', hlsIndex: index })
    } else {
      hlsExternal.set(position, { track, index })
    }
  })

  externalTracks.forEach((track, position) => {
    const label = `[ext] ${trackLabel(track)}`
    if (!hlsMode) {
      items.push({ key: `text:${track.index}`, label, kind: 'text', track })
    } else if (hlsExternal.has(position)) {
      const { index } = hlsExternal.get(position)
      items.push({ key: `hls:${index}`, label, kind: 'hls', hlsIndex: index })
      hlsExternal.delete(position)
    }
  })

  // external HLS tracks that are not in the list (yet): keep them as they are
  hlsExternal.forEach(({ track, index }) => {
    items.push({ key: `hls:${index}`, label: `[ext] ${subtitleLabel(track)}`, kind: 'hls', hlsIndex: index })
  })
  return items
}

const disableNativeTracks = video => {
  const { textTracks } = video
  for (let i = 0; i < textTracks.length; i++) {
    const track = textTracks[i]
    track.mode = 'disabled'
  }
}

const VideoPlayer = ({
  videoSrc,
  downloadSrc = videoSrc,
  captionSrc = '',
  subsSrc = '',
  title,
  onNotSupported,
  hls = false,
  heartbeatSrc = '',
  showTrigger = true,
  initiallyOpen = false,
  onClose,
}) => {
  const classes = useStyles()
  const isMobile = useMediaQuery('@media (max-width:930px)')
  const videoRef = useRef(null)
  const wrapperRef = useRef(null)
  const hlsRef = useRef(null)
  const dynamicTrackRef = useRef(null)
  const onNotSupportedRef = useRef(onNotSupported)
  const { t } = useTranslation()
  const [open, setOpen] = useState(initiallyOpen)
  const [videoElement, setVideoElement] = useState(null)
  const [loading, setLoading] = useState(true)
  const [playing, setPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  const [muted, setMuted] = useState(false)
  const [volume, setVolume] = useState(1)
  const [fullscreen, setFullscreen] = useState(false)
  const [idle, setIdle] = useState(false)
  const [anchorEl, setAnchorEl] = useState(null)
  const [speed, setSpeed] = useState(1)
  const [subtitleAnchorEl, setSubtitleAnchorEl] = useState(null)
  const [subtitleTracks, setSubtitleTracks] = useState([])
  const [subtitleTrack, setSubtitleTrack] = useState(-1)
  const [externalTracks, setExternalTracks] = useState([])
  const [externalKey, setExternalKey] = useState(null) // 'text:<index>' of the selected external track

  const setVideoNode = useCallback(node => {
    videoRef.current = node
    setVideoElement(node)
  }, [])

  useEffect(() => {
    onNotSupportedRef.current = onNotSupported
  }, [onNotSupported])

  useEffect(() => {
    const vid = document.createElement('video')
    const supported = hls ? Hls.isSupported() || canPlayNativeHls(vid) : Boolean(vid.canPlayType(getMimeType(videoSrc)))
    if (!supported) onNotSupportedRef.current?.()
  }, [hls, videoSrc])

  useEffect(() => {
    if (!open || !hls || !videoElement) return undefined

    const video = videoElement

    let hlsPlayer
    let nativeHls = false
    setLoading(true)
    setSubtitleTracks([])
    setSubtitleTrack(-1)

    if (Hls.isSupported()) {
      hlsPlayer = new Hls()
      hlsRef.current = hlsPlayer
      hlsPlayer.on(Hls.Events.MANIFEST_PARSED, (_, data) => {
        setSubtitleTracks(data.subtitleTracks || [])
        setSubtitleTrack(hlsPlayer.subtitleTrack)
        video.play().catch(() => {})
      })
      hlsPlayer.on(Hls.Events.SUBTITLE_TRACKS_UPDATED, (_, data) => {
        setSubtitleTracks(data.subtitleTracks || [])
      })
      hlsPlayer.on(Hls.Events.SUBTITLE_TRACK_SWITCH, (_, data) => {
        setSubtitleTrack(data.id)
      })
      hlsPlayer.on(Hls.Events.ERROR, (_, data) => {
        if (!data.fatal) return

        if (data.type === Hls.ErrorTypes.NETWORK_ERROR) {
          hlsPlayer.startLoad()
        } else if (data.type === Hls.ErrorTypes.MEDIA_ERROR) {
          hlsPlayer.recoverMediaError()
        } else {
          hlsPlayer.stopLoad()
          setLoading(false)
        }
      })
      hlsPlayer.loadSource(videoSrc)
      hlsPlayer.attachMedia(video)
    } else if (canPlayNativeHls(video)) {
      nativeHls = true
      video.src = videoSrc
      video.load()
      video.play().catch(() => {})
    } else {
      onNotSupportedRef.current?.()
    }

    return () => {
      if (hlsRef.current === hlsPlayer) hlsRef.current = null
      if (hlsPlayer) hlsPlayer.destroy()
      setSubtitleAnchorEl(null)
      setSubtitleTracks([])
      setSubtitleTrack(-1)
      if (nativeHls) {
        video.pause()
        video.removeAttribute('src')
        video.load()
      }
    }
  }, [hls, open, videoElement, videoSrc])

  useEffect(() => {
    if (!open || !heartbeatSrc) return undefined

    const timer = window.setInterval(() => {
      fetch(heartbeatSrc, { cache: 'no-store' }).catch(() => {})
    }, 30 * 1000)

    return () => window.clearInterval(timer)
  }, [heartbeatSrc, open])

  const handlePlayPause = useCallback(() => {
    const video = videoRef.current
    if (!video) return
    video.paused ? video.play() : video.pause()
  }, [])

  const togglePlay = () => setPlaying(p => !p)
  const handleTimeUpdate = () => setCurrentTime(videoRef.current.currentTime)
  const handleLoaded = () => {
    setDuration(videoRef.current.duration)
    setLoading(false)
  }
  const handleSeek = (_, val) => {
    videoRef.current.currentTime = val
    handleTimeUpdate()
  }
  const handleVolume = (_, val) => {
    const v = val / 100
    videoRef.current.volume = v
    setVolume(v)
    setMuted(v === 0)
  }
  const toggleMute = () => {
    videoRef.current.muted = !muted
    setMuted(m => !m)
  }

  const skip = useCallback(
    secs => {
      const video = videoRef.current
      if (!video) return
      const target = Math.min(Math.max(video.currentTime + secs, 0), duration)
      video.currentTime = target
      setCurrentTime(target)
    },
    [duration],
  )

  // The wrapper (not the <video>) goes fullscreen: the controls and the subtitle layers must stay visible
  const enterFull = () => {
    const wrapper = wrapperRef.current
    if (wrapper?.requestFullscreen) wrapper.requestFullscreen()
    else if (wrapper?.webkitRequestFullscreen) wrapper.webkitRequestFullscreen()
    else if (videoRef.current?.webkitEnterFullscreen) videoRef.current.webkitEnterFullscreen()
  }
  const exitFull = () => document.exitFullscreen()

  useEffect(() => {
    const onFull = () => setFullscreen(!!document.fullscreenElement)
    document.addEventListener('fullscreenchange', onFull)
    return () => document.removeEventListener('fullscreenchange', onFull)
  }, [])

  // While the video plays, the controls hide after a few seconds without activity (the cursor too); a move
  // of the mouse, a touch or a key brings them back. On pause and while a menu is open they stay.
  const menuOpen = Boolean(anchorEl || subtitleAnchorEl)
  useEffect(() => {
    const wrapper = wrapperRef.current
    if (!open || !playing || !wrapper || menuOpen) {
      setIdle(false)
      return undefined
    }
    let timer = 0
    const wake = () => {
      setIdle(false)
      window.clearTimeout(timer)
      timer = window.setTimeout(() => setIdle(true), IDLE_MS)
    }
    wake()
    const events = ['mousemove', 'pointerdown', 'touchstart', 'wheel']
    events.forEach(name => wrapper.addEventListener(name, wake, { passive: true }))
    document.addEventListener('keydown', wake)
    return () => {
      window.clearTimeout(timer)
      events.forEach(name => wrapper.removeEventListener(name, wake))
      document.removeEventListener('keydown', wake)
    }
  }, [open, playing, menuOpen])

  const openSpeedMenu = e => setAnchorEl(e.currentTarget)
  const closeSpeedMenu = () => setAnchorEl(null)
  const changeSpeed = val => {
    videoRef.current.playbackRate = val
    setSpeed(val)
    closeSpeedMenu()
  }
  const clearExternal = useCallback(() => {
    if (dynamicTrackRef.current) {
      dynamicTrackRef.current.remove()
      dynamicTrackRef.current = null
    }
    setExternalKey(null)
  }, [])

  // Direct playback: shows an external subtitle file as WebVTT through a <track> element.
  // (With HLS the server puts the WebVTT copy of external files into the playlist.)
  const selectExternal = useCallback(
    track => {
      const video = videoRef.current
      if (!video) return
      clearExternal()
      disableNativeTracks(video)
      setExternalKey(`text:${track.index}`)

      const element = document.createElement('track')
      element.kind = 'subtitles'
      element.label = trackLabel(track)
      if (track.lang && track.lang !== 'und') element.srclang = track.lang
      element.src = subtitleFileUrl(track)
      video.appendChild(element)
      dynamicTrackRef.current = element
      element.track.mode = 'showing'
    },
    [clearExternal],
  )

  useEffect(() => {
    if (!open || !subsSrc) return undefined
    const controller = new AbortController()
    fetchExternalSubtitles(subsSrc, controller.signal).then(list => {
      if (!controller.signal.aborted) setExternalTracks(list)
    })
    return () => {
      controller.abort()
      setExternalTracks([])
    }
  }, [open, subsSrc])

  // Direct playback: the same-name .srt/.vtt file used to be shown through a static <track>.
  // When the track list is available, pick that file from the list instead.
  useEffect(() => {
    if (hls || !externalTracks.length || !captionSrc) return
    const index = linkFileIndex(captionSrc)
    const track = externalTracks.find(item => item.index === index)
    if (track) selectExternal(track)
  }, [externalTracks, hls, captionSrc, selectExternal])

  const openSubtitleMenu = e => setSubtitleAnchorEl(e.currentTarget)
  const closeSubtitleMenu = () => setSubtitleAnchorEl(null)
  const changeSubtitleTrack = index => {
    clearExternal()
    const hlsPlayer = hlsRef.current
    if (hlsPlayer) {
      hlsPlayer.subtitleDisplay = index >= 0
      hlsPlayer.subtitleTrack = index
    } else if (videoRef.current) {
      disableNativeTracks(videoRef.current)
    }
    setSubtitleTrack(index)
    closeSubtitleMenu()
  }
  const changeExternalTrack = track => {
    selectExternal(track)
    closeSubtitleMenu()
  }
  const downloadVideo = () => {
    const a = document.createElement('a')
    a.href = downloadSrc
    a.download = ''
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  }

  const closePlayer = () => {
    setOpen(false)
    onClose?.()
  }

  const handleKey = useCallback(
    e => {
      if (!open) return
      switch (e.key) {
        case ' ':
          e.preventDefault()
          handlePlayPause()
          break
        case 'ArrowRight':
          e.preventDefault()
          skip(10)
          break
        case 'ArrowLeft':
          e.preventDefault()
          skip(-10)
          break
        default:
          break
      }
    },
    [open, handlePlayPause, skip],
  )
  useEffect(() => {
    document.addEventListener('keydown', handleKey)
    return () => document.removeEventListener('keydown', handleKey)
  }, [handleKey])

  const subtitleMenu = buildSubtitleMenu(hls, subtitleTracks, externalTracks)
  const hasSubtitleMenu = subtitleMenu.length > 0
  const subtitlesOn = subtitleTrack >= 0 || externalKey !== null
  const isItemSelected = item => {
    if (item.kind === 'hls') return externalKey === null && subtitleTrack === item.hlsIndex
    return externalKey === `${item.kind}:${item.track.index}`
  }
  const selectItem = item => {
    if (item.kind === 'hls') changeSubtitleTrack(item.hlsIndex)
    else changeExternalTrack(item.track)
  }

  return (
    <>
      {showTrigger && (
        <StyledButton
          onClick={() => {
            setLoading(true)
            setOpen(true)
          }}
        >
          <PlayArrowIcon />
          <span>{t('Play')}</span>
        </StyledButton>
      )}
      <StyledDialog
        open={open}
        onClose={closePlayer}
        maxWidth='lg'
        fullWidth
        fullScreen={isMobile}
        classes={{ paper: classes.dialogPaper }}
      >
        <DialogTitle className={classes.header} disableTypography>
          <Typography variant='h6' noWrap>
            {title || 'Video Player'}
          </Typography>
          <IconButton size='medium' onClick={closePlayer} className={classes.iconButton}>
            <CloseIcon fontSize='medium' />
          </IconButton>
        </DialogTitle>
        <DialogContent style={{ padding: 0 }}>
          <Box
            ref={wrapperRef}
            className={`${classes.videoWrapper} ${fullscreen ? classes.videoWrapperFull : ''} ${
              idle ? classes.videoWrapperIdle : ''
            }`}
            onClick={handlePlayPause}
            style={isMobile ? { minHeight: 240 } : {}}
          >
            <video
              autoPlay
              ref={setVideoNode}
              src={hls ? undefined : videoSrc}
              onTimeUpdate={handleTimeUpdate}
              onLoadedMetadata={handleLoaded}
              onPlay={togglePlay}
              onPause={togglePlay}
              className={classes.video}
            >
              <track kind='captions' srcLang='en' label='English captions' src={hls ? undefined : captionSrc} default />
            </video>
            {loading && (
              <Box className={classes.loadingOverlay}>
                <CircularProgress fontSize='medium' />
              </Box>
            )}
            <IconButton
              size='medium'
              className={classes.centralControl}
              style={{
                opacity: playing ? 0 : 1,
              }}
            >
              <PlayArrowIcon fontSize='medium' />
            </IconButton>
            <Box className={classes.controls} onClick={e => e.stopPropagation()}>
              {isMobile && (
                <Box className={classes.timeRow}>
                  <Typography variant='body2'>
                    {formatTime(currentTime)} / {formatTime(duration)}
                  </Typography>
                </Box>
              )}
              <PrettoSlider
                className={classes.slider}
                value={currentTime}
                max={duration}
                onChange={handleSeek}
                size='medium'
              />
              <Box className={classes.controlRow}>
                <PlayerTooltip title={playing ? t('Pause') : t('Play')}>
                  <IconButton size='medium' onClick={handlePlayPause} className={classes.iconButton}>
                    {playing ? <PauseIcon fontSize='medium' /> : <PlayArrowIcon fontSize='medium' />}
                  </IconButton>
                </PlayerTooltip>
                <PlayerTooltip title={t('Rewind-10-Sec')}>
                  <IconButton
                    size='medium'
                    className={classes.iconButton}
                    onClick={e => {
                      e.stopPropagation()
                      skip(-10)
                    }}
                  >
                    <Replay10Icon fontSize='medium' />
                  </IconButton>
                </PlayerTooltip>

                <PlayerTooltip title={t('Forward-10-Sec')}>
                  <IconButton
                    size='medium'
                    className={classes.iconButton}
                    onClick={e => {
                      e.stopPropagation()
                      skip(10)
                    }}
                  >
                    <Forward10Icon fontSize='medium' />
                  </IconButton>
                </PlayerTooltip>
                <PlayerTooltip title={muted ? t('Unmute') : t('Mute')}>
                  <IconButton size='medium' className={classes.iconButton} onClick={toggleMute}>
                    {muted ? <VolumeOffIcon fontSize='medium' /> : <VolumeUpIcon fontSize='medium' />}
                  </IconButton>
                </PlayerTooltip>
                {!isMobile && (
                  <Slider
                    className={classes.slider}
                    value={volume * 100}
                    onChange={handleVolume}
                    size='medium'
                    style={{ width: 70 }}
                  />
                )}
                {!isMobile && (
                  <Box className={classes.timeRow}>
                    <Typography variant='body2'>
                      {formatTime(currentTime)} / {formatTime(duration)}
                    </Typography>
                  </Box>
                )}
                <Box flexGrow={1} />
                {hasSubtitleMenu && (
                  <>
                    <PlayerTooltip title={t('GStreamer.Subtitles')}>
                      <IconButton
                        size='medium'
                        onClick={openSubtitleMenu}
                        className={classes.iconButton}
                        style={subtitlesOn ? { color: '#00e68a' } : undefined}
                      >
                        <SubtitlesIcon fontSize='medium' />
                      </IconButton>
                    </PlayerTooltip>
                    <Menu
                      container={portalContainer}
                      anchorEl={subtitleAnchorEl}
                      open={Boolean(subtitleAnchorEl)}
                      onClose={closeSubtitleMenu}
                      className={classes.subtitleMenu}
                    >
                      <MenuItem selected={!subtitlesOn} onClick={() => changeSubtitleTrack(-1)}>
                        {t('None')}
                      </MenuItem>
                      {subtitleMenu.map(item => (
                        <MenuItem key={item.key} selected={isItemSelected(item)} onClick={() => selectItem(item)}>
                          {item.label}
                        </MenuItem>
                      ))}
                    </Menu>
                  </>
                )}
                <PlayerTooltip title={t('Speed')}>
                  <IconButton size='medium' onClick={openSpeedMenu} className={classes.iconButton}>
                    <SpeedIcon fontSize='medium' />
                  </IconButton>
                </PlayerTooltip>
                <Menu
                  container={portalContainer}
                  anchorEl={anchorEl}
                  open={Boolean(anchorEl)}
                  onClose={closeSpeedMenu}
                  className={classes.speedMenu}
                >
                  {[0.5, 1, 1.5, 2].map(r => (
                    <MenuItem key={r} selected={r === speed} onClick={() => changeSpeed(r)}>
                      {r}x
                    </MenuItem>
                  ))}
                </Menu>
                <PlayerTooltip title={t('PIP')}>
                  <IconButton
                    size='medium'
                    className={classes.iconButton}
                    onClick={() => videoRef.current.requestPictureInPicture()}
                  >
                    <PictureInPictureIcon fontSize='medium' />
                  </IconButton>
                </PlayerTooltip>

                <PlayerTooltip title={t('Download')}>
                  <IconButton size='medium' className={classes.iconButton} onClick={downloadVideo}>
                    <GetAppIcon fontSize='medium' />
                  </IconButton>
                </PlayerTooltip>

                <PlayerTooltip title={fullscreen ? t('ExitFullscreen') : t('Fullscreen')}>
                  <IconButton size='medium' onClick={fullscreen ? exitFull : enterFull} className={classes.iconButton}>
                    {fullscreen ? <FullscreenExitIcon fontSize='medium' /> : <FullscreenIcon fontSize='medium' />}
                  </IconButton>
                </PlayerTooltip>
              </Box>
            </Box>
          </Box>
        </DialogContent>
      </StyledDialog>
    </>
  )
}

export default VideoPlayer

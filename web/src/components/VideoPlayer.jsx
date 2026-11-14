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
import AssPlayer from 'utils/ass'
import TextCues from 'utils/TextCues'
import { useBooleanPlayerPreference, useStringPlayerPreference } from 'utils/PlayerPreferences'
import {
  fetchExternalSubtitles,
  fetchFontFiles,
  fontsListUrl,
  isAssTrack,
  linkFileIndex,
  subtitleFileUrl,
  trackLabel,
} from 'utils/Subs'

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
    '&:hover $subtitleLayer': {
      bottom: 'calc(var(--subs-lift, 112px) + 8px)',
    },
  },
  // Text subtitles (WebVTT); the lift above the controls is measured by the component
  subtitleLayer: {
    position: 'absolute',
    left: 0,
    right: 0,
    bottom: theme.spacing(3),
    padding: theme.spacing(0, 6),
    zIndex: 2,
    textAlign: 'center',
    color: '#fff',
    fontSize: 'clamp(15px, 2.2vw, 34px)',
    lineHeight: 1.3,
    pointerEvents: 'none',
    transition: 'bottom 200ms',
    '& > div': {
      whiteSpace: 'pre-wrap',
      textShadow: '0 0 4px #000, 0 0 4px #000, 1px 1px 2px #000, -1px -1px 2px #000',
    },
    [theme.breakpoints.down('sm')]: {
      padding: theme.spacing(0, 2),
      bottom: 'calc(var(--subs-lift, 112px) + 8px)',
    },
  },
  // Not a flex container: JASSUB puts its own element next to the <video> and positions the canvas
  // relative to it, which only works in the normal flow. object-fit centers the picture.
  videoWrapperFull: {
    height: '100%',
    '& video': { width: '100%', height: '100%', objectFit: 'contain' },
  },
  // No activity: the controls, the buttons over the picture and the cursor are hidden and
  // the text subtitles go back to the bottom. Defined after videoWrapper, so it wins over its :hover rules.
  videoWrapperIdle: {
    cursor: 'none',
    '& video': { cursor: 'none' },
    '& $controls, &:hover $controls, & $centralControl, &:hover $centralControl, & $skipButton, &:hover $skipButton': {
      opacity: 0,
      pointerEvents: 'none',
    },
    '& $subtitleLayer, &:hover $subtitleLayer': {
      bottom: theme.spacing(3),
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
// with [ext] for external files. An ASS/SSA file is offered in one format: with styling
// (rendered by the player) or as text (WebVTT), depending on the setting.
//  kind 'hls'  - a track of the HLS playlist (embedded tracks and the WebVTT copy of external files)
//  kind 'ass'  - an ASS/SSA file shown with its styling
//  kind 'text' - an external file shown as WebVTT (direct playback only)
const buildSubtitleMenu = (hlsMode, subtitleTracks, externalTracks, assStyled) => {
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
    if (assStyled && isAssTrack(track)) {
      items.push({ key: `ass:${track.index}`, label, kind: 'ass', track })
      hlsExternal.delete(position) // its WebVTT copy is not offered
    } else if (!hlsMode) {
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
  const controlsRef = useRef(null)
  const textLayerRef = useRef(null)
  const [controlsHeight, setControlsHeight] = useState(112)
  const hlsRef = useRef(null)
  const assRef = useRef(null)
  const dynamicTrackRef = useRef(null)
  const loadTokenRef = useRef(0)
  const assStyled = useBooleanPlayerPreference('isAssStyled', true)
  const assFonts = useBooleanPlayerPreference('assFonts', true)
  const assRenderer = useStringPlayerPreference('assRenderer', 'auto')
  const fontsRequestRef = useRef(null) // the list of fonts for ASS: GET /subs/{hash}/{id}?fonts
  // read when an ASS track is opened; must not re-create selectExternal when they change
  const assOptionsRef = useRef({})
  assOptionsRef.current = { mode: assRenderer, fonts: assFonts }
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
  const [externalKey, setExternalKey] = useState(null) // 'ass:<index>' or 'text:<index>' of the selected external track

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
      hlsPlayer.subtitleDisplay = false // text subtitles are drawn by TextCues
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
    if (!open || !videoElement || !textLayerRef.current) return undefined
    const cues = new TextCues(videoElement, textLayerRef.current)
    return () => cues.destroy()
  }, [open, videoElement])

  // Height of the controls bar: text subtitles are drawn above it
  useEffect(() => {
    const controls = controlsRef.current
    if (!open || !controls) return undefined
    const update = () => setControlsHeight(controls.offsetHeight || 112)
    update()
    if (typeof ResizeObserver === 'undefined') return undefined
    const observer = new ResizeObserver(update)
    observer.observe(controls)
    return () => observer.disconnect()
  }, [open, videoElement])

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
    loadTokenRef.current += 1
    if (assRef.current) assRef.current.unload()
    if (dynamicTrackRef.current) {
      dynamicTrackRef.current.remove()
      dynamicTrackRef.current = null
    }
    setExternalKey(null)
  }, [])

  // Shows an external subtitle file: variant 'ass' draws it with its styling,
  // 'text' shows it as WebVTT (direct playback; with HLS the server provides the WebVTT copy)
  const selectExternal = useCallback(
    async (track, variant) => {
      const video = videoRef.current
      if (!video) return
      clearExternal()
      const token = loadTokenRef.current

      // switch off the other subtitle sources
      const hlsPlayer = hlsRef.current
      if (hlsPlayer) {
        hlsPlayer.subtitleDisplay = false
        hlsPlayer.subtitleTrack = -1
      }
      setSubtitleTrack(-1)
      disableNativeTracks(video)
      setExternalKey(`${variant}:${track.index}`)

      if (variant === 'ass') {
        try {
          const response = await fetch(subtitleFileUrl(track, true))
          if (!response.ok) throw new Error(`status ${response.status}`)
          const text = await response.text()
          if (token !== loadTokenRef.current || !videoRef.current || !wrapperRef.current) return
          const fontFiles = (await fontsRequestRef.current) || []
          if (token !== loadTokenRef.current || !videoRef.current || !wrapperRef.current) return
          if (!assRef.current) assRef.current = new AssPlayer(video, wrapperRef.current)
          assRef.current.options = { ...assOptionsRef.current, fontFiles }
          await assRef.current.load(text, 'external')
        } catch {
          // the file could not be read: leave the player without subtitles
          if (token === loadTokenRef.current) setExternalKey(null)
        }
      } else if (!hls) {
        const element = document.createElement('track')
        element.kind = 'subtitles'
        element.label = trackLabel(track)
        if (track.lang && track.lang !== 'und') element.srclang = track.lang
        element.src = subtitleFileUrl(track, false)
        video.appendChild(element)
        dynamicTrackRef.current = element
        element.track.mode = 'hidden' // TextCues draws it
      }
    },
    [hls, clearExternal],
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

  // Fonts for ASS subtitles, from the torrent and from the video file, in one list. The request is made when
  // they can be used; a track waits for it. Old servers do not know ?fonts: the list is empty then.
  useEffect(() => {
    if (!open || !subsSrc || !assStyled || !assFonts) return undefined
    const controller = new AbortController()
    fontsRequestRef.current = fetchFontFiles(fontsListUrl(subsSrc), controller.signal)
    return () => {
      controller.abort()
      fontsRequestRef.current = null
    }
  }, [open, subsSrc, assStyled, assFonts])

  // Direct playback: the same-name .srt/.vtt file used to be shown through a static <track>.
  // When the track list is available, pick that file from the list instead.
  useEffect(() => {
    if (hls || !externalTracks.length || !captionSrc) return
    const index = linkFileIndex(captionSrc)
    const track = externalTracks.find(item => item.index === index)
    if (track) selectExternal(track, 'text')
  }, [externalTracks, hls, captionSrc, selectExternal])

  useEffect(() => {
    if (open || !assRef.current) return
    assRef.current.destroy()
    assRef.current = null
  }, [open])

  useEffect(
    () => () => {
      if (assRef.current) assRef.current.destroy()
      assRef.current = null
    },
    [],
  )

  const openSubtitleMenu = e => setSubtitleAnchorEl(e.currentTarget)
  const closeSubtitleMenu = () => setSubtitleAnchorEl(null)
  const changeSubtitleTrack = index => {
    clearExternal()
    const hlsPlayer = hlsRef.current
    if (hlsPlayer) {
      hlsPlayer.subtitleDisplay = false
      hlsPlayer.subtitleTrack = index
    } else if (videoRef.current) {
      disableNativeTracks(videoRef.current)
    }
    setSubtitleTrack(index)
    closeSubtitleMenu()
  }
  const changeExternalTrack = (track, variant) => {
    selectExternal(track, variant)
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

  const subtitleMenu = buildSubtitleMenu(hls, subtitleTracks, externalTracks, assStyled)
  const hasSubtitleMenu = subtitleMenu.length > 0
  const subtitlesOn = subtitleTrack >= 0 || externalKey !== null
  const isItemSelected = item => {
    if (item.kind === 'hls') return externalKey === null && subtitleTrack === item.hlsIndex
    return externalKey === `${item.kind}:${item.track.index}`
  }
  const selectItem = item => {
    if (item.kind === 'hls') changeSubtitleTrack(item.hlsIndex)
    else changeExternalTrack(item.track, item.kind)
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
            style={{ '--subs-lift': `${controlsHeight}px`, ...(isMobile ? { minHeight: 240 } : {}) }}
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
            <div ref={textLayerRef} className={classes.subtitleLayer} />
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
            <Box ref={controlsRef} className={classes.controls} onClick={e => e.stopPropagation()}>
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

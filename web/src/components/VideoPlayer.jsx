import {
  Box,
  CircularProgress,
  Dialog,
  DialogContent,
  DialogTitle,
  IconButton,
  Slider,
  Typography,
} from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import CloseIcon from '@material-ui/icons/Close'
import FastForwardIcon from '@material-ui/icons/FastForward'
import FastRewindIcon from '@material-ui/icons/FastRewind'
import FullscreenIcon from '@material-ui/icons/Fullscreen'
import FullscreenExitIcon from '@material-ui/icons/FullscreenExit'
import PauseIcon from '@material-ui/icons/Pause'
import PlayArrowIcon from '@material-ui/icons/PlayArrow'
import VolumeOffIcon from '@material-ui/icons/VolumeOff'
import VolumeUpIcon from '@material-ui/icons/VolumeUp'
import React, { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { StyledButton } from './TorrentCard/style'

const useStyles = makeStyles(theme => ({
  dialogPaper: {
    backgroundColor: '#fff',
    borderRadius: theme.spacing(1),
    border: `4px solid #00a572`,
  },
  header: {
    backgroundColor: '#00a572',
    color: '#fff',
    margin: 0,
    padding: theme.spacing(1, 2),
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  dialogContent: {
    padding: 0,
    overflowX: 'hidden', // Remove horizontal scroll
  },
  video: {
    width: '100%',
    display: 'block',
    backgroundColor: '#000', // Fallback background
    minHeight: '300px', // Ensure the video area is visible
    cursor: 'pointer', // Indicate clickable video
  },
  playButton: {},
  controls: {
    backgroundColor: '#fff',
    borderTop: `2px solid #00a572`,
    padding: theme.spacing(1),
  },
  progressSlider: {
    margin: 0,
  },
  controlRow: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'start',
    flexWrap: 'nowrap',
  },
  volumeControl: {
    display: 'flex',
    alignItems: 'center',
    position: 'relative',
    cursor: 'pointer',
  },
  // Updated vertical volume slider container positioned above the volume button.
  volumeSliderContainer: {
    position: 'absolute',
    bottom: 'calc(100%)', // Positioned above the volume button
    left: '50%',
    transform: 'translateX(-50%)',
    height: 150,
    width: 30,
    backgroundColor: '#fff',
    border: `1px solid #00a572`,
    borderRadius: theme.spacing(0.5),
    boxShadow: '0px 4px 8px rgba(0, 0, 0, 0.1)',
    padding: theme.spacing(1),
    zIndex: 10,
    display: 'flex',
    justifyContent: 'center',
  },
  // Customize vertical slider appearance.
  volumeSlider: {
    color: '#00a572',
    height: '100%',
    '& .MuiSlider-track': {
      color: '#00a572',
    },
    '& .MuiSlider-thumb': {
      border: '2px solid #00a572',
    },
  },
  // Loading overlay styles.
  loadingOverlay: {
    position: 'absolute',
    top: 0,
    left: 0,
    width: '100%',
    height: '100%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: 'rgba(2, 2, 2, 0.44)',
    zIndex: 2,
  },
  // Ripple container that covers the video.
  rippleContainer: {
    position: 'absolute',
    top: 0,
    left: 0,
    width: '100%',
    height: '100%',
    pointerEvents: 'none',
    overflow: 'hidden',
    zIndex: 3,
  },
  // Ripple circle animation.
  ripple: {
    position: 'absolute',
    width: 20,
    height: 20,
    backgroundColor: 'rgba(0, 165, 114, 0.4)',
    borderRadius: '50%',
    transform: 'scale(0)',
    animation: '$rippleEffect 600ms ease-out',
  },
  '@keyframes rippleEffect': {
    '0%': {
      transform: 'scale(0)',
      opacity: 0.7,
    },
    '100%': {
      transform: 'scale(8)',
      opacity: 0,
    },
  },
  // Ripple icon animation.
  rippleIcon: {
    position: 'absolute',
    top: '50%',
    left: '50%',
    transform: 'translate(-50%, -50%) scale(0.5)',
    color: '#00a572',
    animation: '$rippleIconEffect 600ms ease-out',
  },
  '@keyframes rippleIconEffect': {
    '0%': {
      opacity: 1,
      transform: 'translate(-50%, -50%) scale(0.5)',
    },
    '100%': {
      opacity: 0,
      transform: 'translate(-50%, -50%) scale(1.5)',
    },
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

/**
 * VideoPlayer component
 *
 * Props:
 * - videoSrc: string (URL of the video file)
 * - captionSrc: string (optional URL for captions, e.g. VTT file)
 * - title: string (optional title for the video)
 */
const VideoPlayer = ({ videoSrc, captionSrc = '', title, onNotSupported }) => {
  const classes = useStyles()
  const videoRef = useRef(null)
  const containerRef = useRef(null)
  const { t } = useTranslation()
  const [ripples, setRipples] = useState([])
  const [isSupported, setIsSupported] = useState(false)
  const [open, setOpen] = useState(false)
  const [isPlaying, setIsPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  const [volume, setVolume] = useState(1) // Range: 0 to 1
  const [muted, setMuted] = useState(false)
  const [isFullScreen, setIsFullScreen] = useState(false)
  const [showVolumeSlider, setShowVolumeSlider] = useState(false)
  const [loading, setLoading] = useState(true) // Loading state

  // Unified hover events for volume control.
  const handleVolumeEnter = () => {
    setShowVolumeSlider(true)
  }
  const handleVolumeLeave = () => {
    setShowVolumeSlider(false)
  }

  // Determine MIME type based on file extension.
  const getMimeType = url => {
    const ext = url.split('?')[0].split('.').pop().toLowerCase()
    switch (ext) {
      case 'mp4':
        return 'video/mp4'
      case 'ogg':
        return 'video/ogg'
      case 'webm':
        return 'video/webm'
      default:
        return ''
    }
  }

  // Check if the browser supports the provided video type.
  useEffect(() => {
    if (!videoSrc) {
      setIsSupported(false)
      onNotSupported()
      return
    }
    const mimeType = getMimeType(videoSrc)
    if (!mimeType) {
      setIsSupported(false)
      onNotSupported()
      return
    }
    const videoElem = document.createElement('video')
    const canPlay = videoElem.canPlayType(mimeType)
    setIsSupported(!!canPlay && canPlay !== '')
    if (!(!!canPlay && canPlay !== '')) {
      onNotSupported()
    }
  }, [videoSrc, onNotSupported])

  // Listen for full screen changes.
  useEffect(() => {
    const handleFullScreenChange = () => {
      setIsFullScreen(!!document.fullscreenElement)
    }
    document.addEventListener('fullscreenchange', handleFullScreenChange)
    return () => {
      document.removeEventListener('fullscreenchange', handleFullScreenChange)
    }
  }, [])

  // When the dialog opens, start playing the video.
  const handleDialogEntered = () => {
    if (videoRef.current) {
      videoRef.current
        .play()
        .then(() => setIsPlaying(true))
        .catch(error => {
          console.error('Error attempting to play', error)
          setIsPlaying(false)
        })
    }
  }

  // Toggle between play and pause.
  const handleTogglePlay = () => {
    if (!videoRef.current) return
    if (isPlaying) {
      videoRef.current.pause()
      setIsPlaying(false)
    } else {
      videoRef.current
        .play()
        .then(() => setIsPlaying(true))
        .catch(error => console.error('Error attempting to play', error))
    }
  }

  // Handle click on video: toggle play/pause and create a ripple.
  const handleVideoClick = e => {
    // Determine the new state icon.
    const newType = isPlaying ? 'play' : 'pause'
    handleTogglePlay()

    // Get click coordinates relative to the container.
    const rect = containerRef.current.getBoundingClientRect()
    const x = e.clientX - rect.left
    const y = e.clientY - rect.top
    const id = Date.now()
    setRipples(prev => [...prev, { id, x, y, type: newType }])
    // Remove ripple after animation.
    setTimeout(() => {
      setRipples(prev => prev.filter(r => r.id !== id))
    }, 600)
  }

  const handleSkipBackward = () => {
    if (!videoRef.current) return
    videoRef.current.currentTime = Math.max(videoRef.current.currentTime - 10, 0)
    setCurrentTime(videoRef.current.currentTime)
  }

  const handleSkipForward = () => {
    if (!videoRef.current) return
    videoRef.current.currentTime = Math.min(videoRef.current.currentTime + 10, duration)
    setCurrentTime(videoRef.current.currentTime)
  }

  const handleToggleMute = () => {
    if (!videoRef.current) return
    if (muted) {
      videoRef.current.muted = false
      setMuted(false)
      if (videoRef.current.volume === 0) {
        videoRef.current.volume = 0.5
        setVolume(0.5)
      }
    } else {
      videoRef.current.muted = true
      setMuted(true)
    }
  }

  const handleVolumeSliderChange = (event, newValue) => {
    if (videoRef.current) {
      const newVolume = newValue / 100
      videoRef.current.volume = newVolume
      setVolume(newVolume)
      if (newVolume === 0) {
        videoRef.current.muted = true
        setMuted(true)
      } else {
        videoRef.current.muted = false
        setMuted(false)
      }
    }
  }

  const handleTimeUpdate = e => {
    setCurrentTime(e.target.currentTime)
  }

  const handleLoadedMetadata = e => {
    setDuration(e.target.duration)
    setCurrentTime(e.target.currentTime)
    setLoading(false) // Hide loader when metadata is loaded
  }

  const handleSliderChange = (event, newValue) => {
    if (videoRef.current) {
      videoRef.current.currentTime = newValue
    }
    setCurrentTime(newValue)
  }

  const handleToggleFullScreen = () => {
    if (!document.fullscreenElement) {
      if (videoRef.current.requestFullscreen) {
        videoRef.current.requestFullscreen()
      }
    } else if (document.exitFullscreen) {
      document.exitFullscreen()
    }
  }

  if (!isSupported) {
    return null
  }
  // Uncomment for debugging: console.log("captionSrc", captionSrc);
  return (
    <>
      <StyledButton onClick={() => setOpen(true)}>
        <PlayArrowIcon />
        <span>{t('Play')}</span>
      </StyledButton>

      <Dialog
        open={open}
        onClose={() => setOpen(false)}
        maxWidth='lg'
        maxHeight='lg'
        fullWidth
        onEntered={handleDialogEntered}
        classes={{ paper: classes.dialogPaper }}
      >
        <DialogTitle disableTypography className={classes.header}>
          <h2 style={{ margin: 0, textOverflow: 'ellipsis' }}>{title ?? 'Video Player'}</h2>
          <IconButton onClick={() => setOpen(false)} style={{ color: '#fff' }}>
            <CloseIcon />
          </IconButton>
        </DialogTitle>
        <DialogContent className={classes.dialogContent}>
          <Box position='relative' ref={containerRef}>
            {/* Video with click-to-toggle play/pause */}
            <video
              ref={videoRef}
              src={videoSrc}
              autoPlay
              className={classes.video}
              onClick={handleVideoClick}
              onTimeUpdate={handleTimeUpdate}
              onLoadedMetadata={handleLoadedMetadata}
            >
              {/* {captionSrc && ( */}
              <track default kind='captions' src={captionSrc} label='Captions' />
              {/* )} */}
            </video>
            {loading && (
              <Box className={classes.loadingOverlay}>
                <CircularProgress size={40} />
              </Box>
            )}
            {/* Ripple overlay */}
            <Box className={classes.rippleContainer}>
              {ripples.map(ripple => (
                <React.Fragment key={ripple.id}>
                  <Box className={classes.ripple} style={{ left: ripple.x - 10, top: ripple.y - 10 }} />
                  <Box className={classes.rippleIcon} style={{ left: ripple.x, top: ripple.y }}>
                    {ripple.type === 'play' ? <PlayArrowIcon fontSize='large' /> : <PauseIcon fontSize='large' />}
                  </Box>
                </React.Fragment>
              ))}
            </Box>
          </Box>

          <Box className={classes.controls}>
            {/* Time display with progress slider */}
            <Box display='flex' alignItems='center' justifyContent='space-between'>
              <Typography variant='body2'>{formatTime(currentTime)}</Typography>
              <Box flexGrow={1} mx={2}>
                <Slider
                  className={classes.progressSlider}
                  value={currentTime}
                  max={duration}
                  onChange={handleSliderChange}
                  aria-labelledby='video-progress'
                />
              </Box>
              <Typography variant='body2'>{formatTime(duration)}</Typography>
            </Box>

            <Box className={classes.controlRow}>
              <IconButton onClick={handleSkipBackward}>
                <FastRewindIcon fontSize='large' className={classes.playButton} />
              </IconButton>
              <IconButton onClick={handleTogglePlay}>
                {isPlaying ? (
                  <PauseIcon fontSize='large' className={classes.playButton} />
                ) : (
                  <PlayArrowIcon fontSize='large' className={classes.playButton} />
                )}
              </IconButton>
              <IconButton onClick={handleSkipForward}>
                <FastForwardIcon fontSize='large' className={classes.playButton} />
              </IconButton>
              <Box className={classes.volumeControl} onMouseEnter={handleVolumeEnter} onMouseLeave={handleVolumeLeave}>
                <IconButton onClick={handleToggleMute}>
                  {muted ? (
                    <VolumeOffIcon fontSize='large' className={classes.playButton} />
                  ) : (
                    <VolumeUpIcon fontSize='large' className={classes.playButton} />
                  )}
                </IconButton>
                {showVolumeSlider && (
                  <Box className={classes.volumeSliderContainer}>
                    <Slider
                      orientation='vertical'
                      className={classes.volumeSlider}
                      value={volume * 100}
                      onChange={handleVolumeSliderChange}
                      aria-labelledby='volume-slider'
                    />
                  </Box>
                )}
              </Box>
              <IconButton style={{ marginLeft: 'auto' }} onClick={handleToggleFullScreen}>
                {isFullScreen ? (
                  <FullscreenExitIcon fontSize='large' className={classes.playButton} />
                ) : (
                  <FullscreenIcon fontSize='large' className={classes.playButton} />
                )}
              </IconButton>
            </Box>
          </Box>
        </DialogContent>
      </Dialog>
    </>
  )
}

export default VideoPlayer

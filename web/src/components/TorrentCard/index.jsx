import { forwardRef, memo, useEffect, useRef, useState } from 'react'
import {
  Audiotrack as AudiotrackIcon,
  UnfoldMore as UnfoldMoreIcon,
  PlayArrow as PlayArrowIcon,
  Close as CloseIcon,
  Delete as DeleteIcon,
} from '@material-ui/icons'
import { getPeerString, humanizeSize, humanizeSpeed, removeRedundantCharacters } from 'utils/Utils'
import { playlistTorrHost, streamHost, torrentsHost } from 'utils/Hosts'
import { NoImageIcon } from 'icons'
import DialogTorrentDetailsContent from 'components/DialogTorrentDetailsContent'
import Dialog from '@material-ui/core/Dialog'
import Slide from '@material-ui/core/Slide'
import {
  Button,
  CircularProgress,
  DialogActions,
  DialogTitle,
  ListItemIcon,
  ListItemText,
  Menu,
  MenuItem,
  useMediaQuery,
  useTheme,
} from '@material-ui/core'
import axios from 'axios'
import ptt from 'parse-torrent-title'
import { useTranslation } from 'react-i18next'
import AddDialog from 'components/Add/AddDialog'
import { StyledDialog } from 'style/CustomMaterialUiStyles'
import useOnStandaloneAppOutsideClick from 'utils/useOnStandaloneAppOutsideClick'
import { GETTING_INFO, IN_DB, CLOSED, PRELOAD, WORKING } from 'torrentStates'
import { TORRENT_CATEGORIES } from 'components/categories'
import VideoPlayer from 'components/VideoPlayer'
import { isFilePlayable } from 'components/DialogTorrentDetailsContent/helpers'
import {
  gstreamerHeartbeatUrl,
  gstreamerMasterUrl,
  gstreamerProbeUrl,
  shouldUseGStreamerPlayer,
  useGStreamerRuntime,
} from 'utils/GStreamer'

import {
  StatusIndicators,
  StyledButton,
  TorrentCard,
  TorrentCardButtons,
  TorrentCardDescription,
  TorrentCardPoster,
} from './style'

const Transition = forwardRef((props, ref) => <Slide direction='up' ref={ref} {...props} />)

const wait = milliseconds => new Promise(resolve => setTimeout(resolve, milliseconds))

const requestTorrentFiles = async (hash, isActive, attemptsLeft = 60) => {
  const { data: status } = await axios.post(torrentsHost(), { action: 'get', hash })
  const files = status?.file_stats || []
  if (!isActive() || files.length || attemptsLeft <= 1) return files

  await wait(1000)
  if (!isActive()) return []
  return requestTorrentFiles(hash, isActive, attemptsLeft - 1)
}

const fileName = path => path.split('\\').pop().split('/').pop()

const filesFromMetadata = data => {
  if (!data) return []
  try {
    return JSON.parse(data).TorrServer?.Files || []
  } catch (_) {
    return []
  }
}

const episodeLabel = (path, index) => {
  const name = fileName(path)
  const parsed = ptt.parse(name)
  const season = Number(parsed.season)
  const episode = Number(parsed.episode)
  const code = `${season ? `S${String(season).padStart(2, '0')}` : ''}${
    episode ? `E${String(episode).padStart(2, '0')}` : ''
  }`
  const title = parsed.title || name.replace(/\.[^/.]+$/, '')
  return code ? `${code} - ${title}` : `${index + 1}. ${title}`
}

const probeTrackValue = (track, name) => track?.[name] ?? track?.[`${name[0].toLowerCase()}${name.slice(1)}`]

const probeAudioTracks = probe =>
  (probe?.Tracks || probe?.tracks || []).filter(
    track => String(probeTrackValue(track, 'Type')).toLowerCase() === 'audio',
  )

const audioCodecName = track => {
  const codec = String(probeTrackValue(track, 'Codec') || '')
  const caps = String(probeTrackValue(track, 'CapsName') || '')
  const value = `${caps} ${codec}`.toLowerCase()

  switch (true) {
    case value.includes('eac3') || value.includes('e-ac3') || value.includes('e-ac-3'):
      return 'E-AC3'
    case value.includes('truehd') || value.includes('true-hd') || value.includes('mlp'):
      return 'TrueHD'
    case value.includes('ac3') || value.includes('ac-3') || value.includes('a/52'):
      return 'AC3'
    case value.includes('dts'):
      return 'DTS'
    case value.includes('aac') || value.includes('mpegversion=(int)4') || value.includes('mpegversion=4'):
      return 'AAC'
    case value.includes('opus'):
      return 'Opus'
    case value.includes('vorbis'):
      return 'Vorbis'
    case value.includes('flac'):
      return 'FLAC'
    case value.includes('mp3') || value.includes('layer=(int)3'):
      return 'MP3'
    case value.includes('pcm') || value.includes('audio/x-raw'):
      return 'PCM'
    default: {
      const shortName = (caps || codec).split(',')[0].trim()
      return shortName
        .replace(/^audio\/(?:x-)?/i, '')
        .replace(/[_-]+/g, ' ')
        .toUpperCase()
    }
  }
}

const sameFileList = (left, right) => {
  const leftFiles = left || []
  const rightFiles = right || []
  return (
    leftFiles.length === rightFiles.length &&
    leftFiles.every((file, index) => {
      const other = rightFiles[index]
      return file.id === other?.id && file.path === other.path && file.length === other.length
    })
  )
}

const Torrent = ({ torrent }) => {
  const { t } = useTranslation()
  const [isDetailedInfoOpened, setIsDetailedInfoOpened] = useState(false)
  const [isDeleteTorrentOpened, setIsDeleteTorrentOpened] = useState(false)
  const [unsupportedPlayers, setUnsupportedPlayers] = useState({})
  const [episodeMenuAnchor, setEpisodeMenuAnchor] = useState(null)
  const [selectedPlayer, setSelectedPlayer] = useState(null)
  const [resolvedFileList, setResolvedFileList] = useState([])
  const [isResolvingPlayers, setIsResolvingPlayers] = useState(false)
  const [playerResolveFailed, setPlayerResolveFailed] = useState(false)
  const [openEpisodeMenuAfterResolve, setOpenEpisodeMenuAfterResolve] = useState(false)
  const [audioTracksByFile, setAudioTracksByFile] = useState({})
  const [audioMenuAnchor, setAudioMenuAnchor] = useState(null)
  const [audioMenuPlayer, setAudioMenuPlayer] = useState(null)
  const [isResolvingAudio, setIsResolvingAudio] = useState(false)
  const isMounted = useRef(true)
  const episodeButtonRef = useRef(null)
  const audioButtonRef = useRef(null)
  const gstRuntime = useGStreamerRuntime()

  useEffect(
    () => () => {
      isMounted.current = false
    },
    [],
  )

  const theme = useTheme()
  const fullScreen = useMediaQuery(theme.breakpoints.down('md'))

  const openDetailedInfo = () => setIsDetailedInfoOpened(true)
  const closeDetailedInfo = () => setIsDetailedInfoOpened(false)
  const openDeleteTorrentAlert = () => setIsDeleteTorrentOpened(true)
  const closeDeleteTorrentAlert = () => setIsDeleteTorrentOpened(false)

  const {
    title,
    name,
    category,
    poster,
    torrent_size: torrentSize,
    download_speed: downloadSpeed,
    hash,
    stat,
    data,
    file_stats: torrentFileList,
  } = torrent

  const dropTorrent = () => axios.post(torrentsHost(), { action: 'drop', hash })
  const deleteTorrent = () => axios.post(torrentsHost(), { action: 'rem', hash })

  const getParsedTitle = () => {
    const parse = key => ptt.parse(title || '')?.[key] || ptt.parse(name || '')?.[key]

    const titleStrings = []

    let parsedTitle = removeRedundantCharacters(parse('title'))
    const parsedYear = parse('year')
    const parsedResolution = parse('resolution')
    if (parsedTitle) titleStrings.push(parsedTitle)
    if (parsedYear) titleStrings.push(`(${parsedYear})`)
    if (parsedResolution) titleStrings.push(`[${parsedResolution}]`)
    parsedTitle = titleStrings.join(' ')
    return { parsedTitle }
  }
  const { parsedTitle } = getParsedTitle()

  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false)
  const handleClickOpenEditDialog = () => setIsEditDialogOpen(true)
  const handleCloseEditDialog = () => setIsEditDialogOpen(false)

  const fullPlaylistLink = `${playlistTorrHost()}/${encodeURIComponent(parsedTitle || 'file')}.m3u?link=${hash}&m3u`

  const detailedInfoDialogRef = useOnStandaloneAppOutsideClick(closeDetailedInfo)
  // main categories
  const catIndex = TORRENT_CATEGORIES.findIndex(e => e.key === category)
  const catArray = TORRENT_CATEGORIES.find(e => e.key === category)
  const getFileLink = (path, id) =>
    `${streamHost()}/${encodeURIComponent(fileName(path))}?link=${hash}&index=${id}&play`

  const fileList = torrentFileList?.length
    ? torrentFileList
    : resolvedFileList.length
    ? resolvedFileList
    : filesFromMetadata(data)
  const playableVideoList = fileList.filter(({ path }) => isFilePlayable(path))
  const getVideoCaption = path => {
    // Get base name without extension
    const baseName = path.replace(/\.[^/.]+$/, '')
    // Find a file with the same base name and a subtitle extension
    const captionFile = fileList.find(file => file.path.startsWith(baseName) && /\.(srt|vtt)$/i.test(file.path))
    return captionFile ? getFileLink(captionFile.path, captionFile.id) : ''
  }
  const createPlayer = (file, index) => {
    const hls = shouldUseGStreamerPlayer(file.path, gstRuntime)
    const downloadSrc = getFileLink(file.path, file.id)
    return {
      ...file,
      key: `${file.id}:${hls ? 'gst' : 'stream'}`,
      label: episodeLabel(file.path, index),
      videoSrc: hls ? gstreamerMasterUrl(hash, file.id) : downloadSrc,
      downloadSrc,
      hls,
      heartbeatSrc: hls ? gstreamerHeartbeatUrl(hash) : '',
    }
  }
  const players = playableVideoList.map(createPlayer)
  const availablePlayers = players.filter(player => !unsupportedPlayers[player.key])
  const singlePlayer = players.length === 1 ? players[0] : null
  const audioMenuTracks = audioMenuPlayer ? audioTracksByFile[audioMenuPlayer.id] || [] : []

  const playerWithAudio = (player, audio) => ({
    ...player,
    key: `${player.key}:audio:${audio}`,
    videoSrc: gstreamerMasterUrl(hash, player.id, audio),
    playerTitle: title || player.label,
  })

  const showAudioTracks = (player, tracks, anchor) => {
    if (!tracks.length) {
      setSelectedPlayer(playerWithAudio(player, 0))
      return
    }
    setAudioMenuPlayer(player)
    const target = anchor || audioButtonRef.current
    if (target) {
      setAudioMenuAnchor(target)
    } else {
      window.requestAnimationFrame(() => {
        if (isMounted.current) setAudioMenuAnchor(audioButtonRef.current)
      })
    }
  }

  const resolveAudioTracks = async (player, anchor) => {
    const cached = audioTracksByFile[player.id]
    if (cached !== undefined) {
      showAudioTracks(player, cached, anchor)
      return
    }

    setIsResolvingAudio(true)
    try {
      const { data: probe } = await axios.get(gstreamerProbeUrl(hash, player.id))
      if (!isMounted.current) return

      const tracks = probeAudioTracks(probe)
      setAudioTracksByFile(current => ({ ...current, [player.id]: tracks }))
      showAudioTracks(player, tracks, anchor)
    } catch (_) {
      if (isMounted.current) setSelectedPlayer(playerWithAudio(player, 0))
    } finally {
      if (isMounted.current) setIsResolvingAudio(false)
    }
  }

  const audioTrackLabel = (track, ordinal) => {
    const trackTitle = String(probeTrackValue(track, 'Title') || '').trim()
    const language = String(probeTrackValue(track, 'Language') || '').trim()
    const codec = audioCodecName(track)
    const channels = Number(probeTrackValue(track, 'Channels'))
    const rate = Number(probeTrackValue(track, 'Rate'))
    const details = [
      trackTitle && language ? language.toUpperCase() : '',
      codec,
      channels > 0 ? `${channels} ch` : '',
      rate > 0 ? `${Math.round(rate / 1000)} kHz` : '',
    ].filter(Boolean)
    return {
      primary: trackTitle || language.toUpperCase() || `${t('GStreamer.AudioTrack')} ${ordinal + 1}`,
      secondary: details.join(' / '),
    }
  }

  const selectAudioTrack = (track, ordinal) => {
    if (!audioMenuPlayer) return
    const value = Number(probeTrackValue(track, 'Index'))
    const audio = Number.isInteger(value) && value >= 0 ? value : ordinal
    setAudioMenuAnchor(null)
    setSelectedPlayer(playerWithAudio(audioMenuPlayer, audio))
  }

  useEffect(() => {
    if (!openEpisodeMenuAfterResolve || players.length <= 1 || !episodeButtonRef.current) return
    setEpisodeMenuAnchor(episodeButtonRef.current)
    setOpenEpisodeMenuAfterResolve(false)
  }, [openEpisodeMenuAfterResolve, players.length])

  const markPlayerUnsupported = key => {
    setUnsupportedPlayers(current => ({ ...current, [key]: true }))
    setSelectedPlayer(current => (current?.key === key ? null : current))
  }
  const resolvePlayers = async () => {
    setIsResolvingPlayers(true)
    setPlayerResolveFailed(false)

    try {
      const files = await requestTorrentFiles(hash, () => isMounted.current)
      if (!isMounted.current) return

      const loadedPlayers = files.filter(({ path }) => isFilePlayable(path)).map(createPlayer)
      setResolvedFileList(files)
      if (loadedPlayers.length === 1) {
        if (loadedPlayers[0].hls) {
          await resolveAudioTracks(loadedPlayers[0])
        } else {
          setSelectedPlayer(loadedPlayers[0])
        }
      } else if (loadedPlayers.length > 1) {
        setOpenEpisodeMenuAfterResolve(true)
      } else {
        setPlayerResolveFailed(true)
      }
    } catch (_) {
      if (isMounted.current) setPlayerResolveFailed(true)
    } finally {
      if (isMounted.current) setIsResolvingPlayers(false)
    }
  }

  return (
    <>
      <TorrentCard>
        <TorrentCardPoster isPoster={poster} onClick={handleClickOpenEditDialog}>
          {poster ? <img src={poster} alt='poster' /> : <NoImageIcon />}
        </TorrentCardPoster>

        <TorrentCardButtons>
          <StyledButton onClick={openDetailedInfo}>
            <UnfoldMoreIcon />
            <span>{t('Details')}</span>
          </StyledButton>

          {singlePlayer && !unsupportedPlayers[singlePlayer.key] ? (
            singlePlayer.hls ? (
              <>
                <StyledButton
                  ref={audioButtonRef}
                  disabled={isResolvingAudio || isResolvingPlayers}
                  aria-haspopup='menu'
                  aria-expanded={Boolean(audioMenuAnchor)}
                  onClick={event => resolveAudioTracks(singlePlayer, event.currentTarget)}
                >
                  {isResolvingAudio || isResolvingPlayers ? (
                    <CircularProgress size={20} color='inherit' />
                  ) : (
                    <PlayArrowIcon />
                  )}
                  <span>{t('Play')}</span>
                </StyledButton>
                <Menu
                  anchorEl={audioMenuAnchor}
                  open={Boolean(audioMenuAnchor)}
                  onClose={() => setAudioMenuAnchor(null)}
                  getContentAnchorEl={null}
                  anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
                  transformOrigin={{ vertical: 'top', horizontal: 'left' }}
                  PaperProps={{ style: { maxHeight: '65vh', width: 420, maxWidth: 'calc(100vw - 32px)' } }}
                >
                  {audioMenuTracks.map((track, ordinal) => {
                    const label = audioTrackLabel(track, ordinal)
                    const index = probeTrackValue(track, 'Index') ?? ordinal
                    return (
                      <MenuItem key={index} onClick={() => selectAudioTrack(track, ordinal)}>
                        <ListItemIcon style={{ minWidth: 34 }}>
                          <AudiotrackIcon fontSize='small' />
                        </ListItemIcon>
                        <ListItemText primary={label.primary} secondary={label.secondary} />
                      </MenuItem>
                    )
                  })}
                </Menu>
              </>
            ) : (
              <VideoPlayer
                title={title}
                videoSrc={singlePlayer.videoSrc}
                downloadSrc={singlePlayer.downloadSrc}
                captionSrc={getVideoCaption(singlePlayer.path)}
                heartbeatSrc={singlePlayer.heartbeatSrc}
                onNotSupported={() => markPlayerUnsupported(singlePlayer.key)}
              />
            )
          ) : players.length > 1 && availablePlayers.length ? (
            <>
              <StyledButton
                ref={episodeButtonRef}
                aria-haspopup='menu'
                aria-expanded={Boolean(episodeMenuAnchor)}
                onClick={event => setEpisodeMenuAnchor(event.currentTarget)}
              >
                <PlayArrowIcon />
                <span>{t('Play')}</span>
              </StyledButton>
              <Menu
                anchorEl={episodeMenuAnchor}
                open={Boolean(episodeMenuAnchor)}
                onClose={() => setEpisodeMenuAnchor(null)}
                getContentAnchorEl={null}
                anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
                transformOrigin={{ vertical: 'top', horizontal: 'left' }}
                PaperProps={{ style: { maxHeight: '65vh', width: 420, maxWidth: 'calc(100vw - 32px)' } }}
              >
                {availablePlayers.map(player => (
                  <MenuItem
                    key={player.key}
                    onClick={() => {
                      setEpisodeMenuAnchor(null)
                      setSelectedPlayer(player)
                    }}
                  >
                    <ListItemIcon style={{ minWidth: 34 }}>
                      <PlayArrowIcon fontSize='small' />
                    </ListItemIcon>
                    <ListItemText primary={player.label} secondary={humanizeSize(player.length)} />
                  </MenuItem>
                ))}
              </Menu>
            </>
          ) : gstRuntime.built_in && !playerResolveFailed && players.length === 0 ? (
            <StyledButton disabled={isResolvingPlayers} onClick={resolvePlayers}>
              {isResolvingPlayers ? <CircularProgress size={20} color='inherit' /> : <PlayArrowIcon />}
              <span>{t('Play')}</span>
            </StyledButton>
          ) : (
            <StyledButton
              onClick={() => {
                window.open(fullPlaylistLink, '_blank')
              }}
            >
              <PlayArrowIcon />
              <span>{t('Playlist')}</span>
            </StyledButton>
          )}

          {selectedPlayer && (
            <VideoPlayer
              key={selectedPlayer.key}
              title={selectedPlayer.playerTitle || selectedPlayer.label}
              videoSrc={selectedPlayer.videoSrc}
              downloadSrc={selectedPlayer.downloadSrc}
              captionSrc={selectedPlayer.hls ? '' : getVideoCaption(selectedPlayer.path)}
              hls={selectedPlayer.hls}
              heartbeatSrc={selectedPlayer.heartbeatSrc}
              initiallyOpen
              showTrigger={false}
              onClose={() => setSelectedPlayer(null)}
              onNotSupported={() => markPlayerUnsupported(selectedPlayer.key)}
            />
          )}

          <StyledButton onClick={() => dropTorrent(torrent)}>
            <CloseIcon />
            <span>{t('Drop')}</span>
          </StyledButton>

          <StyledButton onClick={openDeleteTorrentAlert}>
            <DeleteIcon />
            <span>{t('Delete')}</span>
          </StyledButton>
        </TorrentCardButtons>

        <TorrentCardDescription>
          <div className='description-title-wrapper'>
            <div className='description-section-name'>
              {category ? (catIndex >= 0 ? t(catArray.name) : category) : t('Name')}
            </div>
            <div className='description-torrent-title'>{parsedTitle}</div>
          </div>

          <div className='description-statistics-wrapper'>
            <div className='description-statistics-element-wrapper'>
              <div className='description-section-name'>
                <StatusIndicator stat={stat} />
                {t('Size')}
              </div>
              <div className='description-statistics-element-value'>{torrentSize > 0 && humanizeSize(torrentSize)}</div>
            </div>

            <div className='description-statistics-element-wrapper'>
              <div className='description-section-name'>{t('Speed')}</div>
              <div className='description-statistics-element-value'>
                {downloadSpeed > 0 ? humanizeSpeed(downloadSpeed) : '---'}
              </div>
            </div>

            <div className='description-statistics-element-wrapper'>
              <div className='description-section-name'>{t('Peers')}</div>
              <div className='description-statistics-element-value'>{getPeerString(torrent) || '---'}</div>
            </div>
          </div>
        </TorrentCardDescription>
      </TorrentCard>

      <StyledDialog
        open={isDetailedInfoOpened}
        onClose={closeDetailedInfo}
        fullScreen={fullScreen}
        fullWidth
        maxWidth='xl'
        TransitionComponent={Transition}
        ref={detailedInfoDialogRef}
      >
        <DialogTorrentDetailsContent closeDialog={closeDetailedInfo} torrent={torrent} />
      </StyledDialog>

      <Dialog open={isDeleteTorrentOpened} onClose={closeDeleteTorrentAlert}>
        <DialogTitle>{t('DeleteTorrent?')}</DialogTitle>
        <DialogActions>
          <Button variant='outlined' onClick={closeDeleteTorrentAlert} color='secondary'>
            {t('Cancel')}
          </Button>

          <Button
            variant='contained'
            onClick={() => {
              deleteTorrent(torrent)
              closeDeleteTorrentAlert()
            }}
            color='secondary'
            autoFocus
          >
            {t('OK')}
          </Button>
        </DialogActions>
      </Dialog>

      {isEditDialogOpen && (
        <AddDialog
          hash={hash}
          title={title}
          name={name}
          poster={poster}
          handleClose={handleCloseEditDialog}
          category={category}
        />
      )}
    </>
  )
}

export const StatusIndicator = ({ stat }) => {
  const { t } = useTranslation()

  const values = {
    [GETTING_INFO]: t('TorrentGettingInfo'),
    [PRELOAD]: t('TorrentPreload'),
    [WORKING]: t('TorrentWorking'),
    [CLOSED]: t('TorrentClosed'),
    [IN_DB]: t('TorrentInDb'),
  }

  const colors = {
    [GETTING_INFO]: '#2196F3',
    [PRELOAD]: '#FFC107',
    [WORKING]: '#CDDC39',
    [CLOSED]: '#E57373',
    [IN_DB]: '#9E9E9E',
  }

  return (
    <span className='description-status-wrapper'>
      <StatusIndicators color={colors[stat]} title={values[stat]} />
    </span>
  )
}

export default memo(Torrent, (prev, next) => {
  const p = prev.torrent
  const n = next.torrent
  return (
    p.hash === n.hash &&
    p.title === n.title &&
    p.name === n.name &&
    p.poster === n.poster &&
    p.category === n.category &&
    p.stat === n.stat &&
    p.torrent_size === n.torrent_size &&
    p.download_speed === n.download_speed &&
    p.data === n.data &&
    sameFileList(p.file_stats, n.file_stats)
  )
})

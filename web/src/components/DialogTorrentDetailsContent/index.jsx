import Button from '@material-ui/core/Button'
import { AppBar, IconButton, makeStyles, Toolbar, Typography } from '@material-ui/core'
import CloseIcon from '@material-ui/icons/Close'
import styled, { css } from 'styled-components'
import { NoImageIcon } from 'icons'
import { getPeerString, humanizeSize } from 'utils/Utils'
import { viewedHost } from 'utils/Hosts'
import { CopyToClipboard } from 'react-copy-to-clipboard'
import { useEffect, useState } from 'react'

import { useUpdateCache, useCreateCacheMap } from './customHooks'

const useStyles = makeStyles(theme => ({
  appBar: { position: 'relative' },
  title: { marginLeft: theme.spacing(2), flex: 1 },
}))

const DialogContent = styled.div`
  display: grid;
  grid-template-rows: min-content 200px 80px 70px;
`
const Poster = styled.div`
  ${({ poster }) => css`
    height: 400px;
    border-radius: 5px;
    overflow: hidden;

    ${poster
      ? css`
          img {
            border-radius: 5px;
            height: 100%;
          }
        `
      : css`
          width: 300px;
          display: grid;
          place-items: center;
          background: #74c39c;

          svg {
            transform: scale(2.5) translateY(-3px);
          }
        `}
  `}
`
const HeaderSection = styled.section`
  padding: 40px;
  display: grid;
  grid-template-columns: min-content 1fr;
  gap: 30px;
`

const TorrentData = styled.div`
  > :not(:last-child) {
    margin-bottom: 20px;
  }
`

const CacheSection = styled.section`
  padding: 40px;
  background: lightgray;
`

const ButtonSection = styled.section`
  box-shadow: 0px 4px 4px -1px rgb(0 0 0 / 30%);
  display: flex;
  justify-content: space-evenly;
  align-items: center;
  text-transform: uppercase;
`

const ButtonSectionButton = styled.div`
  background: lightblue;
  height: 100%;
  flex: 1;
  display: grid;
  place-items: center;
  cursor: pointer;
  font-size: 15px;

  :not(:last-child) {
    border-right: 1px solid blue;
  }

  :hover {
    background: red;
  }
`

const TorrentFilesSection = styled.div``

const TorrentName = styled.div`
  font-size: 50px;
  font-weight: 200;
  line-height: 1;
`
const TorrentSubName = styled.div`
  color: #7c7b7c;
`

const shortenText = (text, count) => text.slice(0, count) + (text.length > count ? '...' : '')

export default function DialogTorrentDetailsContent({ closeDialog, torrent }) {
  const classes = useStyles()
  const [isLoading, setIsLoading] = useState(true)
  const {
    poster,
    hash,
    title,
    name,
    download_speed: downloadSpeed,
    upload_speed: uploadSpeed,
    stat_string: statString,
    torrent_size: torrentSize,
  } = torrent

  const cache = useUpdateCache(hash)
  const cacheMap = useCreateCacheMap(cache)

  useEffect(() => setIsLoading(false), [cacheMap])

  const { Capacity, PiecesCount, PiecesLength } = cache

  return (
    <>
      <AppBar className={classes.appBar}>
        <Toolbar>
          <IconButton edge='start' color='inherit' onClick={closeDialog} aria-label='close'>
            <CloseIcon />
          </IconButton>
          <Typography variant='h6' className={classes.title}>
            Torrent Details
          </Typography>
          <Button autoFocus color='inherit' onClick={closeDialog}>
            close
          </Button>
        </Toolbar>
      </AppBar>

      {isLoading ? (
        'loading'
      ) : (
        <DialogContent>
          <HeaderSection>
            <Poster poster={poster}>{poster ? <img alt='poster' src={poster} /> : <NoImageIcon />}</Poster>

            <TorrentData>
              <div>
                {name && name !== title ? (
                  <>
                    <TorrentName>{shortenText(name, 50)}</TorrentName>
                    <TorrentSubName>{shortenText(title, 160)}</TorrentSubName>
                  </>
                ) : (
                  <TorrentName>{shortenText(title, 50)}</TorrentName>
                )}
              </div>

              <div>peers: {getPeerString(torrent)}</div>
              <div>loaded: {getPreload(torrent)}</div>
              <div>download speed: {humanizeSize(downloadSpeed)}</div>
              <div>upload speed: {humanizeSize(uploadSpeed)}</div>
              <div>status: {statString}</div>
              <div>torrent size: {humanizeSize(torrentSize)}</div>

              <div>Capacity: {humanizeSize(Capacity)}</div>
              <div>PiecesCount: {PiecesCount}</div>
              <div>PiecesLength: {humanizeSize(PiecesLength)}</div>
            </TorrentData>
          </HeaderSection>

          <CacheSection />

          <ButtonSection>
            <CopyToClipboard text={hash}>
              <ButtonSectionButton>copy hash</ButtonSectionButton>
            </CopyToClipboard>

            <ButtonSectionButton>remove views</ButtonSectionButton>

            <ButtonSectionButton>drop torrent</ButtonSectionButton>

            <ButtonSectionButton>download playlist</ButtonSectionButton>

            <ButtonSectionButton>download playlist after last view</ButtonSectionButton>
          </ButtonSection>

          <TorrentFilesSection />
        </DialogContent>
      )}
    </>
  )
}

function getPreload(torrent) {
  if (torrent.preloaded_bytes > 0 && torrent.preload_size > 0 && torrent.preloaded_bytes < torrent.preload_size) {
    const progress = ((torrent.preloaded_bytes * 100) / torrent.preload_size).toFixed(2)
    return `${humanizeSize(torrent.preloaded_bytes)} / ${humanizeSize(torrent.preload_size)}   ${progress}%`
  }

  if (!torrent.preloaded_bytes) return humanizeSize(0)

  return humanizeSize(torrent.preloaded_bytes)
}

function remViews(hash) {
  try {
    if (hash)
      fetch(viewedHost(), {
        method: 'post',
        body: JSON.stringify({ action: 'rem', hash, file_index: -1 }),
        headers: {
          Accept: 'application/json, text/plain, */*',
          'Content-Type': 'application/json',
        },
      })
  } catch (e) {
    console.error(e)
  }
}

function getViewed(hash, callback) {
  try {
    fetch(viewedHost(), {
      method: 'post',
      body: JSON.stringify({ action: 'list', hash }),
      headers: {
        Accept: 'application/json, text/plain, */*',
        'Content-Type': 'application/json',
      },
    })
      .then(res => res.json())
      .then(callback)
  } catch (e) {
    console.error(e)
  }
}

function getPlayableFile(torrent) {
  if (!torrent || !torrent.file_stats) return null
  return torrent.file_stats.filter(file => extPlayable.includes(getExt(file.path)))
}

function getExt(filename) {
  const ext = filename.split('.').pop()
  if (ext === filename) return ''
  return ext.toLowerCase()
}
const extPlayable = [
  // video
  '3g2',
  '3gp',
  'aaf',
  'asf',
  'avchd',
  'avi',
  'drc',
  'flv',
  'iso',
  'm2v',
  'm2ts',
  'm4p',
  'm4v',
  'mkv',
  'mng',
  'mov',
  'mp2',
  'mp4',
  'mpe',
  'mpeg',
  'mpg',
  'mpv',
  'mxf',
  'nsv',
  'ogg',
  'ogv',
  'ts',
  'qt',
  'rm',
  'rmvb',
  'roq',
  'svi',
  'vob',
  'webm',
  'wmv',
  'yuv',
  // audio
  'aac',
  'aiff',
  'ape',
  'au',
  'flac',
  'gsm',
  'it',
  'm3u',
  'm4a',
  'mid',
  'mod',
  'mp3',
  'mpa',
  'pls',
  'ra',
  's3m',
  'sid',
  'wav',
  'wma',
  'xm',
]

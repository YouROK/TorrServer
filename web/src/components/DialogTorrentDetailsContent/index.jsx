import styled, { css } from 'styled-components'
import { NoImageIcon } from 'icons'
import { getPeerString, humanizeSize } from 'utils/Utils'
// import { viewedHost } from 'utils/Hosts'
import { CopyToClipboard } from 'react-copy-to-clipboard'
import { useEffect, useState } from 'react'
import { Button } from '@material-ui/core'
import { ArrowDownward, ArrowUpward, SwapVerticalCircle, ViewAgenda } from '@material-ui/icons'
import axios from 'axios'
import { torrentsHost } from 'utils/Hosts'

import { useUpdateCache, useCreateCacheMap, useGetSettings } from './customHooks'
import DialogHeader from './DialogHeader'
import TorrentCache from './TorrentCache'
import {
  DetailedTorrentCacheViewWrapper,
  DialogContentGrid,
  TorrentMainSection,
  Poster,
  TorrentData,
  SectionTitle,
  SectionSubName,
  StatisticsWrapper,
  ButtonSection,
  LoadingProgress,
  SectionHeader,
  CacheSection,
  ButtonSectionButton,
  TorrentFilesSection,
} from './style'
import StatisticsField from './StatisticsField'

const shortenText = (text, count) => text.slice(0, count) + (text.length > count ? '...' : '')

export default function DialogTorrentDetailsContent({ closeDialog, torrent }) {
  const [isLoading, setIsLoading] = useState(true)
  const [isDetailedCacheView, setIsDetailedCacheView] = useState(false)
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
  const settings = useGetSettings(cache)

  const dropTorrent = hash => axios.post(torrentsHost(), { action: 'drop', hash })

  const { Capacity, PiecesCount, PiecesLength, Filled } = cache

  useEffect(() => {
    const cacheLoaded = !!Object.entries(cache).length
    const torrentLoaded = torrent.stat_string !== 'Torrent in db' && torrent.stat_string !== 'Torrent getting info'

    if (!cacheLoaded && !isLoading) setIsLoading(true)
    if (cacheLoaded && isLoading && torrentLoaded) setIsLoading(false)
  }, [torrent, cache, isLoading])

  const bufferSize = settings?.PreloadBuffer ? Capacity : 33554432 // Default is 32mb if PreloadBuffer is false

  return (
    <>
      <DialogHeader
        onClose={closeDialog}
        title={isDetailedCacheView ? 'Detailed Cache View' : 'Torrent Details'}
        // eslint-disable-next-line react/jsx-props-no-spreading
        {...(isDetailedCacheView && { onBack: () => setIsDetailedCacheView(false) })}
      />

      {isLoading ? (
        'loading'
      ) : isDetailedCacheView ? (
        <DetailedTorrentCacheViewWrapper>
          <div>PiecesCount: {PiecesCount}</div>
          <div>PiecesLength: {humanizeSize(PiecesLength)}</div>
          <div>status: {statString}</div>
          <TorrentCache cache={cache} cacheMap={cacheMap} />
        </DetailedTorrentCacheViewWrapper>
      ) : (
        <DialogContentGrid>
          <TorrentMainSection>
            <Poster poster={poster}>{poster ? <img alt='poster' src={poster} /> : <NoImageIcon />}</Poster>

            <TorrentData>
              <div>
                {name && name !== title ? (
                  <>
                    <SectionTitle>{shortenText(name, 50)}</SectionTitle>
                    <SectionSubName>{shortenText(title, 160)}</SectionSubName>
                  </>
                ) : (
                  <SectionTitle>{shortenText(title, 50)}</SectionTitle>
                )}
              </div>

              <StatisticsWrapper>
                <StatisticsField
                  title='Download speed'
                  value={humanizeSize(downloadSpeed) || '0 B'}
                  iconBg='#118f00'
                  valueBg='#13a300'
                  icon={ArrowDownward}
                />

                <StatisticsField
                  title='Upload speed'
                  value={humanizeSize(uploadSpeed) || '0 B'}
                  iconBg='#0146ad'
                  valueBg='#0058db'
                  icon={ArrowUpward}
                />

                <StatisticsField
                  title='Peers'
                  value={getPeerString(torrent)}
                  iconBg='#0146ad'
                  valueBg='#0058db'
                  icon={SwapVerticalCircle}
                />

                <StatisticsField
                  title='Torrent size'
                  value={humanizeSize(torrentSize)}
                  iconBg='#0146ad'
                  valueBg='#0058db'
                  icon={ViewAgenda}
                />
              </StatisticsWrapper>
            </TorrentData>
          </TorrentMainSection>

          <CacheSection>
            <SectionHeader>
              <SectionTitle>Buffer</SectionTitle>
              {!settings?.PreloadBuffer && (
                <SectionSubName>Enable &quot;Preload Buffer&quot; in settings to change buffer size</SectionSubName>
              )}
              <LoadingProgress value={Filled} fullAmount={bufferSize} label={humanizeSize(bufferSize)} />
            </SectionHeader>

            <TorrentCache isMini cache={cache} cacheMap={cacheMap} />
            <Button
              style={{ marginTop: '30px' }}
              variant='contained'
              color='primary'
              size='large'
              onClick={() => setIsDetailedCacheView(true)}
            >
              Detailed cache view
            </Button>
          </CacheSection>

          <ButtonSection>
            <CopyToClipboard text={hash}>
              <ButtonSectionButton>
                <div className='hash-group'>
                  <div>copy hash</div>
                  <div className='hash-text'>{hash}</div>
                </div>
              </ButtonSectionButton>
            </CopyToClipboard>

            <ButtonSectionButton>remove views</ButtonSectionButton>

            <ButtonSectionButton onClick={() => dropTorrent(hash)}>drop torrent</ButtonSectionButton>

            <ButtonSectionButton>download playlist</ButtonSectionButton>

            <ButtonSectionButton>download playlist after last view</ButtonSectionButton>
          </ButtonSection>

          <TorrentFilesSection>
            <SectionTitle>Torrent Content</SectionTitle>
          </TorrentFilesSection>
        </DialogContentGrid>
      )}
    </>
  )
}

// function getPreload(torrent) {
//   if (torrent.preloaded_bytes > 0 && torrent.preload_size > 0 && torrent.preloaded_bytes < torrent.preload_size) {
//     const progress = ((torrent.preloaded_bytes * 100) / torrent.preload_size).toFixed(2)
//     return `${humanizeSize(torrent.preloaded_bytes)} / ${humanizeSize(torrent.preload_size)}   ${progress}%`
//   }

//   if (!torrent.preloaded_bytes) return humanizeSize(0)

//   return humanizeSize(torrent.preloaded_bytes)
// }

// function remViews(hash) {
//   try {
//     if (hash)
//       fetch(viewedHost(), {
//         method: 'post',
//         body: JSON.stringify({ action: 'rem', hash, file_index: -1 }),
//         headers: {
//           Accept: 'application/json, text/plain, */*',
//           'Content-Type': 'application/json',
//         },
//       })
//   } catch (e) {
//     console.error(e)
//   }
// }

// function getViewed(hash, callback) {
//   try {
//     fetch(viewedHost(), {
//       method: 'post',
//       body: JSON.stringify({ action: 'list', hash }),
//       headers: {
//         Accept: 'application/json, text/plain, */*',
//         'Content-Type': 'application/json',
//       },
//     })
//       .then(res => res.json())
//       .then(callback)
//   } catch (e) {
//     console.error(e)
//   }
// }

// function getPlayableFile(torrent) {
//   if (!torrent || !torrent.file_stats) return null
//   return torrent.file_stats.filter(file => extPlayable.includes(getExt(file.path)))
// }

// function getExt(filename) {
//   const ext = filename.split('.').pop()
//   if (ext === filename) return ''
//   return ext.toLowerCase()
// }
// const extPlayable = [
//   // video
//   '3g2',
//   '3gp',
//   'aaf',
//   'asf',
//   'avchd',
//   'avi',
//   'drc',
//   'flv',
//   'iso',
//   'm2v',
//   'm2ts',
//   'm4p',
//   'm4v',
//   'mkv',
//   'mng',
//   'mov',
//   'mp2',
//   'mp4',
//   'mpe',
//   'mpeg',
//   'mpg',
//   'mpv',
//   'mxf',
//   'nsv',
//   'ogg',
//   'ogv',
//   'ts',
//   'qt',
//   'rm',
//   'rmvb',
//   'roq',
//   'svi',
//   'vob',
//   'webm',
//   'wmv',
//   'yuv',
//   // audio
//   'aac',
//   'aiff',
//   'ape',
//   'au',
//   'flac',
//   'gsm',
//   'it',
//   'm3u',
//   'm4a',
//   'mid',
//   'mod',
//   'mp3',
//   'mpa',
//   'pls',
//   'ra',
//   's3m',
//   'sid',
//   'wav',
//   'wma',
//   'xm',
// ]

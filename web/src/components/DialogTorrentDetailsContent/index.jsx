import { NoImageIcon } from 'icons'
import { getPeerString, humanizeSize } from 'utils/Utils'
import { CopyToClipboard } from 'react-copy-to-clipboard'
import { useEffect, useState } from 'react'
import { Button, ButtonGroup, Typography } from '@material-ui/core'
import {
  ArrowDownward as ArrowDownwardIcon,
  ArrowUpward as ArrowUpwardIcon,
  SwapVerticalCircle as SwapVerticalCircleIcon,
  ViewAgenda as ViewAgendaIcon,
  Cached as CachedIcon,
} from '@material-ui/icons'
import axios from 'axios'
import { streamHost, torrentsHost, viewedHost } from 'utils/Hosts'
import { GETTING_INFO, IN_DB } from 'torrentStates'

import { useUpdateCache, useCreateCacheMap, useGetSettings } from './customHooks'
import DialogHeader from './DialogHeader'
import TorrentCache from './TorrentCache'
import {
  DetailedTorrentCacheViewWrapper,
  DialogContentGrid,
  MainSection,
  MainSectionButtonGroup,
  Poster,
  SectionTitle,
  SectionSubName,
  StatisticsWrapper,
  ButtonSection,
  LoadingProgress,
  SectionHeader,
  CacheSection,
  ButtonSectionButton,
  TorrentFilesSection,
  Divider,
  SmallLabel,
} from './style'
import StatisticsField from './StatisticsField'

const shortenText = (text, count) => text.slice(0, count) + (text.length > count ? '...' : '')

export default function DialogTorrentDetailsContent({ closeDialog, torrent }) {
  const [isLoading, setIsLoading] = useState(true)
  const [isDetailedCacheView, setIsDetailedCacheView] = useState(false)
  const [viewedFileList, setViewedFileList] = useState()
  const [playableFileList, setPlayableFileList] = useState()

  const {
    poster,
    hash,
    title,
    name,
    stat,
    download_speed: downloadSpeed,
    upload_speed: uploadSpeed,
    stat_string: statString,
    torrent_size: torrentSize,
    file_stats: torrentFileList,
  } = torrent

  const cache = useUpdateCache(hash)
  const cacheMap = useCreateCacheMap(cache)
  const settings = useGetSettings(cache)

  const dropTorrent = () => axios.post(torrentsHost(), { action: 'drop', hash })
  const removeTorrentViews = () =>
    axios.post(viewedHost(), { action: 'rem', hash, file_index: -1 }).then(() => setViewedFileList())
  const preloadBuffer = fileId => fetch(`${streamHost()}?link=${hash}&index=${fileId}&preload`)
  const getFileLink = (path, id) =>
    `${streamHost()}/${encodeURIComponent(path.split('\\').pop().split('/').pop())}?link=${hash}&index=${id}&play`

  const { Capacity, PiecesCount, PiecesLength, Filled } = cache

  useEffect(() => {
    setPlayableFileList(torrentFileList?.filter(file => playableExtList.includes(getExt(file.path))))
  }, [torrentFileList])

  useEffect(() => {
    const cacheLoaded = !!Object.entries(cache).length
    const torrentLoaded = stat !== GETTING_INFO && stat !== IN_DB

    if (!cacheLoaded && !isLoading) setIsLoading(true)
    if (cacheLoaded && isLoading && torrentLoaded) setIsLoading(false)
  }, [stat, cache, isLoading])

  useEffect(() => {
    // getting viewed file list
    axios.post(viewedHost(), { action: 'list', hash }).then(({ data }) => {
      if (data) {
        const lst = data.map(itm => itm.file_index)
        setViewedFileList(lst)
      } else setViewedFileList()
    })
  }, [hash])

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
          <MainSection>
            <Poster poster={poster}>{poster ? <img alt='poster' src={poster} /> : <NoImageIcon />}</Poster>

            <div>
              {name && name !== title ? (
                <>
                  <SectionTitle>{shortenText(name, 50)}</SectionTitle>
                  <SectionSubName mb={20}>{shortenText(title, 160)}</SectionSubName>
                </>
              ) : (
                <SectionTitle mb={20}>{shortenText(title, 50)}</SectionTitle>
              )}

              <StatisticsWrapper>
                <StatisticsField
                  title='Download speed'
                  value={humanizeSize(downloadSpeed) || '0 B'}
                  iconBg='#118f00'
                  valueBg='#13a300'
                  icon={ArrowDownwardIcon}
                />

                <StatisticsField
                  title='Upload speed'
                  value={humanizeSize(uploadSpeed) || '0 B'}
                  iconBg='#0146ad'
                  valueBg='#0058db'
                  icon={ArrowUpwardIcon}
                />

                <StatisticsField
                  title='Peers'
                  value={getPeerString(torrent)}
                  iconBg='#0146ad'
                  valueBg='#0058db'
                  icon={SwapVerticalCircleIcon}
                />

                <StatisticsField
                  title='Torrent size'
                  value={humanizeSize(torrentSize)}
                  iconBg='#0146ad'
                  valueBg='#0058db'
                  icon={ViewAgendaIcon}
                />
              </StatisticsWrapper>

              <Divider />

              <SmallLabel>Download Playlist</SmallLabel>
              <MainSectionButtonGroup>
                <Button variant='contained' color='primary' size='large'>
                  full
                </Button>
                <Button variant='contained' color='primary' size='large'>
                  latest
                </Button>
              </MainSectionButtonGroup>

              <SmallLabel>More</SmallLabel>
              <MainSectionButtonGroup>
                <Button variant='contained' color='primary' size='large'>
                  copy hash
                </Button>
                <Button variant='contained' color='primary' size='large'>
                  remove views
                </Button>
                <Button variant='contained' color='primary' size='large'>
                  drop torrent
                </Button>
              </MainSectionButtonGroup>

              {/* <MainSectionButtonGroup>
                <Button variant='contained' color='primary' size='large'>
                  copy hash
                </Button>
                <Button variant='contained' color='primary' size='large'>
                  remove views
                </Button>
                <Button variant='contained' color='primary' size='large'>
                  drop torrent
                </Button>
                <Button variant='contained' color='primary' size='large'>
                  download full playlist
                </Button>
                <Button variant='contained' color='primary' size='large'>
                  download playlist after last view
                </Button>
              </MainSectionButtonGroup> */}
            </div>
          </MainSection>

          <CacheSection>
            <SectionHeader>
              <SectionTitle mb={20}>Buffer</SectionTitle>
              {!settings?.PreloadBuffer && (
                <SectionSubName>Enable &quot;Preload Buffer&quot; in settings to change buffer size</SectionSubName>
              )}
              <LoadingProgress
                value={Filled}
                fullAmount={bufferSize}
                label={`${humanizeSize(Filled) || '0 B'} / ${humanizeSize(bufferSize)}`}
              />
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

            <ButtonSectionButton onClick={() => removeTorrentViews()}>remove views</ButtonSectionButton>

            <ButtonSectionButton onClick={() => dropTorrent()}>drop torrent</ButtonSectionButton>

            <ButtonSectionButton>download full playlist</ButtonSectionButton>

            <ButtonSectionButton>download playlist after last view</ButtonSectionButton>
          </ButtonSection>

          <TorrentFilesSection>
            <SectionTitle mb={20}>Torrent Content</SectionTitle>

            {!playableFileList?.length
              ? 'No playable files in this torrent'
              : playableFileList.map(({ id, path, length }) => (
                  <ButtonGroup key={id} disableElevation variant='contained' color='primary'>
                    <Button>
                      <a href={getFileLink(path, id)}>
                        <Typography>
                          {path.split('\\').pop().split('/').pop()} | {humanizeSize(length)}{' '}
                          {viewedFileList && viewedFileList?.indexOf(id) !== -1 && '| ✓'}
                        </Typography>
                      </a>
                    </Button>

                    <Button onClick={() => preloadBuffer(id)}>
                      <CachedIcon />
                      <Typography>Preload</Typography>
                    </Button>
                  </ButtonGroup>
                ))}
          </TorrentFilesSection>
        </DialogContentGrid>
      )}
    </>
  )
}

function getExt(filename) {
  const ext = filename.split('.').pop()
  if (ext === filename) return ''
  return ext.toLowerCase()
}
const playableExtList = [
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

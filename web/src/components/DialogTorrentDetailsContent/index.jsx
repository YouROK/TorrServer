import { NoImageIcon } from 'icons'
import { getPeerString, humanizeSize } from 'utils/Utils'
import { CopyToClipboard } from 'react-copy-to-clipboard'
import { useEffect, useState } from 'react'
import { Button, ButtonGroup, Typography } from '@material-ui/core'
import ptt from 'parse-torrent-title'
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
  LoadingProgress,
  SectionHeader,
  CacheSection,
  TorrentFilesSection,
  Divider,
  SmallLabel,
  Table,
} from './style'
import StatisticsField from './StatisticsField'

ptt.addHandler('part', /Part[. ]([0-9])/i, { type: 'integer' })

const shortenText = (text, count) => text.slice(0, count) + (text.length > count ? '...' : '')

export default function DialogTorrentDetailsContent({ closeDialog, torrent }) {
  const [isLoading, setIsLoading] = useState(true)
  const [isDetailedCacheView, setIsDetailedCacheView] = useState(false)
  const [viewedFileList, setViewedFileList] = useState()
  const [playableFileList, setPlayableFileList] = useState()

  const isOnlyOnePlayableFile = playableFileList?.length === 1
  const latestViewedFileId = viewedFileList?.[viewedFileList?.length - 1]
  const latestViewedFile = playableFileList?.find(({ id }) => id === latestViewedFileId)?.path
  const latestViewedFileData = latestViewedFile && ptt.parse(latestViewedFile)

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
        const lst = data.map(itm => itm.file_index).sort((a, b) => a - b)
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
                  iconBg='#cdc118'
                  valueBg='#d8cb18'
                  icon={SwapVerticalCircleIcon}
                />

                <StatisticsField
                  title='Torrent size'
                  value={humanizeSize(torrentSize)}
                  iconBg='#9b01ad'
                  valueBg='#ac03bf'
                  icon={ViewAgendaIcon}
                />
              </StatisticsWrapper>

              <Divider />

              {!isOnlyOnePlayableFile && !!viewedFileList?.length && (
                <>
                  <SmallLabel>Download Playlist</SmallLabel>
                  <SectionSubName mb={10}>
                    <strong>Latest file played:</strong> {latestViewedFileData.title}.
                    {latestViewedFileData.season && (
                      <>
                        {' '}
                        Season: {latestViewedFileData.season}. Episode: {latestViewedFileData.episode}.
                      </>
                    )}
                  </SectionSubName>

                  <MainSectionButtonGroup>
                    <Button variant='contained' color='primary' size='large'>
                      full
                    </Button>
                    <Button variant='contained' color='primary' size='large'>
                      from latest file
                    </Button>
                  </MainSectionButtonGroup>
                </>
              )}

              <SmallLabel mb={10}>Torrent State</SmallLabel>

              <MainSectionButtonGroup>
                <Button onClick={() => removeTorrentViews()} variant='contained' color='primary' size='large'>
                  remove views
                </Button>
                <Button onClick={() => dropTorrent()} variant='contained' color='primary' size='large'>
                  drop torrent
                </Button>
              </MainSectionButtonGroup>

              <SmallLabel mb={10}>Info</SmallLabel>

              <MainSectionButtonGroup>
                {(isOnlyOnePlayableFile || !viewedFileList?.length) && (
                  <Button variant='contained' color='primary' size='large'>
                    download playlist
                  </Button>
                )}
                <CopyToClipboard text={hash}>
                  <Button variant='contained' color='primary' size='large'>
                    copy hash
                  </Button>
                </CopyToClipboard>
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

          <TorrentFilesSection>
            <SectionTitle mb={20}>Torrent Content</SectionTitle>

            <Table>
              <thead>
                <tr>
                  <th style={{ width: '0' }}>viewed</th>
                  <th>name</th>
                  <th style={{ width: '0' }}>season</th>
                  <th style={{ width: '0' }}>episode</th>
                  <th style={{ width: '0' }}>resolution</th>
                  <th style={{ width: '100px' }}>size</th>
                  <th style={{ width: '400px' }}>actions</th>
                </tr>
              </thead>

              <tbody>
                <tr className='viewed-file-row'>
                  <td className='viewed-file-indicator' />
                  <td>Jupiters Legacy</td>
                  <td>3</td>
                  <td>1</td>
                  <td>1080p</td>
                  <td>945,41 MB</td>
                  <td className='button-cell'>
                    <Button variant='outlined' color='primary' size='small'>
                      Preload
                    </Button>
                    <Button variant='outlined' color='primary' size='small'>
                      Open link
                    </Button>
                    <Button variant='outlined' color='primary' size='small'>
                      Copy link
                    </Button>
                  </td>
                </tr>

                <tr className='viewed-file-row'>
                  <td className='viewed-file-indicator' />
                  <td>Jupiters Legacy</td>
                  <td>3</td>
                  <td>2</td>
                  <td>1080p</td>
                  <td>712,47 MB</td>
                  <td className='button-cell'>
                    <Button variant='outlined' color='primary' size='small'>
                      Preload
                    </Button>
                    <Button variant='outlined' color='primary' size='small'>
                      Open link
                    </Button>
                    <Button variant='outlined' color='primary' size='small'>
                      Copy link
                    </Button>
                  </td>
                </tr>

                <tr>
                  <td />
                  <td>Jupiters Legacy</td>
                  <td>3</td>
                  <td>3</td>
                  <td>1080p</td>
                  <td>687,44 MB</td>
                  <td className='button-cell'>
                    <Button variant='outlined' color='primary' size='small'>
                      Preload
                    </Button>
                    <Button variant='outlined' color='primary' size='small'>
                      Open link
                    </Button>
                    <Button variant='outlined' color='primary' size='small'>
                      Copy link
                    </Button>
                  </td>
                </tr>
              </tbody>
            </Table>

            {!playableFileList?.length
              ? 'No playable files in this torrent'
              : playableFileList.map(({ id, path, length }) => {
                  {
                    /* console.log(ptt.parse(path)) */
                  }
                  {
                    /* console.log({ title: ptt.parse(path).title })
                  console.log({ resolution: ptt.parse(path).resolution })
                  console.log({ episode: ptt.parse(path).episode })
                  console.log({ season: ptt.parse(path).season }) */
                  }

                  return (
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
                  )
                })}
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

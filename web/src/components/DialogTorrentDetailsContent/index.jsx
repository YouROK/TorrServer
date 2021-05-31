import { NoImageIcon } from 'icons'
import { humanizeSize } from 'utils/Utils'
import { CopyToClipboard } from 'react-copy-to-clipboard'
import { useEffect, useState } from 'react'
import { Button } from '@material-ui/core'
import ptt from 'parse-torrent-title'
import axios from 'axios'
import { playlistTorrHost, streamHost, torrentsHost, viewedHost } from 'utils/Hosts'
import { GETTING_INFO, IN_DB } from 'torrentStates'
import CircularProgress from '@material-ui/core/CircularProgress'

import { useUpdateCache, useCreateCacheMap, useGetSettings } from './customHooks'
import DialogHeader from './DialogHeader'
import TorrentCache from './TorrentCache'
import {
  DetailedViewWrapper,
  DetailedViewWidgetSection,
  DetailedViewCacheSection,
  DialogContentGrid,
  MainSection,
  MainSectionButtonGroup,
  Poster,
  SectionTitle,
  SectionSubName,
  WidgetWrapper,
  LoadingProgress,
  SectionHeader,
  CacheSection,
  TorrentFilesSection,
  Divider,
  SmallLabel,
  Table,
} from './style'
import {
  DownlodSpeedWidget,
  UploadSpeedWidget,
  PeersWidget,
  SizeWidget,
  PiecesCountWidget,
  PiecesLengthWidget,
  StatusWidget,
} from './widgets'

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
  const fullPlaylistLink = `${playlistTorrHost()}/${encodeURIComponent(name || title || 'file')}.m3u?link=${hash}&m3u`
  const partialPlaylistLink = `${fullPlaylistLink}&fromlast`

  const fileHasEpisodeText = !!playableFileList?.find(({ path }) => ptt.parse(path).episode)
  const fileHasSeasonText = !!playableFileList?.find(({ path }) => ptt.parse(path).season)
  const fileHasResolutionText = !!playableFileList?.find(({ path }) => ptt.parse(path).resolution)

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
        {...(isDetailedCacheView && { onBack: () => setIsDetailedCacheView(false) })}
      />

      <div style={{ minHeight: '80vh' }}>
        {isLoading ? (
          <div style={{ minHeight: '80vh', display: 'grid', placeItems: 'center' }}>
            <CircularProgress />
          </div>
        ) : isDetailedCacheView ? (
          <DetailedViewWrapper>
            <DetailedViewWidgetSection>
              <SectionTitle mb={20}>Data</SectionTitle>
              <WidgetWrapper detailedView>
                <DownlodSpeedWidget data={downloadSpeed} />
                <UploadSpeedWidget data={uploadSpeed} />
                <PeersWidget data={torrent} />
                <SizeWidget data={torrentSize} />
                <PiecesCountWidget data={PiecesCount} />
                <PiecesLengthWidget data={PiecesLength} />
                <StatusWidget data={statString} />
              </WidgetWrapper>
            </DetailedViewWidgetSection>

            <DetailedViewCacheSection>
              <SectionTitle mb={20}>Cache</SectionTitle>
              <TorrentCache cache={cache} cacheMap={cacheMap} />
            </DetailedViewCacheSection>
          </DetailedViewWrapper>
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

                <WidgetWrapper>
                  <DownlodSpeedWidget data={downloadSpeed} />
                  <UploadSpeedWidget data={uploadSpeed} />
                  <PeersWidget data={torrent} />
                  <SizeWidget data={torrentSize} />
                </WidgetWrapper>

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
                      <a style={{ textDecoration: 'none' }} href={fullPlaylistLink}>
                        <Button style={{ width: '100%' }} variant='contained' color='primary' size='large'>
                          full
                        </Button>
                      </a>

                      <a style={{ textDecoration: 'none' }} href={partialPlaylistLink}>
                        <Button style={{ width: '100%' }} variant='contained' color='primary' size='large'>
                          from latest file
                        </Button>
                      </a>
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
                    <a style={{ textDecoration: 'none' }} href={fullPlaylistLink}>
                      <Button style={{ width: '100%' }} variant='contained' color='primary' size='large'>
                        download playlist
                      </Button>
                    </a>
                  )}
                  <CopyToClipboard text={hash}>
                    <Button variant='contained' color='primary' size='large'>
                      copy hash
                    </Button>
                  </CopyToClipboard>
                </MainSectionButtonGroup>
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

              {!playableFileList?.length ? (
                'No playable files in this torrent'
              ) : (
                <>
                  <Table>
                    <thead>
                      <tr>
                        <th style={{ width: '0' }}>viewed</th>
                        <th>name</th>
                        {fileHasSeasonText && <th style={{ width: '0' }}>season</th>}
                        {fileHasEpisodeText && <th style={{ width: '0' }}>episode</th>}
                        {fileHasResolutionText && <th style={{ width: '0' }}>resolution</th>}
                        <th style={{ width: '100px' }}>size</th>
                        <th style={{ width: '400px' }}>actions</th>
                      </tr>
                    </thead>

                    <tbody>
                      {playableFileList.map(({ id, path, length }) => {
                        const { title, resolution, episode, season } = ptt.parse(path)
                        const isViewed = viewedFileList?.includes(id)
                        const link = getFileLink(path, id)

                        return (
                          <tr key={id} className={isViewed ? 'viewed-file-row' : null}>
                            <td className={isViewed ? 'viewed-file-indicator' : null} />
                            <td>{title}</td>
                            {fileHasSeasonText && <td>{season}</td>}
                            {fileHasEpisodeText && <td>{episode}</td>}
                            {fileHasResolutionText && <td>{resolution}</td>}
                            <td>{humanizeSize(length)}</td>
                            <td className='button-cell'>
                              <Button onClick={() => preloadBuffer(id)} variant='outlined' color='primary' size='small'>
                                Preload
                              </Button>

                              <a style={{ textDecoration: 'none' }} href={link} target='_blank' rel='noreferrer'>
                                <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                                  Open link
                                </Button>
                              </a>

                              <CopyToClipboard text={link}>
                                <Button variant='outlined' color='primary' size='small'>
                                  Copy link
                                </Button>
                              </CopyToClipboard>
                            </td>
                          </tr>
                        )
                      })}
                    </tbody>
                  </Table>
                </>
              )}
            </TorrentFilesSection>
          </DialogContentGrid>
        )}
      </div>
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

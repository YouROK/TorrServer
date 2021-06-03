import axios from 'axios'
import { memo } from 'react'
import { playlistTorrHost, torrentsHost, viewedHost } from 'utils/Hosts'
import { CopyToClipboard } from 'react-copy-to-clipboard'
import { Button } from '@material-ui/core'
import ptt from 'parse-torrent-title'

import { SmallLabel, MainSectionButtonGroup } from './style'
import { SectionSubName } from '../style'

const TorrentFunctions = memo(
  ({ hash, viewedFileList, playableFileList, name, title, setViewedFileList }) => {
    const latestViewedFileId = viewedFileList?.[viewedFileList?.length - 1]
    const latestViewedFile = playableFileList?.find(({ id }) => id === latestViewedFileId)?.path
    const isOnlyOnePlayableFile = playableFileList?.length === 1
    const latestViewedFileData = latestViewedFile && ptt.parse(latestViewedFile)
    const dropTorrent = () => axios.post(torrentsHost(), { action: 'drop', hash })
    const removeTorrentViews = () =>
      axios.post(viewedHost(), { action: 'rem', hash, file_index: -1 }).then(() => setViewedFileList())
    const fullPlaylistLink = `${playlistTorrHost()}/${encodeURIComponent(name || title || 'file')}.m3u?link=${hash}&m3u`
    const partialPlaylistLink = `${fullPlaylistLink}&fromlast`

    return (
      <>
        {!isOnlyOnePlayableFile && !!viewedFileList?.length && (
          <>
            <SmallLabel>Download Playlist</SmallLabel>
            <SectionSubName mb={10}>
              <strong>Latest file played:</strong> {latestViewedFileData?.title}.
              {latestViewedFileData?.season && (
                <>
                  {' '}
                  Season: {latestViewedFileData?.season}. Episode: {latestViewedFileData?.episode}.
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
            reset torrent
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
      </>
    )
  },
  () => true,
)

export default TorrentFunctions

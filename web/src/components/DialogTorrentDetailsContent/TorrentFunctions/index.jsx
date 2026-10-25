import axios from 'axios'
import { memo, useEffect, useState } from 'react'
import { playlistTorrHost, torrentsHost, viewedHost } from 'utils/Hosts'
import { CopyToClipboard } from 'react-copy-to-clipboard'
import { Button, Dialog, DialogActions, DialogTitle } from '@material-ui/core'
import ptt from 'parse-torrent-title'
import { useTranslation } from 'react-i18next'

import { SmallLabel, MainSectionButtonGroup } from './style'
import { SectionSubName } from '../style'

const TorrentFunctions = memo(
  ({
    hash,
    viewedFileList,
    playableFileList,
    name,
    title,
    setViewedFileList,
    offlineState,
    offlineError,
    offlinePinned,
    loadedSize,
    torrentSize,
  }) => {
    const { t } = useTranslation()
    const [offline, setOffline] = useState(offlineState || '')
    const [offlineBusy, setOfflineBusy] = useState(false)
    const [pinned, setPinned] = useState(!!offlinePinned)
    useEffect(() => setPinned(!!offlinePinned), [offlinePinned])
    const [confirmState, setConfirmState] = useState(null)
    const [errorText, setErrorText] = useState('')
    useEffect(() => setOffline(offlineState || ''), [offlineState])
    const offlinePercent = torrentSize > 0 ? Math.min(100, Math.floor(((loadedSize || 0) * 100) / torrentSize)) : 0
    // window.confirm/window.alert are replaced with the application dialogs
    // (Material-UI, localized) - review remark N-6
    const callOffline = (body, confirmText) => {
      const run = () => {
        setOfflineBusy(true)
        axios
          .post(torrentsHost(), { hash, ...body })
          .then(({ data }) => {
            setOffline(data?.offline_state || '')
            setPinned(!!data?.offline_pinned)
          })
          .catch(e => setErrorText(e?.response?.data?.error || e.message))
          .finally(() => setOfflineBusy(false))
      }
      if (confirmText) {
        setConfirmState({ body: confirmText, onConfirm: run })
        return
      }
      run()
    }
    const enableOffline = () => callOffline({ action: 'offline', enable: true })
    const removeOffline = () =>
      callOffline(
        { action: 'offline', enable: false },
        t('OfflineRemoveConfirm', 'Delete the offline copy of this torrent from the disk?'),
      )
    const pauseOffline = () => callOffline({ action: 'offline_pause' })
    const resumeOffline = () => callOffline({ action: 'offline_resume' })
    // a pinned copy is never removed by the disk quota rotation
    const togglePin = () => callOffline({ action: 'offline_pin', enable: !pinned })
    const latestViewedFileId = viewedFileList?.[viewedFileList?.length - 1]
    const latestViewedFile = playableFileList?.find(({ id }) => id === latestViewedFileId)?.path
    const isOnlyOnePlayableFile = playableFileList?.length === 1
    const latestViewedFileData = latestViewedFile && ptt.parse(latestViewedFile)
    const dropTorrent = () => axios.post(torrentsHost(), { action: 'drop', hash })
    const removeTorrentViews = () =>
      axios.post(viewedHost(), { action: 'rem', hash, file_index: -1 }).then(() => setViewedFileList())
    const fullPlaylistLink = `${playlistTorrHost()}/${encodeURIComponent(name || title || 'file')}.m3u?link=${hash}&m3u`
    const partialPlaylistLink = `${fullPlaylistLink}&fromlast`
    const magnet = `magnet:?xt=urn:btih:${hash}&dn=${encodeURIComponent(name || title)}`

    return (
      <>
        {!isOnlyOnePlayableFile && !!viewedFileList?.length && (
          <>
            <SmallLabel>{t('DownloadPlaylist')}</SmallLabel>
            <SectionSubName mb={10}>
              {t('LatestFilePlayed')}{' '}
              <strong>
                {latestViewedFileData?.title}.
                {latestViewedFileData?.season && (
                  <>
                    {' '}
                    {t('Season')}: {latestViewedFileData?.season}. {t('Episode')}: {latestViewedFileData?.episode}.
                  </>
                )}
              </strong>
            </SectionSubName>

            <MainSectionButtonGroup>
              <a style={{ textDecoration: 'none' }} href={fullPlaylistLink}>
                <Button style={{ width: '100%' }} variant='contained' color='primary' size='large'>
                  {t('Full')}
                </Button>
              </a>

              <a style={{ textDecoration: 'none' }} href={partialPlaylistLink}>
                <Button style={{ width: '100%' }} variant='contained' color='primary' size='large'>
                  {t('FromLatestFile')}
                </Button>
              </a>
            </MainSectionButtonGroup>
          </>
        )}
        <SmallLabel mb={10}>{t('OfflineMode', 'Offline mode')}</SmallLabel>
        {!!offline && (
          <SectionSubName mb={10}>
            {offline === 'ready' &&
              t('OfflineReady', 'Downloaded completely. Playback does not use the torrent network.')}
            {offline === 'downloading' &&
              `${t('OfflineDownloading', 'Downloading for offline viewing')}: ${offlinePercent}%`}
            {offline === 'paused' && `${t('OfflinePaused', 'Download paused')}: ${offlinePercent}%`}
          </SectionSubName>
        )}
        {offline === 'paused' && (
          <SectionSubName mb={10}>{t('OfflinePausedStreamHint', 'While the download is paused, playback streams from the torrent network; the part already saved on disk is used once the download is resumed or finished.')}</SectionSubName>
        )}
        {!!offlineError && <SectionSubName mb={10}>{offlineError}</SectionSubName>}
        <MainSectionButtonGroup>
          {!offline && (
            <Button
              disabled={offlineBusy}
              onClick={enableOffline}
              variant='contained'
              color='primary'
              size='large'
            >
              {t('OfflineEnable', 'Download for offline viewing')}
            </Button>
          )}
          {offline === 'downloading' && (
            <Button
              disabled={offlineBusy}
              onClick={pauseOffline}
              variant='contained'
              color='primary'
              size='large'
            >
              {t('OfflinePause', 'Pause offline download')}
            </Button>
          )}
          {offline === 'paused' && (
            <Button
              disabled={offlineBusy}
              onClick={resumeOffline}
              variant='contained'
              color='primary'
              size='large'
            >
              {t('OfflineResume', 'Resume offline download')}
            </Button>
          )}
          {!!offline && (
            <Button disabled={offlineBusy} onClick={togglePin} variant='contained' color='primary' size='large'>
              {pinned
                ? t('OfflineUnpin', 'Unpin: allow rotation to remove')
                : t('OfflinePin', 'Pin: never remove automatically')}
            </Button>
          )}
          {!!offline && (
            <Button
              disabled={offlineBusy}
              onClick={removeOffline}
              variant='contained'
              color='primary'
              size='large'
            >
              {offline === 'ready'
                ? t('OfflineRemove', 'Remove offline copy')
                : t('OfflineCancel', 'Cancel offline download')}
            </Button>
          )}
        </MainSectionButtonGroup>
        <SmallLabel mb={10}>{t('TorrentState')}</SmallLabel>
        <MainSectionButtonGroup>
          <Button onClick={() => removeTorrentViews()} variant='contained' color='primary' size='large'>
            {t('RemoveViews')}
          </Button>
          <Button onClick={() => dropTorrent()} variant='contained' color='primary' size='large'>
            {t('DropTorrent')}
          </Button>
        </MainSectionButtonGroup>
        <SmallLabel mb={10}>{t('Info')}</SmallLabel>
        <MainSectionButtonGroup>
          {(isOnlyOnePlayableFile || !viewedFileList?.length) && (
            <a style={{ textDecoration: 'none' }} href={fullPlaylistLink}>
              <Button style={{ width: '100%' }} variant='contained' color='primary' size='large'>
                {t('DownloadPlaylist')}
              </Button>
            </a>
          )}
          <CopyToClipboard text={magnet}>
            <Button variant='contained' color='primary' size='large'>
              {t('CopyHash')}
            </Button>
          </CopyToClipboard>
        </MainSectionButtonGroup>

        <Dialog open={!!confirmState} onClose={() => setConfirmState(null)}>
          <DialogTitle>{confirmState?.body}</DialogTitle>
          <DialogActions>
            <Button variant='outlined' color='secondary' onClick={() => setConfirmState(null)}>
              {t('Cancel')}
            </Button>
            <Button
              variant='contained'
              color='primary'
              onClick={() => {
                const { onConfirm } = confirmState || {}
                setConfirmState(null)
                onConfirm?.()
              }}
            >
              {t('OK')}
            </Button>
          </DialogActions>
        </Dialog>

        <Dialog open={!!errorText} onClose={() => setErrorText('')}>
          <DialogTitle>{errorText}</DialogTitle>
          <DialogActions>
            <Button variant='contained' color='primary' onClick={() => setErrorText('')}>
              {t('OK')}
            </Button>
          </DialogActions>
        </Dialog>
      </>
    )
  },
  // Compare every prop (review remark N-7): the first version compared only
  // the offline fields, so a change of viewedFileList / playableFileList /
  // name / title did not re-render the block (playlists, latest file,
  // magnet). The lists are small, JSON.stringify is cheap for them.
  (prev, next) =>
    prev.hash === next.hash &&
    prev.name === next.name &&
    prev.title === next.title &&
    prev.setViewedFileList === next.setViewedFileList &&
    prev.offlineState === next.offlineState &&
    prev.offlineError === next.offlineError &&
    prev.loadedSize === next.loadedSize &&
    prev.torrentSize === next.torrentSize &&
    JSON.stringify(prev.viewedFileList) === JSON.stringify(next.viewedFileList) &&
    JSON.stringify(prev.playableFileList) === JSON.stringify(next.playableFileList),
)

export default TorrentFunctions

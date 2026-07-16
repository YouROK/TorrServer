import { streamHost } from 'utils/Hosts'
import isEqual from 'lodash/isEqual'
import { humanizeSize, detectStandaloneApp, isMacOS, isAppleDevice } from 'utils/Utils'
import ptt from 'parse-torrent-title'
import { Button } from '@material-ui/core'
import CopyToClipboard from 'react-copy-to-clipboard'
import { useTranslation } from 'react-i18next'
import {
  gstreamerHeartbeatUrl,
  gstreamerMasterUrl,
  shouldUseGStreamerPlayer,
  useGStreamerRuntime,
} from 'utils/GStreamer'

import VideoPlayer from '../../VideoPlayer'
import { TableStyle, ShortTableWrapper, ShortTable } from './style'

const { memo, useState } = require('react')

// russian episode detection support
ptt.addHandler('episode', /(\d{1,4})[- |. ]серия|серия[- |. ](\d{1,4})/i, {
  type: 'integer',
})
ptt.addHandler('season', /sezon[- |. ](\d{1,3})|(\d{1,3})[- |. ]sezon/i, {
  type: 'integer',
})
ptt.addHandler('season', /сезон[- |. ](\d{1,3})|(\d{1,3})[- |. ]сезон/i, {
  type: 'integer',
})

const Table = memo(
  ({ playableFileList, viewedFileList, selectedSeason, seasonAmount, hash }) => {
    const { t } = useTranslation()
    const [unsupportedPlayers, setUnsupportedPlayers] = useState({})
    const gstRuntime = useGStreamerRuntime()
    const preloadBuffer = fileId => fetch(`${streamHost()}?link=${hash}&index=${fileId}&preload`)
    const getFileLink = (path, id) =>
      `${streamHost()}/${encodeURIComponent(path.split('\\').pop().split('/').pop())}?link=${hash}&index=${id}&play`
    const getPlayer = (path, id) => {
      const hls = shouldUseGStreamerPlayer(path, gstRuntime)
      return {
        key: `${id}:${hls ? 'gst' : 'stream'}`,
        src: hls ? gstreamerMasterUrl(hash, id) : getFileLink(path, id),
        hls,
        heartbeatSrc: hls ? gstreamerHeartbeatUrl(hash) : '',
      }
    }
    const markPlayerUnsupported = key => {
      setUnsupportedPlayers(current => ({ ...current, [key]: true }))
    }
    const fileHasEpisodeText = !!playableFileList?.find(({ path }) => ptt.parse(path).episode)
    const fileHasSeasonText = !!playableFileList?.find(({ path }) => ptt.parse(path).season)
    const fileHasResolutionText = !!playableFileList?.find(({ path }) => ptt.parse(path).resolution)

    // if files in list is more then 1 and no season text detected by ptt.parse, show full name
    const shouldDisplayFullFileName = playableFileList?.length > 1 && !fileHasEpisodeText

    const isVlcUsed = JSON.parse(localStorage.getItem('isVlcUsed')) ?? false
    const isInfuseUsed = JSON.parse(localStorage.getItem('isInfuseUsed')) ?? false
    const isSenPlayerUsed = JSON.parse(localStorage.getItem('isSenPlayerUsed')) ?? false
    const isIinaUsed = JSON.parse(localStorage.getItem('isIinaUsed')) ?? false
    const isStandalone = detectStandaloneApp()
    const isMac = isMacOS()
    const isApple = isAppleDevice()
    const shouldShowOpenLink =
      !isStandalone ||
      (!(isApple && isInfuseUsed) && !(isApple && isSenPlayerUsed) && !isVlcUsed && !(isMac && isIinaUsed))

    return !playableFileList?.length ? (
      'No playable files in this torrent'
    ) : (
      <>
        <TableStyle>
          <thead>
            <tr>
              <th style={{ width: '0' }}>{t('Viewed')}</th>
              <th>{t('Name')}</th>
              {fileHasSeasonText && seasonAmount?.length === 1 && <th style={{ width: '0' }}>{t('Season')}</th>}
              {fileHasEpisodeText && <th style={{ width: '0' }}>{t('Episode')}</th>}
              {fileHasResolutionText && <th style={{ width: '0' }}>{t('Resolution')}</th>}
              <th style={{ width: '100px' }}>{t('Size')}</th>
              <th style={{ width: '400px' }}>{t('Actions')}</th>
            </tr>
          </thead>

          <tbody>
            {playableFileList.map(({ id, path, length }) => {
              const { title, resolution, episode, season } = ptt.parse(path)
              const isViewed = viewedFileList?.includes(id)
              const link = getFileLink(path, id)
              const player = getPlayer(path, id)
              const playerSupported = !unsupportedPlayers[player.key]
              const fullLink = new URL(link, window.location.href)
              const infuseLink = `infuse://x-callback-url/play?url=${encodeURIComponent(fullLink)}`
              const senPlayerLink = `senplayer://x-callback-url/play?url=${encodeURIComponent(fullLink)}`
              const iinaLink = `iina://weblink?url=${encodeURIComponent(fullLink)}`

              return (
                (season === selectedSeason || !seasonAmount?.length) && (
                  <tr key={id} className={isViewed ? 'viewed-file-row' : null}>
                    <td data-label='viewed' aria-label='viewed' className={isViewed ? 'viewed-file-indicator' : null} />
                    <td data-label='name'>{shouldDisplayFullFileName ? path : title}</td>
                    {fileHasSeasonText && seasonAmount?.length === 1 && <td data-label='season'>{season}</td>}
                    {fileHasEpisodeText && <td data-label='episode'>{episode}</td>}
                    {fileHasResolutionText && <td data-label='resolution'>{resolution}</td>}
                    <td data-label='size'>{humanizeSize(length)}</td>
                    <td>
                      <div className='button-cell'>
                        <Button onClick={() => preloadBuffer(id)} variant='outlined' color='primary' size='small'>
                          {t('Preload')}
                        </Button>
                        {isApple && isInfuseUsed && (
                          <a style={{ textDecoration: 'none' }} href={infuseLink}>
                            <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                              {t('Infuse')}
                            </Button>
                          </a>
                        )}
                        {isApple && isSenPlayerUsed && (
                          <a style={{ textDecoration: 'none' }} href={senPlayerLink}>
                            <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                              {t('SenPlayer')}
                            </Button>
                          </a>
                        )}
                        {isVlcUsed && (
                          <a style={{ textDecoration: 'none' }} href={`vlc://${fullLink}`}>
                            <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                              VLC
                            </Button>
                          </a>
                        )}
                        {isMac && isIinaUsed && (
                          <a style={{ textDecoration: 'none' }} href={iinaLink}>
                            <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                              IINA
                            </Button>
                          </a>
                        )}
                        {playerSupported ? (
                          <VideoPlayer
                            title={title}
                            videoSrc={player.src}
                            downloadSrc={link}
                            hls={player.hls}
                            heartbeatSrc={player.heartbeatSrc}
                            onNotSupported={() => markPlayerUnsupported(player.key)}
                          />
                        ) : (
                          shouldShowOpenLink && (
                            <a style={{ textDecoration: 'none' }} href={link} target='_blank' rel='noreferrer'>
                              <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                                {t('OpenLink')}
                              </Button>
                            </a>
                          )
                        )}
                        <CopyToClipboard text={fullLink}>
                          <Button variant='outlined' color='primary' size='small'>
                            {t('CopyLink')}
                          </Button>
                        </CopyToClipboard>
                        {playerSupported && shouldShowOpenLink && (
                          <a style={{ textDecoration: 'none' }} href={link} target='_blank' rel='noreferrer'>
                            <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                              {t('OpenLink')}
                            </Button>
                          </a>
                        )}
                      </div>
                    </td>
                  </tr>
                )
              )
            })}
          </tbody>
        </TableStyle>

        <ShortTableWrapper>
          {playableFileList.map(({ id, path, length }) => {
            const { title, resolution, episode, season } = ptt.parse(path)
            const isViewed = viewedFileList?.includes(id)
            const link = getFileLink(path, id)
            const player = getPlayer(path, id)
            const playerSupported = !unsupportedPlayers[player.key]
            const fullLink = new URL(link, window.location.href)
            const infuseLink = `infuse://x-callback-url/play?url=${encodeURIComponent(fullLink)}`
            const senPlayerLink = `senplayer://x-callback-url/play?url=${encodeURIComponent(fullLink)}`
            const iinaLink = `iina://weblink?url=${encodeURIComponent(fullLink)}`

            return (
              (season === selectedSeason || !seasonAmount?.length) && (
                <ShortTable key={id} isViewed={isViewed}>
                  <div className='short-table-name'>{shouldDisplayFullFileName ? path : title}</div>
                  <div className='short-table-data'>
                    {isViewed && (
                      <div className='short-table-field'>
                        <div className='short-table-field-name'>{t('Viewed')}</div>
                        <div className='short-table-field-value'>
                          <div className='short-table-viewed-indicator' />
                        </div>
                      </div>
                    )}
                    {fileHasSeasonText && seasonAmount?.length === 1 && (
                      <div className='short-table-field'>
                        <div className='short-table-field-name'>{t('Season')}</div>
                        <div className='short-table-field-value'>{season}</div>
                      </div>
                    )}
                    {fileHasEpisodeText && (
                      <div className='short-table-field'>
                        <div className='short-table-field-name'>{t('Episode')}</div>
                        <div className='short-table-field-value'>{episode}</div>
                      </div>
                    )}
                    {fileHasResolutionText && (
                      <div className='short-table-field'>
                        <div className='short-table-field-name'>{t('Resolution')}</div>
                        <div className='short-table-field-value'>{resolution}</div>
                      </div>
                    )}
                    <div className='short-table-field'>
                      <div className='short-table-field-name'>{t('Size')}</div>
                      <div className='short-table-field-value'>{humanizeSize(length)}</div>
                    </div>
                  </div>
                  <div className='short-table-buttons'>
                    <Button onClick={() => preloadBuffer(id)} variant='outlined' color='primary' size='small'>
                      {t('Preload')}
                    </Button>

                    {isApple && isInfuseUsed && (
                      <a style={{ textDecoration: 'none' }} href={infuseLink}>
                        <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                          {t('Infuse')}
                        </Button>
                      </a>
                    )}

                    {isApple && isSenPlayerUsed && (
                      <a style={{ textDecoration: 'none' }} href={senPlayerLink}>
                        <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                          {t('SenPlayer')}
                        </Button>
                      </a>
                    )}

                    {isVlcUsed && (
                      <a style={{ textDecoration: 'none' }} href={`vlc://${fullLink}`}>
                        <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                          VLC
                        </Button>
                      </a>
                    )}

                    {isMac && isIinaUsed && (
                      <a style={{ textDecoration: 'none' }} href={iinaLink}>
                        <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                          IINA
                        </Button>
                      </a>
                    )}

                    {player.hls && playerSupported && (
                      <VideoPlayer
                        title={title}
                        videoSrc={player.src}
                        downloadSrc={link}
                        hls
                        heartbeatSrc={player.heartbeatSrc}
                        onNotSupported={() => markPlayerUnsupported(player.key)}
                      />
                    )}

                    {shouldShowOpenLink && (
                      <a style={{ textDecoration: 'none' }} href={link} target='_blank' rel='noreferrer'>
                        <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                          {t('OpenLink')}
                        </Button>
                      </a>
                    )}

                    <CopyToClipboard text={fullLink}>
                      <Button variant='outlined' color='primary' size='small'>
                        {t('CopyLink')}
                      </Button>
                    </CopyToClipboard>
                  </div>
                </ShortTable>
              )
            )
          })}
        </ShortTableWrapper>
      </>
    )
  },
  (prev, next) => isEqual(prev, next),
)

export default Table

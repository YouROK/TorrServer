import { streamHost } from 'utils/Hosts'
import isEqual from 'lodash/isEqual'
import { humanizeSize, detectStandaloneApp, isMacOS, isAppleDevice } from 'utils/Utils'
import ptt from 'parse-torrent-title'
import { Button } from '@material-ui/core'
import CopyToClipboard from 'react-copy-to-clipboard'
import { useTranslation } from 'react-i18next'
import { memo, useState } from 'react'
import QRCodeStyling from 'qr-code-styling'
import { StyledDialog, StyledHeader } from 'style/CustomMaterialUiStyles'

import VideoPlayer from '../../VideoPlayer'
import { TableStyle, ShortTableWrapper, ShortTable } from './style'
import { isFilePlayable } from '../helpers'
import VlcLogoSvg from '../../../images/VLC_logo.svg'
import { useMaterialUITheme } from 'style/materialUISetup'
import { mainColors, themeColors } from 'style/colors'
import { Content } from 'components/Search/style'

window.QrCode = QRCodeStyling;

// russian episode detection support
ptt.addHandler('episode', /(\d{1,4})[- |. ]серия|серия[- |. ](\d{1,4})/i, { type: 'integer' })
ptt.addHandler('season', /sezon[- |. ](\d{1,3})|(\d{1,3})[- |. ]sezon/i, { type: 'integer' })
ptt.addHandler('season', /сезон[- |. ](\d{1,3})|(\d{1,3})[- |. ]сезон/i, { type: 'integer' })

const Table = memo(
  ({ playableFileList, allFileList, viewedFileList, selectedSeason, seasonAmount, hash }) => {
    const { t } = useTranslation()
    const [isDarkMode, currentThemeMode, updateThemeMode, muiTheme] = useMaterialUITheme()
    const [isSupported, setIsSupported] = useState(true)
    const [isQrEnabled, setIsQrEnabled] = useState(false)
    const preloadBuffer = fileId => fetch(`${streamHost()}?link=${hash}&index=${fileId}&preload`)
    const getFileLink = (path, id) =>
      `${streamHost()}/${encodeURIComponent(path.split('\\').pop().split('/').pop())}?link=${hash}&index=${id}&play`

    // Use allFileList if available, otherwise fallback to playableFileList
    const fileListToDisplay = allFileList || playableFileList

    const fileHasEpisodeText = !!fileListToDisplay?.find(({ path }) => ptt.parse(path).episode)
    const fileHasSeasonText = !!fileListToDisplay?.find(({ path }) => ptt.parse(path).season)
    const fileHasResolutionText = !!fileListToDisplay?.find(({ path }) => ptt.parse(path).resolution)

    // if files in list is more then 1 and no season text detected by ptt.parse, show full name
    const shouldDisplayFullFileName = fileListToDisplay?.length > 1 && !fileHasEpisodeText

    // Function to trigger download for non-playable files
    const triggerDownload = (path, id) => {
      const link = getFileLink(path, id)
      const fileName = path.split('\\').pop().split('/').pop()

      // Create a temporary anchor element and trigger download
      const downloadLink = document.createElement('a')
      downloadLink.href = link
      downloadLink.download = fileName
      downloadLink.style.display = 'none'
      document.body.appendChild(downloadLink)
      downloadLink.click()
      document.body.removeChild(downloadLink)
    }

    const isVlcUsed = JSON.parse(localStorage.getItem('isVlcUsed')) ?? false
    const isInfuseUsed = JSON.parse(localStorage.getItem('isInfuseUsed')) ?? false
    const isIinaUsed = JSON.parse(localStorage.getItem('isIinaUsed')) ?? false
    const isStandalone = detectStandaloneApp()
    const isMac = isMacOS()
    const isApple = isAppleDevice()
    const shouldShowOpenLink = !isStandalone || (!(isApple && isInfuseUsed) && !isVlcUsed && !(isMac && isIinaUsed))

    const createVlcQrCode = link => {
      const qrCode = new QRCodeStyling({
        width: 450,
        height: 450,
        type: 'svg',
        data: link,
        image: VlcLogoSvg,
        dotsOptions: {
          color: '#eb6d00',
          type: 'rounded',
        },
        backgroundOptions: {
          color:  themeColors[isDarkMode ? 'dark' : 'light'].app.appSecondaryColor,
        },
        imageOptions: {
          crossOrigin: 'anonymous',
          margin: 20,
        },
      })
      setIsQrEnabled(true)
      setTimeout(() => qrCode.append(document.querySelector('#canvas')))
    }

    return !fileListToDisplay?.length ? (
      'No files in this torrent'
    ) : (
      <>
        <StyledDialog
          open={isQrEnabled}
          onClose={() => setIsQrEnabled(false)}
          fullScreen={false}
          maxWidth='md'
        >
          <StyledHeader>VLC QR Code</StyledHeader>
          <div id='canvas' />
        </StyledDialog>
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
            {fileListToDisplay.map(({ id, path, length }) => {
              const { title, resolution, episode, season } = ptt.parse(path)
              const isViewed = viewedFileList?.includes(id)
              const link = getFileLink(path, id)
              const fullLink = new URL(link, window.location.href)
              const infuseLink = `infuse://x-callback-url/play?url=${encodeURIComponent(fullLink)}`
              const iinaLink = `iina://weblink?url=${encodeURIComponent(fullLink)}`
              const isPlayable = isFilePlayable(path)

              return (
                (season === selectedSeason || !seasonAmount?.length) && (
                  <tr key={id} className={isViewed ? 'viewed-file-row' : null}>
                    <td data-label='viewed' aria-label='viewed' className={isViewed ? 'viewed-file-indicator' : null} />
                    <td data-label='name'>{shouldDisplayFullFileName ? path : title || path}</td>
                    {fileHasSeasonText && seasonAmount?.length === 1 && <td data-label='season'>{season}</td>}
                    {fileHasEpisodeText && <td data-label='episode'>{episode}</td>}
                    {fileHasResolutionText && <td data-label='resolution'>{resolution}</td>}
                    <td data-label='size'>{humanizeSize(length)}</td>
                    <td>
                      <div className='button-cell'>
                        <Button onClick={() => preloadBuffer(id)} variant='outlined' color='primary' size='small'>
                          {t('Preload')}
                        </Button>
                        {isPlayable ? (
                          <>
                            {isApple && isInfuseUsed && (
                              <a style={{ textDecoration: 'none' }} href={infuseLink}>
                                <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                                  {t('Infuse')}
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
                            {isVlcUsed && (
                              <Button
                                style={{ width: '100%' }}
                                variant='outlined'
                                color='primary'
                                size='small'
                                onClick={() => createVlcQrCode(`vlc://${fullLink}`)}
                              >
                                VLC Qr Code
                              </Button>
                            )}
                            {isMac && isIinaUsed && (
                              <a style={{ textDecoration: 'none' }} href={iinaLink}>
                                <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                                  IINA
                                </Button>
                              </a>
                            )}
                            {isSupported ? (
                              <VideoPlayer title={title} videoSrc={link} onNotSupported={() => setIsSupported(false)} />
                            ) : (
                              shouldShowOpenLink && (
                                <a style={{ textDecoration: 'none' }} href={link} target='_blank' rel='noreferrer'>
                                  <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                                    {t('OpenLink')}
                                  </Button>
                                </a>
                              )
                            )}
                            {isSupported && shouldShowOpenLink && (
                              <a style={{ textDecoration: 'none' }} href={link} target='_blank' rel='noreferrer'>
                                <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                                  {t('OpenLink')}
                                </Button>
                              </a>
                            )}
                          </>
                        ) : (
                          <Button
                            onClick={() => triggerDownload(path, id)}
                            variant='contained'
                            color='primary'
                            size='small'
                            style={{ width: '100%' }}
                          >
                            {t('Download')}
                          </Button>
                        )}
                        <CopyToClipboard text={fullLink}>
                          <Button variant='outlined' color='primary' size='small'>
                            {t('CopyLink')}
                          </Button>
                        </CopyToClipboard>
                      </div>
                    </td>
                  </tr>
                )
              )
            })}
          </tbody>
        </TableStyle>

        <ShortTableWrapper>
          {fileListToDisplay.map(({ id, path, length }) => {
            const { title, resolution, episode, season } = ptt.parse(path)
            const isViewed = viewedFileList?.includes(id)
            const link = getFileLink(path, id)
            const fullLink = new URL(link, window.location.href)
            const infuseLink = `infuse://x-callback-url/play?url=${encodeURIComponent(fullLink)}`
            const iinaLink = `iina://weblink?url=${encodeURIComponent(fullLink)}`
            const isPlayable = isFilePlayable(path)

            return (
              (season === selectedSeason || !seasonAmount?.length) && (
                <ShortTable key={id} isViewed={isViewed}>
                  <div className='short-table-name'>{shouldDisplayFullFileName ? path : title || path}</div>
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

                    {isPlayable ? (
                      <>
                        {isApple && isInfuseUsed && (
                          <a style={{ textDecoration: 'none' }} href={infuseLink}>
                            <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                              {t('Infuse')}
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

                        {isVlcUsed && (
                          <Button
                            style={{ width: '100%' }}
                            variant='outlined'
                            color='primary'
                            size='small'
                            onClick={() => createVlcQrCode(`vlc://${fullLink}`)}
                          >
                            VLC Qr Code
                          </Button>
                        )}

                        {isMac && isIinaUsed && (
                          <a style={{ textDecoration: 'none' }} href={iinaLink}>
                            <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                              IINA
                            </Button>
                          </a>
                        )}

                        {shouldShowOpenLink && (
                          <a style={{ textDecoration: 'none' }} href={link} target='_blank' rel='noreferrer'>
                            <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                              {t('OpenLink')}
                            </Button>
                          </a>
                        )}
                      </>
                    ) : (
                      <Button
                        onClick={() => triggerDownload(path, id)}
                        variant='contained'
                        color='primary'
                        size='small'
                        style={{ width: '100%' }}
                      >
                        {t('Download')}
                      </Button>
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

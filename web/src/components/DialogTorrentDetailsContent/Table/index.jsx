import { streamHost } from 'utils/Hosts'
import isEqual from 'lodash/isEqual'
import { humanizeSize, isStandaloneApp } from 'utils/Utils'
import ptt from 'parse-torrent-title'
import { Button } from '@material-ui/core'
import CopyToClipboard from 'react-copy-to-clipboard'
import { useTranslation } from 'react-i18next'

import VideoPlayer from '../../VideoPlayer'
import { TableStyle, ShortTableWrapper, ShortTable } from './style'

const { memo, useState } = require('react')

// russian episode detection support
ptt.addHandler('episode', /(\d{1,4})[- |. ]серия|серия[- |. ](\d{1,4})/i, { type: 'integer' })
ptt.addHandler('season', /sezon[- |. ](\d{1,3})|(\d{1,3})[- |. ]sezon/i, { type: 'integer' })
ptt.addHandler('season', /сезон[- |. ](\d{1,3})|(\d{1,3})[- |. ]сезон/i, { type: 'integer' })

const Table = memo(
  ({ playableFileList, viewedFileList, selectedSeason, seasonAmount, hash }) => {
    const { t } = useTranslation()
    const [isSupported, setIsSupported] = useState(true)
    const preloadBuffer = fileId => fetch(`${streamHost()}?link=${hash}&index=${fileId}&preload`)
    const getFileLink = (path, id) =>
      `${streamHost()}/${encodeURIComponent(path.split('\\').pop().split('/').pop())}?link=${hash}&index=${id}&play`
    const fileHasEpisodeText = !!playableFileList?.find(({ path }) => ptt.parse(path).episode)
    const fileHasSeasonText = !!playableFileList?.find(({ path }) => ptt.parse(path).season)
    const fileHasResolutionText = !!playableFileList?.find(({ path }) => ptt.parse(path).resolution)

    // if files in list is more then 1 and no season text detected by ptt.parse, show full name
    const shouldDisplayFullFileName = playableFileList?.length > 1 && !fileHasEpisodeText

    const isVlcUsed = JSON.parse(localStorage.getItem('isVlcUsed')) ?? false

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
                        {isSupported ? (
                          <VideoPlayer title={title} videoSrc={link} onNotSupported={() => setIsSupported(false)} />
                        ) : (
                          <a style={{ textDecoration: 'none' }} href={link} target='_blank' rel='noreferrer'>
                            <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                              {t('OpenLink')}
                            </Button>
                          </a>
                        )}
                        <CopyToClipboard text={link}>
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
          {playableFileList.map(({ id, path, length }) => {
            const { title, resolution, episode, season } = ptt.parse(path)
            const isViewed = viewedFileList?.includes(id)
            const link = getFileLink(path, id)

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

                    {isVlcUsed && isStandaloneApp ? (
                      <a style={{ textDecoration: 'none' }} href={`vlc://${link}`}>
                        <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                          VLC
                        </Button>
                      </a>
                    ) : (
                      <a style={{ textDecoration: 'none' }} href={link} target='_blank' rel='noreferrer'>
                        <Button style={{ width: '100%' }} variant='outlined' color='primary' size='small'>
                          {t('OpenLink')}
                        </Button>
                      </a>
                    )}

                    <CopyToClipboard text={link}>
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

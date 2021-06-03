import { streamHost } from 'utils/Hosts'
import isEqual from 'lodash/isEqual'
import { humanizeSize } from 'utils/Utils'
import ptt from 'parse-torrent-title'
import { Button } from '@material-ui/core'
import CopyToClipboard from 'react-copy-to-clipboard'

import { TableStyle, ShortTableWrapper, ShortTable } from './style'

const { memo } = require('react')

const Table = memo(
  ({ playableFileList, viewedFileList, selectedSeason, seasonAmount, hash }) => {
    const preloadBuffer = fileId => fetch(`${streamHost()}?link=${hash}&index=${fileId}&preload`)
    const getFileLink = (path, id) =>
      `${streamHost()}/${encodeURIComponent(path.split('\\').pop().split('/').pop())}?link=${hash}&index=${id}&play`
    const fileHasEpisodeText = !!playableFileList?.find(({ path }) => ptt.parse(path).episode)
    const fileHasSeasonText = !!playableFileList?.find(({ path }) => ptt.parse(path).season)
    const fileHasResolutionText = !!playableFileList?.find(({ path }) => ptt.parse(path).resolution)

    return !playableFileList?.length ? (
      'No playable files in this torrent'
    ) : (
      <>
        <TableStyle>
          <thead>
            <tr>
              <th style={{ width: '0' }}>viewed</th>
              <th>name</th>
              {fileHasSeasonText && seasonAmount?.length === 1 && <th style={{ width: '0' }}>season</th>}
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
                (season === selectedSeason || !seasonAmount?.length) && (
                  <tr key={id} className={isViewed ? 'viewed-file-row' : null}>
                    <td data-label='viewed' className={isViewed ? 'viewed-file-indicator' : null} />
                    <td data-label='name'>{title}</td>
                    {fileHasSeasonText && seasonAmount?.length === 1 && <td data-label='season'>{season}</td>}
                    {fileHasEpisodeText && <td data-label='episode'>{episode}</td>}
                    {fileHasResolutionText && <td data-label='resolution'>{resolution}</td>}
                    <td data-label='size'>{humanizeSize(length)}</td>
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
                  <div className='short-table-name'>{title}</div>
                  <div className='short-table-data'>
                    {isViewed && (
                      <div className='short-table-field'>
                        <div className='short-table-field-name'>viewed</div>
                        <div className='short-table-field-value'>
                          <div className='short-table-viewed-indicator' />
                        </div>
                      </div>
                    )}
                    {fileHasSeasonText && seasonAmount?.length === 1 && (
                      <div className='short-table-field'>
                        <div className='short-table-field-name'>season</div>
                        <div className='short-table-field-value'>{season}</div>
                      </div>
                    )}
                    {fileHasEpisodeText && (
                      <div className='short-table-field'>
                        <div className='short-table-field-name'>epoisode</div>
                        <div className='short-table-field-value'>{episode}</div>
                      </div>
                    )}
                    {fileHasResolutionText && (
                      <div className='short-table-field'>
                        <div className='short-table-field-name'>resolution</div>
                        <div className='short-table-field-value'>{resolution}</div>
                      </div>
                    )}
                    <div className='short-table-field'>
                      <div className='short-table-field-name'>size</div>
                      <div className='short-table-field-value'>{humanizeSize(length)}</div>
                    </div>
                  </div>
                  <div className='short-table-buttons'>
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

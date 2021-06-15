import 'fontsource-roboto'
import { forwardRef, memo, useState } from 'react'
import ptt from 'parse-torrent-title'
import {
  UnfoldMore as UnfoldMoreIcon,
  Edit as EditIcon,
  Close as CloseIcon,
  Delete as DeleteIcon,
} from '@material-ui/icons'
import { getPeerString, humanizeSize, shortenText } from 'utils/Utils'
import { torrentsHost } from 'utils/Hosts'
import { NoImageIcon } from 'icons'
import DialogTorrentDetailsContent from 'components/DialogTorrentDetailsContent'
import Dialog from '@material-ui/core/Dialog'
import Slide from '@material-ui/core/Slide'
import { Button, DialogActions, DialogTitle, useMediaQuery, useTheme } from '@material-ui/core'
import axios from 'axios'
import { useTranslation } from 'react-i18next'
import AddDialog from 'components/Add/AddDialog'
import { isFilePlayable } from 'components/DialogTorrentDetailsContent/helpers'

import { StyledButton, TorrentCard, TorrentCardButtons, TorrentCardDescription, TorrentCardPoster } from './style'

const Transition = forwardRef((props, ref) => <Slide direction='up' ref={ref} {...props} />)

const Torrent = ({ torrent }) => {
  const { t } = useTranslation()
  const [isDetailedInfoOpened, setIsDetailedInfoOpened] = useState(false)
  const [isDeleteTorrentOpened, setIsDeleteTorrentOpened] = useState(false)

  const theme = useTheme()
  const fullScreen = useMediaQuery(theme.breakpoints.down('md'))

  const openDetailedInfo = () => setIsDetailedInfoOpened(true)
  const closeDetailedInfo = () => setIsDetailedInfoOpened(false)
  const openDeleteTorrentAlert = () => setIsDeleteTorrentOpened(true)
  const closeDeleteTorrentAlert = () => setIsDeleteTorrentOpened(false)

  const { title, name, poster, torrent_size: torrentSize, download_speed: downloadSpeed, hash } = torrent

  const dropTorrent = () => axios.post(torrentsHost(), { action: 'drop', hash })
  const deleteTorrent = () => axios.post(torrentsHost(), { action: 'rem', hash })

  const getParsedData = () => {
    const parse = key => ptt.parse(title || '')?.[key] || ptt.parse(name || '')?.[key]

    const parsedYear = parse('year')
    const parsedResolution = parse('resolution')
    const parsedTitle = parse('title')

    return { parsedResolution, parsedYear, parsedTitle }
  }

  const { parsedResolution, parsedYear, parsedTitle } = getParsedData()

  const playableFileAmount = torrent.file_stats?.filter(({ path }) => isFilePlayable(path))?.length

  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false)
  const handleClickOpenEditDialog = () => setIsEditDialogOpen(true)
  const handleCloseEditDialog = () => setIsEditDialogOpen(false)

  return (
    <>
      <TorrentCard>
        <TorrentCardPoster isPoster={poster}>
          {poster ? <img src={poster} alt='poster' /> : <NoImageIcon />}
        </TorrentCardPoster>

        <TorrentCardButtons>
          <StyledButton onClick={openDetailedInfo}>
            <UnfoldMoreIcon />
            <span>{t('Details')}</span>
          </StyledButton>

          <StyledButton onClick={handleClickOpenEditDialog}>
            <EditIcon />
            <span>{t('Edit')}</span>
          </StyledButton>

          <StyledButton onClick={() => dropTorrent(torrent)}>
            <CloseIcon />
            <span>{t('Drop')}</span>
          </StyledButton>

          <StyledButton onClick={openDeleteTorrentAlert}>
            <DeleteIcon />
            <span>{t('Delete')}</span>
          </StyledButton>
        </TorrentCardButtons>

        <TorrentCardDescription>
          <div className='description-title-wrapper'>
            <div className='description-section-name'>{t('Name')}</div>
            <div className='description-torrent-title'>{shortenText(parsedTitle, 100)}</div>
          </div>

          <div className='description-statistics-wrapper'>
            {parsedResolution && (
              <div className='description-statistics-element-wrapper'>
                <div className='description-section-name'>{t('Resolution')}</div>
                <div className='description-statistics-element-value'>{parsedResolution}</div>
              </div>
            )}

            {parsedYear && (
              <div className='description-statistics-element-wrapper'>
                <div className='description-section-name'>{t('Year')}</div>
                <div className='description-statistics-element-value'>{parsedYear}</div>
              </div>
            )}

            {playableFileAmount > 1 && (
              <div className='description-statistics-element-wrapper'>
                <div className='description-section-name'>{t('Files')}</div>
                <div className='description-statistics-element-value'>{playableFileAmount}</div>
              </div>
            )}
          </div>

          <div className='description-statistics-wrapper'>
            <div className='description-statistics-element-wrapper'>
              <div className='description-section-name'>{t('Size')}</div>
              <div className='description-statistics-element-value'>{torrentSize > 0 && humanizeSize(torrentSize)}</div>
            </div>

            <div className='description-statistics-element-wrapper'>
              <div className='description-section-name'>{t('Speed')}</div>
              <div className='description-statistics-element-value'>
                {downloadSpeed > 0 ? humanizeSize(downloadSpeed) : '---'}
              </div>
            </div>

            <div className='description-statistics-element-wrapper'>
              <div className='description-section-name'>{t('Peers')}</div>
              <div className='description-statistics-element-value'>{getPeerString(torrent) || '---'}</div>
            </div>
          </div>
        </TorrentCardDescription>
      </TorrentCard>

      <Dialog
        open={isDetailedInfoOpened}
        onClose={closeDetailedInfo}
        fullScreen={fullScreen}
        fullWidth
        maxWidth='xl'
        TransitionComponent={Transition}
      >
        <DialogTorrentDetailsContent closeDialog={closeDetailedInfo} torrent={torrent} />
      </Dialog>

      <Dialog open={isDeleteTorrentOpened} onClose={closeDeleteTorrentAlert}>
        <DialogTitle>{t('DeleteTorrent?')}</DialogTitle>
        <DialogActions>
          <Button variant='outlined' onClick={closeDeleteTorrentAlert} color='primary'>
            {t('Cancel')}
          </Button>

          <Button
            variant='contained'
            onClick={() => {
              deleteTorrent(torrent)
              closeDeleteTorrentAlert()
            }}
            color='primary'
            autoFocus
          >
            {t('OK')}
          </Button>
        </DialogActions>
      </Dialog>

      {isEditDialogOpen && (
        <AddDialog hash={hash} title={title} name={name} poster={poster} handleClose={handleCloseEditDialog} />
      )}
    </>
  )
}

export default memo(Torrent)

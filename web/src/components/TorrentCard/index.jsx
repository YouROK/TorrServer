import 'fontsource-roboto'
import { forwardRef, useState } from 'react'
import { UnfoldMore as UnfoldMoreIcon, Close as CloseIcon, Delete as DeleteIcon } from '@material-ui/icons'
import { getPeerString, humanizeSize, shortenText } from 'utils/Utils'
import { torrentsHost } from 'utils/Hosts'
import { NoImageIcon } from 'icons'
import DialogTorrentDetailsContent from 'components/DialogTorrentDetailsContent'
import Dialog from '@material-ui/core/Dialog'
import Slide from '@material-ui/core/Slide'
import { Button, DialogActions, DialogTitle, useMediaQuery, useTheme } from '@material-ui/core'
import axios from 'axios'
import ptt from 'parse-torrent-title'
import { useTranslation } from 'react-i18next'

import { StyledButton, TorrentCard, TorrentCardButtons, TorrentCardDescription, TorrentCardPoster } from './style'

const Transition = forwardRef((props, ref) => <Slide direction='up' ref={ref} {...props} />)

export default function Torrent({ torrent }) {
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

  const parsedTitle = (title || name) && ptt.parse(title || name).title

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
    </>
  )
}

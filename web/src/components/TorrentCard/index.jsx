import 'fontsource-roboto'
import { forwardRef, useState } from 'react'
import HeightIcon from '@material-ui/icons/Height'
import CloseIcon from '@material-ui/icons/Close'
import DeleteIcon from '@material-ui/icons/Delete'
import { getPeerString, humanizeSize } from 'utils/Utils'
import { torrentsHost } from 'utils/Hosts'
import { NoImageIcon } from 'icons'
import DialogTorrentDetailsContent from 'components/DialogTorrentDetailsContent'
import Dialog from '@material-ui/core/Dialog'
import Slide from '@material-ui/core/Slide'
import { Button, DialogActions, DialogTitle, useMediaQuery, useTheme } from '@material-ui/core'
import axios from 'axios'

import {
  StyledButton,
  TorrentCard,
  TorrentCardButtons,
  TorrentCardDescription,
  TorrentCardDescriptionContent,
  TorrentCardDescriptionLabel,
  TorrentCardPoster,
  TorrentCardDetails,
} from './style'

const Transition = forwardRef((props, ref) => <Slide direction='up' ref={ref} {...props} />)

export default function Torrent({ torrent }) {
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

  return (
    <>
      <TorrentCard>
        <TorrentCardPoster isPoster={poster}>
          {poster ? <img src={poster} alt='poster' /> : <NoImageIcon />}
        </TorrentCardPoster>

        <TorrentCardButtons>
          <StyledButton onClick={openDetailedInfo}>
            <HeightIcon />
            <span>Details</span>
          </StyledButton>

          <StyledButton onClick={() => dropTorrent(torrent)}>
            <CloseIcon />
            <span>Drop</span>
          </StyledButton>

          <StyledButton onClick={openDeleteTorrentAlert}>
            <DeleteIcon />
            <span>Delete</span>
          </StyledButton>
        </TorrentCardButtons>

        <TorrentCardDescription>
          <span>
            <TorrentCardDescriptionLabel>Name</TorrentCardDescriptionLabel>
            <TorrentCardDescriptionContent isTitle>{title || name}</TorrentCardDescriptionContent>
          </span>

          <TorrentCardDetails>
            <span>
              <TorrentCardDescriptionLabel>Size</TorrentCardDescriptionLabel>
              <TorrentCardDescriptionContent>
                {torrentSize > 0 && humanizeSize(torrentSize)}
              </TorrentCardDescriptionContent>
            </span>

            <span>
              <TorrentCardDescriptionLabel>Speed</TorrentCardDescriptionLabel>
              <TorrentCardDescriptionContent>
                {downloadSpeed > 0 ? humanizeSize(downloadSpeed) : '---'}
              </TorrentCardDescriptionContent>
            </span>

            <span>
              <TorrentCardDescriptionLabel>Peers</TorrentCardDescriptionLabel>
              <TorrentCardDescriptionContent>{getPeerString(torrent) || '---'}</TorrentCardDescriptionContent>
            </span>
          </TorrentCardDetails>
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
        <DialogTitle>Delete Torrent?</DialogTitle>
        <DialogActions>
          <Button variant='outlined' onClick={closeDeleteTorrentAlert} color='primary'>
            Cancel
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
            Ok
          </Button>
        </DialogActions>
      </Dialog>
    </>
  )
}

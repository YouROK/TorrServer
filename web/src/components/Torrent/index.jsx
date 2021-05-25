/* eslint-disable camelcase */
import 'fontsource-roboto'
import { useEffect, useRef, useState } from 'react'
import Button from '@material-ui/core/Button'
import HeightIcon from '@material-ui/icons/Height'
import CloseIcon from '@material-ui/icons/Close'
import DeleteIcon from '@material-ui/icons/Delete'
import DialogActions from '@material-ui/core/DialogActions'
import Dialog from '@material-ui/core/Dialog'
import DataUsageIcon from '@material-ui/icons/DataUsage'
import { getPeerString, humanizeSize } from 'utils/Utils'
import { torrentsHost } from 'utils/Hosts'
import { NoImageIcon } from 'icons'
import DialogTorrentInfo from 'components/DialogTorrentInfo'
import DialogCacheInfo from 'components/DialogCacheInfo'

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

export default function Torrent({ torrent }) {
  const [open, setOpen] = useState(false)
  const [showCache, setShowCache] = useState(false)
  const [torrentLocalComponentValue, setTorrentLocalComponentValue] = useState(torrent)
  const timerID = useRef(-1)

  useEffect(() => {
    setTorrentLocalComponentValue(torrent)
  }, [torrent])

  useEffect(() => {
    if (open)
      timerID.current = setInterval(() => {
        getTorrent(torrentLocalComponentValue.hash, (torr, error) => {
          if (error) console.error(error)
          else if (torr) setTorrentLocalComponentValue(torr)
        })
      }, 1000)
    else clearInterval(timerID.current)

    return () => {
      clearInterval(timerID.current)
    }
  }, [torrentLocalComponentValue.hash, open])

  const { title, name, poster, torrent_size, download_speed } = torrentLocalComponentValue

  return (
    <>
      <TorrentCard>
        <TorrentCardPoster isPoster={poster}>
          {poster ? <img src={poster} alt='poster' /> : <NoImageIcon />}
        </TorrentCardPoster>

        <TorrentCardButtons>
          <StyledButton
            onClick={() => {
              setShowCache(true)
              setOpen(true)
            }}
          >
            <DataUsageIcon />
            <span>Cache</span>
          </StyledButton>

          <StyledButton onClick={() => dropTorrent(torrentLocalComponentValue)}>
            <CloseIcon />
            <span>Drop</span>
          </StyledButton>

          <StyledButton onClick={() => deleteTorrent(torrentLocalComponentValue)}>
            <DeleteIcon />
            <span>Delete</span>
          </StyledButton>

          <StyledButton
            onClick={() => {
              setShowCache(false)
              setOpen(true)
            }}
          >
            <HeightIcon />
            <span>Details</span>
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
                {torrent_size > 0 && humanizeSize(torrent_size)}
              </TorrentCardDescriptionContent>
            </span>

            <span>
              <TorrentCardDescriptionLabel>Speed</TorrentCardDescriptionLabel>
              <TorrentCardDescriptionContent>
                {download_speed > 0 ? humanizeSize(download_speed) : '---'}
              </TorrentCardDescriptionContent>
            </span>

            <span>
              <TorrentCardDescriptionLabel>Peers</TorrentCardDescriptionLabel>
              <TorrentCardDescriptionContent>
                {getPeerString(torrentLocalComponentValue) || '---'}
              </TorrentCardDescriptionContent>
            </span>
          </TorrentCardDetails>
        </TorrentCardDescription>
      </TorrentCard>

      <Dialog open={open} onClose={() => setOpen(false)} aria-labelledby='form-dialog-title' fullWidth maxWidth='lg'>
        {!showCache ? (
          <DialogTorrentInfo torrent={(open, torrentLocalComponentValue)} />
        ) : (
          <DialogCacheInfo hash={(open, torrentLocalComponentValue.hash)} />
        )}
        <DialogActions>
          <Button variant='outlined' color='primary' onClick={() => setOpen(false)}>
            OK
          </Button>
        </DialogActions>
      </Dialog>
    </>
  )
}

function getTorrent(hash, callback) {
  try {
    fetch(torrentsHost(), {
      method: 'post',
      body: JSON.stringify({ action: 'get', hash }),
      headers: {
        Accept: 'application/json, text/plain, */*',
        'Content-Type': 'application/json',
      },
    })
      .then(res => res.json())
      .then(
        json => {
          callback(json, null)
        },
        error => {
          callback(null, error)
        },
      )
  } catch (e) {
    console.error(e)
  }
}

function deleteTorrent(torrent) {
  try {
    fetch(torrentsHost(), {
      method: 'post',
      body: JSON.stringify({
        action: 'rem',
        hash: torrent.hash,
      }),
      headers: {
        Accept: 'application/json, text/plain, */*',
        'Content-Type': 'application/json',
      },
    })
  } catch (e) {
    console.error(e)
  }
}

function dropTorrent(torrent) {
  try {
    fetch(torrentsHost(), {
      method: 'post',
      body: JSON.stringify({
        action: 'drop',
        hash: torrent.hash,
      }),
      headers: {
        Accept: 'application/json, text/plain, */*',
        'Content-Type': 'application/json',
      },
    })
  } catch (e) {
    console.error(e)
  }
}

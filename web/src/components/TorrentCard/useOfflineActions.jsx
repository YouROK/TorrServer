import { useState } from 'react'
import axios from 'axios'
import { Button, Dialog, DialogActions, DialogTitle } from '@material-ui/core'
import { useTranslation } from 'react-i18next'
import { useQueryClient } from 'react-query'
import { torrentsHost } from 'utils/Hosts'

// Offline actions of one torrent for the list card (the same API calls as the "Offline mode"
// block of the torrent dialog). `dialogs` must be rendered by the caller (confirm + error).
export const useOfflineActions = hash => {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [busy, setBusy] = useState(false)
  const [confirmText, setConfirmText] = useState('')
  const [pendingBody, setPendingBody] = useState(null)
  const [errorText, setErrorText] = useState('')

  const post = body => {
    setBusy(true)
    return axios
      .post(torrentsHost(), { hash, ...body })
      .catch(e => setErrorText(e?.response?.data?.error || e.message))
      .finally(() => {
        // refresh the torrent list right away instead of waiting for the next poll
        queryClient.invalidateQueries('torrents')
        setBusy(false)
      })
  }

  const enable = () => post({ action: 'offline', enable: true })
  const pause = () => post({ action: 'offline_pause' })
  const resume = () => post({ action: 'offline_resume' })
  // removing deletes files from the disk - ask first
  const remove = () => {
    setConfirmText(t('OfflineRemoveConfirm', 'Delete the offline copy of this torrent from the disk?'))
    setPendingBody({ action: 'offline', enable: false })
  }

  const closeConfirm = () => {
    setConfirmText('')
    setPendingBody(null)
  }

  const dialogs = (
    // No wrapper element: the controls are rendered inside the grid of the card buttons.
    // Clicks in the dialogs are portaled React events: they must not reach the card.
    <>
      <Dialog open={!!pendingBody} onClose={closeConfirm} onClick={e => e.stopPropagation()}>
        <DialogTitle>{confirmText}</DialogTitle>
        <DialogActions>
          <Button variant='outlined' color='secondary' onClick={closeConfirm}>
            {t('Cancel')}
          </Button>
          <Button
            variant='contained'
            color='primary'
            onClick={() => {
              const body = pendingBody
              closeConfirm()
              if (body) post(body)
            }}
          >
            {t('OK')}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={!!errorText} onClose={() => setErrorText('')} onClick={e => e.stopPropagation()}>
        <DialogTitle>{errorText}</DialogTitle>
        <DialogActions>
          <Button variant='contained' color='primary' onClick={() => setErrorText('')}>
            {t('OK')}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  )

  return { busy, enable, pause, resume, remove, dialogs }
}

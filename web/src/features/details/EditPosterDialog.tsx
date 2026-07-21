import { Button, Input, Label, Modal, Spinner, TextField, useMediaQuery } from '@heroui/react'
import { useEffect, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'

import type { TorrentStat } from 'shared/api/types'
import { TORRENTS_QUERY_KEY, updateTorrent } from 'shared/api/torrents'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import { queryMax } from 'shared/theme/breakpoints'
import AppDialog from 'shared/ui/AppDialog'
import { DIALOG_SHEET_M } from 'shared/ui/dialogSizes'
import { useOptionalAppToast } from 'shared/ui/Toast'

export interface EditPosterDialogProps {
  torrent: TorrentStat
  open: boolean
  onClose: () => void
}

/** Lightweight poster-URL editor opened from the Details hero poster. */
export default function EditPosterDialog({ torrent, open, onClose }: EditPosterDialogProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const toast = useOptionalAppToast()
  const isMobile = useMediaQuery(queryMax('mobile'))
  const isFullScreenBreakpoint = useDialogFullScreen()

  const [url, setUrl] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!open) return
    // eslint-disable-next-line react-hooks/set-state-in-effect -- hydrate from torrent when dialog opens
    setUrl(torrent.poster || '')
  }, [open, torrent.poster])

  const handleSave = async () => {
    setSaving(true)
    try {
      // Server SetTorrent replaces title/category too — always send current values.
      await updateTorrent({
        hash: torrent.hash,
        title: torrent.title || torrent.name || torrent.hash,
        category: torrent.category || undefined,
        poster: url.trim(),
      })
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: TORRENTS_QUERY_KEY }),
        queryClient.invalidateQueries({ queryKey: ['torrent', torrent.hash] }),
      ])
      toast?.showToast({ message: t('Saved'), severity: 'success' })
      onClose()
    } catch {
      toast?.showToast({ message: t('Error'), severity: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const footerButtonClassName = isMobile ? 'min-h-11 px-4' : undefined

  return (
    <AppDialog
      open={open}
      onClose={onClose}
      size='md'
      fullScreen={isFullScreenBreakpoint}
      dialogStyle={isFullScreenBreakpoint ? undefined : DIALOG_SHEET_M}
    >
      <Modal.Header>
        <Modal.Heading>{t('AddDialog.AddPosterLinkInput')}</Modal.Heading>
        <Modal.CloseTrigger aria-label={t('Close')} />
      </Modal.Header>
      <Modal.Body>
        <TextField value={url} onChange={setUrl} isDisabled={saving} autoFocus>
          <Label>{t('AddDialog.AddPosterLinkInput')}</Label>
          <Input placeholder='https://…' />
        </TextField>
      </Modal.Body>
      <Modal.Footer>
        <Button onPress={onClose} isDisabled={saving} variant='secondary' className={footerButtonClassName}>
          {t('Cancel')}
        </Button>
        <Button
          variant='primary'
          onPress={() => void handleSave()}
          isDisabled={saving}
          className={footerButtonClassName}
        >
          {saving ? <Spinner size='sm' color='current' /> : t('Save')}
        </Button>
      </Modal.Footer>
    </AppDialog>
  )
}

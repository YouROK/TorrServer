import { Button, Modal } from '@heroui/react'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { TorrentStat } from 'shared/api/types'
import { torrsShareUrl } from 'shared/api/extras'
import { copyToClipboard } from 'shared/lib/clipboard'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import AppDialog from 'shared/ui/AppDialog'
import { useOptionalAppToast } from 'shared/ui/Toast'

export interface ExportLibraryDialogProps {
  open: boolean
  onClose: () => void
  torrents: TorrentStat[]
}

type ExportKind = 'magnets' | 'torrs' | 'json'

function downloadText(filename: string, text: string, mime: string) {
  const blob = new Blob([text], { type: mime })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  anchor.rel = 'noopener'
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)
  URL.revokeObjectURL(url)
}

/** Full-library export (magnets / web+torrs:// / JSON) — uses the complete torrents list from API, not UI filters. */
export default function ExportLibraryDialog({ open, onClose, torrents }: ExportLibraryDialogProps) {
  const { t } = useTranslation()
  const toast = useOptionalAppToast()
  const isFullScreen = useDialogFullScreen()
  const [copiedKind, setCopiedKind] = useState<ExportKind | null>(null)
  const [manualText, setManualText] = useState<string | null>(null)

  const magnets = useMemo(
    () =>
      torrents
        .map(tr => `magnet:?xt=urn:btih:${tr.hash}&dn=${encodeURIComponent(tr.title || tr.name || tr.hash)}`)
        .join('\n'),
    [torrents],
  )

  const torrs = useMemo(() => torrents.map(tr => torrsShareUrl(tr)).join('\n'), [torrents])

  const json = useMemo(
    () =>
      JSON.stringify(
        torrents.map(tr => ({
          hash: tr.hash,
          title: tr.title,
          name: tr.name,
          category: tr.category,
          poster: tr.poster,
          torrs_hash: tr.torrs_hash,
          torrent_size: tr.torrent_size,
        })),
        null,
        2,
      ),
    [torrents],
  )

  const payload = (kind: ExportKind) => {
    if (kind === 'magnets') return magnets
    if (kind === 'torrs') return torrs
    return json
  }

  const copy = async (kind: ExportKind) => {
    const text = payload(kind)
    try {
      await copyToClipboard(text)
      setCopiedKind(kind)
      setManualText(null)
      toast?.showToast({ message: t('Copied'), severity: 'success' })
      window.setTimeout(() => setCopiedKind(current => (current === kind ? null : current)), 2000)
    } catch {
      setManualText(text)
      toast?.showToast({ message: t('ExportCopyFailed'), severity: 'error' })
    }
  }

  const download = (kind: ExportKind) => {
    if (kind === 'magnets') downloadText('torrserver-magnets.txt', magnets, 'text/plain;charset=utf-8')
    else if (kind === 'torrs') downloadText('torrserver-torrs.txt', torrs, 'text/plain;charset=utf-8')
    else downloadText('torrserver-library.json', json, 'application/json;charset=utf-8')
    toast?.showToast({ message: t('ExportDownloaded'), severity: 'success' })
  }

  const copyLabel = (kind: ExportKind, idle: string) => (copiedKind === kind ? t('Copied') : idle)

  return (
    <AppDialog open={open} onClose={onClose} size='sm' fullScreen={isFullScreen}>
      <Modal.Header>
        <Modal.Heading>{t('ExportLibrary')}</Modal.Heading>
        <Modal.CloseTrigger aria-label={t('Close')} />
      </Modal.Header>
      <Modal.Body className='space-y-3'>
        <p className='text-sm text-muted'>{t('ExportLibraryCount', { count: torrents.length })}</p>
        <div className='flex flex-col gap-2'>
          <Button
            variant='secondary'
            className='min-h-11 w-full justify-start'
            isDisabled={!torrents.length}
            onPress={() => void copy('magnets')}
          >
            {copyLabel('magnets', t('ExportMagnets'))}
          </Button>
          <Button
            variant='secondary'
            className='min-h-11 w-full justify-start'
            isDisabled={!torrents.length}
            onPress={() => void copy('torrs')}
          >
            {copyLabel('torrs', t('ExportTorrs'))}
          </Button>
          <Button
            variant='secondary'
            className='min-h-11 w-full justify-start'
            isDisabled={!torrents.length}
            onPress={() => void copy('json')}
          >
            {copyLabel('json', t('ExportJson'))}
          </Button>
        </div>
        <div className='flex flex-col gap-2 border-t border-border/60 pt-3'>
          <p className='text-xs text-muted'>{t('ExportDownloadHint')}</p>
          <Button
            variant='ghost'
            className='min-h-11 w-full justify-start'
            isDisabled={!torrents.length}
            onPress={() => download('magnets')}
          >
            {t('ExportDownloadMagnets')}
          </Button>
          <Button
            variant='ghost'
            className='min-h-11 w-full justify-start'
            isDisabled={!torrents.length}
            onPress={() => download('torrs')}
          >
            {t('ExportDownloadTorrs')}
          </Button>
          <Button
            variant='ghost'
            className='min-h-11 w-full justify-start'
            isDisabled={!torrents.length}
            onPress={() => download('json')}
          >
            {t('ExportDownloadJson')}
          </Button>
        </div>
        {manualText ? (
          <div className='space-y-2 rounded-xl border border-border bg-surface-secondary p-3'>
            <p className='text-xs text-muted'>{t('ExportManualCopyHint')}</p>
            <textarea
              className='h-32 w-full resize-y rounded-lg border border-border bg-surface p-2 font-mono text-xs text-foreground'
              readOnly
              value={manualText}
              onFocus={event => event.currentTarget.select()}
            />
          </div>
        ) : null}
      </Modal.Body>
      <Modal.Footer>
        <Button variant='secondary' onPress={onClose}>
          {t('Close')}
        </Button>
      </Modal.Footer>
    </AppDialog>
  )
}

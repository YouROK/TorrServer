import { Button, Modal, Spinner, Switch, TextArea, TextField } from '@heroui/react'
import { useQueryClient } from '@tanstack/react-query'
import { useCallback, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { addTorrent, TORRENTS_QUERY_KEY } from 'shared/api/torrents'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import { useTorrentsQuery } from 'shared/hooks/useTorrentsQuery'
import { parseLibraryImportText, type LibraryImportItem } from 'shared/lib/libraryImport'
import AppDialog from 'shared/ui/AppDialog'
import { DIALOG_SHEET_M } from 'shared/ui/dialogSizes'
import { useOptionalAppToast } from 'shared/ui/Toast'

export interface ImportLibraryDialogProps {
  open: boolean
  onClose: () => void
}

const ADD_CONCURRENCY = 3

async function mapPool<T>(items: T[], concurrency: number, worker: (item: T) => Promise<void>) {
  let next = 0
  const runners = Array.from({ length: Math.min(concurrency, Math.max(items.length, 0)) }, async () => {
    while (next < items.length) {
      const index = next
      next += 1
      await worker(items[index])
    }
  })
  await Promise.all(runners)
}

/** Import magnets / torrs:// / export JSON back into the library via torrent add API. */
export default function ImportLibraryDialog({ open, onClose }: ImportLibraryDialogProps) {
  const { t } = useTranslation()
  const toast = useOptionalAppToast()
  const queryClient = useQueryClient()
  const isFullScreen = useDialogFullScreen()
  const fileInputRef = useRef<HTMLInputElement>(null)

  const { data: torrents } = useTorrentsQuery({ enabled: open })

  const [raw, setRaw] = useState('')
  const [saveToDb, setSaveToDb] = useState(true)
  const [skipExisting, setSkipExisting] = useState(true)
  const [running, setRunning] = useState(false)
  const [progress, setProgress] = useState<{
    done: number
    total: number
    ok: number
    skipped: number
    failed: number
  } | null>(null)

  const existingHashes = useMemo(() => new Set((torrents || []).map(tr => tr.hash.toLowerCase())), [torrents])

  const parsed = useMemo(() => parseLibraryImportText(raw), [raw])

  const reset = () => {
    setRaw('')
    setSaveToDb(true)
    setSkipExisting(true)
    setRunning(false)
    setProgress(null)
  }

  const handleClose = () => {
    if (running) return
    reset()
    onClose()
  }

  const onPickFile = useCallback(
    async (file: File | undefined) => {
      if (!file) return
      try {
        setRaw(await file.text())
      } catch {
        toast?.showToast({ message: t('Error'), severity: 'error' })
      }
    },
    [t, toast],
  )

  const runImport = async () => {
    const items = parseLibraryImportText(raw)
    if (!items.length) {
      toast?.showToast({ message: t('ImportLibraryEmpty'), severity: 'warning' })
      return
    }

    setRunning(true)
    const known = new Set(existingHashes)
    const stats = { done: 0, total: items.length, ok: 0, skipped: 0, failed: 0 }
    setProgress({ ...stats })

    const shouldSkip = (item: LibraryImportItem) => {
      if (!skipExisting || !item.hashHint) return false
      return known.has(item.hashHint.toLowerCase())
    }

    await mapPool(items, ADD_CONCURRENCY, async item => {
      if (shouldSkip(item)) {
        stats.skipped += 1
      } else {
        try {
          await addTorrent({
            link: item.link,
            title: item.title,
            category: item.category,
            poster: item.poster,
            save_to_db: saveToDb,
          })
          if (item.hashHint) known.add(item.hashHint.toLowerCase())
          stats.ok += 1
        } catch {
          stats.failed += 1
        }
      }
      stats.done += 1
      setProgress({ ...stats })
    })

    await queryClient.invalidateQueries({ queryKey: TORRENTS_QUERY_KEY })
    setRunning(false)

    toast?.showToast({
      message: t('ImportLibraryDone', {
        ok: stats.ok,
        skipped: stats.skipped,
        failed: stats.failed,
      }),
      severity: stats.failed ? 'warning' : 'success',
    })
  }

  return (
    <AppDialog open={open} onClose={handleClose} size='md' fullScreen={isFullScreen} dialogStyle={DIALOG_SHEET_M}>
      <Modal.Header>
        <Modal.Heading>{t('ImportLibrary')}</Modal.Heading>
        <Modal.CloseTrigger aria-label={t('Close')} />
      </Modal.Header>
      <Modal.Body className='space-y-4'>
        <p className='text-sm text-muted'>{t('ImportLibraryHint')}</p>
        <TextField value={raw} onChange={setRaw} isDisabled={running}>
          <TextArea className='min-h-40 font-mono text-xs' rows={10} placeholder={t('ImportLibraryPlaceholder')} />
        </TextField>
        <div className='flex flex-wrap gap-2'>
          <input
            ref={fileInputRef}
            type='file'
            accept='.json,.txt,application/json,text/plain'
            className='hidden'
            onChange={event => {
              const file = event.target.files?.[0]
              event.target.value = ''
              void onPickFile(file)
            }}
          />
          <Button
            variant='secondary'
            className='min-h-11'
            isDisabled={running}
            onPress={() => fileInputRef.current?.click()}
          >
            {t('ImportLibraryFromFile')}
          </Button>
          <Button variant='ghost' className='min-h-11' isDisabled={running || !raw} onPress={() => setRaw('')}>
            {t('Clear')}
          </Button>
        </div>
        <p className='text-sm text-muted'>{t('ImportLibraryParsed', { count: parsed.length })}</p>
        <div className='flex min-h-11 items-start justify-between gap-4'>
          <div className='min-w-0 flex-1'>
            <p className='text-sm font-medium text-foreground'>{t('AddDialog.SaveToDb')}</p>
          </div>
          <Switch
            isSelected={saveToDb}
            onChange={setSaveToDb}
            isDisabled={running}
            aria-label={t('AddDialog.SaveToDb')}
            className='mt-0.5 shrink-0'
          >
            <Switch.Content>
              <Switch.Control>
                <Switch.Thumb />
              </Switch.Control>
            </Switch.Content>
          </Switch>
        </div>
        <div className='flex min-h-11 items-start justify-between gap-4'>
          <div className='min-w-0 flex-1'>
            <p className='text-sm font-medium text-foreground'>{t('ImportLibrarySkipExisting')}</p>
          </div>
          <Switch
            isSelected={skipExisting}
            onChange={setSkipExisting}
            isDisabled={running}
            aria-label={t('ImportLibrarySkipExisting')}
            className='mt-0.5 shrink-0'
          >
            <Switch.Content>
              <Switch.Control>
                <Switch.Thumb />
              </Switch.Control>
            </Switch.Content>
          </Switch>
        </div>
        {progress ? (
          <p className='flex items-center gap-2 text-sm text-muted'>
            {running ? <Spinner size='sm' /> : null}
            {t('ImportLibraryProgress', progress)}
          </p>
        ) : null}
      </Modal.Body>
      <Modal.Footer className='gap-2'>
        <Button variant='secondary' onPress={handleClose} isDisabled={running}>
          {t('Close')}
        </Button>
        <Button variant='primary' onPress={() => void runImport()} isDisabled={running || !parsed.length}>
          {running ? <Spinner size='sm' color='current' /> : t('ImportLibraryRun')}
        </Button>
      </Modal.Footer>
    </AppDialog>
  )
}

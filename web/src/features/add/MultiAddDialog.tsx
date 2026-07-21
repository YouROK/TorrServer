import { Button, Input, Label, ListBox, Modal, Select, Spinner, TextField, useMediaQuery } from '@heroui/react'
import { AlertCircle, Check, Film, Trash2 } from 'lucide-react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'

import type { MultiAddFileState } from 'shared/api/types'
import { TORRENTS_QUERY_KEY, uploadTorrent } from 'shared/api/torrents'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import { useTorrentsQuery } from 'shared/hooks/useTorrentsQuery'
import {
  checkImageURL,
  getMoviePosters,
  parseTorrentInfoHash,
  parseTorrentTitle,
  shortenTitleForPosterSearch,
} from 'shared/lib/torrentHelpers'
import { queryMax } from 'shared/theme/breakpoints'
import { TORRENT_CATEGORIES } from 'shared/torrent/categories'
import AppDialog from 'shared/ui/AppDialog'
import { DIALOG_SHEET_M } from 'shared/ui/dialogSizes'
import { useOptionalAppToast } from 'shared/ui/Toast'

export interface MultiAddDialogProps {
  files: File[]
  open: boolean
  onClose: () => void
}

type PerFileStatus = 'pending' | 'saving' | 'success' | 'error'

interface MultiAddRow extends MultiAddFileState {
  status: PerFileStatus
}

function createInitialState(files: File[]): MultiAddRow[] {
  return files.map(file => ({
    file,
    title: file.name.replace(/\.torrent$/i, ''),
    category: '',
    poster: '',
    isPosterOk: false,
    originalName: file.name,
    parsedTitle: '',
    infoHash: '',
    alreadyExists: false,
    status: 'pending',
  }))
}

/** Batch-add multiple dropped `.torrent` files, each enriched with a parsed title + poster guess. */
export default function MultiAddDialog({ files, open, onClose }: MultiAddDialogProps) {
  const { t, i18n } = useTranslation()
  const queryClient = useQueryClient()
  const toast = useOptionalAppToast()
  const isMobile = useMediaQuery(queryMax('mobile'))
  const isFullScreenBreakpoint = useDialogFullScreen()

  const { data: existingTorrents = [] } = useTorrentsQuery()

  const [fileList, setFileList] = useState<MultiAddRow[]>(() => createInitialState(files))
  const [saving, setSaving] = useState(false)
  const [enriching, setEnriching] = useState(true)

  const handleUpdate = useCallback((index: number, updates: Partial<MultiAddRow>) => {
    setFileList(prev => prev.map((item, i) => (i === index ? { ...item, ...updates } : item)))
  }, [])

  useEffect(() => {
    let cancelled = false
    const posterLang = i18n.language?.startsWith('ru') ? 'ru' : 'en'

    const enrich = async () => {
      setEnriching(true)
      await Promise.all(
        files.map(async (file, index) => {
          const infoHash = await parseTorrentInfoHash(file)
          if (cancelled) return

          if (infoHash) {
            const existing = existingTorrents.find(tor => tor.hash?.toLowerCase() === infoHash.toLowerCase())
            if (existing) {
              handleUpdate(index, {
                infoHash,
                alreadyExists: true,
                title: existing.title || existing.name || file.name,
                category: String(existing.category || ''),
                poster: existing.poster || '',
                isPosterOk: Boolean(existing.poster),
              })
              return
            }
            handleUpdate(index, { infoHash })
          }

          await new Promise<void>(resolve => {
            parseTorrentTitle(file, ({ parsedTitle, originalName }) => {
              if (cancelled) {
                resolve()
                return
              }
              const searchTitle = shortenTitleForPosterSearch(parsedTitle || originalName || file.name)
              handleUpdate(index, {
                originalName: originalName || file.name,
                parsedTitle: parsedTitle || '',
                title: parsedTitle || originalName || file.name.replace(/\.torrent$/i, ''),
              })
              if (!searchTitle) {
                resolve()
                return
              }
              getMoviePosters(searchTitle, posterLang)
                .then(async urls => {
                  if (cancelled || !urls?.length) return
                  const ok = await checkImageURL(urls[0])
                  if (ok) handleUpdate(index, { poster: urls[0], isPosterOk: true })
                })
                .finally(() => resolve())
            })
          })
        }),
      )
      if (!cancelled) setEnriching(false)
    }

    void enrich()
    return () => {
      cancelled = true
    }
  }, [files, existingTorrents, handleUpdate, i18n.language])

  const visibleFiles = useMemo(
    () => fileList.map((item, index) => ({ item, index })).filter(({ item }) => !item.alreadyExists),
    [fileList],
  )
  const skippedCount = fileList.length - visibleFiles.length

  const handleRemove = useCallback(
    (index: number) => {
      setFileList(prev => {
        const next = prev.filter((_, i) => i !== index)
        if (next.filter(item => !item.alreadyExists).length === 0) onClose()
        return next
      })
    },
    [onClose],
  )

  const handleSaveAll = async () => {
    const targets = visibleFiles.filter(({ item }) => item.status !== 'success')
    if (!targets.length) return

    setSaving(true)
    targets.forEach(({ index }) => handleUpdate(index, { status: 'saving' }))

    const outcomes = await Promise.allSettled(
      targets.map(({ item, index }) =>
        uploadTorrent(item.file, { title: item.title, category: item.category, poster: item.poster }).then(
          () => index,
          err => {
            throw { index, err }
          },
        ),
      ),
    )

    let successCount = 0
    let failureCount = 0
    for (const outcome of outcomes) {
      if (outcome.status === 'fulfilled') {
        successCount += 1
        handleUpdate(outcome.value, { status: 'success' })
      } else {
        failureCount += 1
        const failedIndex = (outcome.reason as { index?: number })?.index
        if (typeof failedIndex === 'number') handleUpdate(failedIndex, { status: 'error' })
      }
    }

    if (successCount > 0) {
      await queryClient.invalidateQueries({ queryKey: TORRENTS_QUERY_KEY })
    }

    if (failureCount === 0) {
      toast?.showToast({ message: t('Search.TorrentAdded'), severity: 'success' })
      onClose()
    } else if (successCount > 0) {
      toast?.showToast({ message: t('Search.TorrentAdded'), severity: 'warning' })
    } else {
      toast?.showToast({ message: t('Error'), severity: 'error' })
    }

    setSaving(false)
  }

  const footerButtonClassName = isMobile ? 'min-h-11 px-4' : undefined
  const addableCount = visibleFiles.filter(({ item }) => item.status !== 'success').length

  return (
    <AppDialog
      open={open}
      onClose={onClose}
      size='md'
      fullScreen={isFullScreenBreakpoint}
      dialogClassName='flex flex-col overflow-hidden'
      dialogStyle={isFullScreenBreakpoint ? undefined : DIALOG_SHEET_M}
    >
      <Modal.Header className='shrink-0'>
        <Modal.Heading>
          {t('AddNewTorrent')} ({visibleFiles.length}
          {skippedCount > 0 ? ` · ${skippedCount} ${t('TorrentInDb')}` : ''})
        </Modal.Heading>
        <Modal.CloseTrigger aria-label={t('Close')} />
      </Modal.Header>
      <Modal.Body className='min-h-0 flex-1 overflow-y-auto overscroll-contain'>
        {enriching ? (
          <div className='grid place-items-center py-8'>
            <Spinner size='lg' />
          </div>
        ) : (
          <div className='space-y-4'>
            {fileList.map((item, index) => {
              if (item.alreadyExists) {
                return (
                  <p key={`${item.file.name}-exists-${index}`} className='text-sm text-muted'>
                    {item.title || item.file.name} — {t('TorrentInDb')}
                  </p>
                )
              }
              return (
                <div
                  key={`${item.file.name}-${index}`}
                  className={`grid gap-3 border-b border-separator pb-4 last:border-b-0 last:pb-0 ${
                    item.poster ? 'sm:grid-cols-[64px_1fr_auto]' : 'sm:grid-cols-[1fr_auto]'
                  }`}
                >
                  {item.poster ? (
                    <div className='hidden h-24 w-16 overflow-hidden rounded-lg bg-surface-tertiary sm:block'>
                      <img src={item.poster} alt='' className='h-full w-full object-cover' />
                    </div>
                  ) : null}
                  <div className='space-y-2.5'>
                    <div className='flex items-center gap-2 text-xs text-muted'>
                      <StatusIcon status={item.status} />
                      <span className='truncate'>
                        {index + 1}. {item.file.name}
                      </span>
                    </div>
                    <TextField
                      value={item.title}
                      onChange={value => handleUpdate(index, { title: value })}
                      isDisabled={item.status === 'saving' || item.status === 'success'}
                    >
                      <Label>{t('AddDialog.TitleBlank')}</Label>
                      <Input />
                    </TextField>
                    <Select
                      selectedKey={item.category || 'none'}
                      onSelectionChange={key => handleUpdate(index, { category: key === 'none' ? '' : String(key) })}
                      isDisabled={item.status === 'saving' || item.status === 'success'}
                    >
                      <Label>{t('AddDialog.CategoryHelperText')}</Label>
                      <Select.Trigger>
                        <Select.Value />
                        <Select.Indicator />
                      </Select.Trigger>
                      <Select.Popover>
                        <ListBox>
                          <ListBox.Item id='none'>—</ListBox.Item>
                          {TORRENT_CATEGORIES.map(cat => (
                            <ListBox.Item key={cat.key} id={cat.key}>
                              {t(cat.name)}
                            </ListBox.Item>
                          ))}
                        </ListBox>
                      </Select.Popover>
                    </Select>
                  </div>
                  <Button
                    isIconOnly
                    variant='ghost'
                    aria-label={t('Delete')}
                    onPress={() => handleRemove(index)}
                    isDisabled={item.status === 'saving'}
                    className='min-h-11 min-w-11'
                  >
                    <Trash2 size={16} strokeWidth={1.75} aria-hidden />
                  </Button>
                </div>
              )
            })}
          </div>
        )}
      </Modal.Body>
      <Modal.Footer className='shrink-0'>
        <Button onPress={onClose} isDisabled={saving} variant='secondary' className={footerButtonClassName}>
          {t('Cancel')}
        </Button>
        <Button
          variant='primary'
          onPress={() => void handleSaveAll()}
          isDisabled={saving || enriching || addableCount === 0}
          className={footerButtonClassName}
        >
          {saving ? <Spinner size='sm' color='current' /> : `${t('Add')} (${addableCount})`}
        </Button>
      </Modal.Footer>
    </AppDialog>
  )
}

function StatusIcon({ status }: { status: PerFileStatus }) {
  switch (status) {
    case 'saving':
      return <Spinner size='sm' color='current' />
    case 'success':
      return <Check size={14} strokeWidth={1.75} className='text-accent' aria-hidden />
    case 'error':
      return <AlertCircle size={14} strokeWidth={1.75} className='text-danger' aria-hidden />
    default:
      return <Film size={14} strokeWidth={1.75} aria-hidden />
  }
}

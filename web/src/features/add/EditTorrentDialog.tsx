import { Button, Input, Label, ListBox, Modal, Select, Spinner, TextField, useMediaQuery } from '@heroui/react'
import { Film } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'

import type { TorrentStat } from 'shared/api/types'
import { TORRENTS_QUERY_KEY, updateTorrent } from 'shared/api/torrents'
import { getMoviePosters, shortenTitleForPosterSearch } from 'shared/lib/torrentHelpers'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import { queryMax } from 'shared/theme/breakpoints'
import { TORRENT_CATEGORIES } from 'shared/torrent/categories'
import AppDialog from 'shared/ui/AppDialog'
import { DIALOG_EDIT } from 'shared/ui/dialogSizes'
import { useOptionalAppToast } from 'shared/ui/Toast'

import PosterPicker from './PosterPicker'

export interface EditTorrentDialogProps {
  torrent: TorrentStat | null
  open: boolean
  onClose: () => void
}

/** Edit an existing torrent's title/category/poster — same poster-search UX as AddDialog, prefilled. */
export default function EditTorrentDialog({ torrent, open, onClose }: EditTorrentDialogProps) {
  const { t, i18n } = useTranslation()
  const queryClient = useQueryClient()
  const toast = useOptionalAppToast()
  const isMobile = useMediaQuery(queryMax('mobile'))
  const isFullScreenBreakpoint = useDialogFullScreen()
  const posterRequestRef = useRef(0)

  const [title, setTitle] = useState('')
  const [category, setCategory] = useState('')
  const [poster, setPoster] = useState('')
  const [posterOptions, setPosterOptions] = useState<string[]>([])
  const [postersLoading, setPostersLoading] = useState(false)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!open || !torrent) return
    // eslint-disable-next-line react-hooks/set-state-in-effect -- hydrate edit form from torrent prop
    setTitle(torrent.title || torrent.name || '')
    setCategory(torrent.category || '')
    setPoster(torrent.poster || '')
    setPosterOptions(torrent.poster ? [torrent.poster] : [])
  }, [open, torrent])

  useEffect(() => {
    if (!open) return undefined
    const query = shortenTitleForPosterSearch(title.trim()) || title.trim()
    if (!query || query.length < 2) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- keep current poster when search query too short
      setPosterOptions(prev => (poster ? Array.from(new Set([poster, ...prev])) : prev))
      return undefined
    }

    const requestId = ++posterRequestRef.current
    const timer = window.setTimeout(() => {
      setPostersLoading(true)
      const lang = i18n.language?.startsWith('ru') ? 'ru' : 'en'
      void getMoviePosters(query, lang)
        .then(urls => {
          if (requestId !== posterRequestRef.current) return
          const next = urls || []
          setPosterOptions(poster ? Array.from(new Set([poster, ...next])) : next)
        })
        .finally(() => {
          if (requestId === posterRequestRef.current) setPostersLoading(false)
        })
    }, 450)

    return () => window.clearTimeout(timer)
  }, [open, title, i18n.language, poster])

  const handleSave = async () => {
    if (!torrent) return
    setSaving(true)
    try {
      await updateTorrent({
        hash: torrent.hash,
        title: title.trim() || torrent.name || torrent.hash,
        category: category || undefined,
        poster: poster || '',
      })
      await queryClient.invalidateQueries({ queryKey: TORRENTS_QUERY_KEY })
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
      open={open && Boolean(torrent)}
      onClose={onClose}
      size='lg'
      fullScreen={isFullScreenBreakpoint}
      dialogClassName='flex flex-col overflow-hidden'
      dialogStyle={isFullScreenBreakpoint ? undefined : DIALOG_EDIT}
    >
      <Modal.Header className='shrink-0'>
        <Modal.Heading>{t('EditTorrent')}</Modal.Heading>
        <Modal.CloseTrigger aria-label={t('Close')} />
      </Modal.Header>
      <Modal.Body className='flex min-h-0 flex-1 flex-col gap-4 overflow-hidden'>
        <div className='shrink-0 space-y-4'>
          <p className='break-all font-mono text-xs text-muted' title={torrent?.hash}>
            {torrent?.hash}
          </p>

          <TextField value={title} onChange={setTitle} isDisabled={saving} autoFocus>
            <Label>{t('AddDialog.TitleBlank')}</Label>
            <Input />
          </TextField>

          <Select
            selectedKey={category || 'none'}
            onSelectionChange={key => setCategory(key === 'none' ? '' : String(key))}
            isDisabled={saving}
          >
            <Label>{t('Category')}</Label>
            <Select.Trigger>
              <Select.Value />
              <Select.Indicator />
            </Select.Trigger>
            <Select.Popover>
              <ListBox>
                <ListBox.Item id='none'>{t('Uncategorized')}</ListBox.Item>
                {TORRENT_CATEGORIES.map(cat => (
                  <ListBox.Item key={cat.key} id={cat.key}>
                    {t(cat.name)}
                  </ListBox.Item>
                ))}
              </ListBox>
            </Select.Popover>
          </Select>

          <TextField value={poster} onChange={setPoster} isDisabled={saving}>
            <Label>{t('AddDialog.AddPosterLinkInput')}</Label>
            <Input />
          </TextField>
        </div>

        {/* Always mounted — reserved height so TMDB results never resize the dialog. */}
        <div className='flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden rounded-xl border border-border bg-surface-secondary p-3'>
          <p className='mb-2 flex shrink-0 items-center gap-1.5 text-xs text-muted'>
            <Film size={14} strokeWidth={1.75} aria-hidden />
            {t('AddDialog.AddPosterLinkInput')}
            {postersLoading ? '…' : ''}
          </p>
          <div className='min-h-0 flex-1 overflow-y-auto overscroll-contain'>
            {postersLoading && posterOptions.length === 0 ? (
              <div className='flex flex-wrap gap-2' aria-hidden>
                {Array.from({ length: 8 }, (_, i) => (
                  <div key={i} className='h-[132px] w-[88px] animate-pulse rounded-lg bg-surface-tertiary' />
                ))}
              </div>
            ) : posterOptions.length > 0 ? (
              <PosterPicker
                poster={poster}
                posterOptions={posterOptions}
                onSelect={setPoster}
                disabled={saving}
                layout='wrap'
              />
            ) : (
              <div className='min-h-[132px]' aria-hidden />
            )}
          </div>
        </div>
      </Modal.Body>
      <Modal.Footer className='shrink-0'>
        <Button onPress={onClose} isDisabled={saving} variant='secondary' className={footerButtonClassName}>
          {t('Cancel')}
        </Button>
        <Button
          variant='primary'
          onPress={() => void handleSave()}
          isDisabled={saving || !torrent}
          className={footerButtonClassName}
        >
          {saving ? <Spinner size='sm' color='current' /> : t('Save')}
        </Button>
      </Modal.Footer>
    </AppDialog>
  )
}

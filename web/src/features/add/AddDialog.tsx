import {
  Button,
  Description,
  Input,
  Label,
  ListBox,
  Modal,
  Select,
  Spinner,
  Switch,
  TextArea,
  TextField,
  useMediaQuery,
} from '@heroui/react'
import { UploadCloud } from 'lucide-react'
import { iconAction } from 'shared/ui/iconProps'
import { useCallback, useEffect, useRef, useState } from 'react'
import { useDropzone } from 'react-dropzone'
import { useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'

import { addTorrent, TORRENTS_QUERY_KEY } from 'shared/api/torrents'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import { useTorrentsQuery } from 'shared/hooks/useTorrentsQuery'
import { detectApplePlatform } from 'shared/lib/platform'
import {
  checkTorrentSource,
  getMoviePosters,
  parseSourceInfoHash,
  parseTorrentTitle,
  shortenTitleForPosterSearch,
} from 'shared/lib/torrentHelpers'
import { isTorrsLink, toApiTorrsLink } from 'shared/lib/torrsLink'
import { queryMax } from 'shared/theme/breakpoints'
import { TORRENT_CATEGORIES } from 'shared/torrent/categories'
import AppDialog from 'shared/ui/AppDialog'
import { DIALOG_SHEET_M } from 'shared/ui/dialogSizes'
import { useOptionalAppToast } from 'shared/ui/Toast'

import MultiAddDialog from './MultiAddDialog'
import PosterPicker from './PosterPicker'

export interface AddDialogProps {
  open: boolean
  onClose: () => void
  /** Pre-fill magnet/hash/link (PWA protocol handler / share target). */
  initialSource?: string | null
}

/** Add torrent by magnet/hash/URL or file drop; TMDB poster search + category. */
export default function AddDialog({ open, onClose, initialSource }: AddDialogProps) {
  const { t, i18n } = useTranslation()
  const queryClient = useQueryClient()
  const toast = useOptionalAppToast()
  const isMobile = useMediaQuery(queryMax('mobile'))
  // Matches Details fullscreen policy (breakpoint + Appearance pref) so tablets stay consistent.
  const isFullScreenBreakpoint = useDialogFullScreen()

  const [source, setSource] = useState(initialSource || '')
  const [title, setTitle] = useState('')
  const [category, setCategory] = useState('')
  const [poster, setPoster] = useState('')
  const [posterOptions, setPosterOptions] = useState<string[]>([])
  const [postersLoading, setPostersLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [saveToDb, setSaveToDb] = useState(true)
  const [multiFiles, setMultiFiles] = useState<File[] | null>(null)
  const [hashExists, setHashExists] = useState(false)

  const posterRequestRef = useRef(0)
  const titleTouchedRef = useRef(false)

  const { data: torrents } = useTorrentsQuery()

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- seed source from share/launch props
    if (open && initialSource) setSource(initialSource)
  }, [open, initialSource])

  const resetForm = () => {
    setSource('')
    setTitle('')
    setCategory('')
    setPoster('')
    setPosterOptions([])
    setHashExists(false)
    setSaveToDb(true)
    titleTouchedRef.current = false
  }

  // Debounced duplicate-hash check against the currently known torrent list.
  useEffect(() => {
    const trimmed = source.trim()
    if (!open || !trimmed || !checkTorrentSource(trimmed)) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- clear stale duplicate flag when input invalid
      setHashExists(false)
      return undefined
    }

    let cancelled = false
    const timer = window.setTimeout(() => {
      void parseSourceInfoHash(trimmed).then(infoHash => {
        if (cancelled || !infoHash) {
          if (!cancelled) setHashExists(false)
          return
        }
        const exists = (torrents || []).some(item => item.hash?.toLowerCase() === infoHash.toLowerCase())
        setHashExists(exists)
      })
    }, 300)

    return () => {
      cancelled = true
      window.clearTimeout(timer)
    }
  }, [open, source, torrents])

  // Debounced title auto-fill from the parsed torrent/magnet name, unless the user edited it.
  useEffect(() => {
    const trimmed = source.trim()
    if (!open || !trimmed || !checkTorrentSource(trimmed) || titleTouchedRef.current) return undefined

    let cancelled = false
    const timer = window.setTimeout(() => {
      parseTorrentTitle(trimmed, ({ parsedTitle }) => {
        if (cancelled || !parsedTitle || titleTouchedRef.current) return
        setTitle(parsedTitle)
      })
    }, 350)

    return () => {
      cancelled = true
      window.clearTimeout(timer)
    }
  }, [open, source])

  const handleFiles = useCallback(
    (files: File[]) => {
      // Always filter by extension — iOS may ignore/mis-handle accept MIME filters.
      const torrents = files.filter(file => file.name.toLowerCase().endsWith('.torrent'))
      if (!torrents.length) {
        if (files.length > 0) {
          toast?.showToast({ message: t('AddDialog.WrongTorrentSource'), severity: 'error' })
        }
        return
      }
      setMultiFiles(torrents)
    },
    [t, toast],
  )

  // iOS Safari/Edge grey out .torrent when accept lists only application/x-bittorrent
  // (UTType unknown). Omit accept on iOS; keep MIME+ext filter elsewhere for nicer pickers.
  const dropzoneAccept = detectApplePlatform().isIOS
    ? undefined
    : { 'application/x-bittorrent': ['.torrent'], 'application/octet-stream': ['.torrent'] }

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop: handleFiles,
    ...(dropzoneAccept ? { accept: dropzoneAccept } : {}),
    multiple: true,
    noClick: false,
    disabled: saving,
  })

  // Debounced TMDB poster search driven by the (possibly auto-filled) title.
  useEffect(() => {
    const query = shortenTitleForPosterSearch(title.trim()) || title.trim()
    if (!query || query.length < 2) {
      // eslint-disable-next-line react-hooks/set-state-in-effect -- clear posters when title too short
      setPosterOptions([])
      return undefined
    }

    const requestId = ++posterRequestRef.current
    const timer = window.setTimeout(() => {
      setPostersLoading(true)
      const lang = i18n.language?.startsWith('ru') ? 'ru' : 'en'
      void getMoviePosters(query, lang)
        .then(urls => {
          if (requestId !== posterRequestRef.current) return
          setPosterOptions(urls || [])
          setPoster(prev => prev || urls?.[0] || '')
        })
        .finally(() => {
          if (requestId === posterRequestRef.current) setPostersLoading(false)
        })
    }, 450)

    return () => window.clearTimeout(timer)
  }, [title, i18n.language])

  const handleAdd = async () => {
    const trimmed = source.trim()
    if (!trimmed) return
    if (!checkTorrentSource(trimmed)) {
      toast?.showToast({ message: t('AddDialog.WrongTorrentSource'), severity: 'error' })
      return
    }
    if (hashExists) {
      toast?.showToast({ message: t('AddDialog.HashExists'), severity: 'warning' })
      return
    }

    setSaving(true)
    try {
      await addTorrent({
        link: isTorrsLink(trimmed) ? toApiTorrsLink(trimmed) : trimmed,
        title: title || undefined,
        category: category.trim() || undefined,
        poster: poster || '',
        save_to_db: saveToDb,
      })
      await queryClient.invalidateQueries({ queryKey: TORRENTS_QUERY_KEY })
      toast?.showToast({ message: t('Search.TorrentAdded'), severity: 'success' })
      resetForm()
      onClose()
    } catch {
      toast?.showToast({ message: t('Error'), severity: 'error' })
    } finally {
      setSaving(false)
    }
  }

  const trimmedSource = source.trim()
  const sourceInvalid = Boolean(trimmedSource) && !checkTorrentSource(trimmedSource)
  const footerButtonClassName = isMobile ? 'min-h-11 px-4' : undefined

  if (multiFiles) {
    return (
      <MultiAddDialog
        files={multiFiles}
        open={open}
        onClose={() => {
          setMultiFiles(null)
          onClose()
        }}
      />
    )
  }

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
        <Modal.Heading>{t('AddNewTorrent')}</Modal.Heading>
        <Modal.CloseTrigger aria-label={t('Close')} />
      </Modal.Header>
      <Modal.Body className='min-h-0 flex-1 space-y-4 overflow-y-auto overscroll-contain'>
        <TextField value={source} onChange={setSource} isInvalid={hashExists || sourceInvalid} isDisabled={saving}>
          <Label>{t('AddDialog.TorrentSourceLink')}</Label>
          <TextArea rows={2} autoFocus />
          <Description>
            {hashExists
              ? t('AddDialog.HashExists')
              : sourceInvalid
                ? t('AddDialog.WrongTorrentSource')
                : t('AddDialog.TorrentSourceOptions')}
          </Description>
        </TextField>

        <div
          {...getRootProps()}
          className={`grid min-h-16 cursor-pointer place-items-center gap-1.5 rounded-xl border-2 border-dashed p-4 text-center transition-colors ${
            isDragActive ? 'border-accent bg-accent-soft' : 'border-border hover:border-accent/50'
          }`}
        >
          <input {...getInputProps()} />
          <UploadCloud {...iconAction} className='text-muted' aria-hidden />
          <Description>{t('AddDialog.AppendFile.ClickOrDrag')}</Description>
        </div>

        <TextField
          value={title}
          onChange={value => {
            titleTouchedRef.current = true
            setTitle(value)
          }}
          isDisabled={saving}
        >
          <Label>{t('AddDialog.TitleBlank')}</Label>
          <Input />
          <Description>{t('AddDialog.CustomTorrentTitleHelperText')}</Description>
        </TextField>

        <Select
          selectedKey={category || 'none'}
          onSelectionChange={key => setCategory(key === 'none' ? '' : String(key))}
          isDisabled={saving}
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

        {postersLoading || posterOptions.length > 0 || poster ? (
          <div>
            <Description className='mb-2 block'>
              {t('AddDialog.AddPosterLinkInput')}
              {postersLoading ? '…' : ''}
            </Description>
            <PosterPicker
              poster={poster}
              posterOptions={posterOptions}
              onSelect={setPoster}
              disabled={saving}
              layout='scroll'
            />
            <TextField value={poster} onChange={setPoster} isDisabled={saving}>
              <Label>{t('AddDialog.AddPosterLinkInput')}</Label>
              <Input />
            </TextField>
          </div>
        ) : null}

        <div className='flex min-h-11 items-start justify-between gap-4'>
          <div className='min-w-0 flex-1'>
            <p className='text-sm font-medium text-foreground'>{t('AddDialog.SaveToDb')}</p>
            <p className='mt-0.5 text-xs text-muted'>{t('AddDialog.SaveToDbHint')}</p>
          </div>
          <Switch
            id='add-save-to-db'
            isSelected={saveToDb}
            onChange={setSaveToDb}
            isDisabled={saving}
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
      </Modal.Body>
      <Modal.Footer className='shrink-0'>
        <Button onPress={onClose} isDisabled={saving} variant='secondary' className={footerButtonClassName}>
          {t('Cancel')}
        </Button>
        <Button
          variant='primary'
          onPress={() => void handleAdd()}
          isDisabled={saving || !trimmedSource || hashExists || sourceInvalid}
          className={footerButtonClassName}
        >
          {saving ? <Spinner size='sm' color='current' /> : t('Add')}
        </Button>
      </Modal.Footer>
    </AppDialog>
  )
}

import { useMemo, useState } from 'react'
import { Button, Input, Modal, TextField } from '@heroui/react'
import { useTranslation } from 'react-i18next'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import AppDialog from 'shared/ui/AppDialog'
import {
  requestDownloadAllPlaylists,
  requestOpenExportLibrary,
  requestOpenImportLibrary,
} from 'shared/lib/libraryEvents'
import { requestOpenSettings, type SettingsDeepLinkTab } from 'shared/lib/settingsEvents'

export interface CommandPaletteProps {
  open: boolean
  onClose: () => void
  onAdd: () => void
  onSearch: () => void
  onAbout: () => void
  onDonate: () => void
  onServerStatus: () => void
  onToggleTheme: () => void
}

interface CommandItem {
  id: string
  label: string
  run: () => void
}

/** Cmd/Ctrl+K jump list for common library actions. */
export default function CommandPalette({
  open,
  onClose,
  onAdd,
  onSearch,
  onAbout,
  onDonate,
  onServerStatus,
  onToggleTheme,
}: CommandPaletteProps) {
  const { t } = useTranslation()
  const isFullScreen = useDialogFullScreen()
  const [query, setQuery] = useState('')

  const commands = useMemo<CommandItem[]>(() => {
    const openSettings = (tab: SettingsDeepLinkTab) => {
      requestOpenSettings(tab)
      onClose()
    }
    return [
      {
        id: 'add',
        label: t('AddNewTorrent'),
        run: () => {
          onAdd()
          onClose()
        },
      },
      {
        id: 'search',
        label: t('nav.Search'),
        run: () => {
          onSearch()
          onClose()
        },
      },
      {
        id: 'status',
        label: t('ServerStatus'),
        run: () => {
          onServerStatus()
          onClose()
        },
      },
      { id: 'settings', label: t('nav.Settings'), run: () => openSettings('primary') },
      { id: 'appearance', label: t('SettingsDialog.SectionAppearance'), run: () => openSettings('appearance') },
      {
        id: 'theme',
        label: t('Theme'),
        run: () => {
          onToggleTheme()
          onClose()
        },
      },
      {
        id: 'import',
        label: t('ImportLibrary'),
        run: () => {
          requestOpenImportLibrary()
          onClose()
        },
      },
      {
        id: 'export',
        label: t('ExportLibrary'),
        run: () => {
          requestOpenExportLibrary()
          onClose()
        },
      },
      {
        id: 'playlists',
        label: t('DownloadAllPlaylists'),
        run: () => {
          requestDownloadAllPlaylists()
          onClose()
        },
      },
      {
        id: 'about',
        label: t('About'),
        run: () => {
          onAbout()
          onClose()
        },
      },
      {
        id: 'donate',
        label: t('Donate'),
        run: () => {
          onDonate()
          onClose()
        },
      },
    ]
  }, [t, onAdd, onSearch, onAbout, onDonate, onServerStatus, onToggleTheme, onClose])

  const filtered = useMemo(() => {
    const q = query.trim().toLocaleLowerCase()
    if (!q) return commands
    return commands.filter(cmd => cmd.label.toLocaleLowerCase().includes(q) || cmd.id.includes(q))
  }, [commands, query])

  return (
    <AppDialog
      open={open}
      onClose={onClose}
      size='sm'
      fullScreen={isFullScreen}
      dialogClassName='flex flex-col overflow-hidden'
    >
      <Modal.Header className='shrink-0'>
        <Modal.Heading>{t('CommandPalette')}</Modal.Heading>
        <Modal.CloseTrigger aria-label={t('Close')} />
      </Modal.Header>
      <Modal.Body className='flex min-h-0 flex-1 flex-col gap-3 overflow-hidden'>
        <TextField value={query} onChange={setQuery} autoFocus className='shrink-0'>
          <Input placeholder={t('CommandPaletteHint')} className='min-h-11' />
        </TextField>
        <div
          className={
            isFullScreen
              ? 'flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto overscroll-contain'
              : 'flex max-h-72 flex-col gap-1 overflow-y-auto'
          }
        >
          {filtered.map(cmd => (
            <Button key={cmd.id} variant='ghost' className='min-h-11 w-full justify-start' onPress={cmd.run}>
              {cmd.label}
            </Button>
          ))}
          {filtered.length === 0 ? <p className='py-4 text-center text-sm text-muted'>{t('NoSearchResults')}</p> : null}
        </div>
      </Modal.Body>
    </AppDialog>
  )
}

import { Checkbox, Modal } from '@heroui/react'
import { useTranslation } from 'react-i18next'
import type { TorrentCache as TorrentCacheData } from 'shared/api/types'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import AppDialog from 'shared/ui/AppDialog'
import { DIALOG_CACHE } from 'shared/ui/dialogSizes'

import TorrentCache from './TorrentCache'

export interface CacheMapDialogProps {
  open: boolean
  onClose: () => void
  cache: TorrentCacheData
  isSnakeDebugMode: boolean
  onSnakeDebugModeChange: (value: boolean) => void
}

/** Large cache-map workspace stacked above torrent Details. */
export default function CacheMapDialog({
  open,
  onClose,
  cache,
  isSnakeDebugMode,
  onSnakeDebugModeChange,
}: CacheMapDialogProps) {
  const { t } = useTranslation()
  const isFullScreen = useDialogFullScreen()

  return (
    <AppDialog
      open={open}
      onClose={onClose}
      size='lg'
      fullScreen={isFullScreen}
      dialogClassName='flex flex-col overflow-hidden'
      dialogStyle={isFullScreen ? undefined : DIALOG_CACHE}
    >
      <Modal.Header className='flex shrink-0 items-center gap-2'>
        <Modal.Heading className='min-w-0 flex-1 truncate'>{t('DetailedCacheView.header')}</Modal.Heading>
        <Checkbox isSelected={isSnakeDebugMode} onChange={onSnakeDebugModeChange}>
          <Checkbox.Content>
            <Checkbox.Control>
              <Checkbox.Indicator />
            </Checkbox.Control>
            {t('SnakeDebug')}
          </Checkbox.Content>
        </Checkbox>
        <Modal.CloseTrigger aria-label={t('Close')} />
      </Modal.Header>

      <Modal.Body className='flex min-h-0 flex-1 flex-col overflow-hidden'>
        <div className='flex min-h-0 min-w-0 flex-1 flex-col'>
          <TorrentCache cache={cache} mode='detailed' isSnakeDebugMode={isSnakeDebugMode} />
        </div>
      </Modal.Body>
    </AppDialog>
  )
}

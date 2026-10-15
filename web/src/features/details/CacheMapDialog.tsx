import { Checkbox, Modal } from '@heroui/react'
import { useTranslation } from 'react-i18next'
import type { TorrentCache as TorrentCacheData, TorrentStat } from 'shared/api/types'
import { humanizeSpeed } from 'shared/lib/format'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import AppDialog from 'shared/ui/AppDialog'
import { DIALOG_CACHE } from 'shared/ui/dialogSizes'

import TorrentCache from './TorrentCache'

export interface CacheMapDialogProps {
  open: boolean
  onClose: () => void
  cache: TorrentCacheData
  torrent: TorrentStat
  isSnakeDebugMode: boolean
  onSnakeDebugModeChange: (value: boolean) => void
  hash?: string
}

/** Large cache-map workspace stacked above torrent Details. */
export default function CacheMapDialog({
  open,
  onClose,
  cache,
  torrent,
  isSnakeDebugMode,
  onSnakeDebugModeChange,
  hash,
}: CacheMapDialogProps) {
  const { t } = useTranslation()
  const isFullScreen = useDialogFullScreen()
  const activePeers = torrent.active_peers
  const totalPeers = torrent.total_peers
  const peersLabel =
    activePeers == null ? '—' : totalPeers != null ? `${activePeers}/${totalPeers}` : String(activePeers)
  const seedersLabel = torrent.connected_seeders != null ? String(torrent.connected_seeders) : '—'

  return (
    <AppDialog
      open={open}
      onClose={onClose}
      size='lg'
      fullScreen={isFullScreen}
      dialogClassName='ts-cache-map-modal flex flex-col overflow-hidden'
      dialogStyle={isFullScreen ? undefined : DIALOG_CACHE}
    >
      <Modal.Header className='flex shrink-0 flex-wrap items-center gap-x-3 gap-y-2 px-3 sm:px-0'>
        <Modal.Heading className='min-w-0 shrink truncate sm:max-w-[40%]'>
          {t('DetailedCacheView.header')}
        </Modal.Heading>
        <div className='flex min-w-0 flex-1 flex-wrap items-center justify-center gap-x-4 gap-y-1 text-xs sm:text-sm'>
          <span className='inline-flex items-center gap-1.5 tabular-nums'>
            <span className='size-2 shrink-0 rounded-full bg-accent' aria-hidden />
            <span className='text-muted'>{t('DownloadSpeed')}</span>
            <span className='font-semibold text-foreground'>{humanizeSpeed(torrent.download_speed)}</span>
          </span>
          <span className='inline-flex items-center gap-1.5 tabular-nums'>
            <span className='text-muted'>{t('Peers')}</span>
            <span className='font-semibold text-foreground'>{peersLabel}</span>
          </span>
          <span className='inline-flex items-center gap-1.5 tabular-nums'>
            <span className='text-muted'>{t('Seeders')}</span>
            <span className='font-semibold text-foreground'>{seedersLabel}</span>
          </span>
        </div>
        <Checkbox isSelected={isSnakeDebugMode} onChange={onSnakeDebugModeChange} className='shrink-0'>
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
          <TorrentCache cache={cache} mode='detailed' isSnakeDebugMode={isSnakeDebugMode} hash={hash} />
        </div>
      </Modal.Body>
    </AppDialog>
  )
}

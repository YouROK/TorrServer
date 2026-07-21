import { useQuery } from '@tanstack/react-query'
import { Button, Modal, Spinner } from '@heroui/react'
import { Activity, RefreshCw } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { getRuntimeStatus, RUNTIME_STATUS_QUERY_KEY } from 'shared/api/runtime'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import { humanizeSize, humanizeSpeed } from 'shared/lib/format'
import AppDialog from 'shared/ui/AppDialog'
import { DIALOG_SHEET_L } from 'shared/ui/dialogSizes'
import { iconMenu } from 'shared/ui/iconProps'

export interface ServerStatusDialogProps {
  open: boolean
  onClose: () => void
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className='min-w-0 rounded-xl border border-border bg-surface-secondary px-3 py-2.5 text-center'>
      <p className='truncate text-[11px] text-muted' title={label}>
        {label}
      </p>
      <p className='mt-0.5 truncate text-sm font-bold tabular-nums text-foreground' title={value}>
        {value}
      </p>
    </div>
  )
}

/** Live BitTorrent + integrations status — structured from GET /runtime/status (with raw /stat dump). */
export default function ServerStatusDialog({ open, onClose }: ServerStatusDialogProps) {
  const { t } = useTranslation()
  const isFullScreen = useDialogFullScreen()
  const { data, isLoading, isFetching, isError, refetch } = useQuery({
    queryKey: RUNTIME_STATUS_QUERY_KEY,
    queryFn: ({ signal }) => getRuntimeStatus(signal),
    enabled: open,
    refetchInterval: open ? 2500 : false,
  })

  const bt = data?.bt
  const torrents = bt?.torrents ?? []
  const loadedPct =
    bt?.total_size && bt.total_size > 0 ? Math.min(100, ((bt.loaded_size ?? 0) / bt.total_size) * 100) : 0

  return (
    <AppDialog
      open={open}
      onClose={onClose}
      size='lg'
      fullScreen={isFullScreen}
      dialogClassName='flex flex-col overflow-hidden'
      dialogStyle={
        isFullScreen ? undefined : { ...DIALOG_SHEET_L, height: 'min(80dvh, 46rem)', maxHeight: 'min(80dvh, 46rem)' }
      }
    >
      <Modal.Header className='flex shrink-0 items-center gap-2'>
        <Modal.Heading className='min-w-0 flex-1 truncate'>{t('ServerStatus')}</Modal.Heading>
        <Button
          isIconOnly
          variant='ghost'
          className='shrink-0'
          aria-label={t('Refresh')}
          isDisabled={isFetching}
          onPress={() => void refetch()}
        >
          {isFetching ? <Spinner size='sm' color='current' /> : <RefreshCw {...iconMenu} aria-hidden />}
        </Button>
        <Modal.CloseTrigger aria-label={t('Close')} className='shrink-0' />
      </Modal.Header>

      <Modal.Body className='min-h-0 flex-1 space-y-4 overflow-y-auto overscroll-contain'>
        {isLoading ? (
          <div className='grid place-items-center py-16'>
            <Spinner />
          </div>
        ) : isError ? (
          <p className='py-8 text-center text-sm text-danger'>{t('Error')}</p>
        ) : (
          <>
            <section>
              <div className='mb-2 flex items-center gap-2'>
                <Activity size={16} strokeWidth={1.75} className='text-accent' aria-hidden />
                <h3 className='text-xs font-semibold tracking-wide text-muted uppercase'>{t('ServerStatusClient')}</h3>
              </div>
              <div className='grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-4'>
                <StatCard
                  label={t('ServerStatusListenPort')}
                  value={bt?.listen_port != null ? String(bt.listen_port) : '—'}
                />
                <StatCard label={t('ServerStatusTorrents')} value={String(bt?.torrent_count ?? 0)} />
                <StatCard label={t('ServerStatusStreams')} value={String(bt?.active_streams ?? 0)} />
                <StatCard label={t('ServerStatusBannedIPs')} value={String(bt?.banned_ips ?? 0)} />
                <StatCard label={t('DownloadSpeed')} value={humanizeSpeed(bt?.download_speed)} />
                <StatCard label={t('UploadSpeed')} value={humanizeSpeed(bt?.upload_speed)} />
                <StatCard label={t('Peers')} value={`${bt?.active_peers ?? 0} / ${bt?.total_peers ?? 0}`} />
                <StatCard label={t('ServerStatusSeeders')} value={String(bt?.connected_seeders ?? 0)} />
                <StatCard label={t('BytesRead')} value={humanizeSize(bt?.bytes_read)} />
                <StatCard label={t('BytesWritten')} value={humanizeSize(bt?.bytes_written)} />
                <StatCard label={t('Size')} value={humanizeSize(bt?.total_size)} />
                <StatCard
                  label={t('ServerStatusLoaded')}
                  value={`${humanizeSize(bt?.loaded_size)} · ${loadedPct.toFixed(1)}%`}
                />
              </div>
              {bt?.peer_id ? (
                <p className='mt-2 truncate font-mono text-xs text-muted' title={bt.peer_id}>
                  Peer ID: {bt.peer_id}
                </p>
              ) : null}
            </section>

            <section>
              <h3 className='mb-2 text-xs font-semibold tracking-wide text-muted uppercase'>
                {t('IntegrationsStatus')}
              </h3>
              <div className='grid grid-cols-2 gap-2 sm:grid-cols-4'>
                <StatCard label={t('DLNAStatus')} value={data?.dlna_enabled ? t('Enabled') : t('Disabled')} />
                <StatCard label={t('BonjourStatus')} value={data?.bonjour_enabled ? t('Enabled') : t('Disabled')} />
                <StatCard label={t('WebDAVStatus')} value={data?.webdav_enabled ? t('Enabled') : t('Disabled')} />
                <StatCard
                  label={t('FuseStatus')}
                  value={data?.fuse_enabled ? data.fuse_path || t('Enabled') : t('Disabled')}
                />
              </div>
              {data?.friendly_name ? (
                <p className='mt-2 text-xs text-muted'>
                  {t('SettingsDialog.FriendlyName')}: <span className='text-foreground'>{data.friendly_name}</span>
                </p>
              ) : null}
            </section>

            <section>
              <h3 className='mb-2 text-xs font-semibold tracking-wide text-muted uppercase'>
                {t('ServerStatusTorrents')}
              </h3>
              {torrents.length === 0 ? (
                <p className='text-sm text-muted'>{t('NoTorrentsAdded')}</p>
              ) : (
                <div className='space-y-2'>
                  {torrents.map(tr => {
                    const pct =
                      tr.torrent_size && tr.torrent_size > 0
                        ? Math.min(100, ((tr.loaded_size ?? 0) / tr.torrent_size) * 100)
                        : 0
                    return (
                      <div key={tr.hash} className='rounded-xl border border-border bg-surface px-3 py-2.5'>
                        <p className='truncate text-sm font-semibold text-foreground' title={tr.title || tr.name}>
                          {tr.title || tr.name || tr.hash}
                        </p>
                        <p className='mt-0.5 truncate text-xs text-muted'>
                          {tr.stat_string || '—'} · {pct.toFixed(1)}% · {humanizeSize(tr.loaded_size)} /{' '}
                          {humanizeSize(tr.torrent_size)}
                        </p>
                        <p className='mt-0.5 text-xs tabular-nums text-muted'>
                          ↓ {humanizeSpeed(tr.download_speed)} · ↑ {humanizeSpeed(tr.upload_speed)} · {t('Peers')}{' '}
                          {tr.active_peers ?? 0}/{tr.total_peers ?? 0} · {t('ServerStatusSeeders')}{' '}
                          {tr.connected_seeders ?? 0}
                        </p>
                      </div>
                    )
                  })}
                </div>
              )}
            </section>

            {bt?.raw_stat ? (
              <section>
                <h3 className='mb-2 text-xs font-semibold tracking-wide text-muted uppercase'>
                  {t('ServerStatusRaw')}
                </h3>
                <pre className='max-h-64 overflow-auto rounded-xl border border-border bg-surface-secondary p-3 font-mono text-[11px] leading-relaxed text-muted whitespace-pre-wrap'>
                  {bt.raw_stat}
                </pre>
              </section>
            ) : null}
          </>
        )}
      </Modal.Body>
    </AppDialog>
  )
}

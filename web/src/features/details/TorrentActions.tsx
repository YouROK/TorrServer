import { useMemo, memo, useState, type ReactNode } from 'react'
import { Button, ButtonGroup, Drawer, Modal, Separator, useMediaQuery, useOverlayState } from '@heroui/react'
import {
  EyeOff,
  Hash,
  Link2,
  ListMusic,
  Magnet,
  MoreHorizontal,
  Settings,
  Share2,
  SquareArrowOutUpRight,
  Trash2,
} from 'lucide-react'
import ptt from 'parse-torrent-title'
import { useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import type { PlayableFile, TorrentStat } from 'shared/api/types'
import { playlistAllUrl, torrsShareUrl } from 'shared/api/extras'
import { playlistTorrHost, streamHost } from 'shared/api/hosts'
import { dropTorrent, markTorrentsDroppedInList, removeTorrent, TORRENTS_QUERY_KEY } from 'shared/api/torrents'
import { clearViewedFiles } from 'shared/api/viewed'
import { useExternalPlayers } from 'shared/lib/externalPlayers'
import { copyToClipboard } from 'shared/lib/clipboard'
import { requestOpenSettings } from 'shared/lib/settingsEvents'
import { queryMax } from 'shared/theme/breakpoints'
import { iconMenu } from 'shared/ui/iconProps'
import { useOptionalAppToast } from 'shared/ui/Toast'
import { usePlayLauncher } from 'features/player/usePlayLauncher'

export interface TorrentActionsProps {
  hash: string
  torrsHash?: string
  viewedFileList?: number[]
  playableFileList?: PlayableFile[]
  name?: string
  title?: string
  setViewedFileList: (list?: number[]) => void
  onViewedChange?: () => void
  onDropped?: () => void
  /** After permanent delete — close details. */
  onDeleted?: () => void
  /** Continue Watching: auto-play this file when the list is ready. */
  autoPlayFileId?: number
  autoPlayTimecode?: number
  /** Phone/fullscreen Details — secondary actions in a bottom sheet. */
  compact?: boolean
}

type PendingConfirm = 'drop' | 'delete' | 'clearViews' | null

/** Renders Infuse/VLC/… as equal secondary pills (same look as Copy link beside them). */
function ExternalPlayersGroup({
  players,
  size = 'md',
  compact = false,
  stretch = false,
}: {
  players: { label: string; href: string }[]
  size?: 'sm' | 'md' | 'lg'
  compact?: boolean
  /** Equal-width chips across the row (desktop Stats). */
  stretch?: boolean
}) {
  if (players.length === 0) return null

  return (
    <>
      {players.map(player => (
        <Button
          key={player.label}
          variant='secondary'
          size={compact ? 'sm' : size}
          className={
            compact ? 'min-h-11 max-w-[5rem] px-2 text-xs' : stretch ? 'min-h-11 min-w-0 flex-1 px-2' : 'min-h-11'
          }
          onPress={() => {
            window.location.href = player.href
          }}
        >
          {compact ? null : <SquareArrowOutUpRight {...iconMenu} aria-hidden />}
          <span className='truncate'>{player.label}</span>
        </Button>
      ))}
    </>
  )
}

/**
 * Stats-tab action block: playlist / magnet / hash / drop / clear viewed.
 * Play + Cache live on their own tabs; copy helpers use {@link copyToClipboard}.
 */
function TorrentActions({
  hash,
  torrsHash,
  viewedFileList,
  playableFileList,
  name,
  title,
  setViewedFileList,
  onViewedChange,
  onDropped,
  onDeleted,
  autoPlayFileId,
  autoPlayTimecode,
  compact: compactProp = false,
}: TorrentActionsProps) {
  const { t } = useTranslation()
  const toast = useOptionalAppToast()
  const queryClient = useQueryClient()
  const isPhone = useMediaQuery(queryMax('mobile'))
  const compact = compactProp || isPhone
  const [pendingConfirm, setPendingConfirm] = useState<PendingConfirm>(null)
  const moreState = useOverlayState()
  const confirmState = useOverlayState({
    isOpen: pendingConfirm != null,
    onOpenChange: open => {
      if (!open) setPendingConfirm(null)
    },
  })

  const isSingleFileTorrent = playableFileList?.length === 1
  const latestViewedFileId = viewedFileList?.[viewedFileList.length - 1]
  const latestViewedFilePath = playableFileList?.find(({ id }) => id === latestViewedFileId)?.path
  const latestViewedFileInfo = latestViewedFilePath ? ptt.parse(latestViewedFilePath) : null

  const displayName = name || title || 'file'
  const fullPlaylistLink = `${playlistTorrHost()}/${encodeURIComponent(displayName)}.m3u?link=${hash}&m3u`
  const fromLatestPlaylistLink = `${fullPlaylistLink}&fromlast`
  const magnetLink = `magnet:?xt=urn:btih:${hash}&dn=${encodeURIComponent(name || title || '')}`

  /** Keep launcher mounted for Continue Watching autoplay + player modals. */
  const { playerModals } = usePlayLauncher({
    hash,
    displayName,
    knownPlayableFiles: playableFileList || [],
    onViewedChange,
    autoPlayFileId,
    autoPlayTimecode,
  })

  /** Only offer app deep links when there's exactly one obvious file to hand off. */
  const { buildExternalPlayers, hasAnyExternalPlayer } = useExternalPlayers()
  const singleFileStream = useMemo(() => {
    if (playableFileList?.length !== 1) return null
    const file = playableFileList[0]
    const fileName = file.path.split('\\').pop()?.split('/').pop() || file.path
    const link = `${streamHost()}/${encodeURIComponent(fileName)}?link=${hash}&index=${file.id}&play`
    const fullLink = new URL(link, window.location.href).toString()
    return { link, fullLink, externalPlayers: buildExternalPlayers(fullLink) }
  }, [playableFileList, hash, buildExternalPlayers])
  const externalPlayers = singleFileStream?.externalPlayers ?? []

  const runPendingConfirm = () => {
    if (pendingConfirm === 'drop' || pendingConfirm === 'delete') {
      const previous = queryClient.getQueryData<TorrentStat[]>(TORRENTS_QUERY_KEY)
      if (pendingConfirm === 'drop') {
        markTorrentsDroppedInList(queryClient, hash)
      } else {
        queryClient.setQueryData<TorrentStat[]>(TORRENTS_QUERY_KEY, prev => prev?.filter(item => item.hash !== hash))
      }
      const mutate = pendingConfirm === 'drop' ? dropTorrent : removeTorrent
      const successMessage = pendingConfirm === 'drop' ? t('DropTorrent') : t('Delete')
      const afterSuccess = pendingConfirm === 'drop' ? onDropped : onDeleted
      void mutate(hash)
        .then(() => {
          toast?.showToast({ message: successMessage, severity: 'success' })
          void queryClient.invalidateQueries({ queryKey: TORRENTS_QUERY_KEY })
          afterSuccess?.()
        })
        .catch(() => {
          if (previous) queryClient.setQueryData(TORRENTS_QUERY_KEY, previous)
          toast?.showToast({ message: t('Error'), severity: 'error' })
        })
    } else if (pendingConfirm === 'clearViews') {
      void clearViewedFiles(hash)
        .then(() => {
          setViewedFileList(undefined)
          toast?.showToast({ message: t('RemoveViews'), severity: 'success' })
        })
        .catch(() => toast?.showToast({ message: t('Error'), severity: 'error' }))
    }
    setPendingConfirm(null)
  }

  const confirmHeading =
    pendingConfirm === 'drop' ? t('DropTorrent') : pendingConfirm === 'delete' ? t('Delete') : t('RemoveViews')
  const confirmBody =
    pendingConfirm === 'drop'
      ? t('ConfirmDropTorrent')
      : pendingConfirm === 'delete'
        ? t('ConfirmDeleteTorrent')
        : t('ConfirmRemoveViews')

  const copyMagnetLink = async () => {
    try {
      await copyToClipboard(magnetLink)
      toast?.showToast({ message: t('Copied'), severity: 'success' })
    } catch {
      toast?.showToast({ message: t('Error'), severity: 'error' })
    }
  }

  const copyInfoHash = async () => {
    try {
      await copyToClipboard(hash)
      toast?.showToast({ message: t('Copied'), severity: 'success' })
    } catch {
      toast?.showToast({ message: t('Error'), severity: 'error' })
    }
  }

  const copyTorrsLink = async () => {
    try {
      await copyToClipboard(torrsShareUrl({ hash, torrs_hash: torrsHash }))
      toast?.showToast({ message: t('Copied'), severity: 'success' })
    } catch {
      toast?.showToast({ message: t('Error'), severity: 'error' })
    }
  }

  const copyStreamLink = async () => {
    if (!singleFileStream) return
    try {
      await copyToClipboard(singleFileStream.fullLink)
      toast?.showToast({ message: t('Copied'), severity: 'success' })
    } catch {
      toast?.showToast({ message: t('Error'), severity: 'error' })
    }
  }

  const hasPartialProgress = !isSingleFileTorrent && !!viewedFileList?.length

  const confirmModal = (
    <Modal.Root state={confirmState}>
      <Modal.Backdrop>
        <Modal.Container size='sm' placement='center'>
          <Modal.Dialog className='ts-compact-modal'>
            <Modal.Header className='shrink-0'>
              <Modal.Heading>{confirmHeading}</Modal.Heading>
            </Modal.Header>
            <Modal.Body>{confirmBody}</Modal.Body>
            <Modal.Footer className='shrink-0'>
              <Button variant='secondary' onPress={() => setPendingConfirm(null)} autoFocus>
                {t('Cancel')}
              </Button>
              <Button
                variant={pendingConfirm === 'drop' || pendingConfirm === 'delete' ? 'danger' : 'primary'}
                onPress={runPendingConfirm}
              >
                {t('OK')}
              </Button>
            </Modal.Footer>
          </Modal.Dialog>
        </Modal.Container>
      </Modal.Backdrop>
    </Modal.Root>
  )

  if (compact) {
    const closeMore = () => moreState.close()
    const sheetAction = (label: string, icon: ReactNode, onPress: () => void, danger = false) => (
      <Button
        key={label}
        variant='ghost'
        onPress={() => {
          closeMore()
          onPress()
        }}
        className={`h-auto w-full justify-start gap-3 px-4 py-3 min-h-11 ${danger ? 'text-danger' : ''}`}
      >
        {icon}
        {label}
      </Button>
    )

    return (
      <div className='space-y-2.5 pb-[max(0.75rem,env(safe-area-inset-bottom,0px))]'>
        <div className='flex items-center justify-end gap-2'>
          <Button variant='secondary' size='md' className='min-h-11 min-w-0 flex-1 gap-2' onPress={moreState.open}>
            <MoreHorizontal {...iconMenu} aria-hidden />
            <span className='truncate text-sm'>{t('Actions')}</span>
          </Button>
        </div>

        {externalPlayers.length > 0 ? (
          <div className='flex flex-wrap items-center gap-1.5'>
            <ExternalPlayersGroup players={externalPlayers} compact />
          </div>
        ) : null}

        {isSingleFileTorrent && !hasAnyExternalPlayer ? (
          <button
            type='button'
            onClick={() => requestOpenSettings('app')}
            className='flex items-center gap-1.5 text-xs text-muted transition-colors hover:text-accent'
          >
            <Settings size={14} strokeWidth={1.75} aria-hidden />
            {t('ExternalPlayersHint')}
          </button>
        ) : null}

        {hasPartialProgress ? (
          <div className='rounded-lg border border-border bg-surface-secondary p-3'>
            <p className='mb-2 truncate text-xs text-muted'>
              {t('LatestFilePlayed')}{' '}
              <strong className='text-foreground'>
                {latestViewedFileInfo?.title}
                {latestViewedFileInfo?.season ? (
                  <>
                    {' '}
                    · {t('Season')} {latestViewedFileInfo.season} · {t('Episode')} {latestViewedFileInfo.episode}
                  </>
                ) : null}
              </strong>
            </p>
            <ButtonGroup className='w-full'>
              <Button
                variant='primary'
                size='sm'
                className='min-h-11 flex-1'
                onPress={() => window.open(fullPlaylistLink, '_blank')}
              >
                {t('Full')}
              </Button>
              <Button
                variant='primary'
                size='sm'
                className='min-h-11 flex-1'
                onPress={() => window.open(fromLatestPlaylistLink, '_blank')}
              >
                {t('FromLatestFile')}
              </Button>
            </ButtonGroup>
          </div>
        ) : null}

        <Drawer.Root state={moreState}>
          <Drawer.Backdrop isDismissable>
            <Drawer.Content placement='bottom'>
              <Drawer.Dialog className='ts-sheet-drawer' aria-label={t('Info')}>
                <Drawer.Header>
                  <Drawer.Heading>{t('Info')}</Drawer.Heading>
                  <Drawer.CloseTrigger className='min-h-11 min-w-11' aria-label={t('Close')} />
                </Drawer.Header>
                <Drawer.Body className='flex flex-col gap-0.5 px-0 pt-1'>
                  {singleFileStream
                    ? sheetAction(t('CopyLink'), <Link2 {...iconMenu} aria-hidden />, () => void copyStreamLink())
                    : null}
                  {isSingleFileTorrent || !viewedFileList?.length
                    ? sheetAction(t('DownloadPlaylist'), <ListMusic {...iconMenu} aria-hidden />, () =>
                        window.open(fullPlaylistLink, '_blank'),
                      )
                    : null}
                  {sheetAction(t('CopyMagnet'), <Magnet {...iconMenu} aria-hidden />, () => void copyMagnetLink())}
                  {sheetAction(t('CopyHash'), <Hash {...iconMenu} aria-hidden />, () => void copyInfoHash())}
                  {sheetAction(t('CopyTorrs'), <Share2 {...iconMenu} aria-hidden />, () => void copyTorrsLink())}
                  {sheetAction(t('DownloadAllPlaylists'), <ListMusic {...iconMenu} aria-hidden />, () =>
                    window.open(playlistAllUrl({ category: undefined }), '_blank'),
                  )}
                  {sheetAction(t('RemoveViews'), <EyeOff {...iconMenu} aria-hidden />, () =>
                    setPendingConfirm('clearViews'),
                  )}
                  {sheetAction(
                    t('DropTorrent'),
                    <Trash2 {...iconMenu} aria-hidden />,
                    () => setPendingConfirm('drop'),
                    true,
                  )}
                  {sheetAction(
                    t('Delete'),
                    <Trash2 {...iconMenu} aria-hidden />,
                    () => setPendingConfirm('delete'),
                    true,
                  )}
                </Drawer.Body>
              </Drawer.Dialog>
            </Drawer.Content>
          </Drawer.Backdrop>
        </Drawer.Root>

        {playerModals}
        {confirmModal}
      </div>
    )
  }

  return (
    <div className='space-y-4'>
      {externalPlayers.length > 0 || singleFileStream ? (
        <div className='flex w-full items-center gap-2'>
          <ExternalPlayersGroup players={externalPlayers} stretch />
          {singleFileStream ? (
            <Button variant='secondary' className='min-h-11 min-w-0 flex-1 px-2' onPress={() => void copyStreamLink()}>
              <Link2 {...iconMenu} aria-hidden />
              <span className='truncate'>{t('CopyLink')}</span>
            </Button>
          ) : null}
        </div>
      ) : null}

      {isSingleFileTorrent && !hasAnyExternalPlayer ? (
        <button
          type='button'
          onClick={() => requestOpenSettings('app')}
          className='flex items-center gap-1.5 text-xs text-muted transition-colors hover:text-accent'
        >
          <Settings size={14} strokeWidth={1.75} aria-hidden />
          {t('ExternalPlayersHint')}
        </button>
      ) : null}

      {hasPartialProgress ? (
        <div className='rounded-xl border border-border bg-surface-secondary p-4'>
          <p className='mb-1 flex items-center gap-2 text-sm font-semibold'>
            <ListMusic {...iconMenu} className='text-accent' aria-hidden />
            {t('DownloadPlaylist')}
          </p>
          <p className='mb-3 text-sm text-muted'>
            {t('LatestFilePlayed')}{' '}
            <strong className='text-foreground'>
              {latestViewedFileInfo?.title}
              {latestViewedFileInfo?.season ? (
                <>
                  {' '}
                  · {t('Season')} {latestViewedFileInfo.season} · {t('Episode')} {latestViewedFileInfo.episode}
                </>
              ) : null}
            </strong>
          </p>
          <ButtonGroup className='w-full'>
            <Button
              variant='primary'
              className='min-h-11 flex-1'
              onPress={() => window.open(fullPlaylistLink, '_blank')}
            >
              {t('Full')}
            </Button>
            <Button
              variant='primary'
              className='min-h-11 flex-1'
              onPress={() => window.open(fromLatestPlaylistLink, '_blank')}
            >
              {t('FromLatestFile')}
            </Button>
          </ButtonGroup>
        </div>
      ) : null}

      <div>
        <p className='mb-2 text-sm font-semibold text-muted'>{t('Info')}</p>
        <div className='flex w-full flex-wrap items-stretch gap-2'>
          {isSingleFileTorrent || !viewedFileList?.length ? (
            <Button
              variant='primary'
              className='min-h-11 min-w-0 flex-1'
              onPress={() => window.open(fullPlaylistLink, '_blank')}
            >
              <ListMusic {...iconMenu} aria-hidden />
              <span className='truncate'>{t('DownloadPlaylist')}</span>
            </Button>
          ) : null}
          <ButtonGroup className='min-w-0 flex-[2]'>
            <Button variant='secondary' className='min-h-11 min-w-0 flex-1' onPress={() => void copyMagnetLink()}>
              <Magnet {...iconMenu} aria-hidden />
              <span className='truncate'>{t('CopyMagnet')}</span>
            </Button>
            <Button variant='secondary' className='min-h-11 min-w-0 flex-1' onPress={() => void copyInfoHash()}>
              <Hash {...iconMenu} aria-hidden />
              <span className='truncate'>{t('CopyHash')}</span>
            </Button>
            <Button variant='secondary' className='min-h-11 min-w-0 flex-1' onPress={() => void copyTorrsLink()}>
              <Share2 {...iconMenu} aria-hidden />
              <span className='truncate'>{t('CopyTorrs')}</span>
            </Button>
          </ButtonGroup>
          <Button
            variant='tertiary'
            className='min-h-11 min-w-0 flex-1'
            onPress={() => window.open(playlistAllUrl({ category: undefined }), '_blank')}
          >
            <ListMusic {...iconMenu} aria-hidden />
            <span className='truncate'>{t('DownloadAllPlaylists')}</span>
          </Button>
        </div>
      </div>

      <Separator />

      <div>
        <p className='mb-2 text-sm font-semibold text-muted'>{t('TorrentState')}</p>
        <div className='flex w-full flex-wrap gap-2'>
          <Button variant='outline' className='min-h-11 min-w-0 flex-1' onPress={() => setPendingConfirm('clearViews')}>
            <EyeOff {...iconMenu} aria-hidden />
            <span className='truncate'>{t('RemoveViews')}</span>
          </Button>
          <Button variant='danger' className='min-h-11 min-w-0 flex-1' onPress={() => setPendingConfirm('drop')}>
            <Trash2 {...iconMenu} aria-hidden />
            <span className='truncate'>{t('DropTorrent')}</span>
          </Button>
          <Button variant='danger' className='min-h-11 min-w-0 flex-1' onPress={() => setPendingConfirm('delete')}>
            <Trash2 {...iconMenu} aria-hidden />
            <span className='truncate'>{t('Delete')}</span>
          </Button>
        </div>
      </div>

      {playerModals}
      {confirmModal}
    </div>
  )
}

export default memo(TorrentActions)

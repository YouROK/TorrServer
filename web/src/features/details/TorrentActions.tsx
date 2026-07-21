import { useMemo, memo, useState, type ReactNode } from 'react'
import { Button, ButtonGroup, Dropdown, Modal, Separator, Spinner, useMediaQuery, useOverlayState } from '@heroui/react'
import {
  EyeOff,
  Hash,
  Link2,
  ListMusic,
  Magnet,
  MoreHorizontal,
  Play,
  Settings,
  Share2,
  SquareArrowOutUpRight,
  Trash2,
} from 'lucide-react'
import ptt from 'parse-torrent-title'
import { useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import type { PlayableFile } from 'shared/api/types'
import { playlistAllUrl, torrsShareUrl } from 'shared/api/extras'
import { playlistTorrHost, streamHost } from 'shared/api/hosts'
import { dropTorrent, removeTorrent, TORRENTS_QUERY_KEY } from 'shared/api/torrents'
import { clearViewedFiles } from 'shared/api/viewed'
import { useExternalPlayers } from 'shared/lib/externalPlayers'
import { copyToClipboard } from 'shared/lib/clipboard'
import { requestOpenSettings } from 'shared/lib/settingsEvents'
import { queryMax } from 'shared/theme/breakpoints'
import { iconBtn } from 'shared/ui/controlClasses'
import { iconMenu } from 'shared/ui/iconProps'
import { useOptionalAppToast } from 'shared/ui/Toast'
import { useConfiguredPlayAction } from 'features/player/useConfiguredPlayAction'
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
  /** Switch details sheet to the Files tab (multi-file "Play" entry point). */
  onShowFiles?: () => void
  /** Continue Watching: auto-play this file when the list is ready. */
  autoPlayFileId?: number
  autoPlayTimecode?: number
  /** Phone/fullscreen Details — denser Play row + secondary actions in a menu. */
  compact?: boolean
}

type PendingConfirm = 'drop' | 'delete' | 'clearViews' | null

/** Renders Infuse/VLC/… as a single button or a ButtonGroup when several are enabled. */
function ExternalPlayersGroup({
  players,
  size = 'md',
  compact = false,
}: {
  players: { label: string; href: string }[]
  size?: 'sm' | 'md' | 'lg'
  compact?: boolean
}) {
  if (players.length === 0) return null

  const buttons = players.map(player => (
    <Button
      key={player.label}
      variant='secondary'
      size={compact ? 'sm' : size}
      className={compact ? 'min-h-11 max-w-[5rem] px-2 text-xs' : 'min-h-11'}
      onPress={() => {
        window.location.href = player.href
      }}
    >
      {compact ? null : <SquareArrowOutUpRight {...iconMenu} aria-hidden />}
      <span className='truncate'>{player.label}</span>
    </Button>
  ))

  if (players.length === 1) return buttons[0]
  return <ButtonGroup>{buttons}</ButtonGroup>
}

/**
 * Overview-tab action block: Play / playlist / magnet / hash / drop / clear viewed.
 * Copy helpers go through {@link copyToClipboard} so LAN HTTP phones do not error.
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
  onShowFiles,
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

  const { handlePlay, resolvePlayableFile, isResolving, playerModals } = usePlayLauncher({
    hash,
    displayName,
    knownPlayableFiles: playableFileList || [],
    onViewedChange,
    autoPlayFileId,
    autoPlayTimecode,
  })

  const { runConfiguredPlay } = useConfiguredPlayAction()

  /** Only offer app deep links when there's exactly one obvious file to hand off — otherwise Play's file picker covers it. */
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
    if (pendingConfirm === 'drop') {
      void dropTorrent(hash)
        .then(async () => {
          toast?.showToast({ message: t('DropTorrent'), severity: 'success' })
          await queryClient.invalidateQueries({ queryKey: TORRENTS_QUERY_KEY })
          onDropped?.()
        })
        .catch(() => toast?.showToast({ message: t('Error'), severity: 'error' }))
    } else if (pendingConfirm === 'delete') {
      void removeTorrent(hash)
        .then(async () => {
          toast?.showToast({ message: t('Delete'), severity: 'success' })
          await queryClient.invalidateQueries({ queryKey: TORRENTS_QUERY_KEY })
          onDeleted?.()
        })
        .catch(() => toast?.showToast({ message: t('Error'), severity: 'error' }))
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
  const playLabel: ReactNode =
    !isSingleFileTorrent && (playableFileList?.length ?? 0) > 1
      ? `${t('TorrentContent')} (${playableFileList!.length})`
      : t('Play')

  const onPlayPress = () => {
    runConfiguredPlay({
      hash,
      displayName,
      knownPlayableFiles: playableFileList || [],
      handlePlay,
      resolvePlayableFile,
      copyText: text =>
        copyToClipboard(text)
          .then(() => toast?.showToast({ message: t('Copied'), severity: 'success' }))
          .catch(() => toast?.showToast({ message: t('Error'), severity: 'error' })),
      onBuiltinMultiFile: onShowFiles,
    })
  }

  const confirmModal = (
    <Modal.Root state={confirmState}>
      <Modal.Backdrop>
        <Modal.Container size='sm'>
          <Modal.Dialog>
            <Modal.Header>
              <Modal.Heading>{confirmHeading}</Modal.Heading>
            </Modal.Header>
            <Modal.Body>{confirmBody}</Modal.Body>
            <Modal.Footer>
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
    return (
      <div className='space-y-2.5'>
        <div className='flex items-center gap-1.5'>
          <Button
            variant='primary'
            size='md'
            className='min-h-11 min-w-0 flex-1 gap-2'
            isPending={isResolving}
            onPress={onPlayPress}
          >
            {({ isPending }) => (
              <>
                {isPending ? (
                  <Spinner size='sm' color='current' />
                ) : (
                  <Play {...iconMenu} fill='currentColor' aria-hidden />
                )}
                <span className='truncate'>{playLabel}</span>
              </>
            )}
          </Button>
          <ExternalPlayersGroup players={externalPlayers} compact />
          <Dropdown>
            <Dropdown.Trigger>
              <Button variant='ghost' isIconOnly className={`${iconBtn} shrink-0 text-muted`} aria-label={t('Info')}>
                <MoreHorizontal {...iconMenu} aria-hidden />
              </Button>
            </Dropdown.Trigger>
            <Dropdown.Popover placement='bottom end' className='min-w-[14rem]'>
              <Dropdown.Menu aria-label={t('Info')}>
                {singleFileStream ? (
                  <Dropdown.Item onPress={() => void copyStreamLink()}>
                    <Link2 {...iconMenu} />
                    {t('CopyLink')}
                  </Dropdown.Item>
                ) : null}
                {isSingleFileTorrent || !viewedFileList?.length ? (
                  <Dropdown.Item onPress={() => window.open(fullPlaylistLink, '_blank')}>
                    <ListMusic {...iconMenu} />
                    {t('DownloadPlaylist')}
                  </Dropdown.Item>
                ) : null}
                <Dropdown.Item onPress={() => void copyMagnetLink()}>
                  <Magnet {...iconMenu} />
                  {t('CopyMagnet')}
                </Dropdown.Item>
                <Dropdown.Item onPress={() => void copyInfoHash()}>
                  <Hash {...iconMenu} />
                  {t('CopyHash')}
                </Dropdown.Item>
                <Dropdown.Item onPress={() => void copyTorrsLink()}>
                  <Share2 {...iconMenu} />
                  {t('CopyTorrs')}
                </Dropdown.Item>
                <Dropdown.Item onPress={() => window.open(playlistAllUrl({ category: undefined }), '_blank')}>
                  <ListMusic {...iconMenu} />
                  {t('DownloadAllPlaylists')}
                </Dropdown.Item>
                <Dropdown.Item onPress={() => setPendingConfirm('clearViews')}>
                  <EyeOff {...iconMenu} />
                  {t('RemoveViews')}
                </Dropdown.Item>
                <Dropdown.Item onPress={() => setPendingConfirm('drop')}>
                  <Trash2 {...iconMenu} />
                  {t('DropTorrent')}
                </Dropdown.Item>
                <Dropdown.Item onPress={() => setPendingConfirm('delete')}>
                  <Trash2 {...iconMenu} />
                  {t('Delete')}
                </Dropdown.Item>
              </Dropdown.Menu>
            </Dropdown.Popover>
          </Dropdown>
        </div>

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

        {playerModals}
        {confirmModal}
      </div>
    )
  }

  return (
    <div className='space-y-4'>
      <div className='flex flex-wrap items-center gap-2'>
        <Button
          variant='primary'
          size='lg'
          fullWidth
          className='sm:w-auto'
          isPending={isResolving}
          onPress={onPlayPress}
        >
          {({ isPending }) => (
            <>
              {isPending ? (
                <Spinner size='sm' color='current' />
              ) : (
                <Play {...iconMenu} fill='currentColor' aria-hidden />
              )}
              {playLabel}
            </>
          )}
        </Button>
        <ExternalPlayersGroup players={externalPlayers} />
        {singleFileStream ? (
          <Button variant='secondary' onPress={() => void copyStreamLink()}>
            <Link2 {...iconMenu} aria-hidden />
            {t('CopyLink')}
          </Button>
        ) : null}
      </div>

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
          <ButtonGroup>
            <Button variant='primary' onPress={() => window.open(fullPlaylistLink, '_blank')}>
              {t('Full')}
            </Button>
            <Button variant='primary' onPress={() => window.open(fromLatestPlaylistLink, '_blank')}>
              {t('FromLatestFile')}
            </Button>
          </ButtonGroup>
        </div>
      ) : null}

      <div>
        <p className='mb-2 text-sm font-semibold text-muted'>{t('Info')}</p>
        <div className='flex flex-wrap items-center gap-2'>
          {isSingleFileTorrent || !viewedFileList?.length ? (
            <Button variant='primary' onPress={() => window.open(fullPlaylistLink, '_blank')}>
              <ListMusic {...iconMenu} aria-hidden />
              {t('DownloadPlaylist')}
            </Button>
          ) : null}
          <ButtonGroup>
            <Button variant='secondary' onPress={() => void copyMagnetLink()}>
              <Magnet {...iconMenu} aria-hidden />
              {t('CopyMagnet')}
            </Button>
            <Button variant='secondary' onPress={() => void copyInfoHash()}>
              <Hash {...iconMenu} aria-hidden />
              {t('CopyHash')}
            </Button>
            <Button variant='secondary' onPress={() => void copyTorrsLink()}>
              <Share2 {...iconMenu} aria-hidden />
              {t('CopyTorrs')}
            </Button>
          </ButtonGroup>
          <Button variant='tertiary' onPress={() => window.open(playlistAllUrl({ category: undefined }), '_blank')}>
            <ListMusic {...iconMenu} aria-hidden />
            {t('DownloadAllPlaylists')}
          </Button>
        </div>
      </div>

      <Separator />

      <div>
        <p className='mb-2 text-sm font-semibold text-muted'>{t('TorrentState')}</p>
        <div className='flex flex-wrap gap-2'>
          <Button variant='outline' onPress={() => setPendingConfirm('clearViews')}>
            <EyeOff {...iconMenu} aria-hidden />
            {t('RemoveViews')}
          </Button>
          <Button variant='danger' onPress={() => setPendingConfirm('drop')}>
            <Trash2 {...iconMenu} aria-hidden />
            {t('DropTorrent')}
          </Button>
          <Button variant='danger' onPress={() => setPendingConfirm('delete')}>
            <Trash2 {...iconMenu} aria-hidden />
            {t('Delete')}
          </Button>
        </div>
      </div>

      {playerModals}
      {confirmModal}
    </div>
  )
}

export default memo(TorrentActions)

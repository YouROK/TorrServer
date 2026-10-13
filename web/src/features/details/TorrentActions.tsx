import { useMemo, memo, useState } from 'react'
import { Button, Dropdown, Modal, useMediaQuery, useOverlayState } from '@heroui/react'
import { buttonVariants } from '@heroui/styles'
import {
  ChevronDown,
  EyeOff,
  Hash,
  Link2,
  ListMusic,
  Magnet,
  RotateCcw,
  Settings,
  Share2,
  SquareArrowOutUpRight,
  Trash2,
} from 'lucide-react'
import ptt from 'parse-torrent-title'
import { useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import type { PlayableFile, TorrentStat } from 'shared/api/types'
import { torrsShareUrl } from 'shared/api/extras'
import { dropTorrent, markTorrentsDroppedInList, removeTorrent, TORRENTS_QUERY_KEY } from 'shared/api/torrents'
import { clearViewedFiles } from 'shared/api/viewed'
import { playlistActionHref, useExternalPlayers } from 'shared/lib/externalPlayers'
import { externalLink, useExternalMediaBase } from 'shared/lib/mediaBase'
import { copyToClipboard } from 'shared/lib/clipboard'
import { bindPlaylistButton, browserLinkAttrs } from 'shared/lib/openBrowser'
import {
  magnetFromHash,
  playlistButtonLabelKey,
  streamPlayUrl,
  torrentPlaylistUrl,
  usePlaylistButtonAction,
} from 'shared/lib/posterPlay'
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

/**
 * Files-tab action block: playlist / share / drop.
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
  const shareState = useOverlayState()
  const playersState = useOverlayState()
  const playlistState = useOverlayState()
  const confirmState = useOverlayState({
    isOpen: pendingConfirm != null,
    onOpenChange: open => {
      if (!open) setPendingConfirm(null)
    },
  })

  const fileCount = playableFileList?.length ?? 0
  /** Series controls need a known multi-file list. In DB / getting info (0 files) stays on the film row. */
  const isSeries = fileCount > 1
  const latestViewedFileId = viewedFileList?.[viewedFileList.length - 1]
  const latestViewedFilePath = playableFileList?.find(({ id }) => id === latestViewedFileId)?.path
  const latestViewedFileInfo = latestViewedFilePath ? ptt.parse(latestViewedFilePath) : null

  const displayName = name || title || 'file'
  const fullPlaylistLink = torrentPlaylistUrl(hash, displayName)
  const fromLatestPlaylistLink = `${fullPlaylistLink}&fromlast`
  const magnetLink = magnetFromHash(hash, displayName)

  /** Keep launcher mounted for Continue Watching autoplay + player modals. */
  const { playerModals } = usePlayLauncher({
    hash,
    displayName,
    knownPlayableFiles: playableFileList || [],
    onViewedChange,
    autoPlayFileId,
    autoPlayTimecode,
  })

  /** File row owns single-file players. Series footer opens the whole playlist. */
  const { buildSeriesExternalPlayers, hasAnyExternalPlayer, playerFlags } = useExternalPlayers()
  const playlistButtonAction = usePlaylistButtonAction(playerFlags)
  const mediaBase = useExternalMediaBase()
  const singleFileStream = useMemo(() => {
    if (playableFileList?.length !== 1) return null
    return { fullLink: externalLink(streamPlayUrl(hash, playableFileList[0]), mediaBase).toString() }
  }, [playableFileList, hash, mediaBase])
  const seriesPlayers = useMemo(() => {
    if (!playableFileList || playableFileList.length < 2) return []
    const playlist = externalLink(torrentPlaylistUrl(hash, displayName), mediaBase).toString()
    const episodes = playableFileList.map(file => ({
      url: externalLink(streamPlayUrl(hash, file), mediaBase).toString(),
      filename: file.path.split('\\').pop()?.split('/').pop() || file.path,
    }))
    return buildSeriesExternalPlayers(playlist, episodes)
  }, [playableFileList, hash, displayName, mediaBase, buildSeriesExternalPlayers])
  const showPlayerHint = (playableFileList?.length ?? 0) > 0 && !hasAnyExternalPlayer

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

  const hasPartialProgress = isSeries && !!viewedFileList?.length
  const latestPlayed = hasPartialProgress ? (
    <p className={`truncate text-xs text-muted${compact ? ' mb-1' : ''}`}>
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
  ) : null

  const confirmModal = (
    <Modal.Root state={confirmState}>
      <Modal.Backdrop>
        <Modal.Container size='sm' placement='center'>
          <Modal.Dialog className='ts-compact-modal'>
            <Modal.Header className='shrink-0'>
              <Modal.Heading>{confirmHeading}</Modal.Heading>
            </Modal.Header>
            <Modal.Body>{confirmBody}</Modal.Body>
            <Modal.Footer className='flex w-full shrink-0 justify-center gap-2'>
              <Button variant='secondary' className='min-h-11 flex-1' onPress={() => setPendingConfirm(null)} autoFocus>
                {t('Cancel')}
              </Button>
              <Button
                variant={pendingConfirm === 'drop' || pendingConfirm === 'delete' ? 'danger' : 'primary'}
                className='min-h-11 flex-1'
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

  const iconState = 'min-h-9 min-w-9 shrink-0'
  /** Icon tools stay fixed. Playlist grows on phone; every desktop button shares the row. */
  const slot = compact ? 'shrink-0' : 'flex min-w-0 flex-1 [&>*]:w-full [&>*]:min-w-0'
  const fillBtn = 'min-h-9 min-w-0 flex-1'

  const playlistLabel =
    playlistButtonAction === 'download'
      ? compact
        ? t('PlaylistShort')
        : t('DownloadPlaylist')
      : t(playlistButtonLabelKey(playlistButtonAction))
  const playlistClick = (url: string) => {
    const absolute = externalLink(url, mediaBase).toString()
    return bindPlaylistButton(playlistButtonAction, absolute, playlistActionHref(playlistButtonAction, absolute), kind =>
      toast?.showToast({
        message: t(kind === 'copied' ? 'Copied' : 'Error'),
        severity: kind === 'copied' ? 'success' : 'error',
      }),
    )
  }
  const playlistAnchor = (href: string, variant: 'primary' | 'outline', label: string, text: string) => (
    <a
      href={externalLink(href, mediaBase).toString()}
      {...browserLinkAttrs}
      onClick={playlistClick(href)}
      title={label}
      aria-label={label}
      className={`${buttonVariants({ variant })} ${fillBtn}`}
    >
      <ListMusic {...iconMenu} aria-hidden />
      <span className='truncate'>{text}</span>
    </a>
  )

  const playlistButtons =
    compact && isSeries && hasPartialProgress ? (
      <Dropdown isOpen={playlistState.isOpen} onOpenChange={playlistState.setOpen}>
        <Dropdown.Trigger className='min-w-0 flex-1'>
          <Button variant='primary' className='min-h-9 w-full min-w-0' aria-label={playlistLabel}>
            <ListMusic {...iconMenu} aria-hidden />
            <span className='truncate'>{playlistLabel}</span>
            <ChevronDown {...iconMenu} aria-hidden />
          </Button>
        </Dropdown.Trigger>
        <Dropdown.Popover placement='top start'>
          <Dropdown.Menu aria-label={playlistLabel}>
            <Dropdown.Item
              href={externalLink(fromLatestPlaylistLink, mediaBase).toString()}
              {...browserLinkAttrs}
              onClick={playlistClick(fromLatestPlaylistLink)}
              textValue={t('FromLatestFile')}
            >
              <ListMusic {...iconMenu} />
              {t('FromLatestFile')}
            </Dropdown.Item>
            <Dropdown.Item
              href={externalLink(fullPlaylistLink, mediaBase).toString()}
              {...browserLinkAttrs}
              onClick={playlistClick(fullPlaylistLink)}
              textValue={playlistLabel}
            >
              <ListMusic {...iconMenu} />
              {playlistLabel}
            </Dropdown.Item>
          </Dropdown.Menu>
        </Dropdown.Popover>
      </Dropdown>
    ) : !isSeries || !hasPartialProgress ? (
      playlistAnchor(fullPlaylistLink, 'primary', playlistLabel, playlistLabel)
    ) : (
      <>
        {playlistAnchor(fromLatestPlaylistLink, 'primary', t('FromLatestFile'), t('FromLatestFile'))}
        {playlistAnchor(fullPlaylistLink, 'outline', playlistLabel, playlistLabel)}
      </>
    )

  const playersMenu =
    seriesPlayers.length > 0 ? (
      <div className={slot}>
        <Dropdown isOpen={playersState.isOpen} onOpenChange={playersState.setOpen}>
          <Dropdown.Trigger className={compact ? undefined : 'w-full'}>
            <Button
              variant='secondary'
              isIconOnly={compact}
              className={compact ? iconState : 'min-h-9 w-full min-w-0'}
              aria-label={t('OpenIn')}
            >
              <SquareArrowOutUpRight {...iconMenu} aria-hidden />
              {compact ? null : <span className='truncate'>{t('OpenIn')}</span>}
            </Button>
          </Dropdown.Trigger>
          <Dropdown.Popover placement='top end'>
            <Dropdown.Menu aria-label={t('OpenIn')}>
              {seriesPlayers.map(player => (
                <Dropdown.Item
                  key={player.label}
                  onPress={() => {
                    window.location.href = player.href
                  }}
                >
                  {player.label}
                </Dropdown.Item>
              ))}
            </Dropdown.Menu>
          </Dropdown.Popover>
        </Dropdown>
      </div>
    ) : null

  const shareMenu = (
    <div className={slot}>
      <Dropdown isOpen={shareState.isOpen} onOpenChange={shareState.setOpen}>
        <Dropdown.Trigger className={compact ? undefined : 'w-full'}>
          <Button
            variant='secondary'
            isIconOnly={compact}
            className={compact ? iconState : 'min-h-9 w-full min-w-0 gap-1.5'}
            aria-label={t('Share')}
          >
            <Share2 {...iconMenu} aria-hidden />
            {compact ? null : <span className='truncate'>{t('Share')}</span>}
          </Button>
        </Dropdown.Trigger>
        <Dropdown.Popover placement='top end'>
          <Dropdown.Menu aria-label={t('Share')}>
            {singleFileStream ? (
              <Dropdown.Item onPress={() => void copyStreamLink()}>
                <Link2 {...iconMenu} />
                {t('CopyLink')}
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
          </Dropdown.Menu>
        </Dropdown.Popover>
      </Dropdown>
    </div>
  )

  const clearButton = (
    <span title={t('RemoveViews')} className={slot}>
      <Button
        variant='outline'
        isIconOnly={compact}
        className={compact ? iconState : 'min-h-9 w-full min-w-0'}
        aria-label={t('RemoveViews')}
        onPress={() => setPendingConfirm('clearViews')}
      >
        <EyeOff {...iconMenu} aria-hidden />
        {compact ? null : <span className='truncate'>{t('RemoveViews')}</span>}
      </Button>
    </span>
  )

  const dropButton = (
    <span title={t('DropTorrent')} className={slot}>
      <Button
        variant='outline'
        isIconOnly={compact}
        className={compact ? iconState : 'min-h-9 w-full min-w-0'}
        aria-label={t('DropTorrent')}
        onPress={() => setPendingConfirm('drop')}
      >
        <RotateCcw {...iconMenu} aria-hidden />
        {compact ? null : <span className='truncate'>{t('DropTorrent')}</span>}
      </Button>
    </span>
  )

  const deleteButton = (
    <Button
      variant='danger'
      isIconOnly={compact}
      className={compact ? iconState : fillBtn}
      aria-label={t('Delete')}
      onPress={() => setPendingConfirm('delete')}
    >
      <Trash2 {...iconMenu} aria-hidden />
    </Button>
  )

  return (
    <>
      <div className={compact ? undefined : 'space-y-2'}>
        {showPlayerHint ? (
          <button
            type='button'
            onClick={() => requestOpenSettings('app')}
            className={`flex items-center gap-1.5 text-xs text-muted transition-colors hover:text-accent${compact ? ' mb-1' : ''}`}
          >
            <Settings size={14} strokeWidth={1.75} aria-hidden />
            {t('ExternalPlayersHint')}
          </button>
        ) : null}

        {latestPlayed}

        <div className='flex w-full items-stretch gap-2'>
          {playlistButtons}
          {playersMenu}
          {shareMenu}
          {clearButton}
          {dropButton}
          {deleteButton}
        </div>
      </div>

      {playerModals}
      {confirmModal}
    </>
  )
}

export default memo(TorrentActions)

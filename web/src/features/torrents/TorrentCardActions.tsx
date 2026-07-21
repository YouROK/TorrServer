import { useMemo, useState } from 'react'
import { Button, Dropdown, Modal, Spinner, Tooltip, useOverlayState } from '@heroui/react'
import {
  Hash,
  Info,
  Link2,
  ListMusic,
  Magnet,
  MoreVertical,
  Pencil,
  Play,
  Share2,
  SquareArrowOutUpRight,
  Trash2,
  X,
} from 'lucide-react'
import { useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import type { TorrentStat } from 'shared/api/types'
import { torrsShareUrl } from 'shared/api/extras'
import { dropTorrent, removeTorrent, TORRENTS_QUERY_KEY } from 'shared/api/torrents'
import { copyToClipboard } from 'shared/lib/clipboard'
import { useExternalPlayers } from 'shared/lib/externalPlayers'
import { magnetFromHash, streamPlayUrl, torrentPlaylistUrl } from 'shared/lib/posterPlay'
import { filesFromMetadata } from 'shared/torrent/fileMetadata'
import { isFilePlayable } from 'shared/torrent/playable'
import { iconAction, iconMenu } from 'shared/ui/iconProps'
import { useSyncModalOpen } from 'shared/ui/ModalOpenContext'
import { useOptionalAppToast } from 'shared/ui/Toast'
import { toPlayableFile } from 'shared/torrent/toPlayableFile'
import { useConfiguredPlayAction } from 'features/player/useConfiguredPlayAction'
import { usePlayLauncher } from 'features/player/usePlayLauncher'

export interface TorrentCardActionsProps {
  torrent: TorrentStat
  onDetails: () => void
  onEdit?: () => void
}

type ConfirmKind = 'drop' | 'delete' | null

/**
 * Poster overlay actions: Play + overflow menu (details / playlist / edit / externals / drop / delete).
 * Secondary actions stay in the menu so 44px targets fit a narrow poster tile.
 */
export default function TorrentCardActions({ torrent, onDetails, onEdit }: TorrentCardActionsProps) {
  const { t } = useTranslation()
  const toast = useOptionalAppToast()
  const queryClient = useQueryClient()

  const [confirmKind, setConfirmKind] = useState<ConfirmKind>(null)
  const confirmState = useOverlayState({
    isOpen: confirmKind != null,
    onOpenChange: open => {
      if (!open) setConfirmKind(null)
    },
  })
  const dropdownState = useOverlayState()

  const hash = torrent.hash
  const displayName = torrent.title || torrent.name || hash
  const playlistHref = torrentPlaylistUrl(hash, displayName)
  const magnetHref = magnetFromHash(hash, displayName)

  const knownPlayableFiles = useMemo(() => {
    const stats = torrent.file_stats
    const files = stats?.length ? stats.map(toPlayableFile) : filesFromMetadata(torrent.data)
    return files.filter(file => isFilePlayable(file.path))
  }, [torrent.file_stats, torrent.data])

  const { handlePlay, resolvePlayableFile, isResolving, playerModals } = usePlayLauncher({
    hash,
    displayName,
    knownPlayableFiles,
    torrentData: torrent.data,
  })

  const { runConfiguredPlay } = useConfiguredPlayAction()
  const { buildExternalPlayers } = useExternalPlayers()

  useSyncModalOpen(confirmState.isOpen || dropdownState.isOpen)

  /** Only offer app deep links when there's exactly one obvious file to hand off — otherwise Play's file picker covers it. */
  const externalPlayers = useMemo(() => {
    if (knownPlayableFiles.length !== 1) return []
    const file = knownPlayableFiles[0]
    const link = streamPlayUrl(hash, file)
    const fullLink = new URL(link, window.location.href).toString()
    return buildExternalPlayers(fullLink)
  }, [knownPlayableFiles, hash, buildExternalPlayers])

  const singleFileStreamHref = useMemo(() => {
    if (knownPlayableFiles.length !== 1) return null
    return new URL(streamPlayUrl(hash, knownPlayableFiles[0]), window.location.href).toString()
  }, [knownPlayableFiles, hash])

  const copyText = async (text: string) => {
    try {
      await copyToClipboard(text)
      toast?.showToast({ message: t('Copied'), severity: 'success' })
    } catch {
      toast?.showToast({ message: t('Error'), severity: 'error' })
    }
  }

  const handlePosterPlay = () => {
    runConfiguredPlay({
      hash,
      displayName,
      knownPlayableFiles,
      handlePlay,
      resolvePlayableFile,
      copyText,
    })
  }

  const runConfirmedAction = () => {
    const action = confirmKind
    setConfirmKind(null)
    confirmState.close()

    const mutate = action === 'drop' ? dropTorrent : action === 'delete' ? removeTorrent : null
    const successMessage = action === 'drop' ? t('DropTorrent') : t('Delete')
    if (!mutate) return

    const previous = queryClient.getQueryData<TorrentStat[]>(TORRENTS_QUERY_KEY)
    queryClient.setQueryData<TorrentStat[]>(TORRENTS_QUERY_KEY, prev => prev?.filter(item => item.hash !== hash))

    void mutate(hash)
      .then(async () => {
        toast?.showToast({ message: successMessage, severity: 'success' })
        await queryClient.invalidateQueries({ queryKey: TORRENTS_QUERY_KEY })
      })
      .catch(() => {
        if (previous) queryClient.setQueryData(TORRENTS_QUERY_KEY, previous)
        toast?.showToast({ message: t('Error'), severity: 'error' })
      })
  }

  const overlayButtonClass =
    'inline-flex h-11 w-11 min-h-11 min-w-11 items-center justify-center rounded-full bg-black/55 p-0 text-white backdrop-blur-sm hover-fine:bg-accent [&_svg]:m-0 [&_svg]:block'

  return (
    <>
      <div className='flex items-center justify-center gap-2' onClick={event => event.stopPropagation()}>
        <Tooltip>
          <Tooltip.Trigger>
            <Button
              variant='primary'
              isIconOnly
              className={overlayButtonClass}
              aria-label={t('Play')}
              isPending={isResolving}
              onPress={handlePosterPlay}
            >
              {({ isPending }) =>
                isPending ? (
                  <Spinner size='sm' color='current' />
                ) : (
                  <Play {...iconAction} fill='currentColor' aria-hidden />
                )
              }
            </Button>
          </Tooltip.Trigger>
          <Tooltip.Content>{t('Play')}</Tooltip.Content>
        </Tooltip>

        {/* Details/Playlist live in the menu — the poster itself already opens details, and cramming 4 icon-only
         * buttons into a ~172px tile leaves no room to reach 40px touch targets. */}
        <Dropdown isOpen={dropdownState.isOpen} onOpenChange={dropdownState.setOpen}>
          <Dropdown.Trigger>
            <Button variant='primary' isIconOnly className={overlayButtonClass} aria-label={t('Actions')}>
              <MoreVertical {...iconAction} aria-hidden />
            </Button>
          </Dropdown.Trigger>
          <Dropdown.Popover placement='bottom end'>
            <Dropdown.Menu aria-label={t('Actions')}>
              <Dropdown.Item onPress={onDetails}>
                <Info {...iconMenu} />
                {t('Details')}
              </Dropdown.Item>
              <Dropdown.Item onPress={() => window.open(playlistHref, '_blank')}>
                <ListMusic {...iconMenu} />
                {t('DownloadPlaylist')}
              </Dropdown.Item>
              {onEdit ? (
                <Dropdown.Item onPress={onEdit}>
                  <Pencil {...iconMenu} />
                  {t('EditTorrent')}
                </Dropdown.Item>
              ) : null}
              {singleFileStreamHref ? (
                <Dropdown.Item onPress={() => void copyText(singleFileStreamHref)}>
                  <Link2 {...iconMenu} />
                  {t('CopyLink')}
                </Dropdown.Item>
              ) : null}
              <Dropdown.Item onPress={() => void copyText(magnetHref)}>
                <Magnet {...iconMenu} />
                {t('CopyMagnet')}
              </Dropdown.Item>
              <Dropdown.Item onPress={() => void copyText(hash)}>
                <Hash {...iconMenu} />
                {t('CopyHash')}
              </Dropdown.Item>
              <Dropdown.Item
                onPress={() => {
                  void copyText(torrsShareUrl(torrent))
                }}
              >
                <Share2 {...iconMenu} />
                {t('CopyTorrs')}
              </Dropdown.Item>
              {externalPlayers.map(player => (
                <Dropdown.Item
                  key={player.label}
                  onPress={() => {
                    window.location.href = player.href
                  }}
                >
                  <SquareArrowOutUpRight {...iconMenu} />
                  {player.label}
                </Dropdown.Item>
              ))}
              <Dropdown.Item
                onPress={() => {
                  setConfirmKind('drop')
                  confirmState.open()
                }}
              >
                <X {...iconMenu} />
                {t('DropTorrent')}
              </Dropdown.Item>
              <Dropdown.Item
                variant='danger'
                onPress={() => {
                  setConfirmKind('delete')
                  confirmState.open()
                }}
              >
                <Trash2 {...iconMenu} />
                {t('Delete')}
              </Dropdown.Item>
            </Dropdown.Menu>
          </Dropdown.Popover>
        </Dropdown>
      </div>

      {playerModals}

      <Modal state={confirmState}>
        <Modal.Backdrop isDismissable>
          <Modal.Container size='sm'>
            <Modal.Dialog aria-label={confirmKind === 'delete' ? t('DeleteTorrent?') : t('DropTorrent')}>
              <Modal.Header>
                <Modal.Icon className='bg-danger/15 text-danger'>
                  <Trash2 {...iconAction} aria-hidden />
                </Modal.Icon>
                <Modal.Heading>{confirmKind === 'delete' ? t('DeleteTorrent?') : t('DropTorrent')}</Modal.Heading>
              </Modal.Header>
              <Modal.Body>{confirmKind === 'delete' ? t('ConfirmDeleteTorrent') : t('ConfirmDropTorrent')}</Modal.Body>
              <Modal.Footer className='flex justify-end gap-2'>
                <Button autoFocus variant='secondary' onPress={() => setConfirmKind(null)}>
                  {t('Cancel')}
                </Button>
                <Button variant='danger' onPress={runConfirmedAction}>
                  {t('OK')}
                </Button>
              </Modal.Footer>
            </Modal.Dialog>
          </Modal.Container>
        </Modal.Backdrop>
      </Modal>
    </>
  )
}

import { memo, useState } from 'react'
import { ListItemIcon, ListItemText, Menu, MenuItem } from '@material-ui/core'
import CloudDoneIcon from '@material-ui/icons/CloudDone'
import CloudDownloadIcon from '@material-ui/icons/CloudDownload'
import PauseCircleOutlineIcon from '@material-ui/icons/PauseCircleOutline'
import PauseIcon from '@material-ui/icons/Pause'
import PlayArrowIcon from '@material-ui/icons/PlayArrow'
import DeleteOutlineIcon from '@material-ui/icons/DeleteOutline'
import LockIcon from '@material-ui/icons/Lock'
import LockOpenIcon from '@material-ui/icons/LockOpen'
import { useTranslation } from 'react-i18next'
import { OFFLINE_DOWNLOADING, OFFLINE_PAUSED, OFFLINE_READY } from 'utils/offline'

import { StyledButton } from './style'
import { useOfflineActions } from './useOfflineActions'

// Offline controls of a torrent card: ONE button of the TorrentCardButtons grid ("Offline") that
// opens a menu (a separate button per action does not fit into the card):
//   not offline -> save for offline viewing
//   downloading -> pause, cancel
//   paused      -> resume, cancel
//   ready       -> remove the offline copy
// Any offline state also offers pin/unpin: a pinned copy is skipped by the disk quota rotation.
const OfflineControls = ({ hash, offlineState, pinned }) => {
  const { t } = useTranslation()
  const { busy, enable, pause, resume, pin, remove, dialogs } = useOfflineActions(hash)
  const [anchor, setAnchor] = useState(null)

  const closeMenu = () => setAnchor(null)
  const item = (label, onClick, icon) => (
    <MenuItem
      key={label}
      onClick={() => {
        closeMenu()
        onClick()
      }}
    >
      <ListItemIcon style={{ minWidth: 34 }}>{icon}</ListItemIcon>
      <ListItemText primary={label} />
    </MenuItem>
  )

  let icon = <CloudDownloadIcon />
  if (offlineState === OFFLINE_PAUSED) icon = <PauseCircleOutlineIcon />
  if (offlineState === OFFLINE_READY) icon = <CloudDoneIcon />

  return (
    <>
      <StyledButton
        disabled={busy}
        title={t('OfflineMode', 'Offline mode')}
        onClick={e => {
          // the card may be clickable as a whole - the button must not trigger it
          e.stopPropagation()
          setAnchor(e.currentTarget)
        }}
      >
        {icon}
        <span>{t('OfflineSummaryOffline', 'Offline')}</span>
        {pinned && <LockIcon style={{ fontSize: 14, marginLeft: 4 }} titleAccess={t('OfflinePinned', 'Pinned')} />}
      </StyledButton>

      <Menu
        anchorEl={anchor}
        open={!!anchor}
        onClose={closeMenu}
        onClick={e => e.stopPropagation()}
        getContentAnchorEl={null}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
        transformOrigin={{ vertical: 'top', horizontal: 'left' }}
      >
        {!offlineState &&
          item(t('OfflineEnable', 'Download for offline viewing'), enable, <CloudDownloadIcon fontSize='small' />)}
        {offlineState === OFFLINE_DOWNLOADING &&
          item(t('OfflinePause', 'Pause offline download'), pause, <PauseIcon fontSize='small' />)}
        {offlineState === OFFLINE_PAUSED &&
          item(t('OfflineResume', 'Resume offline download'), resume, <PlayArrowIcon fontSize='small' />)}
        {!!offlineState &&
          (pinned
            ? item(t('OfflineUnpin', 'Unpin: allow rotation to remove'), () => pin(false), <LockOpenIcon fontSize='small' />)
            : item(t('OfflinePin', 'Pin: never remove automatically'), () => pin(true), <LockIcon fontSize='small' />))}
        {!!offlineState &&
          item(
            offlineState === OFFLINE_READY
              ? t('OfflineRemove', 'Remove offline copy')
              : t('OfflineCancel', 'Cancel offline download'),
            remove,
            <DeleteOutlineIcon fontSize='small' />,
          )}
      </Menu>
      {dialogs}
    </>
  )
}

export default memo(OfflineControls)

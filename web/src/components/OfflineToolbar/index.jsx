import { memo, useState } from 'react'
import {
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Paper,
} from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import PauseIcon from '@material-ui/icons/Pause'
import PlayArrowIcon from '@material-ui/icons/PlayArrow'
import ArrowDownwardIcon from '@material-ui/icons/ArrowDownward'
import ArrowUpwardIcon from '@material-ui/icons/ArrowUpward'
import ExpandLessIcon from '@material-ui/icons/ExpandLess'
import ExpandMoreIcon from '@material-ui/icons/ExpandMore'
import { useTranslation } from 'react-i18next'
import { useQueryClient } from 'react-query'
import { humanizeSize, humanizeSpeed } from 'utils/Utils'
import {
  FILTER_ALL,
  FILTER_DOWNLOADING,
  FILTER_OFFLINE,
  FILTER_ONLINE,
  FILTER_PAUSED,
  FILTER_READY,
  getFilterCount,
} from 'utils/offline'

import { pauseAllOffline, resumeAllOffline } from './bulkActions'
import { OFFLINE_QUOTA_KEY, useOfflineQuota } from './useOfflineQuota'

const useStyles = makeStyles(theme => ({
  root: {
    display: 'flex',
    flexWrap: 'wrap',
    alignItems: 'center',
    gap: theme.spacing(1.5, 3),
    padding: theme.spacing(1.5, 2),
  },
  group: { display: 'flex', flexWrap: 'wrap', alignItems: 'center', gap: theme.spacing(1, 2.5) },
  stat: { display: 'flex', alignItems: 'center', gap: theme.spacing(0.5), whiteSpace: 'nowrap' },
  statLabel: { opacity: 0.7 },
  statValue: { fontWeight: 600 },
  speedIcon: { fontSize: 18 },
  chips: { display: 'flex', flexWrap: 'wrap', gap: theme.spacing(1), flexBasis: '100%' },
  actions: { display: 'flex', flexWrap: 'wrap', alignItems: 'center', gap: theme.spacing(1), marginLeft: 'auto' },
}))

const FILTERS = [
  [FILTER_ALL, 'OfflineFilterAll', 'All'],
  [FILTER_ONLINE, 'OfflineFilterOnline', 'Online'],
  [FILTER_OFFLINE, 'OfflineFilterOffline', 'Offline'],
  [FILTER_DOWNLOADING, 'OfflineFilterDownloading', 'Downloading'],
  [FILTER_PAUSED, 'OfflineFilterPaused', 'Paused'],
  [FILTER_READY, 'OfflineFilterReady', 'Ready'],
]

// The panel can be collapsed; the choice is remembered. Storage may be unavailable (blocked by the
// browser, private mode) - then the choice just lives until the page is reloaded.
const COLLAPSED_KEY = 'offlineToolbarCollapsed'
const readCollapsed = () => {
  try {
    return window.localStorage.getItem(COLLAPSED_KEY) === '1'
  } catch {
    return false
  }
}
const saveCollapsed = value => {
  try {
    window.localStorage.setItem(COLLAPSED_KEY, value ? '1' : '0')
  } catch {
    // ignore: not persisted
  }
}

// Summary + bulk controls + status filter. The torrent list renders it with the full (unfiltered)
// list; it returns null by itself when there is no offline torrent, so it can never show up for a
// purely online list.
const OfflineToolbar = ({ torrents, summary, filter, onFilterChange }) => {
  const classes = useStyles()
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [busy, setBusy] = useState(false)
  const [errors, setErrors] = useState([])
  const [collapsed, setCollapsed] = useState(readCollapsed)
  // hooks go before the early return below; nothing is requested while the panel is hidden
  const quota = useOfflineQuota(!!summary && summary.offline > 0)

  if (!summary || summary.offline === 0) return null

  const toggleCollapsed = () => {
    saveCollapsed(!collapsed)
    setCollapsed(!collapsed)
  }

  const runBulk = async action => {
    setBusy(true)
    try {
      // The list is polled once a second: without an immediate refetch the counters (and the
      // enabled state of these buttons) stay stale for a moment after the action. The buttons
      // remain disabled until the fresh list has arrived.
      const getFresh = () => queryClient.refetchQueries('torrents').then(() => queryClient.getQueryData('torrents'))
      const failed = await action(torrents, getFresh)
      await getFresh()
      queryClient.invalidateQueries(OFFLINE_QUOTA_KEY)
      if (failed.length) setErrors(failed)
    } finally {
      setBusy(false)
    }
  }

  const activeFilter = FILTERS.find(([value]) => value === filter)

  const size = bytes => humanizeSize(bytes) || '0'
  // Used by the offline torrents (of the quota, if one is set) and what a new download may still
  // take: the smaller of what is left of the quota and of the free disk space (after the reserve).
  const usedText = quota ? (quota.max_size > 0 ? `${size(quota.used)} / ${size(quota.max_size)}` : size(quota.used)) : '---'
  const availableText = quota && quota.available >= 0 ? size(quota.available) : '---'
  const limitHint = quota?.limited_by
    ? t(quota.limited_by === 'quota' ? 'OfflineAvailableByQuota' : 'OfflineAvailableByDisk', {
        defaultValue: quota.limited_by === 'quota' ? 'Limited by the offline size quota' : 'Limited by the free disk space',
      })
    : t('OfflineAvailableNoLimit', 'No limit is known')

  const errorDialog = (
    <Dialog open={errors.length > 0} onClose={() => setErrors([])}>
      <DialogTitle>{t('OfflineBulkErrors', 'Some torrents could not be processed')}</DialogTitle>
      <DialogContent>
        {errors.slice(0, 5).map(line => (
          <div key={line}>{line}</div>
        ))}
        {errors.length > 5 && <div>… +{errors.length - 5}</div>}
      </DialogContent>
      <DialogActions>
        <Button variant='contained' color='primary' onClick={() => setErrors([])}>
          {t('OK')}
        </Button>
      </DialogActions>
    </Dialog>
  )

  return (
    <>
      <Paper className={classes.root} variant='outlined'>
        <div className={classes.group}>
          <div className={classes.stat}>
            <span className={classes.statLabel}>{t('OfflineSummaryOnline', 'Online')}:</span>
            <span className={classes.statValue}>{summary.online}</span>
          </div>
          <div className={classes.stat}>
            <span className={classes.statLabel}>{t('OfflineSummaryOffline', 'Offline')}:</span>
            <span className={classes.statValue}>{summary.offline}</span>
          </div>
          <div className={classes.stat} title={t('OfflineStorageUsedHint', 'Space taken by the offline torrents')}>
            <span className={classes.statLabel}>{t('OfflineStorageUsed', 'Used')}:</span>
            <span className={classes.statValue}>{usedText}</span>
          </div>
          <div
            className={classes.stat}
            title={`${t('OfflineStorageAvailableHint', 'Space left for new offline downloads')} - ${limitHint}`}
          >
            <span className={classes.statLabel}>{t('OfflineStorageAvailable', 'Available')}:</span>
            <span className={classes.statValue}>{availableText}</span>
          </div>
          <div className={classes.stat} title={t('OfflineSummaryDownloadSpeed', 'Total download speed')}>
            <ArrowDownwardIcon className={classes.speedIcon} />
            <span className={classes.statValue}>
              {summary.downloadSpeed > 0 ? humanizeSpeed(summary.downloadSpeed) : '---'}
            </span>
          </div>
          <div className={classes.stat} title={t('OfflineSummaryUploadSpeed', 'Total upload speed')}>
            <ArrowUpwardIcon className={classes.speedIcon} />
            <span className={classes.statValue}>
              {summary.uploadSpeed > 0 ? humanizeSpeed(summary.uploadSpeed) : '---'}
            </span>
          </div>
          {/* A filter that is still active must stay visible (and removable) when the panel is
              collapsed - otherwise the list looks incomplete for no apparent reason. */}
          {collapsed && filter !== FILTER_ALL && activeFilter && (
            <Chip
              size='small'
              color='primary'
              label={t(activeFilter[1], activeFilter[2])}
              onDelete={() => onFilterChange(FILTER_ALL)}
            />
          )}
        </div>

        <div className={classes.actions}>
          {!collapsed && (
            <>
              <Button
                size='small'
                variant='contained'
                color='primary'
                startIcon={<PlayArrowIcon />}
                disabled={busy || summary.paused === 0}
                onClick={() => runBulk(resumeAllOffline)}
              >
                {t('OfflineResumeAll', 'Resume all offline')}
              </Button>
              <Button
                size='small'
                variant='contained'
                color='primary'
                startIcon={<PauseIcon />}
                disabled={busy || summary.downloading === 0}
                onClick={() => runBulk(pauseAllOffline)}
              >
                {t('OfflinePauseAll', 'Pause all offline')}
              </Button>
            </>
          )}
          <IconButton
            size='small'
            title={collapsed ? t('OfflineToolbarShow', 'Show panel') : t('OfflineToolbarHide', 'Hide panel')}
            onClick={toggleCollapsed}
          >
            {collapsed ? <ExpandMoreIcon /> : <ExpandLessIcon />}
          </IconButton>
        </div>

        {!collapsed && (
          <div className={classes.chips}>
            {FILTERS.map(([value, key, fallback]) => {
              const count = getFilterCount(summary, value)
              const active = filter === value
              // empty groups are hidden, except "all" and the currently selected one
              if (count === 0 && !active && value !== FILTER_ALL) return null
              return (
                <Chip
                  key={value}
                  clickable
                  size='small'
                  color={active ? 'primary' : 'default'}
                  variant={active ? 'default' : 'outlined'}
                  label={`${t(key, fallback)} · ${count}`}
                  onClick={() => onFilterChange(value)}
                />
              )
            })}
          </div>
        )}
      </Paper>
      {errorDialog}
    </>
  )
}

export default memo(OfflineToolbar)

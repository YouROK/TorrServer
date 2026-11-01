// Pure helpers for the offline mode UI (no React, no i18n - easy to unit-test).
// Field names come from the torrent status returned by POST /torrents {action:'list'}:
// offline_state ('' | 'downloading' | 'paused' | 'ready'), loaded_size, torrent_size,
// download_speed, upload_speed.

export const OFFLINE_DOWNLOADING = 'downloading'
export const OFFLINE_PAUSED = 'paused'
export const OFFLINE_READY = 'ready'

export const FILTER_ALL = 'all'
export const FILTER_ONLINE = 'online' // torrents that are not in the offline mode
export const FILTER_OFFLINE = 'offline' // any offline state
export const FILTER_DOWNLOADING = OFFLINE_DOWNLOADING
export const FILTER_PAUSED = OFFLINE_PAUSED
export const FILTER_READY = OFFLINE_READY

export const OFFLINE_COLORS = {
  [OFFLINE_DOWNLOADING]: '#03A9F4',
  [OFFLINE_PAUSED]: '#FFA000',
  [OFFLINE_READY]: '#4CAF50',
}

export const isOfflineTorrent = torrent => !!torrent?.offline_state

// Bytes saved on disk. A ready torrent is complete even if the server did not send loaded_size.
export const getOfflineLoadedSize = torrent => {
  const total = torrent?.torrent_size || 0
  const loaded = torrent?.offline_state === OFFLINE_READY && !torrent?.loaded_size ? total : torrent?.loaded_size || 0
  return total > 0 ? Math.min(loaded, total) : loaded
}

// 0..100 (fractional - for the progress bar width)
export const getOfflineFraction = torrent => {
  if (torrent?.offline_state === OFFLINE_READY) return 100
  const total = torrent?.torrent_size || 0
  if (!(total > 0)) return 0
  return Math.max(0, Math.min(100, (getOfflineLoadedSize(torrent) * 100) / total))
}

export const getTorrentsSummary = torrents => {
  const summary = {
    total: 0,
    online: 0,
    offline: 0,
    downloading: 0,
    paused: 0,
    ready: 0,
    downloadSpeed: 0,
    uploadSpeed: 0,
  }
  for (const torrent of torrents || []) {
    summary.total += 1
    switch (torrent.offline_state) {
      case OFFLINE_DOWNLOADING:
        summary.downloading += 1
        summary.offline += 1
        break
      case OFFLINE_PAUSED:
        summary.paused += 1
        summary.offline += 1
        break
      case OFFLINE_READY:
        summary.ready += 1
        summary.offline += 1
        break
      default:
        summary.online += 1
    }
    summary.downloadSpeed += Number(torrent.download_speed) || 0
    summary.uploadSpeed += Number(torrent.upload_speed) || 0
  }
  return summary
}

export const getFilterCount = (summary, filter) => {
  switch (filter) {
    case FILTER_ONLINE:
      return summary.online
    case FILTER_OFFLINE:
      return summary.offline
    case FILTER_DOWNLOADING:
      return summary.downloading
    case FILTER_PAUSED:
      return summary.paused
    case FILTER_READY:
      return summary.ready
    default:
      return summary.total
  }
}

export const matchesFilter = (torrent, filter) => {
  switch (filter) {
    case FILTER_ONLINE:
      return !torrent.offline_state
    case FILTER_OFFLINE:
      return !!torrent.offline_state
    case FILTER_DOWNLOADING:
    case FILTER_PAUSED:
    case FILTER_READY:
      return torrent.offline_state === filter
    default:
      return true
  }
}

export const filterTorrents = (torrents, filter) =>
  filter === FILTER_ALL ? torrents : (torrents || []).filter(torrent => matchesFilter(torrent, filter))

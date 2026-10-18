import { useEffect, useState } from 'react'
import axios from 'axios'

import { settingsHost, torrentsHost } from './Hosts'

// Downloading a file or a whole torrent into the long-term disk cache.
// index is the file id from file_stats; 0 downloads the whole torrent.

export const downloadToCache = (hash, index = 0) => axios.post(torrentsHost(), { action: 'download', hash, index })

export const cancelDownload = (hash, index = 0) =>
  axios.post(torrentsHost(), { action: 'cancel_download', hash, index })

export const downloadErrorMessage = error => error?.response?.data?.error || error?.message || String(error)

export const loadedPercent = (loaded, total) =>
  total > 0 ? Math.min(100, Math.floor(((loaded || 0) / total) * 100)) : 0

// Every torrent card asks for the setting, so share one request for a while.
const settingsTTL = 10000
let longTermRequest = null
let longTermRequestedAt = 0

const fetchLongTermCache = () => {
  if (!longTermRequest || Date.now() - longTermRequestedAt > settingsTTL) {
    longTermRequestedAt = Date.now()
    longTermRequest = axios
      .post(settingsHost(), { action: 'get' })
      .then(({ data }) => !!(data?.LongTermCache && data?.UseDisk))
      .catch(() => false)
  }
  return longTermRequest
}

export const useLongTermCache = () => {
  const [enabled, setEnabled] = useState(false)
  useEffect(() => {
    let active = true
    fetchLongTermCache().then(value => active && setEnabled(value))
    return () => {
      active = false
    }
  }, [])
  return enabled
}

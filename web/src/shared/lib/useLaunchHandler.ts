import { useEffect, useState } from 'react'

import { hasLaunchQuery, readLaunchSourceFromUrl } from 'shared/lib/launchSource'

interface LaunchQueueFileHandle {
  getFile: () => Promise<File>
}

interface LaunchParams {
  files?: LaunchQueueFileHandle[]
}

interface LaunchQueue {
  setConsumer: (callback: (params: LaunchParams) => void | Promise<void>) => void
}

declare global {
  interface Window {
    launchQueue?: LaunchQueue
  }
}

function consumeLaunchSourceFromUrl(): string | null {
  if (typeof window === 'undefined') return null
  const source = readLaunchSourceFromUrl(window.location.search)
  if (source && hasLaunchQuery(window.location.search)) {
    window.history.replaceState(null, '', window.location.pathname)
  }
  return source
}

/** Handles PWA launch via protocol_handlers (magnet: / web+torrs:), share_target, and file_handlers (.torrent). */
export default function useLaunchHandler() {
  const [launchSource, setLaunchSource] = useState<string | null>(() => consumeLaunchSourceFromUrl())
  const [launchFiles, setLaunchFiles] = useState<File[] | null>(null)

  useEffect(() => {
    if (!window.launchQueue) return undefined
    window.launchQueue.setConsumer(async launchParams => {
      if (!launchParams.files || launchParams.files.length === 0) return
      const files = await Promise.all(launchParams.files.map(fh => fh.getFile()))
      const torrentFiles = files.filter(f => f.name.toLowerCase().endsWith('.torrent'))
      if (torrentFiles.length > 0) setLaunchFiles(torrentFiles)
    })
    return undefined
  }, [])

  return { launchSource, setLaunchSource, launchFiles, setLaunchFiles }
}

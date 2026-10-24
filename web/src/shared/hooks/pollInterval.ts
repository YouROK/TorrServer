import type { TorrentStat } from 'shared/api/types'
import { GETTING_INFO, PRELOAD } from 'shared/torrent/states'

/** List poll. A hidden tab slows down; metadata resolve stays fast. */
export function torrentsRefetchIntervalMs(hidden: boolean, data: TorrentStat[] | undefined): number {
  if (hidden) return 10_000
  if (data?.some(torrent => torrent.stat === GETTING_INFO)) return 500
  return 1000
}

/** Detail poll. Hidden tab, GETTING_INFO, and PRELOAD each have their own cadence. */
export function torrentDetailRefetchIntervalMs(hidden: boolean, stat: number | undefined): number {
  if (hidden) return 5000
  if (stat === GETTING_INFO) return 500
  if (stat === PRELOAD) return 1000
  return 2000
}

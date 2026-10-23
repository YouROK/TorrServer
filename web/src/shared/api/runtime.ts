import axios from 'axios'

import { runtimeStatusHost } from 'shared/api/hosts'

export interface RuntimeTorrentBrief {
  hash: string
  title?: string
  name?: string
  stat?: number
  stat_string?: string
  torrent_size?: number
  loaded_size?: number
  download_speed?: number
  upload_speed?: number
  active_peers?: number
  total_peers?: number
  connected_seeders?: number
  preloaded_bytes?: number
  preload_size?: number
}

export interface RuntimeBTStatus {
  listen_port?: number
  peer_id?: string
  banned_ips?: number
  active_streams?: number
  torrent_count?: number
  total_size?: number
  loaded_size?: number
  active_peers?: number
  total_peers?: number
  connected_seeders?: number
  bytes_read?: number
  bytes_written?: number
  download_speed?: number
  upload_speed?: number
  torrents?: RuntimeTorrentBrief[]
  raw_stat?: string
}

/** Process MemStats + torrstor cache aggregates from GET /runtime/status. */
export interface RuntimeMemory {
  num_goroutine?: number
  gomaxprocs?: number
  heap_alloc?: number
  heap_inuse?: number
  heap_idle?: number
  heap_sys?: number
  heap_released?: number
  heap_objects?: number
  stack_inuse?: number
  stack_sys?: number
  sys?: number
  total_alloc?: number
  mallocs?: number
  frees?: number
  live_objects?: number
  next_gc?: number
  last_gc_unix_ns?: number
  num_gc?: number
  num_forced_gc?: number
  gc_cpu_fraction?: number
  gc_pause_total_ns?: number
  gc_pause_ns_last?: number
  cache_filled?: number
  cache_capacity?: number
  cache_count?: number
}

export interface RuntimeCacheRow {
  hash: string
  filled?: number
  capacity?: number
  pieces?: number
  closed?: boolean
}

export interface RuntimeStatus {
  dlna_enabled?: boolean
  bonjour_enabled?: boolean
  friendly_name?: string
  webdav_enabled?: boolean
  webdav_path?: string
  fuse_path?: string
  fuse_enabled?: boolean
  memory?: RuntimeMemory
  caches?: RuntimeCacheRow[]
  bt?: RuntimeBTStatus
}

export const RUNTIME_STATUS_QUERY_KEY = ['runtime-status'] as const

export const getRuntimeStatus = async (signal?: AbortSignal): Promise<RuntimeStatus> => {
  const { data } = await axios.get<RuntimeStatus>(runtimeStatusHost(), { signal })
  return data
}

/**
 * Absolute TorrServer API base URLs.
 *
 * Defaults to the page origin (`protocol` + `hostname` + `port`) so production
 * embeds and reverse proxies work without config. Set `VITE_SERVER_HOST` in
 * Vite when the UI is served from a different host than the API (dev proxy /
 * split deploy).
 *
 * Prefer these helpers over hard-coding paths — Vite's proxy and Go embed both
 * expect the same `/torrents`, `/stream`, … routes.
 */

const { protocol, hostname, port } = window.location

const torrserverHost = import.meta.env.VITE_SERVER_HOST || `${protocol}//${hostname}${port ? `:${port}` : ''}`

export const torrentsHost = () => `${torrserverHost}/torrents`
export const viewedHost = () => `${torrserverHost}/viewed`
export const cacheHost = () => `${torrserverHost}/cache`
export const torrentUploadHost = () => `${torrserverHost}/torrent/upload`
export const settingsHost = () => `${torrserverHost}/settings`
export const streamHost = () => `${torrserverHost}/stream`
export const playlistAllHost = () => `${torrserverHost}/playlistall/all.m3u`
export const ffpHost = (hash: string, id: number | string) => `${torrserverHost}/ffp/${hash}/${id}`
export const downloadTestHost = (sizeMb: number) => `${torrserverHost}/download/${sizeMb}`
export const shutdownHost = () => `${torrserverHost}/shutdown`
export const echoHost = () => `${torrserverHost}/echo`
/** Alias: per-torrent M3U playlists are served under `/stream`. */
export const playlistTorrHost = streamHost
export const torznabSearchHost = () => `${torrserverHost}/torznab/search/`
export const torznabCapsHost = () => `${torrserverHost}/torznab/caps`
export const searchHost = () => `${torrserverHost}/search/`
export const torznabTestHost = () => `${torrserverHost}/torznab/test`
export const tmdbSettingsHost = () => `${torrserverHost}/tmdb/settings`
export const gstSettingsHost = () => `${torrserverHost}/gst/settings`
export const gstEchoHost = () => `${torrserverHost}/gst/echo`
export const storageSettingsHost = () => `${torrserverHost}/storage/settings`
export const runtimeStatusHost = () => `${torrserverHost}/runtime/status`

/** Resolved API origin (no trailing path). Useful for building ad-hoc GST URLs. */
export const getTorrServerHost = () => torrserverHost

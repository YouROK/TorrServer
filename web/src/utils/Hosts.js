let torrserverHost = process.env.REACT_APP_SERVER_HOST || ''

export const torrentsHost = () => `${torrserverHost}/torrents`
export const viewedHost = () => `${torrserverHost}/viewed`
export const cacheHost = () => `${torrserverHost}/cache`
export const torrentUploadHost = () => `${torrserverHost}/torrent/upload`
export const settingsHost = () => `${torrserverHost}/settings`
export const streamHost = () => `${torrserverHost}/stream`
export const shutdownHost = () => `${torrserverHost}/shutdown`
export const echoHost = () => `${torrserverHost}/echo`
export const playlistAllHost = () => `${torrserverHost}/playlistall/all.m3u`
export const playlistTorrHost = () => `${torrserverHost}/stream`

export const getTorrServerHost = () => torrserverHost
export const setTorrServerHost = host => {
  torrserverHost = host
}

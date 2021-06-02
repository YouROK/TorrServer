export function humanizeSize(size) {
  if (!size) return ''
  const i = Math.floor(Math.log(size) / Math.log(1024))
  return `${(size / Math.pow(1024, i)).toFixed(2) * 1} ${['B', 'kB', 'MB', 'GB', 'TB'][i]}`
}

export function getPeerString(torrent) {
  if (!torrent || !torrent.connected_seeders) return ''
  return `[${torrent.connected_seeders}] ${torrent.active_peers} / ${torrent.total_peers}`
}

export const shortenText = (text, sympolAmount) =>
  text ? text.slice(0, sympolAmount) + (text.length > sympolAmount ? '...' : '') : ''

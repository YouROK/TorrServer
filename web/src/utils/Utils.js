export function humanizeSize(size) {
    if (!size) return ''
    var i = Math.floor(Math.log(size) / Math.log(1024))
    return (size / Math.pow(1024, i)).toFixed(2) * 1 + ' ' + ['B', 'kB', 'MB', 'GB', 'TB'][i]
}

export function getPeerString(torrent) {
    if (!torrent || !torrent.connected_seeders) return '[0] 0 / 0'
    return '[' + torrent.connected_seeders + '] ' + torrent.active_peers + ' / ' + torrent.total_peers
}

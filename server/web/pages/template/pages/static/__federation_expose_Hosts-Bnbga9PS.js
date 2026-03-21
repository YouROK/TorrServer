var define_process_env_default = {};
const { protocol, hostname, port } = window.location;
let torrserverHost = define_process_env_default.REACT_APP_SERVER_HOST || `${protocol}//${hostname}${port ? `:${port}` : ""}`;
const torrentsHost = () => `${torrserverHost}/api/torrents`;
const viewedHost = () => `${torrserverHost}/api/viewed`;
const cacheHost = () => `${torrserverHost}/api/cache`;
const torrentUploadHost = () => `${torrserverHost}/api/torrent/upload`;
const settingsHost = () => `${torrserverHost}/api/settings`;
const streamHost = () => `${torrserverHost}/api/stream`;
const shutdownHost = () => `${torrserverHost}/api/shutdown`;
const echoHost = () => `${torrserverHost}/api/echo`;
const playlistTorrHost = () => `${torrserverHost}/api/stream`;
const torznabSearchHost = () => `${torrserverHost}/api/torznab/search`;
const searchHost = () => `${torrserverHost}/api/search`;
const torznabTestHost = () => `${torrserverHost}/api/torznab/test`;
const downloadZipHost = () => `${torrserverHost}/api/downloadzip`;
const getTorrServerHost = () => torrserverHost;
const setTorrServerHost = (host) => {
  torrserverHost = host;
};
export {
  cacheHost,
  downloadZipHost,
  echoHost,
  getTorrServerHost,
  playlistTorrHost,
  searchHost,
  setTorrServerHost,
  settingsHost,
  shutdownHost,
  streamHost,
  torrentUploadHost,
  torrentsHost,
  torznabSearchHost,
  torznabTestHost,
  viewedHost
};

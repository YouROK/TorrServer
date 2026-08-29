export default {
  CacheSize: 64,
  ReaderReadAHead: 95,
  PreloadCache: 50,
  UseDisk: false,
  TorrentsSavePath: '',
  RemoveCacheOnDrop: false,
  ForceEncrypt: false,
  RetrackersMode: 1,
  TrackersListURL: 'https://raw.githubusercontent.com/ngosang/trackerslist/master/trackers_best_ip.txt',
  DefaultTrackers: `http://retracker.local/announce
http://bt4.t-ru.org/ann?magnet
http://retracker.mgts.by:80/announce
http://tracker.city9x.com:2710/announce
http://tracker.electro-torrent.pl:80/announce
http://tracker.internetwarriors.net:1337/announce
http://tracker2.itzmx.com:6961/announce
udp://opentor.org:2710
udp://public.popcorn-tracker.org:6969/announce
udp://tracker.opentrackr.org:1337/announce
http://bt.svao-ix.ru/announce
udp://explodie.org:6969/announce
wss://tracker.btorrent.xyz
wss://tracker.openwebtorrent.com`,
  TorrentDisconnectTimeout: 30,
  EnableDebug: false,
  EnableDLNA: false,
  EnableBonjour: true,
  FriendlyName: '',
  EnableRutorSearch: false,
  EnableIPv6: false,
  DisableTCP: false,
  DisableUTP: false,
  DisableUPNP: false,
  DisableDHT: false,
  DisablePEX: false,
  DisableUpload: false,
  EnableLPD: true,
  LPDIPv6: false,
  DownloadRateLimit: 0,
  UploadRateLimit: 0,
  ConnectionsLimit: 25,
  PeersListenPort: 0,
  ResponsiveMode: true,
  SslPort: 0,
  SslCert: '',
  SslKey: '',
  ShowFSActiveTorr: true,
  StoreSettingsInJson: true,
}

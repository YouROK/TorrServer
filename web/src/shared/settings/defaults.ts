import type { BTSets } from 'shared/api/types'

/**
 * Client-side BTSets seed for forms before `/settings` loads.
 * "Reset to defaults" in the UI must call the server `reset` action — do not treat
 * this object as authoritative server defaults.
 */
const defaultSettings: BTSets = {
  CacheSize: 64,
  ReaderReadAHead: 95,
  PreloadCache: 50,
  UseDisk: false,
  TorrentsSavePath: '',
  RemoveCacheOnDrop: false,
  ForceEncrypt: false,
  RetrackersMode: 1,
  TorrentDisconnectTimeout: 30,
  EnableDebug: false,
  EnableDLNA: false,
  EnableBonjour: true,
  FriendlyName: '',
  EnableRutorSearch: false,
  EnableTorznabSearch: false,
  TorznabUrls: [],
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
  TrackTimecode: false,
  StoreSettingsInJson: true,
  TMDBSettings: {
    APIKey: '',
    APIURL: 'https://api.themoviedb.org/3',
    ImageURL: 'https://image.tmdb.org',
    ImageURLRu: 'https://imagetmdb.com',
  },
}

export default defaultSettings

declare module 'parse-torrent' {
  interface ParsedTorrent {
    infoHash?: string
    name?: string
    files?: Array<{ name?: string; length?: number }>
    [key: string]: unknown
  }

  export function remote(
    torrent: Buffer | string | File | Blob,
    callback: (err: Error | null, parsed?: ParsedTorrent) => void,
  ): void
  export function remote(
    torrent: Buffer | string | File | Blob,
    opts: object,
    callback: (err: Error | null, parsed?: ParsedTorrent) => void,
  ): void

  /** v11+: async parse of magnet / infohash / .torrent buffer. */
  export default function parseTorrent(
    torrent: Buffer | string | File | ArrayBufferView | object,
  ): Promise<ParsedTorrent>
}

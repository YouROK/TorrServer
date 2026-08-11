import type { TorrentCache } from 'shared/api/types'

/** Cheap FNV-ish mix for poll equality — avoids huge string concat on sparse piece maps. */
const mix = (h: number, v: number) => {
  let x = (h ^ (v >>> 0)) >>> 0
  x = Math.imul(x ^ (x >>> 16), 0x45d9f3b)
  x = Math.imul(x ^ (x >>> 16), 0x45d9f3b)
  return (x ^ (x >>> 16)) >>> 0
}

export const cheapReadersFingerprint = (readers: TorrentCache['Readers']): number => {
  if (!readers?.length) return 0
  let h = 2166136261
  for (const r of readers) {
    h = mix(h, r.Reader ?? -1)
    h = mix(h, r.Start ?? -1)
    h = mix(h, r.End ?? -1)
    h = mix(h, r.Active === false ? 0 : 1)
  }
  return mix(h, readers.length)
}

export const cheapPiecesFingerprint = (pieces: TorrentCache['Pieces']): number => {
  if (!pieces) return 0
  let h = 2166136261
  let count = 0
  if (Array.isArray(pieces)) {
    for (let i = 0; i < pieces.length; i++) {
      const p = pieces[i]
      if (!p) continue
      count += 1
      h = mix(h, i)
      h = mix(h, p.Size ?? 0)
      h = mix(h, p.Length ?? 0)
      h = mix(h, p.Priority ?? 0)
      h = mix(h, p.Completed ? 1 : 0)
    }
  } else {
    for (const [key, p] of Object.entries(pieces)) {
      if (!p) continue
      count += 1
      h = mix(h, Number(key) || 0)
      h = mix(h, p.Size ?? 0)
      h = mix(h, p.Length ?? 0)
      h = mix(h, p.Priority ?? 0)
      h = mix(h, p.Completed ? 1 : 0)
    }
  }
  return mix(h, count)
}

export const cacheVisualEqual = (a: TorrentCache, b: TorrentCache): boolean =>
  a.Filled === b.Filled &&
  a.Capacity === b.Capacity &&
  a.PiecesCount === b.PiecesCount &&
  a.PiecesLength === b.PiecesLength &&
  cheapReadersFingerprint(a.Readers) === cheapReadersFingerprint(b.Readers) &&
  cheapPiecesFingerprint(a.Pieces) === cheapPiecesFingerprint(b.Pieces)

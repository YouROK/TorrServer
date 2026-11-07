// Reads family / full / PostScript names of a font file with a couple of small Range
// requests instead of downloading the whole file. Supports TrueType, OpenType, TTC
// collections and WOFF (zlib). WOFF2 (brotli) cannot be read this way.

const TAG_TTF = 0x00010000
const TAG_OTTO = 0x4f54544f
const TAG_TRUE = 0x74727565
const TAG_TTCF = 0x74746366
const TAG_WOFF = 0x774f4646
const TAG_WOFF2 = 0x774f4632
const TAG_NAME = 0x6e616d65

const MAX_NAME_TABLE = 256 * 1024
const MAX_WHOLE_FILE = 32 * 1024 * 1024

// Names that are collected: family, subfamily, full name, PostScript name,
// typographic family and subfamily
const NAME_IDS = [1, 2, 4, 6, 16, 17]

export class UnreadableFontError extends Error {}

// Range reader. If the server ignores Range and answers 200, the body is used as a whole.
const createReader = (url, signal) => {
  let whole = null
  const get = async (start, length) => {
    if (whole) return whole.slice(start, start + length)
    const response = await fetch(url, { headers: { Range: `bytes=${start}-${start + length - 1}` }, signal })
    if (response.status === 206) return response.arrayBuffer()
    if (response.status === 200) {
      const size = Number(response.headers.get('Content-Length') || 0)
      if (size > MAX_WHOLE_FILE) throw new UnreadableFontError('server ignores Range and the file is too big')
      whole = await response.arrayBuffer()
      return whole.slice(start, start + length)
    }
    throw new UnreadableFontError(`status ${response.status}`)
  }
  return { get }
}

const decodeName = (view, platform, offset, length) => {
  const chars = []
  if (platform === 3 || platform === 0) {
    for (let i = 0; i + 1 < length; i += 2) chars.push(view.getUint16(offset + i))
  } else {
    // Macintosh and others: single bytes, only ASCII is reliable
    for (let i = 0; i < length; i++) chars.push(view.getUint8(offset + i))
  }
  return String.fromCharCode(...chars)
    .replace(/\0/g, '')
    .trim()
}

const parseNameTable = buffer => {
  const view = new DataView(buffer)
  if (view.byteLength < 6) return null
  const count = view.getUint16(2)
  const stringOffset = view.getUint16(4)
  const names = {}
  NAME_IDS.forEach(id => {
    names[id] = []
  })
  for (let i = 0; i < count; i++) {
    const record = 6 + i * 12
    if (record + 12 > view.byteLength) break
    const platform = view.getUint16(record)
    const nameId = view.getUint16(record + 6)
    const length = view.getUint16(record + 8)
    const offset = stringOffset + view.getUint16(record + 10)
    if (names[nameId] && offset + length <= view.byteLength) {
      const value = decodeName(view, platform, offset, length)
      if (value && !names[nameId].includes(value)) names[nameId].push(value)
    }
  }
  return names
}

const inflate = async buffer => {
  if (!window.DecompressionStream) throw new UnreadableFontError('WOFF needs DecompressionStream')
  const stream = new Blob([buffer]).stream().pipeThrough(new window.DecompressionStream('deflate'))
  return new Response(stream).arrayBuffer()
}

const toInfo = names => {
  const style = [...names[17], ...names[2]].join(' ').toLowerCase()
  return {
    families: [...names[1], ...names[16]],
    fulls: names[4],
    postscript: names[6],
    bold: /bold|black|heavy/.test(style),
    italic: /italic|oblique/.test(style),
  }
}

// Reads one font starting at the given offset of the file (offset != 0 for TTC members)
const readFont = async (reader, offset) => {
  const header = new DataView(await reader.get(offset, 16))
  const isWoff = header.getUint32(0) === TAG_WOFF
  const numTables = header.getUint16(isWoff ? 12 : 4)
  if (numTables === 0 || numTables > 100) throw new UnreadableFontError('invalid table directory')
  const dirStart = isWoff ? 44 : offset + 12
  const dirSize = numTables * (isWoff ? 20 : 16)
  const directory = new DataView(await reader.get(dirStart, dirSize))

  for (let i = 0; i < numTables; i++) {
    const entry = i * (isWoff ? 20 : 16)
    if (directory.getUint32(entry) === TAG_NAME) {
      // sfnt record: tag, checksum, offset, length; WOFF record: tag, offset, compLength, origLength, checksum
      const tableOffset = directory.getUint32(entry + (isWoff ? 4 : 8))
      const storedLength = directory.getUint32(entry + (isWoff ? 8 : 12))
      if (storedLength > MAX_NAME_TABLE) throw new UnreadableFontError('name table is too big')
      // eslint-disable-next-line no-await-in-loop
      let table = await reader.get(tableOffset, storedLength)
      // eslint-disable-next-line no-await-in-loop
      if (isWoff && storedLength < directory.getUint32(entry + 12)) table = await inflate(table)
      const names = parseNameTable(table)
      if (!names) throw new UnreadableFontError('invalid name table')
      return toInfo(names)
    }
  }
  return null
}

// Returns an array of font infos (several for a TTC collection), [] for fonts without a name table.
// Throws UnreadableFontError for data that is not a font (for example zero bytes).
export const readFontNames = async (url, signal) => {
  const reader = createReader(url, signal)
  const head = new DataView(await reader.get(0, 16))
  if (head.byteLength < 12) throw new UnreadableFontError('file is too short')
  const tag = head.getUint32(0)

  if (tag === TAG_TTCF) {
    const count = Math.min(head.getUint32(8), 4)
    const offsets = new DataView(await reader.get(12, count * 4))
    const infos = []
    for (let i = 0; i < count; i++) {
      // eslint-disable-next-line no-await-in-loop
      const info = await readFont(reader, offsets.getUint32(i * 4))
      if (info) infos.push(info)
    }
    return infos
  }
  if (tag === TAG_WOFF2) throw new UnreadableFontError('WOFF2 is not supported')
  if (tag === TAG_TTF || tag === TAG_OTTO || tag === TAG_TRUE || tag === TAG_WOFF) {
    const info = await readFont(reader, 0)
    return info ? [info] : []
  }
  throw new UnreadableFontError(tag === 0 ? 'zero bytes instead of a font' : 'not a font file')
}

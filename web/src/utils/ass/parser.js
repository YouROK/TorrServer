// Minimal ASS/SSA parser: [Script Info], styles and dialogue events with override tags.
// It only understands what the simple DOM renderer can draw (see renderer.js):
// alignment, \pos, \move, \fad, colors, alpha, font, size, bold/italic/underline/strike,
// border and shadow, spacing, \r, \q, simple karaoke (\k). Vector drawings (\p) are skipped.

const DEFAULT_STYLE = {
  name: 'Default',
  fontName: 'Arial',
  fontSize: 20,
  c1: { r: 255, g: 255, b: 255 },
  c2: { r: 255, g: 0, b: 0 },
  c3: { r: 0, g: 0, b: 0 },
  c4: { r: 0, g: 0, b: 0 },
  a1: 0,
  a2: 0,
  a3: 0,
  a4: 128,
  bold: false,
  italic: false,
  underline: false,
  strike: false,
  scaleX: 100,
  scaleY: 100,
  spacing: 0,
  borderStyle: 1,
  outline: 2,
  shadow: 2,
  alignment: 2,
  marginL: 10,
  marginR: 10,
  marginV: 10,
}

const num = (value, fallback = 0) => {
  const n = parseFloat(value)
  return Number.isFinite(n) ? n : fallback
}

const flag = value => {
  const n = parseInt(value, 10)
  return Number.isFinite(n) && n !== 0
}

// "&HAABBGGRR", "&HBBGGRR&" or a decimal number -> { color: {r, g, b}, alpha }
export const parseColor = value => {
  if (value === undefined || value === null) return null
  let s = String(value).trim()
  if (!s) return null
  let n
  if (/^&?h/i.test(s)) {
    s = s.replace(/^&?h/i, '').replace(/&$/, '')
    n = parseInt(s, 16)
  } else {
    n = parseInt(s, 10)
  }
  if (!Number.isFinite(n)) return null
  // unsigned 32-bit value: 0xAABBGGRR
  const u = ((n % 4294967296) + 4294967296) % 4294967296
  return {
    color: { r: u % 256, g: Math.floor(u / 256) % 256, b: Math.floor(u / 65536) % 256 },
    alpha: Math.floor(u / 16777216) % 256,
  }
}

// "&H80&" -> 128
const parseAlpha = value => {
  const n = parseInt(String(value).replace(/^&?h/i, '').replace(/&$/, ''), 16)
  return Number.isFinite(n) ? ((n % 256) + 256) % 256 : null
}

// "h:mm:ss.cc" -> seconds
export const parseTime = value => {
  const m = /^\s*(\d+):(\d+):(\d+)(?:[.,](\d+))?\s*$/.exec(value || '')
  if (!m) return NaN
  const frac = m[4] ? parseInt(m[4], 10) / 10 ** m[4].length : 0
  return parseInt(m[1], 10) * 3600 + parseInt(m[2], 10) * 60 + parseInt(m[3], 10) + frac
}

// SSA (V4) alignment -> ASS numpad alignment
const ssaAlignment = a => {
  if (a >= 9 && a <= 11) return a - 5 // middle: 9,10,11 -> 4,5,6
  if (a >= 5 && a <= 7) return a + 2 // top: 5,6,7 -> 7,8,9
  return a >= 1 && a <= 3 ? a : 2
}

const splitFormat = line =>
  line
    .slice(line.indexOf(':') + 1)
    .split(',')
    .map(s => s.trim().toLowerCase())

const buildStyle = (fields, format, isSsa) => {
  const get = key => fields[format.indexOf(key)]
  const st = { ...DEFAULT_STYLE }
  const name = get('name')
  if (name) st.name = name.trim()
  if (get('fontname')) st.fontName = get('fontname').trim()
  st.fontSize = num(get('fontsize'), st.fontSize)

  const c1 = parseColor(get('primarycolour'))
  const c2 = parseColor(get('secondarycolour'))
  const c3 = parseColor(get('outlinecolour') ?? get('tertiarycolour'))
  const c4 = parseColor(get('backcolour'))
  if (c1) [st.c1, st.a1] = [c1.color, c1.alpha]
  if (c2) [st.c2, st.a2] = [c2.color, c2.alpha]
  if (c3) [st.c3, st.a3] = [c3.color, c3.alpha]
  if (c4) [st.c4, st.a4] = [c4.color, c4.alpha]

  st.bold = flag(get('bold'))
  st.italic = flag(get('italic'))
  st.underline = flag(get('underline'))
  st.strike = flag(get('strikeout'))
  st.scaleX = num(get('scalex'), 100)
  st.scaleY = num(get('scaley'), 100)
  st.spacing = num(get('spacing'), 0)
  st.borderStyle = num(get('borderstyle'), 1)
  st.outline = num(get('outline'), st.outline)
  st.shadow = num(get('shadow'), st.shadow)
  const al = num(get('alignment'), 2)
  st.alignment = isSsa ? ssaAlignment(al) : al >= 1 && al <= 9 ? al : 2
  st.marginL = num(get('marginl'), st.marginL)
  st.marginR = num(get('marginr'), st.marginR)
  st.marginV = num(get('marginv'), st.marginV)
  return st
}

// Longer names must come before their prefixes (fade/fad, bord/b, alpha/an/a, ...).
const TAGS = [
  'fscx',
  'fscy',
  'fsp',
  'fs',
  'fade',
  'fad',
  'fn',
  'fe',
  'frx',
  'fry',
  'frz',
  'fr',
  'fax',
  'fay',
  'xbord',
  'ybord',
  'bord',
  'xshad',
  'yshad',
  'shad',
  'blur',
  'be',
  'alpha',
  'an',
  '1a',
  '2a',
  '3a',
  '4a',
  '1c',
  '2c',
  '3c',
  '4c',
  'iclip',
  'clip',
  'pos',
  'move',
  'org',
  'kf',
  'ko',
  'k',
  'K',
  'a',
  'b',
  'c',
  'i',
  'u',
  's',
  'r',
  'q',
  'p',
  't',
]

// Splits an override block into [name, argument] pairs. \t(...) and other
// parenthesised arguments are taken whole (they may contain backslashes).
const scanTags = block => {
  const out = []
  let i = 0
  while (i < block.length) {
    if (block[i] === '\\') {
      i++
      const rest = block.slice(i, i + 6)
      const name = TAGS.find(t => rest.startsWith(t))
      if (name) {
        i += name.length
        let arg
        if (block[i] === '(') {
          let depth = 0
          let j = i
          for (; j < block.length; j++) {
            if (block[j] === '(') depth++
            else if (block[j] === ')') {
              depth--
              if (depth === 0) break
            }
          }
          arg = block.slice(i + 1, j)
          i = j + 1
        } else {
          let j = block.indexOf('\\', i)
          if (j < 0) j = block.length
          arg = block.slice(i, j)
          i = j
        }
        out.push([name, arg.trim()])
      }
    } else {
      i++
    }
  }
  return out
}

const styleToState = st => ({
  fn: st.fontName,
  fs: st.fontSize,
  b: st.bold,
  i: st.italic,
  u: st.underline,
  s: st.strike,
  c1: st.c1,
  c2: st.c2,
  c3: st.c3,
  c4: st.c4,
  a1: st.a1,
  a2: st.a2,
  a3: st.a3,
  a4: st.a4,
  bord: st.outline,
  shad: st.shadow,
  fscx: st.scaleX,
  fscy: st.scaleY,
  fsp: st.spacing,
})

const COLOR_TAGS = { c: 'c1', '1c': 'c1', '2c': 'c2', '3c': 'c3', '4c': 'c4' }
const ALPHA_TAGS = { '1a': 'a1', '2a': 'a2', '3a': 'a3', '4a': 'a4' }

const csv = arg => arg.split(',').map(s => s.trim())

const parseEventText = (raw, style, styles, wrapStyle) => {
  const segments = []
  const ev = { an: null, pos: null, move: null, fad: null, wrap: wrapStyle }
  let state = styleToState(style)
  let drawing = false
  let kAcc = 0 // ms, start of the next karaoke syllable
  let kind = null // { start, dur, type } of the current syllable
  const softBreak = wrapStyle === 2 ? '\n' : ' '

  const pushText = text => {
    if (!text || drawing) return
    const clean = text.replace(/\\N/g, '\n').replace(/\\n/g, softBreak).replace(/\\h/g, '\u00A0')
    if (!clean) return
    segments.push({ text: clean, ...state, k: kind })
  }

  const applyTag = (name, arg) => {
    switch (name) {
      case 'an': {
        const n = parseInt(arg, 10)
        if (ev.an === null && n >= 1 && n <= 9) ev.an = n
        break
      }
      case 'a': {
        const n = parseInt(arg, 10)
        if (ev.an === null && Number.isFinite(n)) ev.an = ssaAlignment(n)
        break
      }
      case 'pos': {
        const [x, y] = csv(arg).map(Number)
        if (!ev.pos && Number.isFinite(x) && Number.isFinite(y)) ev.pos = [x, y]
        break
      }
      case 'move': {
        const v = csv(arg).map(Number)
        if (!ev.move && v.length >= 4 && v.slice(0, 4).every(Number.isFinite)) {
          ev.move = { x1: v[0], y1: v[1], x2: v[2], y2: v[3], t1: v[4] || 0, t2: v[5] || 0 }
          ev.pos = [v[0], v[1]]
        }
        break
      }
      case 'fad': {
        const [a, b] = csv(arg).map(Number)
        if (!ev.fad && Number.isFinite(a) && Number.isFinite(b)) ev.fad = [a, b]
        break
      }
      case 'q': {
        const n = parseInt(arg, 10)
        if (n >= 0 && n <= 3) ev.wrap = n
        break
      }
      case 'fn':
        state = { ...state, fn: arg || style.fontName }
        break
      case 'fs':
        state = { ...state, fs: num(arg, 0) > 0 ? num(arg) : style.fontSize }
        break
      case 'fscx':
        state = { ...state, fscx: num(arg, 0) > 0 ? num(arg) : style.scaleX }
        break
      case 'fscy':
        state = { ...state, fscy: num(arg, 0) > 0 ? num(arg) : style.scaleY }
        break
      case 'fsp':
        state = { ...state, fsp: arg === '' ? style.spacing : num(arg) }
        break
      case 'b': {
        const n = parseInt(arg, 10)
        state = { ...state, b: arg === '' ? style.bold : n === 1 || n >= 700 }
        break
      }
      case 'i':
        state = { ...state, i: arg === '' ? style.italic : flag(arg) }
        break
      case 'u':
        state = { ...state, u: arg === '' ? style.underline : flag(arg) }
        break
      case 's':
        state = { ...state, s: arg === '' ? style.strike : flag(arg) }
        break
      case 'bord':
      case 'xbord':
      case 'ybord':
        state = { ...state, bord: arg === '' ? style.outline : Math.max(0, num(arg)) }
        break
      case 'shad':
      case 'xshad':
      case 'yshad':
        state = { ...state, shad: arg === '' ? style.shadow : Math.max(0, num(arg)) }
        break
      case 'alpha': {
        const a = parseAlpha(arg)
        if (a !== null) state = { ...state, a1: a, a2: a, a3: a, a4: a }
        break
      }
      case 'r': {
        const target = styles[arg] || style
        state = styleToState(target)
        break
      }
      case 'p':
        drawing = parseInt(arg, 10) > 0
        break
      case 'k':
      case 'K':
      case 'kf':
      case 'ko': {
        const dur = Math.max(0, num(arg)) * 10
        kind = { start: kAcc, dur, type: name === 'ko' ? 'ko' : name === 'k' ? 'k' : 'kf' }
        kAcc += dur
        break
      }
      default:
        if (COLOR_TAGS[name]) {
          const key = COLOR_TAGS[name]
          const parsed = arg === '' ? null : parseColor(arg)
          state = { ...state, [key]: parsed ? parsed.color : style[key] }
        } else if (ALPHA_TAGS[name]) {
          const key = ALPHA_TAGS[name]
          const a = arg === '' ? style[key] : parseAlpha(arg)
          if (a !== null) state = { ...state, [key]: a }
        }
        // everything else (\t, \clip, \frz, \blur, \org, ...) is not supported and ignored
        break
    }
  }

  const re = /\{([^}]*)\}/g
  let last = 0
  let m = re.exec(raw)
  while (m) {
    pushText(raw.slice(last, m.index))
    scanTags(m[1]).forEach(([name, arg]) => applyTag(name, arg))
    last = m.index + m[0].length
    m = re.exec(raw)
  }
  pushText(raw.slice(last))

  return { segments, ev }
}

export const parseAss = source => {
  const text = String(source || '')
    .replace(/^\uFEFF/, '')
    .replace(/\r\n?/g, '\n')
  const info = {}
  const styles = {}
  const events = []
  let section = ''
  let styleFormat = null
  let eventFormat = null
  let isSsa = false
  const dialogues = []

  text.split('\n').forEach(line => {
    const trimmed = line.trim()
    if (!trimmed || trimmed[0] === ';') return
    const sec = /^\[(.+)\]$/.exec(trimmed)
    if (sec) {
      section = sec[1].toLowerCase()
      isSsa = section === 'v4 styles'
      return
    }
    const colon = trimmed.indexOf(':')
    if (colon < 0) return
    const key = trimmed.slice(0, colon).trim().toLowerCase()
    const value = trimmed.slice(colon + 1).trim()

    if (section === 'script info') {
      info[key] = value
    } else if (section === 'v4+ styles' || section === 'v4 styles') {
      if (key === 'format') styleFormat = splitFormat(trimmed)
      else if (key === 'style' && styleFormat) {
        const fields = value.split(',')
        const st = buildStyle(fields, styleFormat, isSsa)
        styles[st.name] = st
      }
    } else if (section === 'events') {
      if (key === 'format') eventFormat = splitFormat(trimmed)
      else if (key === 'dialogue') dialogues.push(value)
    }
  })

  let playResX = num(info.playresx, 0)
  let playResY = num(info.playresy, 0)
  if (!playResX && !playResY) {
    playResX = 384
    playResY = 288
  } else if (!playResX) {
    playResX = Math.round((playResY * 4) / 3)
  } else if (!playResY) {
    playResY = Math.round((playResX * 3) / 4)
  }
  const wrapStyle = num(info.wrapstyle, 0)
  const scaledBorder = /^yes$/i.test(info.scaledborderandshadow || '')

  const format = eventFormat || [
    'layer',
    'start',
    'end',
    'style',
    'name',
    'marginl',
    'marginr',
    'marginv',
    'effect',
    'text',
  ]
  const textIndex = format.indexOf('text')
  const styleList = Object.values(styles)
  const fallbackStyle = styles.Default || styleList[0] || DEFAULT_STYLE

  dialogues.forEach((value, order) => {
    // Text is the last field and may contain commas
    const parts = []
    let rest = value
    for (let i = 0; i < textIndex; i++) {
      const comma = rest.indexOf(',')
      if (comma < 0) break
      parts.push(rest.slice(0, comma))
      rest = rest.slice(comma + 1)
    }
    parts.push(rest)
    const get = key => {
      const idx = format.indexOf(key)
      return idx >= 0 ? parts[idx] : undefined
    }

    const start = parseTime(get('start'))
    const end = parseTime(get('end'))
    if (!Number.isFinite(start) || !Number.isFinite(end) || end <= start) return

    const styleName = (get('style') || '').trim().replace(/^\*/, '')
    const style = styles[styleName] || fallbackStyle
    const { segments, ev } = parseEventText(get('text') || '', style, styles, wrapStyle)
    if (!segments.length) return

    const mL = num(get('marginl'), 0)
    const mR = num(get('marginr'), 0)
    const mV = num(get('marginv'), 0)
    events.push({
      order,
      start,
      end,
      layer: num(get('layer'), 0),
      style,
      an: ev.an || style.alignment,
      marginL: mL || style.marginL,
      marginR: mR || style.marginR,
      marginV: mV || style.marginV,
      pos: ev.pos,
      move: ev.move,
      fad: ev.fad,
      wrap: ev.wrap,
      segments,
      karaoke: segments.some(s => s.k),
    })
  })

  events.sort((a, b) => a.start - b.start || a.order - b.order)

  // Fonts used by the drawn text (style fonts and \fn overrides)
  const fonts = new Set()
  events.forEach(ev =>
    ev.segments.forEach(seg => {
      const name = String(seg.fn || '')
        .replace(/^@/, '')
        .trim()
      if (name) fonts.add(name)
    }),
  )
  return { info, playResX, playResY, scaledBorder, styles, events, fonts: [...fonts] }
}

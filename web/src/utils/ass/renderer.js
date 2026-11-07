import { parseAss } from './parser'

// Simple ASS renderer on DOM/CSS (no WebAssembly). It draws over the visible area of
// the <video> element and supports: styles, alignment, margins, outline and shadow,
// opaque box (BorderStyle 3), layers, stacking of lines with the same alignment,
// \pos, \move, \fad and simple karaoke (the syllable colour changes instantly).
// Not supported: \t animations, \clip, drawings (\p), rotation, blur, ScaleX != ScaleY.

const rgba = (c, alpha) => `rgba(${c.r},${c.g},${c.b},${((255 - alpha) / 255).toFixed(3)})`
const clamp = (v, min, max) => Math.min(Math.max(v, min), max)
const fontFamily = name => `"${String(name).replace(/^@/, '').replace(/["\\]/g, '')}", sans-serif`

// ASS font size is the height of the line (ascent + descent of the font), CSS font-size is the em square.
// The ratio line height / em differs from font to font, so it is measured for every font family.
const FALLBACK_CELL = 1.15
const cellCache = new Map()
let measureContext = null

const cellRatio = (name, bold, italic) => {
  const key = `${name}|${bold}|${italic}`
  if (cellCache.has(key)) return cellCache.get(key)
  let ratio = FALLBACK_CELL
  try {
    if (!measureContext) measureContext = document.createElement('canvas').getContext('2d')
    measureContext.font = `${italic ? 'italic ' : ''}${bold ? 700 : 400} 100px ${fontFamily(name)}`
    const metrics = measureContext.measureText('Hg')
    const height = (metrics.fontBoundingBoxAscent || 0) + (metrics.fontBoundingBoxDescent || 0)
    if (height > 0) ratio = height / 100
  } catch {
    // measuring is not available: use the typical ratio
  }
  cellCache.set(key, ratio)
  return ratio
}

const styleSpan = (span, seg, style, sx, sy, border) => {
  const s = span.style
  s.fontFamily = fontFamily(seg.fn)
  // line height as in ASS; the em size is derived from it with the measured ratio of the font
  const lineHeight = seg.fs * (seg.fscy / 100) * sy
  s.fontSize = `${lineHeight / cellRatio(seg.fn, seg.b, seg.i)}px`
  s.lineHeight = `${lineHeight}px`
  s.fontWeight = seg.b ? '700' : '400'
  s.fontStyle = seg.i ? 'italic' : 'normal'
  const deco = []
  if (seg.u) deco.push('underline')
  if (seg.s) deco.push('line-through')
  s.textDecoration = deco.length ? deco.join(' ') : 'none'
  if (seg.fsp) s.letterSpacing = `${seg.fsp * sx}px`
  s.color = rgba(seg.c1, seg.a1)

  const bord = seg.bord * border
  const shad = seg.shad * border
  if (style.borderStyle === 3) {
    // opaque box
    s.backgroundColor = rgba(seg.c3, seg.a3)
    s.padding = `${bord}px`
    s.boxDecorationBreak = 'clone'
    s.setProperty('-webkit-box-decoration-break', 'clone')
  } else {
    if (bord > 0) {
      s.setProperty('-webkit-text-stroke', `${2 * bord}px ${rgba(seg.c3, seg.a3)}`)
      s.setProperty('paint-order', 'stroke fill')
    }
    if (shad > 0) s.textShadow = `${shad}px ${shad}px 0 ${rgba(seg.c4, seg.a4)}`
  }
}

export default class AssRenderer {
  constructor(video, container) {
    this.video = video
    this.container = container
    this.script = null
    this.nodes = new Map()
    this.geo = null
    this.raf = 0

    this.layer = document.createElement('div')
    Object.assign(this.layer.style, {
      position: 'absolute',
      left: '0px',
      top: '0px',
      width: '0px',
      height: '0px',
      overflow: 'hidden',
      pointerEvents: 'none',
      zIndex: '2',
    })
    container.appendChild(this.layer)

    this.onTick = () => this.tick()
    this.onPlay = () => this.startLoop()
    this.onPause = () => {
      this.stopLoop()
      this.tick()
    }
    this.onGeometry = () => this.updateGeometry()

    this.tickEvents = ['timeupdate', 'seeking', 'seeked']
    this.tickEvents.forEach(name => video.addEventListener(name, this.onTick))
    ;['play', 'playing'].forEach(name => video.addEventListener(name, this.onPlay))
    ;['pause', 'ended'].forEach(name => video.addEventListener(name, this.onPause))
    ;['loadedmetadata', 'resize'].forEach(name => video.addEventListener(name, this.onGeometry))
    document.addEventListener('fullscreenchange', this.onGeometry)

    if (typeof ResizeObserver !== 'undefined') {
      this.observer = new ResizeObserver(this.onGeometry)
      this.observer.observe(video)
    } else {
      window.addEventListener('resize', this.onGeometry)
    }

    this.updateGeometry()
    if (!video.paused) this.startLoop()
  }

  // Parses an ASS file and shows it. Returns the number of drawable events.
  load(text) {
    this.clear()
    this.script = parseAss(text)
    this.tick()
    return this.script.events.length
  }

  // Hides the current subtitles but keeps the renderer for the next track.
  unload() {
    this.clear()
    this.script = null
  }

  destroy() {
    this.stopLoop()
    this.tickEvents.forEach(name => this.video.removeEventListener(name, this.onTick))
    ;['play', 'playing'].forEach(name => this.video.removeEventListener(name, this.onPlay))
    ;['pause', 'ended'].forEach(name => this.video.removeEventListener(name, this.onPause))
    ;['loadedmetadata', 'resize'].forEach(name => this.video.removeEventListener(name, this.onGeometry))
    document.removeEventListener('fullscreenchange', this.onGeometry)
    if (this.observer) this.observer.disconnect()
    else window.removeEventListener('resize', this.onGeometry)
    this.clear()
    this.layer.remove()
    this.script = null
  }

  startLoop() {
    if (this.raf) return
    const step = () => {
      this.tick()
      this.raf = window.requestAnimationFrame(step)
    }
    this.raf = window.requestAnimationFrame(step)
  }

  stopLoop() {
    if (this.raf) window.cancelAnimationFrame(this.raf)
    this.raf = 0
  }

  clear() {
    this.nodes.forEach(node => node.el.remove())
    this.nodes.clear()
  }

  // Visible video area inside the <video> element (object-fit: contain)
  updateGeometry() {
    const v = this.video
    const cw = v.clientWidth
    const ch = v.clientHeight
    if (!cw || !ch) return
    let w = cw
    let h = ch
    if (v.videoWidth && v.videoHeight) {
      const ratio = v.videoWidth / v.videoHeight
      if (cw / ch > ratio) w = ch * ratio
      else h = cw / ratio
    }
    const left = v.offsetLeft + (cw - w) / 2
    const top = v.offsetTop + (ch - h) / 2
    const g = this.geo
    const same = g && ['left', 'top', 'w', 'h'].every(key => Math.abs(g[key] - { left, top, w, h }[key]) < 0.5)
    if (same) return
    this.geo = { left, top, w, h }
    Object.assign(this.layer.style, {
      left: `${left}px`,
      top: `${top}px`,
      width: `${w}px`,
      height: `${h}px`,
    })
    // sizes depend on the geometry: rebuild everything
    this.clear()
    this.tick()
  }

  scales() {
    const { w, h } = this.geo
    const sx = w / this.script.playResX
    const sy = h / this.script.playResY
    const border = this.script.scaledBorder ? sy : h / (this.video.videoHeight || this.script.playResY)
    return { sx, sy, border }
  }

  tick() {
    if (!this.script || !this.geo) return
    const t = this.video.currentTime
    const { events } = this.script

    const active = []
    for (let i = 0; i < events.length; i++) {
      const e = events[i]
      if (e.start > t) break
      if (e.end > t) active.push(i)
    }

    let changed = false
    const activeSet = new Set(active)
    this.nodes.forEach((node, idx) => {
      if (!activeSet.has(idx)) {
        node.el.remove()
        this.nodes.delete(idx)
        changed = true
      }
    })
    active.forEach(idx => {
      if (this.nodes.has(idx)) return
      const node = this.createNode(idx, events[idx])
      this.nodes.set(idx, node)
      this.layer.appendChild(node.el)
      changed = true
    })
    if (changed) this.layout()
    this.nodes.forEach(node => this.animate(node, t))
  }

  createNode(idx, ev) {
    const { sx, sy, border } = this.scales()
    const el = document.createElement('div')
    const col = (ev.an - 1) % 3 // 0 left, 1 center, 2 right
    const row = ev.an >= 7 ? 0 : ev.an >= 4 ? 1 : 2 // 0 top, 1 middle, 2 bottom

    Object.assign(el.style, {
      position: 'absolute',
      margin: '0',
      padding: '0',
      fontSize: '0px',
      lineHeight: '0px',
      textAlign: ['left', 'center', 'right'][col],
      whiteSpace: ev.wrap === 2 ? 'pre' : 'pre-wrap',
      overflowWrap: 'break-word',
      zIndex: String(ev.layer),
    })

    if (ev.pos) {
      el.style.width = 'max-content'
      el.style.maxWidth = `${this.geo.w}px`
      el.style.transform = `translate(${[0, -50, -100][col]}%, ${[0, -50, -100][row]}%)`
    } else {
      el.style.left = `${ev.marginL * sx}px`
      el.style.width = `${Math.max(0, this.geo.w - (ev.marginL + ev.marginR) * sx)}px`
    }

    const karaoke = []
    ev.segments.forEach(seg => {
      const span = document.createElement('span')
      span.textContent = seg.text
      styleSpan(span, seg, ev.style, sx, sy, border)
      el.appendChild(span)
      if (seg.k) karaoke.push({ span, seg, state: { started: null } })
    })

    return { idx, ev, el, karaoke, state: { placed: false } }
  }

  // Stacks lines that share a layer and alignment so they do not overlap (as libass does)
  layout() {
    const { sy } = this.scales()
    const H = this.geo.h
    const groups = new Map()
    this.nodes.forEach(node => {
      if (node.ev.pos) return
      const key = `${node.ev.layer}:${node.ev.an}`
      if (!groups.has(key)) groups.set(key, [])
      groups.get(key).push(node)
    })

    groups.forEach(list => {
      list.sort((a, b) => a.idx - b.idx)
      const { an } = list[0].ev
      let edge = null
      list.forEach(node => {
        const h = node.el.offsetHeight
        const mv = node.ev.marginV * sy
        let top
        if (an >= 7) {
          top = edge === null ? mv : Math.max(mv, edge)
          edge = top + h
        } else if (an >= 4) {
          const base = (H - h) / 2
          top = edge === null ? base : Math.max(base, edge)
          edge = top + h
        } else {
          const base = H - mv - h
          top = edge === null ? base : Math.min(base, edge - h)
          edge = top
        }
        const { el } = node
        el.style.top = `${Math.max(0, top)}px`
      })
    })
  }

  animate(node, t) {
    const { ev, el, state: placement } = node
    const elapsed = (t - ev.start) * 1000
    const duration = (ev.end - ev.start) * 1000

    if (ev.fad) {
      const [fadeIn, fadeOut] = ev.fad
      let opacity = 1
      if (fadeIn > 0 && elapsed < fadeIn) opacity = elapsed / fadeIn
      if (fadeOut > 0 && elapsed > duration - fadeOut) opacity = Math.min(opacity, (duration - elapsed) / fadeOut)
      el.style.opacity = String(clamp(opacity, 0, 1))
    }

    if (ev.pos && (ev.move || !placement.placed)) {
      const { sx, sy } = this.scales()
      let [x, y] = ev.pos
      if (ev.move) {
        const { x1, y1, x2, y2, t1 } = ev.move
        const t2 = ev.move.t2 || duration
        const p = t2 > t1 ? clamp((elapsed - t1) / (t2 - t1), 0, 1) : 1
        x = x1 + (x2 - x1) * p
        y = y1 + (y2 - y1) * p
      }
      el.style.left = `${x * sx}px`
      el.style.top = `${y * sy}px`
      placement.placed = true
    }

    node.karaoke.forEach(item => {
      const { span, seg, state } = item
      const started = elapsed >= seg.k.start
      if (state.started === started) return
      state.started = started
      span.style.color = started ? rgba(seg.c1, seg.a1) : rgba(seg.c2, seg.a2)
    })
  }
}

import { parseAss } from './parser'
import AssRenderer from './renderer'
import JassubRenderer, { jassubSupported, loadJassub } from './jassub'
import { resolveFonts } from './torfonts'

const log = (...args) => {
  // eslint-disable-next-line no-console
  console.info('[ASS]', ...args)
}

const TV_AGENT = /SMART-TV|SmartTV|Tizen|Web0S|webOS|Android TV|AFT[A-Z]|BRAVIA|HbbTV|NetCast|VIDAA/i

// 'auto': libass where it can run (not on TVs), otherwise the simple renderer
export const chooseRenderer = mode => {
  if (mode === 'simple' || !jassubSupported()) return 'simple'
  if (mode === 'full') return 'full'
  return TV_AGENT.test(navigator.userAgent) ? 'simple' : 'full'
}

// Shows ASS subtitles over a video with the best available renderer:
// libass (JASSUB) or the simple DOM renderer, with fonts from the torrent.
//
// options: { mode: 'auto' | 'full' | 'simple', fonts: boolean, fontFiles: [{ path, url, source }] }
export default class AssPlayer {
  constructor(video, container, options = {}) {
    this.video = video
    this.container = container
    this.options = options
    this.simple = null
    this.jassub = null
    this.token = 0
    this.kind = null
  }

  // fontSource: where the fonts of this track should come from first when the same font is in several places
  // ('external': files of the torrent, 'embedded': attached to the video file)
  async load(text, fontSource) {
    this.token += 1
    const { token } = this
    this.unload()

    const script = parseAss(text)
    let faces = []
    if (this.options.fonts !== false && this.options.fontFiles?.length) {
      try {
        faces = await resolveFonts(this.options.fontFiles, script.fonts, fontSource)
      } catch (error) {
        log('fonts from the torrent are not available:', error.message)
      }
    }
    if (token !== this.token) return

    if (chooseRenderer(this.options.mode) === 'full') {
      try {
        const JASSUB = await loadJassub()
        if (token !== this.token) return
        const urls = [...new Set(faces.map(face => face.url))]
        this.jassub = new JassubRenderer(JASSUB, this.video, text, urls, error => {
          if (token !== this.token) return
          log('libass failed, using the simple renderer:', error.message)
          this.showSimple(script, faces, token)
        })
        this.kind = 'full'
        log(`libass (JASSUB), ${urls.length} font file(s) from the torrent`)
        return
      } catch (error) {
        log('libass is not available, using the simple renderer:', error.message)
        if (this.jassub) {
          this.jassub.destroy()
          this.jassub = null
        }
      }
    }
    await this.showSimple(script, faces, token)
  }

  async showSimple(script, faces, token) {
    if (this.jassub) {
      this.jassub.destroy()
      this.jassub = null
    }
    if (!this.simple) this.simple = new AssRenderer(this.video, this.container)
    await this.simple.setFonts(faces)
    if (token !== this.token) return
    this.simple.loadScript(script)
    this.kind = 'simple'
    log(`simple renderer, ${faces.length} font face(s) from the torrent`)
  }

  unload() {
    if (this.jassub) {
      this.jassub.destroy()
      this.jassub = null
    }
    if (this.simple) this.simple.unload()
    this.kind = null
  }

  destroy() {
    this.token += 1
    this.unload()
    if (this.simple) {
      this.simple.destroy()
      this.simple = null
    }
  }
}

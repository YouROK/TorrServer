import AssRenderer from './renderer'
import JassubRenderer, { jassubSupported, loadJassub } from './jassub'

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
// libass (JASSUB) or the simple DOM renderer.
//
// options: { mode: 'auto' | 'full' | 'simple' }
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

  async load(text) {
    this.token += 1
    const { token } = this
    this.unload()

    if (chooseRenderer(this.options.mode) === 'full') {
      try {
        const JASSUB = await loadJassub()
        if (token !== this.token) return
        this.jassub = new JassubRenderer(JASSUB, this.video, text, error => {
          if (token !== this.token) return
          log('libass failed, using the simple renderer:', error.message)
          this.showSimple(text, token)
        })
        this.kind = 'full'
        log('libass (JASSUB)')
        return
      } catch (error) {
        log('libass is not available, using the simple renderer:', error.message)
        if (this.jassub) {
          this.jassub.destroy()
          this.jassub = null
        }
      }
    }
    this.showSimple(text, token)
  }

  showSimple(text, token) {
    if (this.jassub) {
      this.jassub.destroy()
      this.jassub = null
    }
    if (token !== this.token) return
    if (!this.simple) this.simple = new AssRenderer(this.video, this.container)
    this.simple.load(text)
    this.kind = 'simple'
    log('simple renderer')
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

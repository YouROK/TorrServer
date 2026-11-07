// Full ASS rendering with libass compiled to WebAssembly (JASSUB, https://github.com/ThaUnknown/jassub).
// The library files are not part of the bundle: scripts/copy-jassub.js copies them from node_modules
// to public/jassub, and they are loaded only when an ASS track is shown.

const LIB_DIR = 'jassub'

const assetUrl = name => new URL(`${process.env.PUBLIC_URL || '.'}/${LIB_DIR}/${name}`, document.baseURI).href

export const jassubSupported = () => typeof WebAssembly === 'object' && typeof Worker !== 'undefined'

let loading = null

// Loads the UMD build of JASSUB once; resolves to the JASSUB constructor
export const loadJassub = () => {
  if (window.JASSUB) return Promise.resolve(window.JASSUB)
  if (!loading) {
    loading = new Promise((resolve, reject) => {
      const script = document.createElement('script')
      script.src = assetUrl('jassub.umd.js')
      script.async = true
      script.onload = () => (window.JASSUB ? resolve(window.JASSUB) : reject(new Error('JASSUB is not defined')))
      script.onerror = () => {
        loading = null
        script.remove()
        reject(new Error('jassub.umd.js was not loaded (is the library installed? see web/scripts/copy-jassub.js)'))
      }
      document.head.appendChild(script)
    })
  }
  return loading
}

export default class JassubRenderer {
  constructor(JASSUB, video, text, onError) {
    this.destroyed = false
    this.instance = new JASSUB({
      video,
      subContent: text,
      workerUrl: assetUrl('jassub-worker.js'),
      wasmUrl: assetUrl('jassub-worker.wasm'),
      modernWasmUrl: assetUrl('jassub-worker-modern.wasm'),
      availableFonts: { 'liberation sans': assetUrl('default.woff2') },
      defaultFont: 'liberation sans',
    })

    const report = error => {
      if (!this.destroyed) onError(error)
    }
    const { worker } = this.instance
    if (worker && worker.addEventListener) {
      worker.addEventListener('error', event => report(new Error(event.message || 'JASSUB worker failed')))
    }

    // the video changes its size in fullscreen: make JASSUB measure it again once the layout settles
    this.timers = []
    this.onFullscreen = () => {
      this.timers.forEach(id => window.clearTimeout(id))
      this.timers = [50, 300].map(delay =>
        window.setTimeout(() => {
          if (!this.destroyed && this.instance && typeof this.instance.resize === 'function') {
            try {
              this.instance.resize()
            } catch {
              // not supported by this version
            }
          }
        }, delay),
      )
    }
    document.addEventListener('fullscreenchange', this.onFullscreen)
  }

  destroy() {
    this.destroyed = true
    document.removeEventListener('fullscreenchange', this.onFullscreen)
    this.timers.forEach(id => window.clearTimeout(id))
    try {
      this.instance.destroy()
    } catch {
      // already gone
    }
    this.instance = null
  }
}

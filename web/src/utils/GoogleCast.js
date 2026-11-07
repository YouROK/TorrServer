import { externalLink } from './MediaBase'

const SDK_URL = 'https://www.gstatic.com/cv/js/sender/v1/cast_sender.js?loadCastFramework=1'
const SDK_CALLBACK = '__onGCastApiAvailable'
let pending
let context

// One SDK and CastContext per page, including when several torrent players are mounted.
export const loadCastSdk = () => {
  if (!window.isSecureContext) return Promise.reject(new Error('HTTPS'))
  if (context) return Promise.resolve(context)
  if (pending) return pending

  pending = new Promise((resolve, reject) => {
    const previous = window[SDK_CALLBACK]
    let script
    let timer
    let settled = false
    const finish = available => {
      if (settled) return
      settled = true
      window.clearTimeout(timer)
      if (window[SDK_CALLBACK] === callback) window[SDK_CALLBACK] = previous
      try {
        if (!available || !window.cast?.framework || !window.chrome?.cast) throw new Error('Unavailable')
        const api = window.chrome.cast
        context = window.cast.framework.CastContext.getInstance()
        context.setOptions({
          receiverApplicationId: api.media.DEFAULT_MEDIA_RECEIVER_APP_ID,
          autoJoinPolicy: api.AutoJoinPolicy.PAGE_SCOPED,
        })
        resolve(context)
      } catch {
        context = null
        script?.remove()
        reject(new Error('Unavailable'))
      }
    }
    const callback = available => {
      try {
        previous?.(available)
      } finally {
        finish(available)
      }
    }
    window[SDK_CALLBACK] = callback
    if (window.cast?.framework && window.chrome?.cast) {
      finish(true)
    } else {
      script = document.createElement('script')
      script.src = SDK_URL
      script.async = true
      script.onerror = () => finish(false)
      timer = window.setTimeout(() => finish(false), 15000)
      document.body.appendChild(script)
    }
  }).catch(error => {
    pending = null
    throw error
  })
  return pending
}

export const getVideoMimeType = source => {
  const extension = source.split('?')[0].split('.').pop().toLowerCase()
  return { mp4: 'video/mp4', webm: 'video/webm', ogg: 'video/ogg', ogv: 'video/ogg' }[extension] || ''
}

export const castMedia = (source, base, hls = false) => {
  let url
  try {
    const original = new URL(source, window.location.href)
    const target = base ? new URL(base) : original
    if (
      !['http:', 'https:'].includes(original.protocol) ||
      !['http:', 'https:'].includes(target.protocol) ||
      original.username ||
      original.password ||
      target.username ||
      target.password
    )
      throw new Error()
    url = externalLink(source, base)
    const host = url.hostname.toLowerCase().replace(/\.$/, '')
    if (
      !['http:', 'https:'].includes(url.protocol) ||
      host === 'localhost' ||
      host.endsWith('.localhost') ||
      host === '[::1]' ||
      host === '[::]' ||
      /^\[::ffff:7f[0-9a-f]{2}:/.test(host) ||
      host === '0.0.0.0' ||
      host.startsWith('127.')
    )
      throw new Error()
  } catch {
    throw new Error('Url')
  }
  const mime = hls ? 'application/x-mpegURL' : getVideoMimeType(url.pathname)
  if (!mime) throw new Error('Media')
  return { url: url.href, mime }
}

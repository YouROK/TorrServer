import { useEffect, useRef, useState } from 'react'

import { castMedia, loadCastSdk } from './GoogleCast'

// A page may contain many players, but only the latest explicit request may LOAD.
let latestRequest = 0
const initialState = { ready: false, available: false, busy: false, error: '', connected: false }

const useGoogleCast = ({ active, source, base, hls, title, video }) => {
  const [state, setState] = useState(initialState)
  const api = useRef(null)
  const owner = useRef(null)
  const generation = useRef(0)

  const ownsMedia = () => {
    const owned = owner.current
    const session = api.current?.context.getCurrentSession()
    const media = session?.getMediaSession()
    return Boolean(
      owned && session === owned.session && media?.mediaSessionId === owned.id && media?.media?.contentId === owned.url,
    )
  }

  useEffect(() => {
    generation.current += 1
    owner.current = null
    setState(initialState)
    if (!active) return undefined
    let disposed = false
    let detach = () => {}
    loadCastSdk()
      .then(context => {
        if (disposed) return
        const { framework } = window.cast
        const player = new framework.RemotePlayer()
        const controller = new framework.RemotePlayerController(player)
        const update = () => {
          if (disposed) return
          const owned = ownsMedia()
          const connected = owned && player.isConnected
          if (!owned) owner.current = null
          setState(previous => ({
            ...previous,
            ready: true,
            available: context.getCastState() !== framework.CastState.NO_DEVICES_AVAILABLE,
            connected,
            paused: player.isPaused,
            currentTime: Number.isFinite(player.currentTime) ? player.currentTime : 0,
            duration: Number.isFinite(player.duration) ? player.duration : 0,
            canPause: player.canPause,
            canSeek: player.canSeek,
          }))
        }
        const sessionChanged = () => {
          if (api.current?.loadingSession && context.getCurrentSession() !== api.current.loadingSession) {
            generation.current += 1
            api.current.activeRequest = null
            api.current.loadingSession = null
            setState(previous => ({ ...previous, busy: false }))
          }
          update()
        }
        api.current = { context, player, controller, update }
        context.addEventListener(framework.CastContextEventType.CAST_STATE_CHANGED, update)
        context.addEventListener(framework.CastContextEventType.SESSION_STATE_CHANGED, sessionChanged)
        controller.addEventListener(framework.RemotePlayerEventType.ANY_CHANGE, update)
        detach = () => {
          context.removeEventListener(framework.CastContextEventType.CAST_STATE_CHANGED, update)
          context.removeEventListener(framework.CastContextEventType.SESSION_STATE_CHANGED, sessionChanged)
          controller.removeEventListener(framework.RemotePlayerEventType.ANY_CHANGE, update)
        }
        update()
      })
      .catch(error => {
        if (!disposed) setState({ ...initialState, error: error.message })
      })
    return () => {
      disposed = true
      generation.current += 1
      owner.current = null
      api.current = null
      detach()
    }
  }, [active, source, base, hls]) // A changed/closed player never inherits a previous media session.

  const start = async () => {
    if (!active || state.busy) return
    if (!api.current) {
      setState(previous => ({ ...previous, error: window.isSecureContext ? 'Unavailable' : 'HTTPS' }))
      return
    }
    const currentApi = api.current
    if (currentApi.activeRequest) return
    const token = generation.current
    const request = ++latestRequest
    currentApi.activeRequest = request
    const fresh = () => token === generation.current && request === latestRequest && api.current === currentApi
    let loadingMedia = false
    try {
      const { url, mime } = castMedia(source, base, hls)
      setState(previous => ({ ...previous, busy: true, error: '' }))
      // No asynchronous media lookup before this call: keep the browser user activation.
      await currentApi.context.requestSession()
      if (!fresh()) return
      const session = currentApi.context.getCurrentSession()
      if (!session) throw new Error('Load')
      currentApi.loadingSession = session
      loadingMedia = true
      const media = new window.chrome.cast.media.MediaInfo(url, mime)
      media.metadata = new window.chrome.cast.media.GenericMediaMetadata()
      media.metadata.title = title || ''
      const load = new window.chrome.cast.media.LoadRequest(media)
      load.currentTime = Number.isFinite(video?.currentTime) ? Math.max(0, video.currentTime) : 0
      load.autoplay = !video?.paused
      await session.loadMedia(load)
      if (!fresh() || currentApi.context.getCurrentSession() !== session) return
      const accepted = session.getMediaSession()
      if (accepted?.media?.contentId !== url) throw new Error('Load')
      owner.current = { session, id: accepted.mediaSessionId, url }
      video?.pause()
      currentApi.update()
    } catch (error) {
      if (fresh()) {
        const code = typeof error === 'string' ? error : error?.code
        const canceled = code === window.chrome?.cast?.ErrorCode.CANCEL
        setState(previous => ({
          ...previous,
          error: canceled
            ? ''
            : ['Url', 'Media'].includes(error?.message)
            ? error.message
            : loadingMedia
            ? 'Load'
            : 'Session',
        }))
      }
    } finally {
      if (api.current === currentApi && currentApi.activeRequest === request) {
        currentApi.activeRequest = null
        currentApi.loadingSession = null
        setState(previous => ({ ...previous, busy: false }))
      }
    }
  }

  const toggle = () => {
    if (ownsMedia() && api.current.player.canPause) api.current.controller.playOrPause()
  }
  const seek = time => {
    if (!ownsMedia() || !api.current.player.canSeek || !Number.isFinite(time)) return
    const { player, controller } = api.current
    player.currentTime = Math.min(Math.max(0, time), player.duration || 0)
    controller.seek()
  }
  const disconnect = () => {
    if (!ownsMedia()) return
    owner.current.session.endSession(true)
    owner.current = null
    api.current.update()
  }
  return { ...state, start, toggle, seek, disconnect }
}

export default useGoogleCast

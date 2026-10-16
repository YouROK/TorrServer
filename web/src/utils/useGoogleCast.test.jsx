import ReactDOM from 'react-dom'
import { act } from 'react-dom/test-utils'

import useGoogleCast from './useGoogleCast'

jest.mock('./GoogleCast', () => ({
  ...jest.requireActual('./GoogleCast'),
  loadCastSdk: () =>
    global.window.isSecureContext
      ? Promise.resolve(global.window.cast.framework.CastContext.getInstance())
      : Promise.reject(new Error('HTTPS')),
}))

let current
let container
let context
let session
let player
let controller
let contextListeners
let remoteListeners
let mediaListeners
let resolveLoad
let rejectLoad
let props

const Harness = props => {
  current = useGoogleCast(props)
  return (
    <button type='button' onClick={current.start}>
      Cast
    </button>
  )
}
const render = async changes => {
  props = { ...props, ...changes }
  await act(async () => {
    ReactDOM.render(<Harness {...props} />, container)
  })
}
const emitRemote = () => remoteListeners.forEach(listener => listener())

beforeEach(() => {
  Object.defineProperty(window, 'isSecureContext', { configurable: true, value: true })
  container = document.createElement('div')
  document.body.appendChild(container)
  contextListeners = new Map()
  remoteListeners = new Map()
  mediaListeners = new Map()
  session = {
    addEventListener: (type, fn) => mediaListeners.set(type, fn),
    removeEventListener: type => mediaListeners.delete(type),
    getMediaSession: jest.fn(() => session.media),
    loadMedia: jest.fn(request => {
      session.request = request
      return new Promise((resolve, reject) => {
        resolveLoad = resolve
        rejectLoad = reject
      })
    }),
    endSession: jest.fn(),
  }
  context = {
    setOptions: jest.fn(),
    getCastState: () => 'AVAILABLE',
    getCurrentSession: () => session,
    requestSession: jest.fn(() => Promise.resolve()),
    addEventListener: (type, fn) => contextListeners.set(type, fn),
    removeEventListener: type => contextListeners.delete(type),
  }
  player = { isConnected: true, isPaused: false, currentTime: 12, duration: 90, canPause: true, canSeek: true }
  controller = {
    addEventListener: (type, fn) => remoteListeners.set(type, fn),
    removeEventListener: type => remoteListeners.delete(type),
    playOrPause: jest.fn(),
    seek: jest.fn(),
  }
  window.cast = {
    framework: {
      CastContext: { getInstance: () => context },
      CastContextEventType: { CAST_STATE_CHANGED: 'cast', SESSION_STATE_CHANGED: 'session' },
      SessionEventType: { MEDIA_SESSION: 'media_session' },
      CastState: { NO_DEVICES_AVAILABLE: 'NONE' },
      RemotePlayer: function RemotePlayer() {
        return player
      },
      RemotePlayerController: function RemotePlayerController() {
        return controller
      },
      RemotePlayerEventType: { ANY_CHANGE: 'remote' },
    },
  }
  window.chrome = {
    cast: {
      AutoJoinPolicy: { PAGE_SCOPED: 'page' },
      ErrorCode: { CANCEL: 'cancel' },
      media: {
        DEFAULT_MEDIA_RECEIVER_APP_ID: 'default',
        MediaInfo: function MediaInfo(url, mime) {
          this.contentId = url
          this.contentType = mime
        },
        LoadRequest: function LoadRequest(media) {
          this.media = media
        },
        GenericMediaMetadata: function GenericMediaMetadata() {},
      },
    },
  }
  props = {
    active: true,
    source: 'https://media.example/sample.mp4?play=1',
    title: 'Sample',
    video: { currentTime: 24, paused: false, pause: jest.fn() },
  }
})

afterEach(() => {
  act(() => {
    ReactDOM.unmountComponentAtNode(container)
  })
  container.remove()
  jest.useRealTimers()
})

const startLoad = async () => {
  await render()
  // The SDK is already available in this test; opening the UI must never LOAD.
  expect(session.loadMedia).not.toHaveBeenCalled()
  await act(async () => {
    current.start()
  })
}
const acceptLoad = async () => {
  session.media = { mediaSessionId: 1, media: session.request.media }
  await act(async () => resolveLoad())
}

test('explicit Cast maps selected media, title and local position; pauses only after acceptance', async () => {
  await startLoad()
  expect(context.requestSession).toHaveBeenCalledTimes(1)
  expect(session.request.media).toMatchObject({
    contentId: 'https://media.example/sample.mp4?play=1',
    contentType: 'video/mp4',
    metadata: { title: 'Sample' },
  })
  expect(session.request).toMatchObject({ currentTime: 24, autoplay: true })
  expect(props.video.pause).not.toHaveBeenCalled()
  await acceptLoad()
  expect(props.video.pause).toHaveBeenCalledTimes(1)
  expect(current.connected).toBe(true)
  expect(current.currentTime).toBe(12)
})

test('HLS uses the current source and preserves paused intent with nonfinite position', async () => {
  await render({
    hls: true,
    source: '/proxy/master.m3u8?play=1',
    base: 'https://media.example',
    video: { currentTime: NaN, paused: true, pause: jest.fn() },
  })
  await act(async () => {
    current.start()
  })
  expect(session.request).toMatchObject({
    currentTime: 0,
    autoplay: false,
    media: { contentId: 'https://media.example/proxy/master.m3u8?play=1', contentType: 'application/x-mpegURL' },
  })
})

test('canceled picker and failed LOAD preserve local playback without exposing a URL', async () => {
  context.requestSession.mockRejectedValueOnce('cancel')
  await render()
  await act(async () => {
    await current.start()
  })
  expect(session.loadMedia).not.toHaveBeenCalled()
  expect(props.video.pause).not.toHaveBeenCalled()
  expect(current.error).toBe('')
  await act(async () => {
    current.start()
  })
  await act(async () => rejectLoad(new Error('private URL')))
  expect(props.video.pause).not.toHaveBeenCalled()
  expect(current.error).toBe('Load')
  expect(current.busy).toBe(false)
})

test.each(['close', 'source', 'disconnect'])('ignores stale LOAD after %s', async change => {
  await startLoad()
  if (change === 'close') await render({ active: false })
  if (change === 'source') await render({ source: 'https://media.example/next.mp4' })
  if (change === 'disconnect') {
    const oldSession = session
    session = null
    act(() => contextListeners.get('session')())
    session = oldSession
  }
  await acceptLoad()
  expect(props.video.pause).not.toHaveBeenCalled()
  expect(current.connected).toBe(false)
})

test('SDK controls are scoped to accepted media and detach after another sender replaces it', async () => {
  await startLoad()
  current.toggle()
  current.seek(35)
  current.disconnect()
  expect(controller.playOrPause).not.toHaveBeenCalled()
  expect(session.endSession).not.toHaveBeenCalled()
  await acceptLoad()
  act(() => {
    current.toggle()
    current.seek(35)
  })
  expect(controller.playOrPause).toHaveBeenCalledTimes(1)
  expect(player.currentTime).toBe(35)
  expect(controller.seek).toHaveBeenCalledTimes(1)
  session.media = { mediaSessionId: 2, media: { contentId: 'https://media.example/other.mp4' } }
  act(emitRemote)
  expect(current.connected).toBe(false)
  current.toggle()
  current.disconnect()
  expect(controller.playOrPause).toHaveBeenCalledTimes(1)
  expect(session.endSession).not.toHaveBeenCalled()
})

test('explicit disconnect ends owned session; close only removes listeners', async () => {
  await startLoad()
  await acceptLoad()
  act(() => current.disconnect())
  expect(session.endSession).toHaveBeenCalledWith(true)
  await render({ active: false })
  expect(contextListeners.size).toBe(0)
  expect(remoteListeners.size).toBe(0)
  expect(session.endSession).toHaveBeenCalledTimes(1)
})

test('retains accepted ownership when remote connection state arrives after LOAD', async () => {
  await startLoad()
  player.isConnected = false
  await acceptLoad()
  expect(current.connected).toBe(false)
  player.isConnected = true
  act(emitRemote)
  expect(current.connected).toBe(true)
})

test('closing an owned player detaches every listener without stopping the receiver', async () => {
  await startLoad()
  await acceptLoad()
  await render({ active: false })
  expect(session.endSession).not.toHaveBeenCalled()
  expect(contextListeners.size).toBe(0)
  expect(remoteListeners.size).toBe(0)
})

test('a bad URL fails before requesting a receiver and leaves the local video playing', async () => {
  await render({ source: 'http://localhost/sample.mp4' })
  await act(async () => {
    await current.start()
  })
  expect(current.error).toBe('Url')
  expect(context.requestSession).not.toHaveBeenCalled()
  expect(props.video.pause).not.toHaveBeenCalled()
})

test('disconnect abandons a pending LOAD without letting its completion clear a newer request', async () => {
  await startLoad()
  const oldResolve = resolveLoad
  const oldSession = session
  session = null
  act(() => contextListeners.get('session')())
  expect(current.busy).toBe(false)
  session = oldSession
  await act(async () => {
    current.start()
  })
  expect(current.busy).toBe(true)
  await act(async () => oldResolve())
  expect(current.busy).toBe(true)
  expect(props.video.pause).not.toHaveBeenCalled()
  await acceptLoad()
  expect(current.busy).toBe(false)
  expect(props.video.pause).toHaveBeenCalledTimes(1)
})

test.each(['empty', 'old'])('waits for matching media_session after accepted LOAD with %s media', async previous => {
  session.media = previous === 'old' ? { mediaSessionId: 7, media: { contentId: props.source } } : null
  await startLoad()
  await act(async () => resolveLoad())
  expect(current.error).toBe('')
  expect(current.busy).toBe(true)
  expect(props.video.pause).not.toHaveBeenCalled()
  session.media = { mediaSessionId: 8, media: session.request.media }
  await act(async () => mediaListeners.get('media_session')({ mediaSession: session.media }))
  expect(current.connected).toBe(true)
  expect(current.busy).toBe(false)
  expect(props.video.pause).toHaveBeenCalledTimes(1)
  expect(mediaListeners.size).toBe(0)
})

test('media arriving before LOAD resolves cannot pause local playback until LOAD succeeds', async () => {
  await startLoad()
  session.media = { mediaSessionId: 8, media: session.request.media }
  await act(async () => mediaListeners.get('media_session')({ mediaSession: session.media }))
  expect(props.video.pause).not.toHaveBeenCalled()
  await act(async () => resolveLoad())
  expect(props.video.pause).toHaveBeenCalledTimes(1)
})

test('matching media event is accepted before getMediaSession catches up, without controlling old media', async () => {
  await startLoad()
  await act(async () => resolveLoad())
  const delivered = { mediaSessionId: 8, media: session.request.media }
  await act(async () => mediaListeners.get('media_session')({ mediaSession: delivered }))
  expect(props.video.pause).toHaveBeenCalledTimes(1)
  expect(current.connected).toBe(false)
  session.media = delivered
  act(emitRemote)
  expect(current.connected).toBe(true)
})

test.each(['close', 'disconnect', 'timeout'])('cleans up an accepted LOAD waiting for media after %s', async reason => {
  jest.useFakeTimers()
  await startLoad()
  await act(async () => resolveLoad())
  expect(current.busy).toBe(true)
  if (reason === 'close') await render({ active: false })
  if (reason === 'disconnect') {
    session = null
    await act(async () => contextListeners.get('session')())
  }
  if (reason === 'timeout') await act(async () => jest.runOnlyPendingTimers())
  expect(mediaListeners.size).toBe(0)
  expect(current.busy).toBe(false)
  expect(props.video.pause).not.toHaveBeenCalled()
  expect(current.error).toBe(reason === 'timeout' ? 'Load' : '')
})

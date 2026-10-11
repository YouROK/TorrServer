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
  session = {
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

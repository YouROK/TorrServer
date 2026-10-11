import ReactDOM from 'react-dom'
import { act, Simulate } from 'react-dom/test-utils'

import VideoPlayer from './VideoPlayer'

jest.mock('react-i18next', () => ({
  ...jest.requireActual('react-i18next'),
  useTranslation: () => ({ t: key => key }),
}))
jest.mock('utils/MediaBase', () => ({ ...jest.requireActual('utils/MediaBase'), useExternalMediaBase: () => null }))
jest.mock('utils/GoogleCast', () => ({
  ...jest.requireActual('utils/GoogleCast'),
  loadCastSdk: () =>
    global.window.isSecureContext
      ? Promise.resolve(global.window.cast.framework.CastContext.getInstance())
      : Promise.reject(new Error('HTTPS')),
}))

let container
let session
let request
let resolveLoad
let rejectLoad
let context
let remote
let controller
let pause

beforeEach(() => {
  Object.defineProperty(window, 'isSecureContext', { configurable: true, value: true })
  pause = jest.spyOn(HTMLMediaElement.prototype, 'pause').mockImplementation(() => {})
  jest.spyOn(HTMLMediaElement.prototype, 'canPlayType').mockReturnValue('probably')
  container = document.createElement('div')
  document.body.appendChild(container)
  session = {
    getMediaSession: () => session.media,
    loadMedia: jest.fn(value => {
      request = value
      return new Promise((resolve, reject) => {
        resolveLoad = resolve
        rejectLoad = reject
      })
    }),
    endSession: jest.fn(),
  }
  context = {
    getCurrentSession: () => session,
    requestSession: jest.fn(() => Promise.resolve()),
    getCastState: () => 'AVAILABLE',
    addEventListener: () => {},
    removeEventListener: () => {},
  }
  remote = { isConnected: true, isPaused: false, currentTime: 12, duration: 90, canPause: true, canSeek: true }
  controller = { addEventListener: () => {}, removeEventListener: () => {}, playOrPause: jest.fn(), seek: jest.fn() }
  window.cast = {
    framework: {
      CastContext: { getInstance: () => context },
      CastContextEventType: { CAST_STATE_CHANGED: 'cast', SESSION_STATE_CHANGED: 'session' },
      CastState: { NO_DEVICES_AVAILABLE: 'NONE' },
      RemotePlayer: function RemotePlayer() {
        return remote
      },
      RemotePlayerController: function RemotePlayerController() {
        return controller
      },
      RemotePlayerEventType: { ANY_CHANGE: 'remote' },
    },
  }
  window.chrome = {
    cast: {
      ErrorCode: { CANCEL: 'cancel' },
      media: {
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
})

afterEach(() => {
  act(() => {
    ReactDOM.unmountComponentAtNode(container)
  })
  container.remove()
  jest.restoreAllMocks()
})

const render = async () => {
  await act(async () => {
    ReactDOM.render(
      <VideoPlayer videoSrc='https://media.example/sample.mp4' title='Sample' initiallyOpen showTrigger={false} />,
      container,
    )
  })
}
const click = async label => {
  const button = document.querySelector(`button[aria-label="${label}"]`)
  expect(button).not.toBeNull()
  await act(async () => {
    Simulate.click(button)
  })
}
const accept = async () => {
  session.media = { mediaSessionId: 1, media: request.media }
  await act(async () => resolveLoad())
}

test('player Cast control sends current video and only pauses local video after successful LOAD', async () => {
  await render()
  expect(session.loadMedia).not.toHaveBeenCalled()
  await click('GoogleCast.Start')
  expect(request.media.contentId).toBe('https://media.example/sample.mp4')
  expect(pause).not.toHaveBeenCalled()
  await accept()
  expect(pause).toHaveBeenCalledTimes(1)
  expect(document.querySelector('button[aria-label="GoogleCast.Disconnect"]')).not.toBeNull()
})

test.each(['cancel', 'load'])('player preserves local playback on %s failure', async failure => {
  if (failure === 'cancel') context.requestSession.mockRejectedValueOnce('cancel')
  await render()
  await click('GoogleCast.Start')
  if (failure === 'load') await act(async () => rejectLoad(new Error('receiver failed')))
  expect(pause).not.toHaveBeenCalled()
  expect(document.querySelector('button[aria-label="GoogleCast.Disconnect"]')).toBeNull()
  expect(document.body.textContent.includes('GoogleCast.Load')).toBe(failure === 'load')
})

test('connected controls and keyboard seek use receiver state; disconnect stops only on request', async () => {
  await render()
  await click('GoogleCast.Start')
  await accept()
  await click('Pause')
  expect(controller.playOrPause).toHaveBeenCalledTimes(1)
  act(() => {
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight' }))
  })
  expect(remote.currentTime).toBe(22)
  expect(controller.seek).toHaveBeenCalledTimes(1)
  await click('GoogleCast.Disconnect')
  expect(session.endSession).toHaveBeenCalledWith(true)
})

test('insecure HTTP retains local video and explains HTTPS without a working Cast control', async () => {
  Object.defineProperty(window, 'isSecureContext', { configurable: true, value: false })
  await render()
  expect(document.querySelector('video').src).toBe('https://media.example/sample.mp4')
  expect(document.body.textContent).toContain('GoogleCast.HTTPS')
  expect(context.requestSession).not.toHaveBeenCalled()
  expect(session.loadMedia).not.toHaveBeenCalled()
})

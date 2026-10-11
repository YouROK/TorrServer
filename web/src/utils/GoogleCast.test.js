import { loadCastSdk, castMedia } from './GoogleCast'

const SDK_CALLBACK = '__onGCastApiAvailable'

beforeEach(() => {
  Object.defineProperty(window, 'isSecureContext', { configurable: true, value: true })
  jest.resetModules()
  document.body.innerHTML = ''
  delete window.cast
  delete window.chrome
  delete window[SDK_CALLBACK]
})

test('HTTP never requests the SDK', async () => {
  Object.defineProperty(window, 'isSecureContext', { configurable: true, value: false })
  await expect(loadCastSdk()).rejects.toThrow('HTTPS')
  expect(document.querySelector('script')).toBeNull()
})

test('concurrent consumers share initialization and preserve the previous callback', async () => {
  // Reset the shared loader between SDK lifecycle cases.
  // eslint-disable-next-line global-require
  const { loadCastSdk: load } = require('./GoogleCast')
  const previous = jest.fn()
  window[SDK_CALLBACK] = previous
  const first = load()
  const second = load()
  expect(document.querySelectorAll('script')).toHaveLength(1)
  const context = { setOptions: jest.fn() }
  window.cast = { framework: { CastContext: { getInstance: () => context } } }
  window.chrome = {
    cast: { media: { DEFAULT_MEDIA_RECEIVER_APP_ID: 'default' }, AutoJoinPolicy: { PAGE_SCOPED: 'page' } },
  }
  window[SDK_CALLBACK](true)
  await expect(first).resolves.toBe(context)
  await expect(second).resolves.toBe(context)
  expect(context.setOptions).toHaveBeenCalledTimes(1)
  expect(context.setOptions).toHaveBeenCalledWith({ receiverApplicationId: 'default', autoJoinPolicy: 'page' })
  expect(previous).toHaveBeenCalledWith(true)
})

test.each(['unavailable', 'error', 'timeout'])('SDK %s settles and allows a later retry', async reason => {
  jest.useFakeTimers()
  // eslint-disable-next-line global-require
  const { loadCastSdk: load } = require('./GoogleCast')
  const pending = load().catch(error => error)
  if (reason === 'unavailable') window[SDK_CALLBACK](false)
  if (reason === 'error') document.querySelector('script').dispatchEvent(new Event('error'))
  if (reason === 'timeout') jest.runOnlyPendingTimers()
  await expect(pending).resolves.toMatchObject({ message: 'Unavailable' })
  const retry = load().catch(error => error)
  window[SDK_CALLBACK](false)
  await expect(retry).resolves.toMatchObject({ message: 'Unavailable' })
  jest.useRealTimers()
})

test('media URL rewriting preserves proxy path and query with direct and HLS MIME', () => {
  expect(castMedia('https://ui.example/proxy/sample.mp4?play=1', 'https://media.example', false)).toEqual({
    url: 'https://media.example/proxy/sample.mp4?play=1',
    mime: 'video/mp4',
  })
  expect(castMedia('/proxy/master.m3u8?play=1', 'https://media.example', true)).toEqual({
    url: 'https://media.example/proxy/master.m3u8?play=1',
    mime: 'application/x-mpegURL',
  })
})

test.each([
  'ftp://media.example/sample.mp4',
  'https://user:secret@media.example/sample.mp4',
  'http://localhost/sample.mp4',
  'http://127.0.0.2/sample.mp4',
  'http://[::1]/sample.mp4',
  'http://0.0.0.0/sample.mp4',
])('rejects receiver-unreachable or unsafe URL %s', source => {
  expect(() => castMedia(source)).toThrow('Url')
})

test('rejects credentials in the external media base and unsupported containers', () => {
  expect(() => castMedia('/sample.mp4', 'https://user:secret@media.example')).toThrow('Url')
  expect(() => castMedia('https://media.example/sample.mkv')).toThrow('Media')
})

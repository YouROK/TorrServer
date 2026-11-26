import ReactDOM from 'react-dom'
import { act, Simulate } from 'react-dom/test-utils'
import Hls from 'hls.js'

import Table from './index'

jest.mock('i18n', () => ({ t: key => key }))
jest.mock('react-i18next', () => ({ useTranslation: () => ({ t: key => key }) }))
jest.mock('hls.js', () => ({ isSupported: jest.fn(() => false) }))
jest.mock('utils/GStreamer', () => ({
  useGStreamerRuntime: () => ({}),
  shouldUseGStreamerPlayer: path => path.endsWith('.mkv'),
  gstreamerMasterUrl: (hash, id) => `/gst/${hash}/${id}/master.m3u8`,
  gstreamerHeartbeatUrl: hash => `/gst/${hash}/heartbeat`,
}))
jest.mock('utils/MediaBase', () => ({ useExternalMediaBase: () => '', externalLink: link => link }))
jest.mock('style/CustomMaterialUiStyles', () => ({
  StyledDialog: ({ open, children }) => (open ? <div role='dialog'>{children}</div> : null),
}))
jest.mock('components/TorrentCard/style', () => ({
  StyledButton: ({ children, onClick }) => (
    <button type='button' onClick={onClick}>
      {children}
    </button>
  ),
}))
jest.mock('./style', () => ({
  TableStyle: ({ children }) => <table data-layout='desktop'>{children}</table>,
  ShortTableWrapper: ({ children }) => <section data-layout='mobile'>{children}</section>,
  ShortTable: ({ children }) => <article>{children}</article>,
}))

let container
let openWindow
const files = [{ id: 17, path: 'Sample.mp4', length: 100 }]
const render = (playableFileList = files) =>
  act(() => {
    ReactDOM.render(<Table hash='sample-hash' playableFileList={playableFileList} viewedFileList={[]} />, container)
  })
const layout = name => container.querySelector(`[data-layout="${name}"]`)
const button = (root, text) => Array.from(root.querySelectorAll('button')).find(el => el.textContent === text)
const click = element => act(() => Simulate.click(element))

beforeEach(() => {
  jest.useFakeTimers()
  localStorage.clear()
  container = document.createElement('div')
  document.body.appendChild(container)
  jest.spyOn(HTMLMediaElement.prototype, 'canPlayType').mockReturnValue('probably')
  jest.spyOn(HTMLMediaElement.prototype, 'load').mockImplementation(() => {})
  jest.spyOn(HTMLMediaElement.prototype, 'play').mockResolvedValue(undefined)
  jest.spyOn(HTMLMediaElement.prototype, 'pause').mockImplementation(() => {})
  openWindow = jest.spyOn(window, 'open').mockImplementation(() => {})
  document.execCommand = jest.fn(() => true)
  global.fetch = jest.fn(() => Promise.resolve({ ok: true }))
})
afterEach(() => {
  act(() => {
    ReactDOM.unmountComponentAtNode(container)
  })
  container.remove()
  jest.restoreAllMocks()
  jest.useRealTimers()
  delete document.execCommand
  delete global.fetch
})

test.each(['desktop', 'mobile'])(
  '%s MP4 uses the shared player, closes inline and preserves direct/copy/external actions',
  name => {
    localStorage.setItem('isVlcUsed', 'true')
    render()
    const root = layout(name)
    const play = button(root, 'Play')
    expect(play).toBeDefined()
    const open = button(root, 'OpenLink').closest('a')
    expect(open.target).toBe('_blank')
    expect(open.getAttribute('href')).toContain('index=17&play')
    expect(root.querySelector('a[href^="vlc://"]')).not.toBeNull()
    click(button(root, 'CopyLink'))
    expect(document.execCommand).toHaveBeenCalledWith('copy')
    click(play)
    expect(root.querySelector('[role="dialog"] video').getAttribute('src')).toContain('Sample.mp4')
    expect(root.querySelector('[role="dialog"] video').getAttribute('src')).toContain('index=17')
    expect(openWindow).not.toHaveBeenCalled()
    // The shared player's header close button returns to the same file card/table row.
    click(root.querySelector('[role="dialog"] button'))
    expect(root.querySelector('[role="dialog"]')).toBeNull()
    expect(button(root, 'Play')).toBeDefined()
    expect(button(root, 'OpenLink')).toBeDefined()
    click(button(root, 'Preload'))
    expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('index=17&preload'))
  },
)

test.each(['desktop', 'mobile'])('%s HLS still uses the shared native HLS player and heartbeat', name => {
  render([{ id: 3, path: 'Sample.mkv', length: 100 }])
  const root = layout(name)
  expect(button(root, 'Play')).toBeDefined()
  click(button(root, 'Play'))
  expect(root.querySelector('video').getAttribute('src')).toBe('/gst/sample-hash/3/master.m3u8')
  expect(Hls.isSupported).toHaveBeenCalled()
  act(() => {
    jest.advanceTimersByTime(30000)
  })
  expect(global.fetch).toHaveBeenCalledWith('/gst/sample-hash/heartbeat', { cache: 'no-store' })
  click(root.querySelector('[role="dialog"] button'))
  expect(root.querySelector('[role="dialog"]')).toBeNull()
  global.fetch.mockClear()
  act(() => {
    jest.advanceTimersByTime(30000)
  })
  expect(global.fetch).not.toHaveBeenCalled()
})

test.each(['desktop', 'mobile'])('%s unsupported media retains Open/Copy/VLC fallback', name => {
  HTMLMediaElement.prototype.canPlayType.mockReturnValue('')
  localStorage.setItem('isVlcUsed', 'true')
  render()
  const root = layout(name)
  expect(button(root, 'Play')).toBeUndefined()
  expect(button(root, 'OpenLink').closest('a').getAttribute('href')).toContain('index=17&play')
  expect(button(root, 'CopyLink')).toBeDefined()
  expect(root.querySelector('a[href^="vlc://"]')).not.toBeNull()
})

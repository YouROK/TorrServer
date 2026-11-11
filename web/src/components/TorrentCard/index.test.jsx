import ReactDOM from 'react-dom'
import { act, Simulate } from 'react-dom/test-utils'
import axios from 'axios'

import TorrentCard from './index'

jest.mock('axios')
jest.mock('i18n', () => ({ t: key => key }))
jest.mock('react-i18next', () => ({
  useTranslation: () => ({ t: key => key }),
}))
jest.mock('components/DialogTorrentDetailsContent', () => () => null)
jest.mock('components/Add/AddDialog', () => () => null)
jest.mock('style/CustomMaterialUiStyles', () => ({ StyledDialog: () => null }))
jest.mock('utils/useOnStandaloneAppOutsideClick', () => () => null)
jest.mock('utils/PlayerPreferences', () => ({
  useBooleanPlayerPreference: () => false,
}))
jest.mock('utils/GStreamer', () => ({
  useGStreamerRuntime: () => ({}),
  shouldUseGStreamerPlayer: () => false,
}))
jest.mock('utils/MediaBase', () => ({
  useExternalMediaBase: () => '',
  externalLink: link => link,
}))
jest.mock('components/VideoPlayer', () => ({ videoSrc, onClose }) => (
  <button type='button' data-player={videoSrc} onClick={onClose}>
    Close player
  </button>
))
jest.mock('./style', () => {
  const React = jest.requireActual('react')
  const Box = React.forwardRef(({ children }, ref) => (
    <div ref={ref}>
      {children}
    </div>
  ))
  return {
    TorrentCard: Box,
    TorrentCardButtons: Box,
    TorrentCardDescription: Box,
    TorrentCardPoster: Box,
    StatusIndicators: Box,
    StyledButton: React.forwardRef(({ children, ...props }, ref) => (
      <button type='button' ref={ref} {...props}>
        {children}
      </button>
    )),
  }
})
jest.mock('@material-ui/core', () => {
  const actual = jest.requireActual('@material-ui/core')
  return {
    ...actual,
    Menu: ({ open, children, onClose }) =>
      open ? (
        <div role='menu'>
          <button type='button' onClick={onClose}>
            Close menu
          </button>
          {children}
        </div>
      ) : null,
  }
})

const files = [
  { id: 17, path: 'Sample.S01E02.mp4', length: 200 },
  { id: 3, path: 'Sample.S01E01.mp4', length: 100 },
]
const torrent = { hash: 'hash-a', title: 'Sample', file_stats: files }
const deferred = () => {
  let resolve
  let reject
  const promise = new Promise((yes, no) => {
    resolve = yes
    reject = no
  })
  return { promise, resolve, reject }
}
let container
const render = (value = torrent) =>
  act(() => {
    ReactDOM.render(<TorrentCard torrent={value} />, container)
  })
const click = text =>
  act(() => Simulate.click(Array.from(container.querySelectorAll('button')).find(el => el.textContent === text)))
const items = () => Array.from(container.querySelectorAll('[role="menuitem"]'))
const flush = async () =>
  act(async () => {
    await Promise.resolve()
  })

beforeEach(() => {
  container = document.createElement('div')
  document.body.appendChild(container)
  axios.post.mockReset()
  axios.post.mockResolvedValue({ data: [] })
})
afterEach(() => {
  act(() => {
    ReactDOM.unmountComponentAtNode(container)
  })
  container.remove()
})

test('loads viewed IDs only on open, keeps file order and uses a compact accessible mark', async () => {
  axios.post.mockResolvedValue({ data: [{ file_index: 3 }] })
  render()
  expect(axios.post).not.toHaveBeenCalled()
  click('Play')
  await flush()
  expect(axios.post).toHaveBeenCalledWith(expect.stringContaining('/viewed'), {
    action: 'list',
    hash: 'hash-a',
  })
  expect(items().map(el => el.textContent)).toEqual(['S01E02 - Sample200 B', 'S01E01 - Sample100 B'])
  expect(items()[0].querySelector('[aria-label="Viewed"]')).toBeNull()
  expect(items()[1].querySelector('[aria-label="Viewed"]').getAttribute('aria-hidden')).toBe('false')
  expect(container.textContent).not.toContain('Viewed')
  act(() => Simulate.click(items()[1]))
  expect(container.querySelector('[data-player]').getAttribute('data-player')).toContain('index=3')
  click('Close player')
  axios.post.mockResolvedValue({ data: [{ file_index: 17 }] })
  click('Play')
  await flush()
  expect(axios.post).toHaveBeenCalledTimes(2)
  expect(items()[0].querySelector('[aria-label="Viewed"]')).not.toBeNull()
  expect(items()[1].querySelector('[aria-label="Viewed"]')).toBeNull()
})

test('loading and failures remain unknown, do not block playback, and retry on reopen', async () => {
  const request = deferred()
  axios.post.mockReturnValueOnce(request.promise)
  render()
  click('Play')
  expect(container.querySelector('[role="status"]')).not.toBeNull()
  await act(async () => {
    request.reject(new Error('offline'))
  })
  expect(container.querySelector('[role="status"]')).not.toBeNull()
  expect(items()).toHaveLength(2)
  act(() => Simulate.click(items()[0]))
  expect(container.querySelector('[data-player]')).not.toBeNull()
  click('Close player')
  click('Play')
  await flush()
  expect(container.querySelector('[role="status"]')).toBeNull()
  expect(container.querySelector('[aria-label="Viewed"]')).toBeNull()
})

test('ignores stale responses on close/reopen and hash changes', async () => {
  const old = deferred()
  const current = deferred()
  axios.post.mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
  render()
  click('Play')
  click('Close menu')
  click('Play')
  await act(async () => {
    current.resolve({ data: [{ file_index: 17 }] })
  })
  await act(async () => {
    old.resolve({ data: [{ file_index: 3 }] })
  })
  expect(items()[0].querySelector('[aria-label="Viewed"]')).not.toBeNull()
  const next = deferred()
  axios.post.mockReturnValueOnce(next.promise)
  render({ ...torrent, hash: 'hash-b' })
  expect(container.querySelector('[aria-label="Viewed"]')).toBeNull()
  await act(async () => {
    next.resolve({ data: [] })
  })
  expect(container.querySelector('[aria-label="Viewed"]')).toBeNull()
})

test('a pending response for the previous hash cannot mark the new torrent', async () => {
  const previous = deferred()
  const next = deferred()
  axios.post.mockReturnValueOnce(previous.promise).mockReturnValueOnce(next.promise)
  render()
  click('Play')
  render({ ...torrent, hash: 'hash-b' })
  await act(async () => {
    previous.resolve({ data: [{ file_index: 3 }] })
  })
  expect(container.querySelector('[aria-label="Viewed"]')).toBeNull()
  expect(container.querySelector('[role="status"]')).not.toBeNull()
  await act(async () => {
    next.resolve({ data: [{ file_index: 17 }] })
  })
  expect(items()[0].querySelector('[aria-label="Viewed"]')).not.toBeNull()
})

test.each([null, []])('accepts an empty history response: %p', async data => {
  axios.post.mockResolvedValue({ data })
  render()
  click('Play')
  await flush()
  expect(container.querySelector('[role="status"]')).toBeNull()
  expect(container.querySelector('[aria-label="Viewed"]')).toBeNull()
})

test.each([{}, [null], [{ file_index: '3' }]])('keeps malformed history unknown: %p', async data => {
  axios.post.mockResolvedValue({ data })
  render()
  click('Play')
  await flush()
  expect(container.querySelector('[role="status"]').getAttribute('aria-label')).toContain('Error')
  expect(items()).toHaveLength(2)
})

test('pending history is harmless after unmount', async () => {
  const request = deferred()
  axios.post.mockReturnValueOnce(request.promise)
  render()
  click('Play')
  act(() => {
    ReactDOM.unmountComponentAtNode(container)
  })
  const error = jest.spyOn(console, 'error').mockImplementation(() => {})
  await act(async () => {
    request.resolve({ data: [{ file_index: 3 }] })
  })
  expect(error).not.toHaveBeenCalled()
  error.mockRestore()
})

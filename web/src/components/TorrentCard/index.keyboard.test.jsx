import ReactDOM from 'react-dom'
import { act, Simulate } from 'react-dom/test-utils'
import axios from 'axios'

import TorrentCard from './index'

jest.mock('axios')
jest.mock('i18n', () => ({ t: key => key }))
jest.mock('react-i18next', () => ({ useTranslation: () => ({ t: key => key }) }))
jest.mock('components/DialogTorrentDetailsContent', () => () => null)
jest.mock('components/Add/AddDialog', () => () => null)
jest.mock('style/CustomMaterialUiStyles', () => ({ StyledDialog: () => null }))
jest.mock('utils/useOnStandaloneAppOutsideClick', () => () => null)
jest.mock('utils/PlayerPreferences', () => ({ useBooleanPlayerPreference: () => false }))
jest.mock('utils/GStreamer', () => ({
  useGStreamerRuntime: () => ({}),
  shouldUseGStreamerPlayer: () => false,
}))
jest.mock('utils/MediaBase', () => ({ useExternalMediaBase: () => '', externalLink: link => link }))
jest.mock('components/VideoPlayer', () => ({ videoSrc }) => <div data-player={videoSrc} />)
jest.mock('./style', () => {
  const React = jest.requireActual('react')
  const Box = ({ children }) => <div>{children}</div>
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

// Keep the actual MUI Menu/Popover/MenuList/MenuItem/ButtonBase and their keyboard handlers.
const torrent = {
  hash: 'sample-hash',
  title: 'Sample',
  file_stats: [
    { id: 17, path: 'Sample.S01E02.mp4', length: 200 },
    { id: 3, path: 'Sample.S01E01.mp4', length: 100 },
  ],
}
let container
const keyDown = key =>
  act(() => {
    document.activeElement.dispatchEvent(new KeyboardEvent('keydown', { key, bubbles: true, cancelable: true }))
  })

beforeEach(() => {
  jest.useFakeTimers()
  container = document.createElement('div')
  document.body.appendChild(container)
  // jsdom has no layout; supply a nonzero anchor rectangle to the actual Popover.
  jest.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
    top: 0,
    left: 0,
    bottom: 20,
    right: 100,
    width: 100,
    height: 20,
  })
  axios.post.mockReset()
})
afterEach(() => {
  act(() => {
    ReactDOM.unmountComponentAtNode(container)
  })
  container.remove()
  jest.restoreAllMocks()
  jest.useRealTimers()
})

test.each(['loading', 'failure'])('real Menu keeps keyboard playback available during viewed %s', async state => {
  let rejectHistory
  axios.post.mockReturnValue(
    new Promise((resolve, reject) => {
      rejectHistory = reject
    }),
  )
  act(() => {
    ReactDOM.render(<TorrentCard torrent={torrent} />, container)
  })
  const play = Array.from(container.querySelectorAll('button')).find(element => element.textContent === 'Play')
  act(() => {
    play.focus()
    Simulate.click(play)
  })
  if (state === 'failure') {
    await act(async () => {
      rejectHistory(new Error('offline'))
    })
  }
  act(() => {
    jest.advanceTimersByTime(300)
  })
  const menu = document.querySelector('[role="menu"]')
  const files = Array.from(menu.querySelectorAll('[role="menuitem"]')).filter(
    element => element.getAttribute('aria-disabled') !== 'true',
  )
  expect(menu.querySelector('[role="status"]').getAttribute('aria-label')).toContain(
    state === 'failure' ? 'Error' : 'Loading',
  )
  expect(files).toHaveLength(2)
  expect(document.activeElement).toBe(files[0])
  expect(files[0].getAttribute('tabindex')).toBe('0')
  keyDown('ArrowDown')
  expect(document.activeElement).toBe(files[1])
  keyDown('Enter')
  expect(container.querySelector('[data-player]').getAttribute('data-player')).toContain('index=3&play')
  act(() => {
    jest.advanceTimersByTime(300)
  })
  expect(document.querySelector('[role="menu"]')).toBeNull()
})

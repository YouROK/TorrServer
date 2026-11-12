import ReactDOM from 'react-dom'
import { act } from 'react-dom/test-utils'

import fitCacheMap from './fitCacheMap'

import TorrentCache from './index'

let mockWidth = 1120

jest.mock('react-i18next', () => ({ useTranslation: () => ({ t: key => key }) }))
jest.mock('react-measure', () => {
  const React = jest.requireActual('react')
  return ({ onResize, children }) => {
    React.useEffect(() => {
      onResize({ bounds: { width: mockWidth, height: 0 } })
    }, []) // eslint-disable-line react-hooks/exhaustive-deps
    return children({ measureRef: () => {} })
  }
})
jest.mock('components/App', () => {
  const React = jest.requireActual('react')
  return { DarkModeContext: React.createContext({ isDarkMode: false }) }
})
jest.mock('style/materialUISetup', () => ({ THEME_MODES: { DARK: 'dark', LIGHT: 'light' } }))
jest.mock('../customHooks', () => {
  const React = jest.requireActual('react')
  // same shape as the real hook: one entry per piece of the torrent
  return {
    useCreateCacheMap: cache => {
      const [cacheMap, setCacheMap] = React.useState([])
      React.useEffect(() => {
        const { PiecesCount, Pieces, Readers } = cache
        const map = []
        for (let i = 0; i < PiecesCount; i++) {
          const { Size, Length, Priority } = Pieces[i] || {}
          const piece = { id: i, percentage: (Size / Length) * 100 || 0, priority: Priority || 0 }
          Readers.forEach(r => {
            if (i === r.Reader) piece.isReader = true
            if (i >= r.Start && i < r.End) piece.isReaderRange = true
          })
          map.push(piece)
        }
        setCacheMap(map)
      }, [cache])
      return cacheMap
    },
  }
})
jest.mock('./style', () => ({
  SnakeWrapper: ({ children }) => <div>{children}</div>,
  ScrollNotification: ({ children }) => <div>{children}</div>,
}))

const READER_COLOR = '#000' // snakeSettings.light.default.readerColor

let container
let drawn // calls of the fake 2d context
let acceptCanvas // (canvas) => true when the fake browser is able to allocate it

const createContext = canvas => ({
  lineWidth: 0,
  fillStyle: '',
  strokeStyle: '',
  clearRect: jest.fn(),
  translate: jest.fn(() => {
    if (!acceptCanvas(canvas)) throw new DOMException('Canvas is already in error state.')
  }),
  setTransform: jest.fn(),
  fillRect: jest.fn(() => drawn.fills++),
  strokeRect: jest.fn(function strokeRect() {
    drawn.strokes.push(this.strokeStyle)
  }),
  fillText: jest.fn(),
  createLinearGradient: jest.fn(() => ({ addColorStop: jest.fn() })),
})

const makeCache = (piecesCount, extra = {}) => ({
  Capacity: 5 * 1024 ** 3,
  PiecesLength: 16384,
  PiecesCount: piecesCount,
  Pieces: {},
  Readers: [],
  ...extra,
})

const render = (cache, isMini = false) =>
  act(() => {
    ReactDOM.render(<TorrentCache cache={cache} isMini={isMini} />, container)
  })
const canvas = () => container.querySelector('canvas')
const area = c => c.width * c.height

beforeEach(() => {
  mockWidth = 1120
  drawn = { fills: 0, strokes: [] }
  acceptCanvas = () => true
  container = document.createElement('div')
  document.body.appendChild(container)
  jest.spyOn(HTMLCanvasElement.prototype, 'getContext').mockImplementation(function getContext() {
    return createContext(this)
  })
})
afterEach(() => {
  act(() => {
    ReactDOM.unmountComponentAtNode(container)
  })
  container.remove()
  jest.restoreAllMocks()
})

describe('TorrentCache canvas size', () => {
  it('draws one cell per piece for a small torrent', () => {
    render(makeCache(130))

    // 1120px / 17px per piece = 65 pieces per row, 130 pieces = 2 rows
    expect(canvas().width).toBe(1120)
    expect(canvas().height).toBe(2 * 17)
    expect(drawn.fills).toBe(130)
  })

  it('never draws more mini blocks than the torrent has pieces when the cache is bigger than the torrent', () => {
    // 6 GB cache / 16 KiB pieces would be ~384 000 blocks for a 12 000 piece torrent
    render(makeCache(12000, { Capacity: 6291456000 }), true)

    expect(canvas().height).toBeLessThan(10000)
    expect(drawn.fills).toBeLessThanOrEqual(12100)
    expect(drawn.fills).toBeGreaterThanOrEqual(12000)
  })

  it('merges pieces when the map would not fit into the allowed canvas size', () => {
    render(makeCache(100000))

    expect(canvas().height).toBeLessThanOrEqual(32767)
    expect(area(canvas())).toBeLessThanOrEqual(16 * 1024 * 1024)
    expect(drawn.fills).toBeGreaterThan(0)
    expect(drawn.fills).toBeLessThan(100000)
  })

  it('keeps the reader marker on a merged cell', () => {
    render(makeCache(100000, { Readers: [{ Reader: 50000, Start: 49990, End: 50010 }] }))

    expect(drawn.fills).toBeLessThan(100000)
    expect(drawn.strokes).toContain(READER_COLOR)
  })
})

describe('TorrentCache when the browser rejects the canvas', () => {
  it('retries with a smaller map instead of throwing when drawing fails on a big canvas', () => {
    acceptCanvas = c => area(c) <= 5 * 1024 * 1024

    expect(() => render(makeCache(100000))).not.toThrow()

    expect(canvas()).not.toBeNull()
    expect(area(canvas())).toBeLessThanOrEqual(5 * 1024 * 1024)
    expect(drawn.fills).toBeGreaterThan(0)
  })

  it('retries when the browser cannot create a 2d context at all for a big canvas', () => {
    HTMLCanvasElement.prototype.getContext.mockImplementation(function getContext() {
      return area(this) > 5 * 1024 * 1024 ? null : createContext(this)
    })

    expect(() => render(makeCache(100000))).not.toThrow()

    expect(area(canvas())).toBeLessThanOrEqual(5 * 1024 * 1024)
    expect(drawn.fills).toBeGreaterThan(0)
  })

  it('gives up shrinking at a limit instead of looping forever', () => {
    acceptCanvas = () => false

    expect(() => render(makeCache(100000))).not.toThrow()

    expect(drawn.fills).toBe(0)
  })
})

describe('fitCacheMap', () => {
  const piece = (percentage, priority = 0, extra = {}) => ({ percentage, priority, ...extra })
  const layout = { piecesInOneRow: 10, pieceSizeWithGap: 17, canvasWidth: 170 }

  it('returns the very same array when the map fits', () => {
    const source = Array.from({ length: 100 }, () => piece(0))

    expect(fitCacheMap({ source, ...layout })).toBe(source)
  })

  it('does not touch the map when there is no room for a single row', () => {
    const source = Array.from({ length: 100 }, () => piece(0))

    expect(fitCacheMap({ source, piecesInOneRow: 0, pieceSizeWithGap: 17, canvasWidth: 0 })).toBe(source)
  })

  it('merges neighbouring pieces into cells that fit the canvas', () => {
    const source = Array.from({ length: 1000000 }, () => piece(0))
    const merged = fitCacheMap({ source, ...layout })
    const height = Math.ceil(merged.length / layout.piecesInOneRow) * layout.pieceSizeWithGap

    expect(merged.length).toBeLessThan(source.length)
    expect(height).toBeLessThanOrEqual(32767)
    expect(layout.canvasWidth * height).toBeLessThanOrEqual(16 * 1024 * 1024)
  })

  it('averages the fill, keeps the highest priority and the reader flags of the merged pieces', () => {
    const source = Array.from({ length: 1000000 }, () => piece(0))
    source[0] = piece(100, 2)
    source[1] = piece(50, 5, { isReader: true })
    source[2] = piece(0, 0, { isReaderRange: true })

    const [first] = fitCacheMap({ source, ...layout })
    const perCell = Math.ceil(source.length / (Math.floor(32767 / 17) * 10))

    expect(first.percentage).toBeCloseTo(150 / perCell)
    expect(first.priority).toBe(5)
    expect(first.isReader).toBe(true)
    expect(first.isReaderRange).toBe(true)
  })

  it('gives a smaller map for a bigger shrink', () => {
    const source = Array.from({ length: 1000000 }, () => piece(0))
    const normal = fitCacheMap({ source, ...layout })
    const shrunk = fitCacheMap({ source, ...layout, shrink: 4 })

    expect(shrunk.length).toBeLessThan(normal.length)
  })
})

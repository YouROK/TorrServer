import { describe, expect, it } from 'vitest'

import { parseLibraryImportText, withImportTimestamps } from './libraryImport'

describe('parseLibraryImportText', () => {
  it('parses export JSON with torrs_hash and metadata', () => {
    const items = parseLibraryImportText(
      JSON.stringify([
        {
          hash: 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
          title: 'Show A',
          category: 'movie',
          poster: 'https://example.com/a.jpg',
          torrs_hash: 'torrs://packedtoken',
          timestamp: 1700000002,
        },
        {
          hash: 'bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb',
          title: 'Show B',
          timestamp: 1700000001,
        },
      ]),
    )

    expect(items).toHaveLength(2)
    expect(items[0]).toMatchObject({
      link: 'torrs://packedtoken',
      title: 'Show A',
      category: 'movie',
      poster: 'https://example.com/a.jpg',
      hashHint: 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
      timestamp: 1700000002,
    })
    expect(items[1].link).toContain('magnet:?xt=urn:btih:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb')
    expect(items[1].title).toBe('Show B')
    expect(items[1].timestamp).toBe(1700000001)
  })

  it('parses magnets and torrs lines, skipping blanks and comments', () => {
    const items = parseLibraryImportText(`
# header
magnet:?xt=urn:btih:cccccccccccccccccccccccccccccccccccccccc&dn=C
torrs://token-d
web+torrs://token-e

aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
`)

    expect(items).toHaveLength(4)
    expect(items[0].link.startsWith('magnet:')).toBe(true)
    expect(items[1].link).toBe('torrs://token-d')
    expect(items[2].link).toBe('torrs://token-e')
    expect(items[3].link).toBe('aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa')
  })

  it('rewrites bare-infohash web+torrs export lines to magnets', () => {
    const hash = 'e3e497e1efd8db8df93822cad554e13740f51997'
    const items = parseLibraryImportText(`web+torrs://${hash}\ntorrs://${hash.toUpperCase()}`)
    expect(items).toHaveLength(1)
    expect(items[0].link).toBe(`magnet:?xt=urn:btih:${hash}&dn=${hash}`)
    expect(items[0].hashHint).toBe(hash)
  })

  it('dedupes by hash hint', () => {
    const hash = 'dddddddddddddddddddddddddddddddddddddddd'
    const items = parseLibraryImportText(
      [`magnet:?xt=urn:btih:${hash}&dn=One`, `magnet:?xt=urn:btih:${hash}&dn=Two`].join('\n'),
    )
    expect(items).toHaveLength(1)
  })

  it('assigns descending timestamps so line order stays newest-first', () => {
    const a = 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
    const b = 'bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb'
    const items = parseLibraryImportText(`magnet:?xt=urn:btih:${a}\nmagnet:?xt=urn:btih:${b}`)
    expect(items[0].timestamp!).toBeGreaterThan(items[1].timestamp!)
  })
})

describe('withImportTimestamps', () => {
  it('keeps explicit timestamps and fills the rest newest-first', () => {
    const filled = withImportTimestamps([{ link: 'a', timestamp: 50 }, { link: 'b' }, { link: 'c' }], 100)
    expect(filled.map(i => i.timestamp)).toEqual([50, 99, 98])
  })
})

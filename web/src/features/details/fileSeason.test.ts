import { describe, expect, it } from 'vitest'

import { filterFilesForSeason } from './fileSeason'

describe('filterFilesForSeason', () => {
  it('keeps the selected season and files that have no season', () => {
    const files = [
      { path: 'Show.S01E01.mkv' },
      { path: 'Show.S01E02.mkv' },
      { path: 'Show.S02E01.mkv' },
      { path: 'extras/poster.jpg' },
    ]
    const season1 = filterFilesForSeason(files, 1, [1, 2])
    expect(season1.map(file => file.path)).toEqual(['Show.S01E01.mkv', 'Show.S01E02.mkv', 'extras/poster.jpg'])
    expect(filterFilesForSeason(files, 2, [1, 2]).map(file => file.path)).toEqual([
      'Show.S02E01.mkv',
      'extras/poster.jpg',
    ])
    expect(filterFilesForSeason(files, 1, [])).toEqual(files)
    expect(filterFilesForSeason(undefined, 1, [1])).toEqual([])
  })

  it('filters a 1000-file season pack without dropping the chosen season', () => {
    const files = Array.from({ length: 1000 }, (_, i) => ({
      path: `Show.S${String((i % 4) + 1).padStart(2, '0')}E${i}.mkv`,
    }))
    const season3 = filterFilesForSeason(files, 3, [1, 2, 3, 4])
    expect(season3).toHaveLength(250)
    expect(season3.every(file => file.path.includes('.S03E'))).toBe(true)
  })
})

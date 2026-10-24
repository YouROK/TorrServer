import { describe, expect, it } from 'vitest'

import { buildDirectoryTree, directoryTreeIsUseful, filesInSelectedFolder, flattenFolders } from './fileTree'

describe('file tree', () => {
  it('filters a 1200-file pack by nested folder and keeps special names', () => {
    const files = Array.from({ length: 1200 }, (_, i) => ({
      id: i + 1,
      path: `Show/Season ${(i % 12) + 1}/episode ${i} [1080p].mkv`,
      length: 1000 + i,
    }))
    files.push(
      { id: 2000, path: 'loose.mkv', length: 3 },
      { id: 2001, path: String.raw`Show\Season 1\фильм (2020) #1.mkv`, length: 4 },
    )
    const tree = buildDirectoryTree(files)
    expect(directoryTreeIsUseful(tree)).toBe(true)
    expect(flattenFolders(tree).some(folder => folder.label === 'Season 1' && folder.depth === 1)).toBe(true)

    const season1 = filesInSelectedFolder(files, tree, 'root/Show/Season 1')
    expect(season1).toHaveLength(101)
    expect(season1.every(file => file.path.replace(/\\/g, '/').includes('/Season 1/'))).toBe(true)
    expect(season1.some(file => file.path.includes('фильм'))).toBe(true)

    expect(filesInSelectedFolder(files, tree, 'root')).toHaveLength(files.length)
    expect(filesInSelectedFolder(files, tree, 'root/missing')).toBe(files)
  })

  it('hides the sidebar for a single release folder', () => {
    const files = [
      { id: 1, path: 'Movie/video.mkv', length: 1 },
      { id: 2, path: 'Movie/subs.srt', length: 2 },
    ]
    const tree = buildDirectoryTree(files)
    expect(directoryTreeIsUseful(tree)).toBe(false)
    expect(filesInSelectedFolder(files, tree, 'root/Movie')).toBe(files)
  })
})

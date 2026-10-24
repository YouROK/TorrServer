import { describe, expect, it } from 'vitest'

import { filesFromMetadata } from './fileMetadata'

describe('filesFromMetadata', () => {
  it('reads persisted file lists and ignores bad JSON', () => {
    expect(filesFromMetadata(undefined)).toEqual([])
    expect(filesFromMetadata('{')).toEqual([])
    expect(
      filesFromMetadata(
        JSON.stringify({
          TorrServer: {
            Files: [
              { Id: 2, Path: 'Season 1/a.mkv', Length: 10 },
              { id: 3, path: 'b.mkv', length: 4 },
            ],
          },
        }),
      ),
    ).toEqual([
      { id: 2, path: 'Season 1/a.mkv', length: 10 },
      { id: 3, path: 'b.mkv', length: 4 },
    ])
  })
})

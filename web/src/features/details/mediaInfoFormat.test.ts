import { describe, expect, it } from 'vitest'

import { formatFfpBitrate, formatFfpBytes, formatFfpDuration, fpsFromRate, groupStreams } from './mediaInfoFormat'

describe('mediaInfoFormat', () => {
  it('formats duration', () => {
    expect(formatFfpDuration(13699)).toBe('3:48:19')
    expect(formatFfpDuration('90')).toBe('1:30')
  })

  it('formats bitrate and size', () => {
    expect(formatFfpBitrate('8000000')).toBe('8 Mbps')
    expect(formatFfpBytes(String(180 * 1024 ** 3))).toMatch(/GB/)
  })

  it('parses fps and groups streams', () => {
    expect(fpsFromRate('24000/1001')).toBe('23.98')
    const groups = groupStreams({
      streams: [{ codec_type: 'video' }, { codec_type: 'audio' }, { codec_type: 'audio' }, { codec_type: 'subtitle' }],
    })
    expect(groups.video).toHaveLength(1)
    expect(groups.audio).toHaveLength(2)
    expect(groups.subtitle).toHaveLength(1)
  })
})

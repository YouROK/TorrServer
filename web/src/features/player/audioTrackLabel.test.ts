import { describe, expect, it } from 'vitest'

import { extractAudioTracks, formatAudioTrackDisplay, humanizeCodec, probeField } from './audioTrackLabel'

describe('audioTrackLabel', () => {
  it('probeField is case-insensitive', () => {
    expect(probeField({ Language: 'ru' }, 'language')).toBe('ru')
    expect(probeField({ language: 'en' }, 'Language')).toBe('en')
  })

  it('extractAudioTracks keeps only audio', () => {
    const tracks = extractAudioTracks({
      Tracks: [
        { Type: 'video', Codec: 'h264' },
        { Type: 'audio', Codec: 'ac3', Title: 'DUB' },
        { type: 'Audio', codec: 'aac' },
        { Type: 'subtitle', Codec: 'srt' },
      ],
    })
    expect(tracks).toHaveLength(2)
    expect(probeField(tracks[0], 'Title')).toBe('DUB')
  })

  it('humanizeCodec strips Gst caps params', () => {
    expect(humanizeCodec('audio/x-ac3, framed=(boolean)true, rate=(int)48000')).toBe('AC3')
    expect(humanizeCodec('audio/x-eac3, framed=(boolean)true')).toBe('E-AC3')
    expect(humanizeCodec('AC3')).toBe('AC3')
  })

  it('formats two-line display for the track picker', () => {
    const display = formatAudioTrackDisplay(
      {
        Type: 'audio',
        Title: 'DUB | HDRezka Studio | 18+',
        Language: 'ru',
        Codec: 'AC3',
        Channels: 2,
        Rate: 48000,
      },
      0,
    )
    expect(display.title).toBe('DUB | HDRezka Studio | 18+')
    expect(display.meta).toBe('RU · AC3 · 2 ch · 48 kHz')
  })

  it('falls back when Title is missing without duplicating lang in meta', () => {
    const display = formatAudioTrackDisplay({ Type: 'audio', Language: 'en', Codec: 'E-AC3', Channels: 6 }, 1)
    expect(display.title).toBe('EN')
    expect(display.meta).toBe('E-AC3 · 6 ch')
  })

  it('never shows raw caps in title or meta', () => {
    const display = formatAudioTrackDisplay(
      {
        Type: 'audio',
        Language: 'ru',
        Codec: 'audio/x-ac3, framed=(boolean)true, rate=(int)48000',
        Channels: 2,
        Rate: 48000,
      },
      0,
    )
    expect(display.title).toBe('RU')
    expect(display.meta).toBe('AC3 · 2 ch · 48 kHz')
    expect(display.meta).not.toContain('framed')
    expect(display.title).not.toContain('framed')
  })

  it('uses Audio N when no title/lang/codec', () => {
    const display = formatAudioTrackDisplay({ Type: 'audio' }, 2)
    expect(display.title).toBe('Audio 3')
    expect(display.meta).toBe('')
  })
})

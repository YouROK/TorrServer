import { useEffect, useState } from 'react'
import axios from 'axios'
import { settingsHost } from 'utils/Hosts'

// --- pure ---
// Rule: { Name, Enabled, Mode: 'replace' | 'extract' | 'capture', Pattern, Replacement, Flags }
// Rules are applied in order to a file name (no folders, no extension):
//  - replace: name.replace(new RegExp(Pattern, Flags), Replacement)
//  - extract: the group named "title" (or group 1, or the whole match) becomes the name; no match -> unchanged
//  - capture: the name is not changed, only the named groups are saved
// Named groups of a matching rule (e.g. (?<season>\d+), (?<episode>\d+)) are saved as fields and can be used
// in the name template. A rule that is disabled, invalid, not matching or would produce an empty name is skipped.

// Name template: {field}, {field:2} (zero padded to 2 digits), {field|default}, {field:2|1};
// <...> is an optional group that is dropped when any field inside it is empty.
// Fields: title (the name after the rules), name (file name without extension), season, episode (season is 1
// when only the episode is known), special (OP, ED, OVA, ...), year, resolution, ... (everything
// parse-torrent-title finds), plus the named groups captured by the rules.
// Roman numerals captured as season/episode are converted to numbers, a captured resolution gets a "p".
export const DEFAULT_FILE_NAME_TEMPLATE = '{title}< S{season:2}><E{episode:2}>< - {special}>'

// Built-in rules: read-only in the settings UI, applied before the user's own rules.
// The user can switch the whole category off or copy rules from it to the custom rules.
export const BUILTIN_FILE_NAME_RULES = [
  {
    Name: 'Remember resolution: 1080p',
    Enabled: true,
    Mode: 'capture',
    Pattern: '(?:^|[^a-z0-9])(?<resolution>\\d{3,4}p)(?![a-z0-9])',
    Replacement: '',
    Flags: 'i',
  },
  {
    Name: 'Remember resolution: 1920x1080',
    Enabled: true,
    Mode: 'capture',
    Pattern: '(?:^|[^a-z0-9])\\d{3,4}x(?<resolution>\\d{3,4})(?![a-z0-9])',
    Replacement: '',
    Flags: 'i',
  },
  {
    Name: 'Remember year',
    Enabled: true,
    Mode: 'capture',
    Pattern: '[\\s.(\\[](?<year>19\\d{2}|20[0-2]\\d)(?=[\\s.)\\]]|$)',
    Replacement: '',
    Flags: '',
  },
  {
    Name: 'Remember episode: Ep.01, Episode 1, Серия 1',
    Enabled: true,
    Mode: 'capture',
    Pattern: '(?:^|[\\s._-])(?:Episode|Ep|Серия)[\\s._-]*(?<episode>\\d{1,4})(?![a-z0-9])',
    Replacement: '',
    Flags: 'i',
  },
  {
    Name: 'Remove release group at the start',
    Enabled: true,
    Mode: 'replace',
    Pattern: '^(?:\\s*\\[[^\\]]*\\])+\\s*',
    Replacement: '',
    Flags: 'i',
  },
  {
    Name: 'Remove technical tags: [1080p] (BDRip ...) [hash]',
    Enabled: true,
    Mode: 'replace',
    Pattern:
      '\\s*[\\[(][^\\])]*?(?:\\d{3,4}p|\\d{3,4}x\\d{3,4}|x\\.?26[45]|\\b[0-9a-f]{8}\\b|\\b(?:bd|bdrip|web|webrip|web-dl|hevc|avc|flac|aac)\\b)[^\\])]*[\\])]',
    Replacement: '',
    Flags: 'gi',
  },
  {
    Name: 'Scene name: title, season, episode before .S01E01',
    Enabled: true,
    Mode: 'extract',
    Pattern: '^(?!.*\\s)(?<title>.+?)\\.S(?<season>\\d{1,2})(?:E(?<episode>\\d{1,3}))?(?:\\..*)?$',
    Replacement: '',
    Flags: 'i',
  },
  {
    Name: 'Scene name: title before the year',
    Enabled: true,
    Mode: 'extract',
    Pattern: '^(?!.*\\s)(?<title>.+?)\\.(?:19\\d{2}|20[0-2]\\d)(?:\\..*)?$',
    Replacement: '',
    Flags: '',
  },
  {
    Name: 'Scene name: title before 1080p / BluRay / WEB-DL / x264',
    Enabled: true,
    Mode: 'extract',
    Pattern: '^(?!.*\\s)(?<title>.+?)\\.(?:\\d{3,4}p|BluRay|BDRip|WEB-?DL|WEBRip|HDTV|DVDRip|x26[45]|HEVC)(?:\\..*)?$',
    Replacement: '',
    Flags: 'i',
  },
  {
    Name: 'Dots between words to spaces',
    Enabled: true,
    Mode: 'replace',
    Pattern: '([^\\s.\\d])\\.(?=[^\\s.])',
    Replacement: '$1 ',
    Flags: 'g',
  },
  {
    Name: 'Title before S01E02 (with spaces)',
    Enabled: true,
    Mode: 'extract',
    Pattern: '^(?<title>.+?)\\s+S(?<season>\\d{1,2})\\s?E(?<episode>\\d{1,3})(?![0-9]).*$',
    Replacement: '',
    Flags: 'i',
  },
  {
    Name: 'Title before 1x02',
    Enabled: true,
    Mode: 'extract',
    Pattern: '^(?<title>.+?)\\s+(?<season>\\d{1,2})x(?<episode>\\d{2,3})(?![0-9x]).*$',
    Replacement: '',
    Flags: 'i',
  },
  {
    Name: 'Season word: Season 2, Сезон 2',
    Enabled: true,
    Mode: 'replace',
    Pattern: '(?:^|\\s)(?:Season|Сезон)[\\s._-]*(?<season>\\d{1,2})(?=\\s|$)',
    Replacement: ' ',
    Flags: 'i',
  },
  {
    Name: 'Episode word: Episode 5, Серия 5',
    Enabled: true,
    Mode: 'replace',
    Pattern: '(?:^|\\s)(?:Episode|Ep|Серия)[\\s._-]*(?<episode>\\d{1,4})(?=\\s|$)',
    Replacement: ' ',
    Flags: 'i',
  },
  {
    Name: 'Trailing season and episode: - S01E01',
    Enabled: true,
    Mode: 'replace',
    Pattern: '\\s+[-–—]\\s+S(?<season>\\d{1,2})E(?<episode>\\d{1,3})\\s*$',
    Replacement: '',
    Flags: 'i',
  },
  {
    Name: 'Trailing special: - ED1, - OP2, - OVA',
    Enabled: true,
    Mode: 'replace',
    Pattern:
      '\\s+[-–—]\\s+(?<special>(?:NC)?(?:OP|ED)\\d{0,2}[a-z]?|(?:PV|CM|SP|OVA|ONA|OAD|OAV|Special|Trailer|Menu|Extra)\\s?\\d{0,2})$',
    Replacement: '',
    Flags: 'i',
  },
  {
    Name: 'Trailing episode: - 02',
    Enabled: true,
    Mode: 'replace',
    Pattern: '\\s+[-–—]\\s+(?<episode>\\d{1,4})(?:v\\d)?\\s*$',
    Replacement: '',
    Flags: 'i',
  },
  {
    Name: 'Trailing episode in brackets: [09]',
    Enabled: true,
    Mode: 'replace',
    Pattern: '\\s*\\[(?<episode>\\d{1,4})(?:v\\d)?\\]\\s*$',
    Replacement: '',
    Flags: 'i',
  },
  {
    Name: 'Trailing two-digit episode number',
    Enabled: true,
    Mode: 'replace',
    Pattern: '\\s+(?<episode>\\d{2})(?:v\\d)?\\s*$',
    Replacement: '',
    Flags: 'i',
  },
  {
    Name: 'Trailing season: S2',
    Enabled: true,
    Mode: 'replace',
    Pattern: '\\s+S(?<season>\\d{1,2})$',
    Replacement: '',
    Flags: 'i',
  },
  {
    Name: 'Trailing season: TV-02',
    Enabled: true,
    Mode: 'replace',
    Pattern: '\\s+TV[- ]?(?<season>\\d{1,2})$',
    Replacement: '',
    Flags: 'i',
  },
  {
    Name: 'Trailing season: Roman numeral (II, III, IV, VI-IX)',
    Enabled: true,
    Mode: 'replace',
    Pattern: '\\s+(?<season>II|III|IV|VI|VII|VIII|IX)$',
    Replacement: '',
    Flags: '',
  },
  {
    Name: 'Trailing season: single digit (Title 2, Title 0)',
    Enabled: true,
    Mode: 'replace',
    Pattern: '\\s+(?<season>\\d)$',
    Replacement: '',
    Flags: '',
  },
  {
    Name: 'Trailing dash',
    Enabled: true,
    Mode: 'replace',
    Pattern: '\\s+[-–—]\\s*$',
    Replacement: '',
    Flags: 'i',
  },
]

const regexCache = new Map()

const compileRule = (pattern, flags) => {
  const key = `${flags}/${pattern}`
  if (!regexCache.has(key)) {
    if (regexCache.size > 200) regexCache.clear()
    let re = null
    try {
      re = new RegExp(pattern, flags)
    } catch (e) {
      re = null
    }
    regexCache.set(key, re)
  }
  return regexCache.get(key)
}

export const getFileNameRuleError = rule => {
  if (!rule || !rule.Pattern) return ''
  try {
    // eslint-disable-next-line no-new
    new RegExp(rule.Pattern, rule.Flags || '')
    return ''
  } catch (e) {
    return e.message
  }
}

const ROMAN = { I: 1, II: 2, III: 3, IV: 4, V: 5, VI: 6, VII: 7, VIII: 8, IX: 9, X: 10 }

// season and episode are stored as plain numbers: "02" -> "2", "III" -> "3"
const normalizeField = (key, value) => {
  if (key === 'resolution') return /^\d+$/.test(value) ? `${value}p` : value
  if (key !== 'season' && key !== 'episode') return value
  if (/^\d+$/.test(value)) return String(Number(value))
  return ROMAN[value.toUpperCase()] ? String(ROMAN[value.toUpperCase()]) : value
}

// Returns { title, fields }: the name after the rules and the named groups captured on the way.
export const parseFileName = (name, rules) => {
  let title = name
  const fields = {}
  if (!Array.isArray(rules)) return { title, fields }

  rules.forEach(rule => {
    if (!rule || !rule.Enabled || !rule.Pattern) return
    const re = compileRule(rule.Pattern, rule.Flags || '')
    if (!re) return

    re.lastIndex = 0
    const match = re.exec(title)
    if (!match) return

    let next
    if (rule.Mode === 'extract') {
      next = match.groups?.title ?? match[1] ?? match[0]
    } else if (rule.Mode === 'capture') {
      next = title
    } else {
      re.lastIndex = 0
      next = title.replace(re, rule.Replacement || '')
    }

    next = (next || '').trim()
    if (!next) return

    title = next
    Object.entries(match.groups || {}).forEach(([key, value]) => {
      if (key !== 'title' && value !== undefined) fields[key] = normalizeField(key, value)
    })
  })

  return { title, fields }
}

const TEMPLATE_TOKEN = /<([^<>]*)>|\{(\w+)(?::(\d+))?(?:\|([^}]*))?\}/g
const PLACEHOLDER = /\{(\w+)(?::(\d+))?(?:\|([^}]*))?\}/g

const fieldToString = value => {
  const flat = Array.isArray(value) ? value.join('-') : value
  return flat === undefined || flat === null ? '' : String(flat).trim()
}

const fillPlaceholders = (text, values) => {
  let empty = false
  const out = text.replace(PLACEHOLDER, (_, key, pad, fallback) => {
    let value = fieldToString(values[key])
    if (value === '' && fallback !== undefined) value = fallback
    if (value === '') {
      empty = true
    } else if (pad && /^\d+$/.test(value)) {
      value = String(Number(value)).padStart(Number(pad), '0')
    }
    return value
  })
  return { out, empty }
}

export const renderFileNameTemplate = (template, values) =>
  template
    .replace(TEMPLATE_TOKEN, (token, group) => {
      if (group !== undefined) {
        const { out, empty } = fillPlaceholders(group, values)
        return empty ? '' : out
      }
      return fillPlaceholders(token, values).out
    })
    .replace(/\s+/g, ' ')
    .trim()

// Rules that are actually applied: built-in first, then custom. Both categories are on unless switched off
// (the server stores the "disabled" flags, so a client that knows nothing about them leaves both categories on).
export const getEffectiveFileNameRules = settings => [
  ...(settings?.FileNameBuiltinRulesDisabled ? [] : BUILTIN_FILE_NAME_RULES),
  ...(settings?.FileNameCustomRulesDisabled || !Array.isArray(settings?.FileNameRules) ? [] : settings.FileNameRules),
]

// { rules, template, asFileName, usePtt } for the server settings object (BTSets)
export const getFileNameConfig = settings => ({
  asFileName: Boolean(settings?.FileNameAsTitle),
  usePtt: Boolean(settings?.FileNameUsePtt),
  rules: getEffectiveFileNameRules(settings),
  template: (settings?.FileNameTemplate || '').trim() || DEFAULT_FILE_NAME_TEMPLATE,
})
// --- store ---

let currentConfig = getFileNameConfig()
let loadPromise = null
const listeners = new Set()

// settings: the server settings object (FileNameBuiltinRulesDisabled, FileNameCustomRulesDisabled,
// FileNameRules, FileNameTemplate, FileNameAsTitle, FileNameUsePtt)
export const setFileNameSettings = settings => {
  currentConfig = getFileNameConfig(settings)
  listeners.forEach(listener => listener(currentConfig))
}

const loadFileNameConfig = () => {
  if (!loadPromise) {
    loadPromise = axios
      .post(settingsHost(), { action: 'get' })
      .then(({ data }) => setFileNameSettings(data))
      .catch(() => {
        loadPromise = null
      })
  }
  return loadPromise
}

// Config stored on the server; loaded once per page and shared by all components.
export const useFileNameConfig = () => {
  const [config, setConfig] = useState(currentConfig)

  useEffect(() => {
    listeners.add(setConfig)
    loadFileNameConfig()
    setConfig(currentConfig)
    return () => {
      listeners.delete(setConfig)
    }
  }, [])

  return config
}

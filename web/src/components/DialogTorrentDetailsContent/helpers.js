import ptt from 'parse-torrent-title'
import { DEFAULT_FILE_NAME_TEMPLATE, parseFileName, renderFileNameTemplate } from 'utils/FileNameRules'

const getExt = filename => {
  const ext = filename.split('.').pop()
  if (ext === filename) return ''
  return ext.toLowerCase()
}
const playableExtList = [
  // video
  '3g2',
  '3gp',
  'aaf',
  'asf',
  'avchd',
  'avi',
  'drc',
  'dv',
  'flv',
  'iso',
  'm2ts',
  'm2v',
  'm4p',
  'm4v',
  'mk3d',
  'mkv',
  'mng',
  'mov',
  'mp2',
  'mp4',
  'mpe',
  'mpeg',
  'mpg',
  'mpv',
  'mts',
  'mxf',
  'nsv',
  'ogv',
  'qt',
  'rm',
  'rmvb',
  'roq',
  'svi',
  'ts',
  'vob',
  'webm',
  'wmv',
  'yuv',
  // audio
  'aac',
  'ac3',
  'aiff',
  'ape',
  'au',
  'dff',
  'dsf',
  'flac',
  'gsm',
  'it',
  'm3u',
  'm4a',
  'mid',
  'mod',
  'mp3',
  'mpa',
  'mpga',
  'oga',
  'ogg',
  'opus',
  'pls',
  'ra',
  's3m',
  'sid',
  'spx',
  'wav',
  'weba',
  'wma',
  'wv',
  'wvc',
  'xm',
]

export const isFilePlayable = fileName => playableExtList.includes(getExt(fileName))

const getBaseName = path => path.split('\\').pop().split('/').pop()

const stripExt = name => {
  const ext = getExt(name)
  return ext ? name.slice(0, -(ext.length + 1)) : name
}

const hasValue = value => value !== undefined && value !== null && value !== ''

const toNumber = value => (typeof value === 'string' && /^\d+$/.test(value) ? Number(value) : value)

// Parses one file. Returns { title, info }: the file name after the rules and the file's season, episode,
// resolution, year, ... Where info comes from depends on the "parser" option (config.usePtt):
//  - rules (default): from the named groups captured by the rules; a missing season is taken from the folder names
//  - legacy: from parse-torrent-title (it misreads many anime releases, e.g. 1920x1080 as season 20, episode 10)
const parseFile = (path, { rules, usePtt, asFileName } = {}) => {
  // "file name as title": the name is used as is, nothing is parsed
  if (asFileName) return { title: stripExt(getBaseName(path)), info: {} }

  const { title, fields } = parseFileName(stripExt(getBaseName(path)), rules)
  if (usePtt) return { title, info: ptt.parse(path) }

  const info = { ...fields }
  if (!hasValue(info.season)) {
    // "Show TV-02 [WEB-DL 1080p]/Show - 11.mkv": the season is written in the folder name
    path
      .split(/[\\/]/)
      .slice(0, -1)
      .reverse()
      .some(dir => {
        const { season } = parseFileName(dir, rules).fields
        if (hasValue(season)) info.season = season
        return hasValue(season)
      })
  }
  info.season = toNumber(info.season)
  info.episode = toNumber(info.episode)
  // an episode without a season is season 1
  if (hasValue(info.episode) && !hasValue(info.season)) info.season = 1
  return { title, info }
}

// Season, episode, resolution, ... of a file; the same data is used for the file list columns, the season tabs
// and the video title.
export const getFileInfo = (path, config) => parseFile(path, config).info

export const getFileInfos = (paths, config) => new Map(paths.map(path => [path, getFileInfo(path, config)]))

// Video title (the same everywhere: file list, players, playlist menu): the file name goes through the rules
// (utils/FileNameRules), then the template is filled with the title and the season/episode/etc. of the file.
// With the "file name as title" option the title is just the file name without extension.
// The legacy parser (config.usePtt) keeps the original naming: titles and the season/episode/resolution come
// from parse-torrent-title, the rules and the template are not used (the "file name as title" option still wins).
export const isLegacyNaming = config => Boolean(config?.usePtt && !config?.asFileName)

export const getFileDisplayName = (path, config = {}) => {
  const baseName = stripExt(getBaseName(path))
  if (config.asFileName) return baseName
  if (config.usePtt) return ptt.parse(path).title || baseName

  const { title, info } = parseFile(path, config)
  const values = { ...info, title, name: baseName }
  return renderFileNameTemplate(config.template || DEFAULT_FILE_NAME_TEMPLATE, values) || title || baseName
}

const countNames = names => names.reduce((counts, name) => counts.set(name, (counts.get(name) || 0) + 1), new Map())

// Video titles for all files of a torrent as Map(path -> title). Titles are unique within the list: files whose
// titles collide fall back to the plain file name, and if that still collides, to the full path.
// (Not in the legacy mode, which keeps the original naming.)
export const getFileDisplayNames = (paths, config) => {
  if (isLegacyNaming(config)) {
    // as before: the full path when the list has several files and none of them has an episode
    const parsed = paths.map(path => ptt.parse(path))
    const showPath = paths.length > 1 && !parsed.some(({ episode }) => episode)
    return new Map(paths.map((path, i) => [path, showPath ? path : parsed[i].title]))
  }

  const names = paths.map(path => getFileDisplayName(path, config))
  const counts = countNames(names)
  const fileNames = names.map((name, i) => (counts.get(name) > 1 ? stripExt(getBaseName(paths[i])) : name))
  const fileCounts = countNames(fileNames)
  return new Map(paths.map((path, i) => [path, fileCounts.get(fileNames[i]) > 1 ? stripExt(path) : fileNames[i]]))
}

import axios from 'axios'

import i18n from '../i18n'
import { torrentsHost } from './Hosts'

export function humanizeSize(size) {
  if (!size) return ''
  const i = Math.floor(Math.log(size) / Math.log(1024))
  return `${(size / Math.pow(1024, i)).toFixed(2) * 1} ${
    [i18n.t('B'), i18n.t('KB'), i18n.t('MB'), i18n.t('GB'), i18n.t('TB')][i]
  }`
}

export function humanizeSpeed(speed) {
  if (!speed) return ''
  const i = Math.floor(Math.log(speed * 8) / Math.log(1000))
  return `${((speed * 8) / Math.pow(1000, i)).toFixed(0) * 1} ${
    [i18n.t('bps'), i18n.t('kbps'), i18n.t('Mbps'), i18n.t('Gbps'), i18n.t('Tbps')][i]
  }`
}

export function getPeerString(torrent) {
  if (!torrent || !torrent.active_peers) return null
  const seeders = typeof torrent.connected_seeders !== 'undefined' ? torrent.connected_seeders : 0
  return `${torrent.active_peers} / ${torrent.total_peers} · ${seeders}`
}

export const shortenText = (text, sympolAmount) =>
  text ? text.slice(0, sympolAmount) + (text.length > sympolAmount ? '…' : '') : ''

export const removeRedundantCharacters = string => {
  let newString = string
  const brackets = [
    ['(', ')'],
    ['[', ']'],
    ['{', '}'],
  ]

  brackets.forEach(el => {
    const leftBracketRegexFormula = `\\${el[0]}`
    const leftBracketRegex = new RegExp(leftBracketRegexFormula, 'g')
    const leftBracketAmount = [...newString.matchAll(leftBracketRegex)].length
    const rightBracketRegexFormula = `\\${el[1]}`
    const rightBracketRegex = new RegExp(rightBracketRegexFormula, 'g')
    const rightBracketAmount = [...newString.matchAll(rightBracketRegex)].length

    if (leftBracketAmount !== rightBracketAmount) {
      const removeFormula = `(\\${el[0]})(?!.*\\1).*`
      const removeRegex = new RegExp(removeFormula, 'g')
      newString = newString.replace(removeRegex, '')
    }
  })

  const hasThreeDotsAtTheEnd = !!newString.match(/\.{3}$/g)

  const trimmedString = newString.replace(/[\\.| ]+$/g, '').trim()

  return hasThreeDotsAtTheEnd ? `${trimmedString}..` : trimmedString
}

export const getTorrents = async () => {
  try {
    const { data } = await axios.post(torrentsHost(), { action: 'list' })
    return data
  } catch (error) {
    throw new Error(null)
  }
}

export const detectStandaloneApp = () => {
  if (typeof window === 'undefined') return false

  const matchMedia = window.matchMedia?.bind(window)
  const isDisplayModeStandalone = mode => {
    try {
      return !!matchMedia && matchMedia(mode).matches
    } catch {
      return false
    }
  }

  const byDisplayMode =
    isDisplayModeStandalone('(display-mode: standalone)') ||
    isDisplayModeStandalone('screen and (display-mode: standalone)')

  return byDisplayMode || window.navigator?.standalone === true
}

export const isStandaloneApp = detectStandaloneApp()

const detectApplePlatform = () => {
  if (typeof window === 'undefined' || typeof navigator === 'undefined') {
    return { isMac: false, isIOS: false }
  }

  const userAgent = navigator.userAgent || ''
  const platform = navigator.userAgentData?.platform || ''

  const isMac = userAgent.includes('Macintosh') || (platform && platform.toLowerCase().includes('mac'))

  const isIOS = /iPad|iPhone|iPod/.test(userAgent) || (userAgent.includes('Macintosh') && navigator.maxTouchPoints > 1)

  return { isMac, isIOS }
}

export const isAppleDevice = () => {
  const { isMac, isIOS } = detectApplePlatform()
  return isMac || isIOS
}

export const isMacOS = () => {
  const { isMac, isIOS } = detectApplePlatform()
  return isMac && !isIOS
}

export const isDesktop = () => {
  if (typeof window === 'undefined' || typeof navigator === 'undefined') {
    return false
  }

  const userAgent = navigator.userAgent || ''

  const isMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(userAgent)
  const isTabletWithTouch = /Macintosh/i.test(userAgent) && navigator.maxTouchPoints > 1

  return !isMobile && !isTabletWithTouch
}

/**
 * Formats bytes to classic size units (B, KB, MB, GB, TB)
 * Uses binary (1024) base for conversion
 */
export function formatSizeToClassicUnits(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  const value = bytes / Math.pow(1024, i)
  return `${value.toFixed(i === 0 ? 0 : 2)} ${sizes[i]}`
}

/**
 * Parses a human-readable size string (e.g., "1.5 GiB", "500 MiB", "1.5 GCiB") to bytes
 * Supports both binary (KiB, MiB, GiB, TiB) and decimal (KB, MB, GB, TB) units
 * Also handles the server format which may include "CiB" suffix (e.g., "GCiB")
 */
export function parseSizeToBytes(sizeStr) {
  if (!sizeStr || typeof sizeStr !== 'string') return 0

  // Handle plain bytes format (e.g., "1024 B")
  if (sizeStr.trim().match(/^\d+\s*B$/i)) {
    return parseInt(sizeStr.trim().match(/^\d+/)[0], 10)
  }

  // Match number and unit - handle both "GiB" and "GCiB" formats
  // Pattern matches: "1.5 GiB", "1.5 GCiB", "500 MiB", "500 MCiB", etc.
  const match = sizeStr.trim().match(/^([\d.]+)\s*([KMGT]?)(i?B|CiB)$/i)
  if (!match) return 0

  const value = parseFloat(match[1])
  const unit = match[2].toUpperCase()
  const suffix = match[3].toUpperCase()

  if (isNaN(value)) return 0

  // Check if it's binary (iB or CiB suffix) or decimal (just B)
  const isBinary = suffix.includes('I') || suffix.includes('C')
  const base = isBinary ? 1024 : 1000
  const multipliers = { '': 1, K: 1, M: 2, G: 3, T: 4 }
  const multiplier = multipliers[unit] || 1

  return Math.round(value * Math.pow(base, multiplier))
}

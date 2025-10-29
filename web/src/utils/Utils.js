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

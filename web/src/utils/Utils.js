import axios from 'axios'

import { torrentsHost } from './Hosts'

export function humanizeSize(size) {
  if (!size) return ''
  const i = Math.floor(Math.log(size) / Math.log(1024))
  return `${(size / Math.pow(1024, i)).toFixed(2) * 1} ${['B', 'KB', 'MB', 'GB', 'TB'][i]}`
}

export function getPeerString(torrent) {
  if (!torrent || !torrent.connected_seeders) return null
  return `[${torrent.connected_seeders}] ${torrent.active_peers} / ${torrent.total_peers}`
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

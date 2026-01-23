import axios from 'axios'
import parseTorrent from 'parse-torrent'
import ptt from 'parse-torrent-title'

export const getMoviePosters = (movieName, language = 'en') => {
  const url = `${window.location.protocol}//api.themoviedb.org/3/search/multi`
  const imgHost = `${window.location.protocol}//${language === 'ru' ? 'imagetmdb.com' : 'image.tmdb.org'}`

  return axios
    .get(url, {
      params: {
        api_key: process.env.REACT_APP_TMDB_API_KEY,
        language,
        include_image_language: `${language},null,en`,
        query: movieName,
      },
    })
    .then(({ data: { results } }) =>
      results.filter(el => el.poster_path).map(el => `${imgHost}/t/p/w300${el.poster_path}`),
    )
    .catch(() => null)
}

export const checkImageURL = async url => {
  if (!url || !url.match(/.(\.jpg|\.jpeg|\.png|\.gif|\.svg||\.webp).*$/i)) return false
  return true
}

const magnetRegex = /^magnet:\?xt=urn:[a-z0-9].*/i
export const hashRegex = /^\b[0-9a-f]{32}\b$|^\b[0-9a-f]{40}\b$|^\b[0-9a-f]{64}\b$/i
const torrentRegex = /^.*\.(torrent)$/i
const linkRegex = /^(http(s?)):\/\/.*/i
const torrsRegex = /^(torrs):\/\/.*/i

export const checkTorrentSource = source =>
  source.match(hashRegex) !== null ||
  source.match(magnetRegex) !== null ||
  source.match(torrentRegex) !== null ||
  source.match(linkRegex) !== null ||
  source.match(torrsRegex) !== null

export const parseTorrentTitle = (parsingSource, callback) => {
  parseTorrent.remote(parsingSource, (err, { name, files } = {}) => {
    if (!name || err) return callback({ parsedTitle: null, originalName: null })

    const torrentName = ptt.parse(name).title
    const nameOfFileInsideTorrent = files ? ptt.parse(files[0].name).title : null

    let newTitle = torrentName
    if (nameOfFileInsideTorrent) {
      // taking shorter title because in most cases it is more accurate
      newTitle = torrentName.length < nameOfFileInsideTorrent.length ? torrentName : nameOfFileInsideTorrent
    }

    callback({ parsedTitle: newTitle, originalName: name })
  })
}

import axios from 'axios'
import { torrentsHost } from 'utils/Hosts'
import { OFFLINE_DOWNLOADING, OFFLINE_PAUSED } from 'utils/offline'

const errorText = e => e?.response?.data?.error || e?.message || String(e)

// The server answers "... try again" when it loses a race while (re)starting a torrent
// ("cannot start the offline download, try again"): repeat such a request a few times.
const RETRY_DELAYS = [1500, 3000, 5000]
const withRetry = (request, attempt = 0) =>
  request().catch(e => {
    if (attempt >= RETRY_DELAYS.length || !/try again/i.test(errorText(e))) throw e
    return new Promise(resolve => setTimeout(resolve, RETRY_DELAYS[attempt])).then(() =>
      withRetry(request, attempt + 1),
    )
  })

// Runs worker over items with at most `limit` requests in flight. Never rejects.
// (a promise chain instead of `await` in a loop: the project lints with no-await-in-loop)
const runPool = (items, worker, limit = 3) => {
  const results = new Array(items.length)
  let next = 0
  const runNext = () => {
    if (next >= items.length) return Promise.resolve()
    const idx = next
    next += 1
    return Promise.resolve()
      .then(() => worker(items[idx]))
      .then(() => {
        results[idx] = { ok: true }
      })
      .catch(e => {
        results[idx] = { ok: false, error: errorText(e) }
      })
      .then(runNext)
  }
  return Promise.all(Array.from({ length: Math.min(limit, items.length) }, runNext)).then(() => results)
}

const sleep = ms => new Promise(resolve => setTimeout(resolve, ms))

// Extra passes over torrents that are still in the source state after a pass (see bulk()).
const MAX_EXTRA_PASSES = 2

const stillThere = (list, fromState, skip) =>
  (list || []).filter(torrent => torrent.offline_state === fromState && torrent.hash && !skip.has(torrent.hash))

const runPass = (targets, action, limit) =>
  runPool(targets, torrent => withRetry(() => axios.post(torrentsHost(), { action, hash: torrent.hash })), limit).then(
    results =>
      results
        .map((result, i) =>
          result.ok
            ? null
            : {
                hash: targets[i].hash,
                text: `${targets[i].title || targets[i].name || targets[i].hash}: ${result.error}`,
              },
        )
        .filter(Boolean),
  )

// Applies `action` to every torrent in `fromState`. After a pass the fresh list is fetched; torrents
// that are STILL in `fromState` without having reported an error (the server answered "ok" but the
// state did not change yet) get another pass - the same as pressing the button once more.
// Resolves with the list of "<title>: <reason>" strings (empty = no errors).
const bulk = (torrents, fromState, action, limit, getFresh, skip = new Set(), pass = 0) => {
  const targets = stillThere(torrents, fromState, skip)
  if (!targets.length) return Promise.resolve([])
  return runPass(targets, action, limit).then(failed => {
    failed.forEach(f => skip.add(f.hash))
    const errors = failed.map(f => f.text)
    return getFresh().then(fresh => {
      const left = stillThere(fresh, fromState, skip)
      if (!left.length) return errors
      // eslint-disable-next-line no-console
      console.warn(
        `offline bulk "${action}": ${left.length} torrent(s) still in state "${fromState}" after pass ${pass + 1}`,
        left.map(t => ({
          hash: t.hash,
          title: t.title,
          offline_state: t.offline_state,
          stat: t.stat_string,
          error: t.offline_error,
        })),
      )
      if (pass >= MAX_EXTRA_PASSES) return errors
      return sleep(1000)
        .then(() => bulk(fresh, fromState, action, limit, getFresh, skip, pass + 1))
        .then(more => errors.concat(more))
    })
  })
}

// getFresh: () => Promise<torrents> - refetches the torrent list and resolves with it.
// Both resolve with the list of error strings (empty = everything succeeded).
// Resuming starts a torrent in the BitTorrent client (heavy, racy on the server): one at a time.
export const pauseAllOffline = (torrents, getFresh) =>
  bulk(torrents, OFFLINE_DOWNLOADING, 'offline_pause', 3, getFresh)
export const resumeAllOffline = (torrents, getFresh) =>
  bulk(torrents, OFFLINE_PAUSED, 'offline_resume', 1, getFresh)

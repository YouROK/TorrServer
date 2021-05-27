import { useEffect, useRef, useState } from 'react'
import { cacheHost } from 'utils/Hosts'
import axios from 'axios'

export const useUpdateCache = hash => {
  const [cache, setCache] = useState({})
  const componentIsMounted = useRef(true)
  const timerID = useRef(null)

  useEffect(
    () => () => {
      // this function is required to notify "updateCache" when NOT to make state update
      componentIsMounted.current = false
    },
    [],
  )

  useEffect(() => {
    if (hash) {
      timerID.current = setInterval(() => {
        const updateCache = newCache => componentIsMounted.current && setCache(newCache)

        axios
          .post(cacheHost(), { action: 'get', hash })
          .then(({ data }) => updateCache(data))
          // empty cache if error
          .catch(() => updateCache({}))
      }, 100)
    } else clearInterval(timerID.current)

    return () => {
      clearInterval(timerID.current)
    }
  }, [hash])

  return cache
}

export const useCreateCacheMap = cache => {
  const [cacheMap, setCacheMap] = useState([])

  useEffect(() => {
    if (!cache.PiecesCount || !cache.Pieces) return

    const { Pieces, PiecesCount, Readers } = cache

    const map = []

    for (let i = 0; i < PiecesCount; i++) {
      const newPiece = { id: i }

      const currentPiece = Pieces[i]
      if (currentPiece) {
        if (currentPiece.Completed && currentPiece.Size === currentPiece.Length) newPiece.isComplete = true
        else {
          newPiece.inProgress = true
          newPiece.percentage = (currentPiece.Size / currentPiece.Length).toFixed(2)
        }
      }

      Readers.forEach(r => {
        if (i === r.Reader) newPiece.isActive = true
        if (i >= r.Start && i <= r.End) newPiece.isReaderRange = true
      })

      map.push(newPiece)
    }

    setCacheMap(map)
  }, [cache])

  return cacheMap
}

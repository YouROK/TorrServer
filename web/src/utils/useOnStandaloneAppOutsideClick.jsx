import { useEffect, useRef } from 'react'
import { isStandaloneApp } from 'utils/Utils'

export default function useOnStandaloneAppOutsideClick(onClickOutside) {
  const ref = useRef()

  useEffect(() => {
    if (!isStandaloneApp) return

    const handleClickOutside = event => {
      if (ref.current && !ref.current.contains(event.target)) {
        onClickOutside && onClickOutside()
      }
    }

    document.addEventListener('click', handleClickOutside, true)

    return () => {
      document.removeEventListener('click', handleClickOutside, true)
    }
  })

  return ref
}

import { useEffect, useRef } from 'react'
import { isStandaloneApp } from 'utils/Utils'

export default function useOnStandaloneAppOutsideClick(onClickOutside) {
  const ref = useRef()

  useEffect(() => {
    if (!isStandaloneApp) return

    const handleClickOutside = event => {
      if (ref.current && !ref.current.contains(event.target)) {
        const { target, path, composedPath } = event
        const eventPath = path || (composedPath && composedPath()) || []

        const isWithinMuiMenu =
          target.closest('[role="listbox"]') ||
          target.closest('[role="menu"]') ||
          target.closest('.MuiMenu-root') ||
          target.closest('.MuiPopover-root') ||
          target.closest('.MuiPaper-root[role="presentation"]') ||
          target.closest('.MuiSelect-root') ||
          target.closest('[class*="MuiMenu"]') ||
          target.closest('[class*="MuiPopover"]') ||
          target.closest('[class*="MuiSelect-menu"]') ||
          target.closest('[class*="MuiMenuItem"]') ||
          eventPath.some(
            el =>
              el &&
              (el.classList?.contains('MuiMenu-root') ||
                el.classList?.contains('MuiPopover-root') ||
                el.classList?.contains('MuiPaper-root') ||
                el.classList?.contains('MuiMenuItem-root') ||
                el.getAttribute?.('role') === 'listbox' ||
                el.getAttribute?.('role') === 'menu'),
          )

        if (!isWithinMuiMenu) {
          onClickOutside && onClickOutside()
        }
      }
    }

    document.addEventListener('click', handleClickOutside, true)

    return () => {
      document.removeEventListener('click', handleClickOutside, true)
    }
  })

  return ref
}

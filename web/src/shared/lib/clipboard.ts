/**
 * Copy text to the clipboard.
 *
 * `navigator.clipboard` only works in a secure context (HTTPS or localhost).
 * TorrServer is often opened as `http://192.168.x.x` on phones — that fails.
 * Fall back to a hidden textarea + `document.execCommand('copy')`, which still
 * works from a user gesture on iOS Safari / Android Chrome over plain HTTP.
 *
 * When copying from inside a modal, the textarea must be mounted **inside** the
 * dialog — React Aria's focus trap steals focus from `document.body` nodes and
 * `execCommand('copy')` silently fails (looks like the button did nothing).
 *
 * @param text - Plain string to place on the system clipboard.
 * @throws If both the Clipboard API and the `execCommand` fallback fail.
 */
export async function copyToClipboard(text: string): Promise<void> {
  if (typeof navigator !== 'undefined' && typeof navigator.clipboard?.writeText === 'function') {
    try {
      await navigator.clipboard.writeText(text)
      return
    } catch {
      // Fall through — insecure context, permissions, or modal focus races.
    }
  }

  copyWithExecCommand(text)
}

function resolveCopyMountParent(): HTMLElement {
  if (typeof document === 'undefined') {
    throw new Error('Clipboard unavailable')
  }

  const active = document.activeElement as { closest?: (selector: string) => Element | null } | null
  const fromActive = active?.closest?.('[role="dialog"]')
  if (fromActive) return fromActive as HTMLElement

  const openDialog = typeof document.querySelector === 'function' ? document.querySelector('[role="dialog"]') : null
  if (openDialog) return openDialog as HTMLElement

  return document.body
}

function copyWithExecCommand(text: string): void {
  if (typeof document === 'undefined') {
    throw new Error('Clipboard unavailable')
  }

  const parent = resolveCopyMountParent()
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.setAttribute('readonly', '')
  textarea.setAttribute('aria-hidden', 'true')
  // Keep in-flow inside the dialog so the focus trap allows focus + select.
  textarea.style.position = 'fixed'
  textarea.style.top = '0'
  textarea.style.left = '0'
  textarea.style.width = '2em'
  textarea.style.height = '2em'
  textarea.style.padding = '0'
  textarea.style.border = 'none'
  textarea.style.outline = 'none'
  textarea.style.boxShadow = 'none'
  textarea.style.background = 'transparent'
  textarea.style.opacity = '0'
  textarea.style.zIndex = '0'

  parent.appendChild(textarea)

  const selection = document.getSelection()
  const previousRange = selection && selection.rangeCount > 0 ? selection.getRangeAt(0) : null

  textarea.focus({ preventScroll: true })
  textarea.select()
  textarea.setSelectionRange(0, text.length)

  try {
    const ok = document.execCommand('copy')
    if (!ok) throw new Error('Clipboard copy failed')
  } finally {
    parent.removeChild(textarea)
    if (selection) {
      selection.removeAllRanges()
      if (previousRange) {
        try {
          selection.addRange(previousRange)
        } catch {
          // Range may be from a detached node after modal updates.
        }
      }
    }
  }
}

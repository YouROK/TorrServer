import { afterEach, describe, expect, it, vi } from 'vitest'

import { copyToClipboard } from './clipboard'

function installDom(execOk = true) {
  const textarea = {
    value: '',
    style: {} as Record<string, string>,
    focus: vi.fn(),
    select: vi.fn(),
    setSelectionRange: vi.fn(),
    setAttribute: vi.fn(),
  }

  const body = {
    appendChild: vi.fn(),
    removeChild: vi.fn(),
  }

  const doc = {
    body,
    activeElement: null,
    createElement: vi.fn(() => textarea),
    getSelection: vi.fn(() => null),
    querySelector: vi.fn(() => null),
    execCommand: vi.fn(() => execOk),
  }

  vi.stubGlobal('document', doc)
  return { doc, textarea, body }
}

describe('copyToClipboard', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it('uses navigator.clipboard in a secure context', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    vi.stubGlobal('isSecureContext', true)
    vi.stubGlobal('navigator', { clipboard: { writeText } })

    await copyToClipboard('magnet:?xt=urn:btih:abc')

    expect(writeText).toHaveBeenCalledWith('magnet:?xt=urn:btih:abc')
  })

  it('falls back to execCommand when clipboard API is unavailable', async () => {
    vi.stubGlobal('isSecureContext', false)
    vi.stubGlobal('navigator', {})
    const { doc, body } = installDom(true)

    await copyToClipboard('hash-value')

    expect(doc.execCommand).toHaveBeenCalledWith('copy')
    expect(body.appendChild).toHaveBeenCalled()
    expect(body.removeChild).toHaveBeenCalled()
  })

  it('mounts the fallback textarea inside an open dialog when focused there', async () => {
    vi.stubGlobal('isSecureContext', false)
    vi.stubGlobal('navigator', {})

    const dialog = {
      appendChild: vi.fn(),
      removeChild: vi.fn(),
    }
    const button = {
      closest: vi.fn((selector: string) => (selector === '[role="dialog"]' ? dialog : null)),
    }
    const textarea = {
      value: '',
      style: {} as Record<string, string>,
      focus: vi.fn(),
      select: vi.fn(),
      setSelectionRange: vi.fn(),
      setAttribute: vi.fn(),
    }
    const doc = {
      body: { appendChild: vi.fn(), removeChild: vi.fn() },
      activeElement: button,
      createElement: vi.fn(() => textarea),
      getSelection: vi.fn(() => null),
      querySelector: vi.fn(),
      execCommand: vi.fn(() => true),
    }
    vi.stubGlobal('document', doc)

    await copyToClipboard('inside-modal')

    expect(dialog.appendChild).toHaveBeenCalled()
    expect(dialog.removeChild).toHaveBeenCalled()
    expect(doc.body.appendChild).not.toHaveBeenCalled()
  })

  it('falls back when clipboard.writeText rejects', async () => {
    vi.stubGlobal('isSecureContext', true)
    vi.stubGlobal('navigator', {
      clipboard: { writeText: vi.fn().mockRejectedValue(new Error('denied')) },
    })
    const { doc } = installDom(true)

    await copyToClipboard('stream-link')

    expect(doc.execCommand).toHaveBeenCalledWith('copy')
  })

  it('throws when execCommand fallback also fails', async () => {
    vi.stubGlobal('isSecureContext', false)
    vi.stubGlobal('navigator', {})
    installDom(false)

    await expect(copyToClipboard('x')).rejects.toThrow(/Clipboard copy failed/)
  })
})

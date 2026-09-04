import axios from 'axios'

import { getTelegramInitData } from 'shared/lib/telegramWebApp'

const STORAGE_KEY = 'torrserver.basicAuth'
const TG_HEADER = 'X-Telegram-Init-Data'

export interface BasicCredentials {
  username: string
  password: string
}

/** `Basic …` header value for axios / fetch / hls.js. */
export function encodeBasicAuthorization(username: string, password: string): string {
  // btoa expects Latin1; UTF-8 usernames/passwords via percent-escape roundtrip.
  const binary = unescape(encodeURIComponent(`${username}:${password}`))
  return `Basic ${btoa(binary)}`
}

export function getStoredCredentials(): BasicCredentials | null {
  if (typeof sessionStorage === 'undefined') return null
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as Partial<BasicCredentials>
    if (typeof parsed.username !== 'string' || typeof parsed.password !== 'string') return null
    if (!parsed.username) return null
    return { username: parsed.username, password: parsed.password }
  } catch {
    return null
  }
}

export function getAuthorizationHeader(): string | null {
  const creds = getStoredCredentials()
  if (!creds) return null
  return encodeBasicAuthorization(creds.username, creds.password)
}

/** Persist credentials and set axios default Authorization. */
export function setCredentials(username: string, password: string): void {
  if (typeof sessionStorage !== 'undefined') {
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify({ username, password }))
  }
  applyAuthToAxios()
}

export function clearCredentials(): void {
  if (typeof sessionStorage !== 'undefined') {
    sessionStorage.removeItem(STORAGE_KEY)
  }
  delete axios.defaults.headers.common.Authorization
}

/** Apply stored Basic and/or Telegram Mini App initData to axios defaults. */
export function applyAuthToAxios(): void {
  const header = getAuthorizationHeader()
  if (header) axios.defaults.headers.common.Authorization = header
  else delete axios.defaults.headers.common.Authorization

  const initData = getTelegramInitData()
  if (initData) axios.defaults.headers.common[TG_HEADER] = initData
  else delete axios.defaults.headers.common[TG_HEADER]

  // Helps CheckAuth skip WWW-Authenticate for SPA probes.
  axios.defaults.headers.common['X-Requested-With'] = 'XMLHttpRequest'
}

/** Same-origin media URL with userinfo so `<video src>` sends Basic without native dialog. */
export function withAuthMediaUrl(url: string): string {
  const creds = getStoredCredentials()
  if (!creds) return url
  try {
    const origin = typeof window !== 'undefined' ? window.location.origin : undefined
    const parsed = new URL(url, origin)
    if (origin && parsed.origin !== origin) return url
    parsed.username = creds.username
    parsed.password = creds.password
    return parsed.toString()
  } catch {
    return url
  }
}

/** fetch() with Authorization / Telegram initData when available. */
export function authFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  const headers = new Headers(init?.headers)
  const header = getAuthorizationHeader()
  if (header && !headers.has('Authorization')) headers.set('Authorization', header)
  const initData = getTelegramInitData()
  if (initData && !headers.has(TG_HEADER)) headers.set(TG_HEADER, initData)
  return fetch(input, { ...init, headers })
}

/** Clear session Basic and reload so AuthGate can show Login again. */
export function logoutBasicAuth(): void {
  clearCredentials()
  window.location.assign(window.location.pathname || '/')
}

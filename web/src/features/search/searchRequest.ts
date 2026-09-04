/** Pure helpers for aborting in-flight search requests and ignoring stale responses. */

export type MutableRef<T> = { current: T }

export function beginSearchRequest(
  abortRef: MutableRef<AbortController | null>,
  genRef: MutableRef<number>,
): { ac: AbortController; gen: number } {
  abortRef.current?.abort()
  const ac = new AbortController()
  abortRef.current = ac
  const gen = ++genRef.current
  return { ac, gen }
}

/** True when a response belongs to the latest search and was not aborted. */
export function isCurrentSearch(gen: number, currentGen: number, signal: AbortSignal): boolean {
  return gen === currentGen && !signal.aborted
}

import { useCallback } from 'react'

import type { PlayableFile } from 'shared/api/types'
import { buildExternalPlayerHref, useExternalPlayers } from 'shared/lib/externalPlayers'
import { useLocalJsonPref } from 'shared/hooks/useLocalPref'
import { detectApplePlatform } from 'shared/lib/platform'
import {
  coercePosterPlayAction,
  defaultPosterPlayAction,
  POSTER_PLAY_ACTION_KEY,
  streamPlayUrl,
  torrentPlaylistUrl,
  type PosterPlayAction,
} from 'shared/lib/posterPlay'

export interface RunConfiguredPlayArgs {
  hash: string
  displayName: string
  knownPlayableFiles: PlayableFile[]
  /** Built-in player (file picker / VideoPlayer). */
  handlePlay: () => void
  /** Resolve one playable file (picker when several), then hand off. */
  resolvePlayableFile: (onFile: (file: PlayableFile) => void) => void
  copyText: (text: string) => void | Promise<void>
  /**
   * Details overview: when action is built-in and there are several files, open the Files tab
   * instead of the launcher picker (poster keeps using {@link handlePlay}).
   */
  onBuiltinMultiFile?: () => void
}

/**
 * Shared Play-button preference (Settings → Players): built-in / copy link / external.
 * Used by library posters and the details overview Play button.
 */
export function useConfiguredPlayAction() {
  const isIOS = detectApplePlatform().isIOS
  const [stored] = useLocalJsonPref<PosterPlayAction>(POSTER_PLAY_ACTION_KEY, defaultPosterPlayAction(isIOS))
  const { playerFlags } = useExternalPlayers()
  const playAction = coercePosterPlayAction(stored, playerFlags, isIOS)

  const runConfiguredPlay = useCallback(
    ({
      hash,
      displayName,
      knownPlayableFiles,
      handlePlay,
      resolvePlayableFile,
      copyText,
      onBuiltinMultiFile,
    }: RunConfiguredPlayArgs) => {
      if (playAction === 'builtin') {
        if (onBuiltinMultiFile && knownPlayableFiles.length !== 1) {
          onBuiltinMultiFile()
          return
        }
        handlePlay()
        return
      }

      if (playAction === 'copyLink') {
        if (knownPlayableFiles.length === 1) {
          const href = new URL(streamPlayUrl(hash, knownPlayableFiles[0]), window.location.href).toString()
          void copyText(href)
          return
        }
        const playlist = new URL(torrentPlaylistUrl(hash, displayName), window.location.href).toString()
        void copyText(playlist)
        return
      }

      resolvePlayableFile(file => {
        const fullLink = new URL(streamPlayUrl(hash, file), window.location.href).toString()
        window.location.href = buildExternalPlayerHref(playAction, fullLink)
      })
    },
    [playAction],
  )

  return { playAction, runConfiguredPlay }
}

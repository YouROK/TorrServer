import { Button, ButtonGroup, Dropdown, Spinner, Tooltip, useMediaQuery, useOverlayState } from '@heroui/react'
import { ArrowDownToLine, FileVideo, Link2, MoreHorizontal, Play, SquareArrowOutUpRight } from 'lucide-react'
import type { ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import type { ExternalPlayerLink } from 'shared/lib/externalPlayers'
import { copyToClipboard } from 'shared/lib/clipboard'
import { queryMax } from 'shared/theme/breakpoints'
import { iconBtn } from 'shared/ui/controlClasses'
import { iconAction, iconMenu } from 'shared/ui/iconProps'
import { useOptionalAppToast } from 'shared/ui/Toast'

export type { ExternalPlayerLink }

export interface FileRowActionsProps {
  /** Localized preload / buffer button label (TorrentDetails preload wording). */
  preloadLabel: string
  onPreload: () => void
  /** False when the in-app player cannot handle this file — show open-link Play fallback. */
  playerSupported: boolean
  onPlay: () => void
  isPlayPending?: boolean
  openLinkHref?: string
  showOpenLink?: boolean
  /** Direct stream URL copied by "Copy link". */
  copyText: string
  /** Deep links for Infuse / VLC / etc. — always visible; never buried in the More menu. */
  externalPlayers: ExternalPlayerLink[]
  onProbeMedia?: () => void
}

/** Desktop / tablet external-player chip — mid weight between Play and utilities. */
const playerBtn = 'min-h-11 shrink-0 px-3 font-medium'

const actionIcon = { ...iconAction, 'aria-hidden': true as const }
const menuIcon = { ...iconMenu }
/** Ghost utility segments — quieter than secondary player chips; keep 44px hit target. */
const utilitySegBtn = `${iconBtn} text-muted hover-fine:bg-surface-tertiary hover-fine:text-foreground`
const moreBtn = `${iconBtn} text-muted hover-fine:text-foreground`

/**
 * Per-file action strip for the details Files tab.
 *
 * Layout contract:
 * - Play is always primary and reachable (doctrine).
 * - External players stay on-screen (not only in a menu).
 * - Open / Copy / Preload / MediaInfo: one ghost ButtonGroup on desktop; `⋯` on mobile.
 * - Mobile: one compact row (Play icon + external chips + More) — never a full-width
 *   stacked Play that doubles row height.
 */
export default function FileRowActions({
  preloadLabel,
  onPreload,
  playerSupported,
  onPlay,
  isPlayPending = false,
  openLinkHref,
  showOpenLink,
  copyText,
  externalPlayers,
  onProbeMedia,
}: FileRowActionsProps) {
  const { t } = useTranslation()
  const toast = useOptionalAppToast()
  const isMobile = useMediaQuery(queryMax('mobile'))
  const moreMenu = useOverlayState()

  const copyDirectLink = async () => {
    try {
      await copyToClipboard(copyText)
      toast?.showToast({ message: t('Copied'), severity: 'success' })
    } catch {
      toast?.showToast({ message: t('Error'), severity: 'error' })
    }
  }

  const openExternal = () => {
    if (openLinkHref) window.open(openLinkHref, '_blank', 'noopener,noreferrer')
  }

  const showOpenIcon = Boolean(showOpenLink && openLinkHref && playerSupported)

  const playButton = (className: string, iconOnly = false) =>
    playerSupported ? (
      <Button
        variant='primary'
        size='sm'
        isIconOnly={iconOnly}
        className={className}
        isPending={isPlayPending}
        onPress={onPlay}
        aria-label={t('Play')}
      >
        {({ isPending }) => (
          <>
            {isPending ? <Spinner size='sm' color='current' /> : <Play {...iconMenu} fill='currentColor' aria-hidden />}
            {iconOnly ? null : t('Play')}
          </>
        )}
      </Button>
    ) : showOpenLink && openLinkHref ? (
      <Button
        variant='primary'
        size='sm'
        isIconOnly={iconOnly}
        className={className}
        onPress={openExternal}
        aria-label={t('Play')}
      >
        <Play {...iconMenu} fill='currentColor' aria-hidden />
        {iconOnly ? null : t('Play')}
      </Button>
    ) : null

  const utilityButton = (opts: { label: string; onPress: () => void; children: ReactNode }) => (
    <Tooltip.Root>
      <Tooltip.Trigger>
        <Button
          variant='ghost'
          size='sm'
          isIconOnly
          className={utilitySegBtn}
          aria-label={opts.label}
          onPress={opts.onPress}
        >
          {opts.children}
        </Button>
      </Tooltip.Trigger>
      <Tooltip.Content>{opts.label}</Tooltip.Content>
    </Tooltip.Root>
  )

  const secondaryDesktop = (
    <ButtonGroup className='shrink-0'>
      {showOpenIcon
        ? utilityButton({
            label: t('OpenLink'),
            onPress: openExternal,
            children: <SquareArrowOutUpRight {...actionIcon} />,
          })
        : null}
      {utilityButton({
        label: t('CopyLink'),
        onPress: () => void copyDirectLink(),
        children: <Link2 {...actionIcon} />,
      })}
      {utilityButton({
        label: preloadLabel,
        onPress: onPreload,
        children: <ArrowDownToLine {...actionIcon} />,
      })}
      {onProbeMedia
        ? utilityButton({
            label: t('MediaInfo'),
            onPress: onProbeMedia,
            children: <FileVideo {...actionIcon} />,
          })
        : null}
    </ButtonGroup>
  )

  const secondaryMobile = (
    <Dropdown isOpen={moreMenu.isOpen} onOpenChange={moreMenu.setOpen}>
      <Dropdown.Trigger>
        <Button variant='ghost' size='sm' isIconOnly className={moreBtn} aria-label={t('Actions')}>
          <MoreHorizontal {...actionIcon} />
        </Button>
      </Dropdown.Trigger>
      <Dropdown.Popover placement='bottom end'>
        <Dropdown.Menu aria-label={t('Actions')}>
          {showOpenIcon ? (
            <Dropdown.Item
              onPress={() => {
                moreMenu.close()
                openExternal()
              }}
            >
              <SquareArrowOutUpRight {...menuIcon} />
              {t('OpenLink')}
            </Dropdown.Item>
          ) : null}
          <Dropdown.Item
            onPress={() => {
              moreMenu.close()
              void copyDirectLink()
            }}
          >
            <Link2 {...menuIcon} />
            {t('CopyLink')}
          </Dropdown.Item>
          <Dropdown.Item
            onPress={() => {
              moreMenu.close()
              onPreload()
            }}
          >
            <ArrowDownToLine {...menuIcon} />
            {preloadLabel}
          </Dropdown.Item>
          {onProbeMedia ? (
            <Dropdown.Item
              onPress={() => {
                moreMenu.close()
                onProbeMedia()
              }}
            >
              <FileVideo {...menuIcon} />
              {t('MediaInfo')}
            </Dropdown.Item>
          ) : null}
        </Dropdown.Menu>
      </Dropdown.Popover>
    </Dropdown>
  )

  if (isMobile) {
    return (
      <div className='flex shrink-0 items-center gap-1'>
        {playButton(iconBtn, true)}
        {externalPlayers.map(player => (
          <Button
            key={player.label}
            variant='secondary'
            size='sm'
            className='min-h-11 max-w-[4.5rem] shrink-0 px-1.5 text-[11px] font-medium'
            onPress={() => {
              window.location.href = player.href
            }}
          >
            <span className='truncate'>{player.label}</span>
          </Button>
        ))}
        {secondaryMobile}
      </div>
    )
  }

  const externalButtons = externalPlayers.map(player => (
    <Button
      key={player.label}
      variant='secondary'
      size='sm'
      className={playerBtn}
      onPress={() => {
        window.location.href = player.href
      }}
    >
      {player.label}
    </Button>
  ))

  return (
    <div className='flex flex-nowrap items-center gap-1.5'>
      {playButton('min-h-11 shrink-0 px-3')}
      {externalPlayers.length > 1 ? <ButtonGroup className='shrink-0'>{externalButtons}</ButtonGroup> : externalButtons}
      {secondaryDesktop}
    </div>
  )
}

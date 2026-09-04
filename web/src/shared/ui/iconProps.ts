/**
 * Shared Lucide sizing — keep stroke/weight consistent across chrome, actions, menus.
 * Settings tabs use size 17 ad-hoc; action buttons use {@link iconAction}.
 */
export const iconNav = { size: 20, strokeWidth: 1.75 } as const
export const iconNavMobile = { size: 22, strokeWidth: 1.75 } as const
export const iconAction = { size: 18, strokeWidth: 1.75 } as const
export const iconMenu = { size: 16, strokeWidth: 1.75 } as const
export const iconChrome = { size: 18, strokeWidth: 1.75 } as const
export const iconEmpty = { size: 36, strokeWidth: 1.25 } as const
/** Slightly larger empty-state glyph (offline / hero empties). */
export const iconEmptyLg = { size: 44, strokeWidth: 1.25 } as const
/** Compact player chrome inside size-9 hit targets. */
export const iconPlayerCompact = { size: 16, strokeWidth: 1.75 } as const

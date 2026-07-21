/** Bottom tab content band (icons + labels). Safe-area is added as padding on the in-flow nav. */
export const BOTTOM_NAV_CONTENT_PX = 56

/**
 * Clearance for fixed overlays (toasts / banners) above the mobile bottom nav.
 * Nav itself must NOT be `position:fixed` on iOS PWA — that leaves a system white strip
 * under `bottom: 0`. It is an in-flow flex/grid child; overlays still need this offset.
 */
export const BOTTOM_NAV_OFFSET = `calc(${BOTTOM_NAV_CONTENT_PX}px + env(safe-area-inset-bottom, 0px))`

/** Toast / snackbar clearance above the bottom nav. */
export const BOTTOM_NAV_TOAST_OFFSET = `calc(${BOTTOM_NAV_CONTENT_PX}px + env(safe-area-inset-bottom, 0px) + 12px)`

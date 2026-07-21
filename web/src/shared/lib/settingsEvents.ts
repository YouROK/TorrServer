/** Fired after BT settings are successfully saved so open dialogs can refetch. */
export const SETTINGS_CHANGED_EVENT = 'torrserver:settings-changed'

export const notifySettingsChanged = (): void => {
  try {
    window.dispatchEvent(new Event(SETTINGS_CHANGED_EVENT))
  } catch {
    // ignore (SSR / non-browser)
  }
}

/**
 * Lets far-away components (card/details actions) ask the Shell to open Settings on a specific
 * tab — e.g. a "set up external players" hint — without prop-drilling dialog state through
 * TorrentsPage/DetailsDialog.
 */
export const OPEN_SETTINGS_EVENT = 'torrserver:open-settings'

export type SettingsDeepLinkTab =
  'primary' | 'network' | 'features' | 'storage' | 'appearance' | 'app' | 'gstreamer' | 'torznab'

export const requestOpenSettings = (tab: SettingsDeepLinkTab): void => {
  try {
    window.dispatchEvent(new CustomEvent<SettingsDeepLinkTab>(OPEN_SETTINGS_EVENT, { detail: tab }))
  } catch {
    // ignore (SSR / non-browser)
  }
}

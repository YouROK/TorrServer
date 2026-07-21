import type { OfflineAwareProps } from 'shared/api/types'

/** Shared action surface for Sidebar (desktop) and BottomNav (mobile). */
export interface ShellNavProps extends OfflineAwareProps {
  /** Whether a non-"all" category filter is currently applied — used to highlight the Category nav item. */
  isCategoryFilterActive?: boolean
  onAdd: () => void
  onSearch: () => void
  onCategories: () => void
  onSettings: () => void
  onAbout: () => void
  onDonate: () => void
  onServerStatus: () => void
  onCloseServer: () => void
  onRemoveAll: () => void
  /** Present when the session used the web Basic login form. */
  onLogout?: () => void
}

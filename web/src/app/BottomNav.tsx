import { type ReactNode } from 'react'
import { Button, Modal, useOverlayState } from '@heroui/react'
import {
  Activity,
  CreditCard,
  Ellipsis,
  FolderPlus,
  Info,
  Layers,
  LogOut,
  Power,
  Search,
  Settings,
  Trash2,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useSyncModalOpen } from 'shared/ui/ModalOpenContext'
import { iconNav, iconNavMobile } from 'shared/ui/iconProps'

import { BOTTOM_NAV_CONTENT_PX } from './bottomNavLayout'
import type { ShellNavProps } from './navTypes'

/** In-flow bottom tab bar for narrow viewports (not position:fixed — iOS PWA safe-area). Overflow actions live in a sheet. */
export default function BottomNav({
  isOffline,
  isLoading,
  isCategoryFilterActive,
  onAdd,
  onSearch,
  onCategories,
  onSettings,
  onAbout,
  onDonate,
  onServerStatus,
  onCloseServer,
  onRemoveAll,
  onLogout,
}: ShellNavProps) {
  const { t } = useTranslation()
  const disabled = isOffline || isLoading
  const moreState = useOverlayState()
  useSyncModalOpen(moreState.isOpen)

  const closeMore = () => moreState.close()

  const tab = (label: string, icon: ReactNode, onPress: () => void, isDisabled = false, active = false) => (
    <Button
      key={label}
      variant='ghost'
      isDisabled={isDisabled}
      onPress={onPress}
      aria-label={label}
      aria-current={active ? 'true' : undefined}
      className={`flex h-full min-w-0 flex-1 flex-col items-center justify-end gap-0.5 rounded-none px-1 pb-1 pt-1.5 text-xs font-medium ${
        active ? 'text-accent' : ''
      }`}
    >
      {icon}
      <span className='truncate'>{label}</span>
    </Button>
  )

  const sheetAction = (label: string, icon: ReactNode, onPress: () => void, isDisabled = false, danger = false) => (
    <Button
      key={label}
      variant='ghost'
      isDisabled={isDisabled}
      onPress={() => {
        closeMore()
        onPress()
      }}
      className={`h-auto w-full justify-start gap-3 px-4 py-3 min-h-11 ${danger ? 'text-danger' : ''}`}
    >
      {icon}
      {label}
    </Button>
  )

  return (
    <>
      <div
        className='ts-bottom-nav shrink-0 border-t border-border pb-[env(safe-area-inset-bottom,0px)]'
        style={{ backgroundColor: 'var(--surface)' }}
      >
        <div className='mx-auto flex max-w-lg items-stretch' style={{ height: BOTTOM_NAV_CONTENT_PX }}>
          {tab(t('Add'), <FolderPlus {...iconNavMobile} />, onAdd, disabled)}
          {tab(t('nav.Search'), <Search {...iconNavMobile} />, onSearch, disabled)}
          {tab(t('Category'), <Layers {...iconNavMobile} />, onCategories, false, isCategoryFilterActive)}
          {tab(t('nav.More'), <Ellipsis {...iconNavMobile} />, moreState.open)}
        </div>
      </div>

      <Modal state={moreState}>
        <Modal.Backdrop isDismissable onClick={closeMore}>
          <Modal.Container placement='bottom' size='md'>
            <Modal.Dialog aria-label={t('nav.More')}>
              <Modal.Body className='flex flex-col gap-1 pb-[env(safe-area-inset-bottom,0px)] pt-2'>
                {sheetAction(t('RemoveAll'), <Trash2 {...iconNav} />, onRemoveAll, disabled)}
                {sheetAction(t('ServerStatus'), <Activity {...iconNav} />, onServerStatus, disabled)}
                {sheetAction(t('nav.Settings'), <Settings {...iconNav} />, onSettings, disabled)}
                {sheetAction(t('About'), <Info {...iconNav} />, onAbout)}
                {sheetAction(t('Donate'), <CreditCard {...iconNav} />, onDonate)}
                {onLogout ? sheetAction(t('Logout'), <LogOut {...iconNav} />, onLogout) : null}
                {sheetAction(t('CloseServer'), <Power {...iconNav} />, onCloseServer, disabled, true)}
              </Modal.Body>
            </Modal.Dialog>
          </Modal.Container>
        </Modal.Backdrop>
      </Modal>
    </>
  )
}

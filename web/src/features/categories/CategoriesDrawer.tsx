import { Drawer, Separator, useOverlayState } from '@heroui/react'
import { Check, CircleDashed, LayoutGrid } from 'lucide-react'
import { iconAction, iconMenu } from 'shared/ui/iconProps'
import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'

import { TORRENT_CATEGORIES } from 'shared/torrent/categories'
import { useSyncModalOpen } from 'shared/ui/ModalOpenContext'

export interface CategoriesDrawerProps {
  open: boolean
  onClose: () => void
  selectedCategory: string
  onSelectCategory: (category: string) => void
}

/** Side sheet for filtering the torrent library by category, plus "All" / "Uncategorized". */
export default function CategoriesDrawer({ open, onClose, selectedCategory, onSelectCategory }: CategoriesDrawerProps) {
  const { t } = useTranslation()
  useSyncModalOpen(open)

  const state = useOverlayState({
    isOpen: open,
    onOpenChange: next => {
      if (!next) onClose()
    },
  })

  useEffect(() => {
    if (open) state.setOpen(true)
  }, [open, state])

  const select = (category: string) => {
    onSelectCategory(category)
    onClose()
  }

  const itemClass = (active: boolean) =>
    `flex w-full items-center gap-3 rounded-lg px-3 py-3 text-left text-sm font-medium transition-colors ${
      active ? 'bg-accent-soft text-accent-soft-foreground' : 'text-foreground hover-fine:bg-surface-secondary'
    }`

  return (
    <Drawer.Root state={state}>
      <Drawer.Backdrop>
        <Drawer.Content placement='left'>
          <Drawer.Dialog aria-label={t('Category')}>
            <Drawer.Header className='pt-[max(0.75rem,env(safe-area-inset-top,0px))]'>
              <Drawer.Heading className='flex items-center gap-2'>
                <LayoutGrid {...iconMenu} aria-hidden />
                {t('Category')}
              </Drawer.Heading>
              <Drawer.CloseTrigger className='min-h-11 min-w-11' />
            </Drawer.Header>
            <Drawer.Body className='w-[280px] space-y-1 pt-1'>
              <button type='button' className={itemClass(selectedCategory === 'all')} onClick={() => select('all')}>
                {selectedCategory === 'all' ? (
                  <Check {...iconAction} className='shrink-0' aria-hidden />
                ) : (
                  <LayoutGrid {...iconAction} className='shrink-0' aria-hidden />
                )}
                <span>{t('All')}</span>
              </button>

              {TORRENT_CATEGORIES.map(category => (
                <button
                  key={category.key}
                  type='button'
                  className={itemClass(selectedCategory === category.key)}
                  onClick={() => select(category.key)}
                >
                  {selectedCategory === category.key ? (
                    <Check {...iconAction} className='shrink-0' aria-hidden />
                  ) : (
                    <span className='shrink-0'>{category.icon}</span>
                  )}
                  <span>{t(category.name)}</span>
                </button>
              ))}

              <Separator className='my-2' />

              <button type='button' className={itemClass(selectedCategory === '')} onClick={() => select('')}>
                {selectedCategory === '' ? (
                  <Check {...iconAction} className='shrink-0' aria-hidden />
                ) : (
                  <CircleDashed {...iconAction} className='shrink-0' aria-hidden />
                )}
                <span>{t('Uncategorized')}</span>
              </button>
            </Drawer.Body>
          </Drawer.Dialog>
        </Drawer.Content>
      </Drawer.Backdrop>
    </Drawer.Root>
  )
}

import { Button, Link, ListBox, Modal, Select, Spinner } from '@heroui/react'
import axios from 'axios'
import { Activity, CreditCard, Gauge, Heart, SquareArrowOutUpRight } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { runSpeedTest } from 'shared/api/extras'
import { echoHost } from 'shared/api/hosts'
import { publicUrl } from 'shared/lib/publicUrl'
import { useDialogFullScreen } from 'shared/hooks/useDialogFullScreen'
import AppDialog from 'shared/ui/AppDialog'
import { DIALOG_SHEET_M } from 'shared/ui/dialogSizes'
import { iconMenu } from 'shared/ui/iconProps'
import { useOptionalAppToast } from 'shared/ui/Toast'

import { CONTRIBUTORS } from './contributors'

export interface AboutDialogProps {
  open: boolean
  onClose: () => void
  onOpenServerStatus?: () => void
  onOpenDonate?: () => void
}

/** Download sizes for `/download/:size` (path unit is MB). */
const SPEED_TEST_SIZES_MB = [
  { id: '10', mb: 10, label: '10 MB' },
  { id: '100', mb: 100, label: '100 MB' },
  { id: '1024', mb: 1024, label: '1 GB' },
] as const

function AboutLink({ name, href }: { name: string; href: string }) {
  return (
    <Link
      href={href}
      target='_blank'
      rel='noopener noreferrer'
      className='flex items-center justify-between gap-2 rounded-md px-1 py-1.5 text-sm'
    >
      <span>{name}</span>
      <SquareArrowOutUpRight size={14} strokeWidth={1.75} className='shrink-0 text-muted' aria-hidden />
    </Link>
  )
}

function contributorInitials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean)
  if (parts.length >= 2) return `${parts[0][0]}${parts[1][0]}`.toUpperCase()
  return name.slice(0, 2).toUpperCase()
}

/** About dialog: version, links, speedtest, Donate entry, and Special Thanks contributor grid. */
export default function AboutDialog({ open, onClose, onOpenServerStatus, onOpenDonate }: AboutDialogProps) {
  const { t } = useTranslation()
  const toast = useOptionalAppToast()
  const isFullScreenBreakpoint = useDialogFullScreen()
  const [version, setVersion] = useState<string | null>(null)
  const [speedTesting, setSpeedTesting] = useState(false)
  const [speedResult, setSpeedResult] = useState<string | null>(null)
  const [speedSizeId, setSpeedSizeId] = useState<(typeof SPEED_TEST_SIZES_MB)[number]['id']>('10')

  const speedSizeMb = useMemo(
    () => SPEED_TEST_SIZES_MB.find(option => option.id === speedSizeId)?.mb ?? 10,
    [speedSizeId],
  )

  useEffect(() => {
    if (!open) return
    // eslint-disable-next-line react-hooks/set-state-in-effect -- reset About panel when reopened
    setVersion(null)
    setSpeedResult(null)
    axios
      .get(echoHost())
      .then(({ data }) => setVersion(String(data)))
      .catch(() => setVersion(''))
  }, [open])

  const handleSpeedTest = async () => {
    setSpeedTesting(true)
    setSpeedResult(null)
    try {
      const result = await runSpeedTest(speedSizeMb)
      const transferred =
        result.bytes >= 1024 * 1024 * 1024
          ? `${(result.bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
          : `${(result.bytes / (1024 * 1024)).toFixed(1)} MB`
      const label = `${result.mbps.toFixed(1)} Mbps · ${(result.elapsedMs / 1000).toFixed(1)}s · ${transferred}`
      setSpeedResult(label)
      toast?.showToast({ message: label, severity: 'success' })
    } catch {
      toast?.showToast({ message: t('Error'), severity: 'error' })
    } finally {
      setSpeedTesting(false)
    }
  }

  return (
    <AppDialog
      open={open}
      onClose={onClose}
      size='md'
      fullScreen={isFullScreenBreakpoint}
      dialogStyle={isFullScreenBreakpoint ? undefined : DIALOG_SHEET_M}
    >
      <Modal.Header>
        <Modal.Heading>{t('About')}</Modal.Heading>
        <Modal.CloseTrigger aria-label={t('Close')} />
      </Modal.Header>
      <Modal.Body>
        <div className='flex flex-col items-center gap-3 pb-2 pt-1 text-center'>
          <img src={publicUrl('icon.png')} alt='TorrServer' className='size-[72px] rounded-2xl shadow-lg' />
          <h2 className='text-lg font-semibold text-foreground'>
            TorrServer{' '}
            {version === null ? (
              <span className='inline-block h-4 w-12 animate-pulse rounded bg-surface-secondary align-middle' />
            ) : (
              version
            )}
          </h2>
          <p className='text-muted'>{t('ThanksToEveryone')}</p>
        </div>

        <div className='mt-2 rounded-lg border border-border bg-surface-secondary p-3'>
          <p className='mb-1 px-1 text-xs font-semibold uppercase tracking-wide text-muted'>{t('Links')}</p>
          <AboutLink name={t('ProjectSource')} href='https://github.com/YouROK/TorrServer' />
          <AboutLink name={t('Releases')} href='https://github.com/YouROK/TorrServer/releases' />
          <AboutLink name={t('NasReleases')} href='https://github.com/vladlenas' />
          <AboutLink name={t('ApiDocs')} href={publicUrl('swagger/index.html')} />
        </div>

        <div className='mt-3 rounded-lg border border-border bg-surface-secondary p-3'>
          <p className='mb-2 px-1 text-xs font-semibold uppercase tracking-wide text-muted'>{t('ServerStatus')}</p>
          <div className='px-1'>
            <Button
              size='sm'
              variant='secondary'
              className='min-h-10 gap-2'
              onPress={() => {
                onClose()
                onOpenServerStatus?.()
              }}
              isDisabled={!onOpenServerStatus}
            >
              <Activity {...iconMenu} aria-hidden />
              {t('ServerStatus')}
            </Button>
          </div>
        </div>

        <div className='mt-3 rounded-lg border border-border bg-surface-secondary p-3'>
          <p className='mb-2 px-1 text-xs font-semibold uppercase tracking-wide text-muted'>{t('Donate')}</p>
          <div className='px-1'>
            <Button
              size='sm'
              variant='secondary'
              className='min-h-10 gap-2'
              onPress={() => {
                onClose()
                onOpenDonate?.()
              }}
              isDisabled={!onOpenDonate}
            >
              <CreditCard {...iconMenu} aria-hidden />
              {t('Support')}
            </Button>
          </div>
        </div>

        <div className='mt-3 rounded-lg border border-border bg-surface-secondary p-3'>
          <p className='mb-2 px-1 text-xs font-semibold uppercase tracking-wide text-muted'>{t('SpeedTest')}</p>
          <div className='flex flex-wrap items-center gap-2 px-1'>
            <Select
              aria-label={t('SpeedTestSize')}
              selectedKey={speedSizeId}
              onSelectionChange={key => {
                if (key == null) return
                setSpeedSizeId(String(key) as (typeof SPEED_TEST_SIZES_MB)[number]['id'])
                setSpeedResult(null)
              }}
              className='w-[7.5rem] shrink-0'
              isDisabled={speedTesting}
            >
              <Select.Trigger className='min-h-10'>
                <Select.Value />
                <Select.Indicator />
              </Select.Trigger>
              <Select.Popover>
                <ListBox>
                  {SPEED_TEST_SIZES_MB.map(option => (
                    <ListBox.Item key={option.id} id={option.id}>
                      {option.label}
                    </ListBox.Item>
                  ))}
                </ListBox>
              </Select.Popover>
            </Select>
            <Button size='sm' variant='secondary' isPending={speedTesting} onPress={() => void handleSpeedTest()}>
              {({ isPending }) => (
                <>
                  {isPending ? <Spinner size='sm' color='current' /> : <Gauge {...iconMenu} aria-hidden />}
                  {t('SpeedTestRun')}
                </>
              )}
            </Button>
            {speedResult ? (
              <span className='w-full text-sm tabular-nums text-muted sm:w-auto'>{speedResult}</span>
            ) : null}
          </div>
        </div>

        <div className='mt-3 rounded-lg border border-border bg-surface-secondary p-3'>
          <p className='mb-3 flex items-center gap-1.5 px-1 text-xs font-semibold uppercase tracking-wide text-muted'>
            <Heart size={14} strokeWidth={1.75} className='fill-accent/25 text-accent' aria-hidden />
            {t('SpecialThanks')}
          </p>
          <ul className='flex flex-wrap gap-2'>
            {CONTRIBUTORS.map(person => (
              <li key={person.url}>
                <Link
                  href={person.url}
                  target='_blank'
                  rel='noopener noreferrer'
                  className='inline-flex items-center gap-2 rounded-full border border-border/80 bg-surface px-2 py-1 text-sm text-foreground shadow-sm transition-colors hover-fine:border-accent/40 hover-fine:bg-accent-soft/40'
                >
                  <span
                    className='grid size-6 shrink-0 place-items-center rounded-full bg-accent-soft text-[10px] font-bold tracking-wide text-accent'
                    aria-hidden
                  >
                    {contributorInitials(person.name)}
                  </span>
                  <span className='pr-0.5'>{person.name}</span>
                </Link>
              </li>
            ))}
          </ul>
        </div>
      </Modal.Body>
      <Modal.Footer>
        <Button variant='secondary' onPress={onClose} autoFocus>
          {t('Close')}
        </Button>
      </Modal.Footer>
    </AppDialog>
  )
}

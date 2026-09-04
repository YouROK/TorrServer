import { useCallback, useEffect, useState } from 'react'
import { Shield } from 'lucide-react'
import { Button, Description, Label, Spinner, TextArea, TextField } from '@heroui/react'
import axios from 'axios'
import { useTranslation } from 'react-i18next'

import { getWAF, setWAF, type WAFWarning } from 'shared/api/waf'

import SettingsSection from './SettingsSection'

export interface WAFSettingsPanelProps {
  onDirtyChange?: (dirty: boolean) => void
  footerButtonClassName?: string
}

type Status = { message: string; type: 'success' | 'error' | '' }

const emptyLists = { whitelist: '', blacklist: '', referers: '' }

function axiosErrorMessage(err: unknown, fallback: string): string {
  if (axios.isAxiosError(err)) {
    const data = err.response?.data as { error?: string } | undefined
    if (data?.error) return String(data.error)
    if (err.message) return err.message
  }
  if (err instanceof Error && err.message) return err.message
  return fallback
}

/** HTTP WAF lists (not BTSets) — own Save, hot-reloads server filters. */
export default function WAFSettingsPanel({ onDirtyChange, footerButtonClassName }: WAFSettingsPanelProps) {
  const { t } = useTranslation()
  const [whitelist, setWhitelist] = useState('')
  const [blacklist, setBlacklist] = useState('')
  const [referers, setReferers] = useState('')
  const [initial, setInitial] = useState(emptyLists)
  const [warnings, setWarnings] = useState<WAFWarning[]>([])
  const [readOnly, setReadOnly] = useState(false)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [loaded, setLoaded] = useState(false)
  const [loadError, setLoadError] = useState('')
  const [status, setStatus] = useState<Status>({ message: '', type: '' })

  const applySnapshot = useCallback(
    (data: { whitelist: string; blacklist: string; referers: string; warnings: WAFWarning[]; read_only: boolean }) => {
      const next = {
        whitelist: data.whitelist || '',
        blacklist: data.blacklist || '',
        referers: data.referers || '',
      }
      setWhitelist(next.whitelist)
      setBlacklist(next.blacklist)
      setReferers(next.referers)
      setInitial(next)
      setWarnings(data.warnings || [])
      setReadOnly(Boolean(data.read_only))
      setLoaded(true)
      setLoadError('')
    },
    [],
  )

  const loadSettings = useCallback(
    async (signal?: AbortSignal) => {
      setLoading(true)
      setLoadError('')
      try {
        applySnapshot(await getWAF(signal))
      } catch (err) {
        if ((err as Error).name === 'AbortError' || (axios.isAxiosError(err) && err.code === 'ERR_CANCELED')) return
        setLoaded(false)
        setLoadError(axiosErrorMessage(err, t('WAF.LoadFailed')))
      } finally {
        if (!signal?.aborted) setLoading(false)
      }
    },
    [applySnapshot, t],
  )

  useEffect(() => {
    const ac = new AbortController()
    // eslint-disable-next-line react-hooks/set-state-in-effect -- load WAF lists on mount
    void loadSettings(ac.signal)
    return () => ac.abort()
  }, [loadSettings])

  const dirty =
    loaded && (whitelist !== initial.whitelist || blacklist !== initial.blacklist || referers !== initial.referers)

  useEffect(() => {
    onDirtyChange?.(dirty)
  }, [dirty, onDirtyChange])

  useEffect(() => {
    const warnBeforeUnload = (event: BeforeUnloadEvent) => {
      if (!dirty) return
      event.preventDefault()
      event.returnValue = ''
    }
    window.addEventListener('beforeunload', warnBeforeUnload)
    return () => window.removeEventListener('beforeunload', warnBeforeUnload)
  }, [dirty])

  const handleSave = async () => {
    setSaving(true)
    setStatus({ message: '', type: '' })
    try {
      applySnapshot(await setWAF({ whitelist, blacklist, referers }))
      setStatus({ message: t('WAF.Saved'), type: 'success' })
    } catch (err) {
      const forbidden = axios.isAxiosError(err) && err.response?.status === 403
      if (forbidden) setReadOnly(true)
      setStatus({
        message: forbidden ? t('WAF.ReadOnlyHint') : axiosErrorMessage(err, t('WAF.SaveFailed')),
        type: 'error',
      })
    } finally {
      setSaving(false)
    }
  }

  const formatWarning = (warning: WAFWarning) =>
    t(warning.line ? 'WAF.WarningMessage' : 'WAF.WarningMessageNoLine', {
      list: t(`WAF.Lists.${warning.list}`),
      line: warning.line,
      reason: t(`WAF.WarningCodes.${warning.code}`),
    })

  if (loading) {
    return (
      <div className='grid min-h-40 place-items-center'>
        <Spinner size='lg' />
      </div>
    )
  }

  if (!loaded) {
    return (
      <div className='space-y-4'>
        <p className='text-sm text-danger' role='alert'>
          {loadError || t('WAF.LoadFailed')}
        </p>
        <Button variant='outline' onPress={() => void loadSettings()} className={footerButtonClassName}>
          {t('WAF.Retry')}
        </Button>
      </div>
    )
  }

  return (
    <div className='space-y-6'>
      <p className='text-sm leading-relaxed text-muted'>{t('WAF.Subtitle')}</p>
      <p className='text-sm leading-relaxed text-muted'>{t('WAF.StorageHint')}</p>
      <p className='text-sm leading-relaxed text-muted'>{t('WAF.SeparateSaveHint')}</p>

      {readOnly ? (
        <p className='rounded-lg border border-border bg-surface-secondary px-3 py-2 text-sm' role='status'>
          {t('WAF.ReadOnlyHint')}
        </p>
      ) : null}

      <SettingsSection icon={<Shield />} title={t('WAF.Title')} description={t('WAF.IPRulesHint')}>
        <TextField value={whitelist} onChange={setWhitelist} isDisabled={readOnly}>
          <Label>{t('WAF.Whitelist')}</Label>
          <TextArea rows={4} className='font-mono text-xs' />
          <Description>{t('WAF.WhitelistHint')}</Description>
        </TextField>
        <TextField value={blacklist} onChange={setBlacklist} isDisabled={readOnly}>
          <Label>{t('WAF.Blacklist')}</Label>
          <TextArea rows={4} className='font-mono text-xs' />
          <Description>{t('WAF.BlacklistHint')}</Description>
        </TextField>
      </SettingsSection>

      <SettingsSection title={t('WAF.ReferersSection')} description={t('WAF.ReferersHint')}>
        <TextField value={referers} onChange={setReferers} isDisabled={readOnly}>
          <Label>{t('WAF.Referers')}</Label>
          <TextArea rows={3} className='font-mono text-xs' />
          <Description>{t('WAF.ReferersFieldHint')}</Description>
        </TextField>
      </SettingsSection>

      {warnings.length > 0 ? (
        <div className='rounded-lg border border-warning/40 bg-surface-secondary px-3 py-2 text-sm' role='alert'>
          <p id='waf-warning-heading' className='font-medium'>
            {t('WAF.Warnings')}
          </p>
          <ul aria-labelledby='waf-warning-heading' className='mt-2 list-disc space-y-1 pl-5'>
            {warnings.map(warning => (
              <li key={`${warning.list}-${warning.line}-${warning.code}`}>{formatWarning(warning)}</li>
            ))}
          </ul>
        </div>
      ) : null}

      {status.message ? (
        <p className={`text-sm ${status.type === 'success' ? 'text-accent' : 'text-danger'}`} role='alert'>
          {status.message}
        </p>
      ) : null}

      <div className='flex justify-end'>
        <Button
          variant='primary'
          onPress={() => void handleSave()}
          isDisabled={saving || readOnly || !dirty}
          className={footerButtonClassName}
        >
          {saving ? <Spinner size='sm' color='current' /> : t('WAF.SaveLists')}
        </Button>
      </div>
    </div>
  )
}

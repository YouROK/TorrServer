import { useCallback, useEffect, useState } from 'react'
import { Alert, Button, Input, Label, Spinner, TextField } from '@heroui/react'
import axios from 'axios'
import { useTranslation } from 'react-i18next'

import { authFetch } from 'shared/api/authCredentials'
import { sslHost } from 'shared/api/hosts'
import type { BTSets } from 'shared/api/types'

const EXPIRING_SOON_MS = 30 * 24 * 60 * 60 * 1000

interface SSLCertInfo {
  source: 'none' | 'self-signed' | 'uploaded' | 'user'
  cert_file?: string
  key_file?: string
  subject?: string
  issuer?: string
  dns_names?: string[]
  ips?: string[]
  not_after?: string
  trusted?: boolean
  error?: string
}

interface SSLStatus {
  enabled: boolean
  port?: string
  http_port?: string
  http_enabled: boolean
  force_https: boolean
  http_media: boolean
  read_only: boolean
  cert_from_flags: boolean
  cert: SSLCertInfo
}

export interface HTTPSSettingsPanelProps {
  onUpdate: <K extends keyof BTSets>(key: K, value: BTSets[K]) => void
}

function axiosErrorMessage(err: unknown): string {
  if (axios.isAxiosError(err)) {
    const data = err.response?.data as { error?: string } | undefined
    if (data?.error) return String(data.error)
    if (err.message) return err.message
  }
  if (err instanceof Error && err.message) return err.message
  return String(err)
}

/** Live HTTPS cert management when TorrServer was started with `--ssl`. Returns null otherwise. */
export default function HTTPSSettingsPanel({ onUpdate }: HTTPSSettingsPanelProps) {
  const { t } = useTranslation()
  const [status, setStatus] = useState<SSLStatus | null | undefined>(undefined)
  const [busy, setBusy] = useState(false)
  const [message, setMessage] = useState<{ text: string; type: 'success' | 'error' | '' }>({ text: '', type: '' })
  const [certFile, setCertFile] = useState<File | null>(null)
  const [keyFile, setKeyFile] = useState<File | null>(null)
  const [inputKey, setInputKey] = useState(0)
  const [paths, setPaths] = useState({ cert: '', key: '' })

  const applyStatus = useCallback(
    (data: SSLStatus, syncPaths: boolean) => {
      setStatus(data)
      const user = data.cert.source === 'user'
      setPaths({ cert: user ? data.cert.cert_file || '' : '', key: user ? data.cert.key_file || '' : '' })
      if (syncPaths) {
        onUpdate('SslCert', data.cert.cert_file || '')
        onUpdate('SslKey', data.cert.key_file || '')
      }
    },
    [onUpdate],
  )

  useEffect(() => {
    const ac = new AbortController()
    axios
      .get<SSLStatus>(`${sslHost()}/status`, { signal: ac.signal })
      .then(({ data }) => applyStatus(data, false))
      .catch(err => {
        if (axios.isAxiosError(err) && err.code === 'ERR_CANCELED') return
        setStatus(null)
      })
    return () => ac.abort()
  }, [applyStatus])

  const run = async (request: () => Promise<{ data: SSLStatus }>, successText: string) => {
    setBusy(true)
    setMessage({ text: '', type: '' })
    try {
      const { data } = await request()
      applyStatus(data, true)
      setMessage({ text: successText, type: 'success' })
      return true
    } catch (err) {
      setMessage({ text: axiosErrorMessage(err), type: 'error' })
      return false
    } finally {
      setBusy(false)
    }
  }

  const upload = async () => {
    if (!certFile || !keyFile) return
    const form = new FormData()
    form.append('cert', certFile)
    form.append('key', keyFile)
    if (await run(() => axios.post(`${sslHost()}/upload`, form), t('HTTPSSettings.Uploaded'))) {
      setCertFile(null)
      setKeyFile(null)
      setInputKey(k => k + 1)
    }
  }

  const downloadCert = async () => {
    try {
      const res = await authFetch(`${sslHost()}/cert`)
      if (!res.ok) throw new Error(t('HTTPSSettings.Download'))
      const blob = await res.blob()
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'torrserver.crt'
      a.click()
      URL.revokeObjectURL(url)
    } catch (err) {
      setMessage({ text: axiosErrorMessage(err), type: 'error' })
    }
  }

  if (!status?.enabled) return null

  const { cert } = status
  const selfSigned = cert.source === 'self-signed'
  const expires = cert.not_after ? new Date(cert.not_after) : null
  const expiringSoon = Boolean(expires && expires.getTime() - Date.now() < EXPIRING_SOON_MS)
  const names = [...(cert.dns_names || []), ...(cert.ips || [])]
  const locked = busy || status.read_only || status.cert_from_flags
  const { protocol, hostname } = window.location
  const plainHTTPPage = protocol === 'http:' && !['localhost', '127.0.0.1', '[::1]'].includes(hostname)

  let mode = t('HTTPSSettings.ModeBoth', { port: status.port, httpPort: status.http_port })
  if (!status.http_enabled) mode = t('HTTPSSettings.ModeHTTPSOnly', { port: status.port })
  else if (status.http_media) mode = t('HTTPSSettings.ModeHTTPMedia', { port: status.port })
  else if (status.force_https) mode = t('HTTPSSettings.ModeForceHTTPS', { port: status.port })

  return (
    <div className='space-y-4'>
      <p className='text-sm font-medium'>{t('HTTPS')}</p>
      <dl className='space-y-1 text-sm'>
        <div className='flex justify-between gap-4'>
          <dt className='text-muted'>{t('HTTPSSettings.Mode')}</dt>
          <dd className='text-right'>{mode}</dd>
        </div>
        <div className='flex justify-between gap-4'>
          <dt className='text-muted'>{t('HTTPSSettings.Certificate')}</dt>
          <dd className='text-right'>{t(`HTTPSSettings.Source.${cert.source}`)}</dd>
        </div>
        {cert.subject ? (
          <div className='flex justify-between gap-4'>
            <dt className='text-muted'>{t('HTTPSSettings.Subject')}</dt>
            <dd className='text-right break-all'>{cert.subject}</dd>
          </div>
        ) : null}
        {cert.issuer && !selfSigned ? (
          <div className='flex justify-between gap-4'>
            <dt className='text-muted'>{t('HTTPSSettings.Issuer')}</dt>
            <dd className='text-right break-all'>{cert.issuer}</dd>
          </div>
        ) : null}
        {expires ? (
          <div className='flex justify-between gap-4'>
            <dt className='text-muted'>{t('HTTPSSettings.ValidUntil')}</dt>
            <dd className={expiringSoon ? 'text-right text-danger' : 'text-right'}>{expires.toLocaleDateString()}</dd>
          </div>
        ) : null}
        {cert.source !== 'none' ? (
          <div className='flex justify-between gap-4'>
            <dt className='text-muted'>{t('HTTPSSettings.Trusted')}</dt>
            <dd className='text-right'>{cert.trusted ? t('HTTPSSettings.Yes') : t('HTTPSSettings.No')}</dd>
          </div>
        ) : null}
      </dl>
      {names.length > 0 ? (
        <p className='text-sm text-muted'>
          {t('HTTPSSettings.Names')}: {names.join(', ')}
        </p>
      ) : null}
      {cert.cert_file ? (
        <p className='text-sm break-all text-muted'>
          {t('HTTPSSettings.CertFileInUse')}: {cert.cert_file}
          <br />
          {t('HTTPSSettings.KeyFileInUse')}: {cert.key_file}
        </p>
      ) : null}
      {cert.error ? <Alert status='danger'>{cert.error}</Alert> : null}
      <p className='text-sm text-muted'>{t('HTTPSSettings.ModeHint')}</p>
      {selfSigned ? <p className='text-sm text-muted'>{t('HTTPSSettings.SelfSignedHint')}</p> : null}
      {status.cert_from_flags ? <p className='text-sm text-muted'>{t('HTTPSSettings.FromFlagsHint')}</p> : null}

      <div className='flex flex-wrap gap-2'>
        {cert.source !== 'none' ? (
          <Button size='sm' variant='secondary' onPress={() => void downloadCert()}>
            {t('HTTPSSettings.Download')}
          </Button>
        ) : null}
        {selfSigned ? (
          <Button
            size='sm'
            variant='secondary'
            isDisabled={locked}
            onPress={() => void run(() => axios.post(`${sslHost()}/regenerate`), t('HTTPSSettings.Regenerated'))}
          >
            {t('HTTPSSettings.Regenerate')}
          </Button>
        ) : (
          <Button
            size='sm'
            variant='secondary'
            isDisabled={locked}
            onPress={() => void run(() => axios.post(`${sslHost()}/selfsigned`), t('HTTPSSettings.SwitchedToSelfSigned'))}
          >
            {t('HTTPSSettings.UseSelfSigned')}
          </Button>
        )}
      </div>

      <div className='space-y-2'>
        <p className='text-sm text-muted'>{t('HTTPSSettings.UploadHint')}</p>
        {plainHTTPPage ? <Alert status='warning'>{t('HTTPSSettings.PlainHTTPWarning')}</Alert> : null}
        <div className='flex flex-wrap items-center gap-2' key={inputKey}>
          <Button size='sm' variant='secondary' isDisabled={locked} className='relative'>
            {certFile ? certFile.name : t('HTTPSSettings.ChooseCert')}
            <input
              type='file'
              accept='.pem,.crt,.cer'
              className='absolute inset-0 cursor-pointer opacity-0'
              disabled={locked}
              onChange={e => setCertFile(e.target.files?.[0] || null)}
            />
          </Button>
          <Button size='sm' variant='secondary' isDisabled={locked} className='relative'>
            {keyFile ? keyFile.name : t('HTTPSSettings.ChooseKey')}
            <input
              type='file'
              accept='.pem,.key'
              className='absolute inset-0 cursor-pointer opacity-0'
              disabled={locked}
              onChange={e => setKeyFile(e.target.files?.[0] || null)}
            />
          </Button>
          <Button size='sm' variant='primary' isDisabled={locked || !certFile || !keyFile} onPress={() => void upload()}>
            {t('HTTPSSettings.Upload')}
          </Button>
        </div>
      </div>

      <div className='space-y-2'>
        <p className='text-sm text-muted'>{t('HTTPSSettings.PathsHint')}</p>
        <TextField
          value={paths.cert}
          onChange={value => setPaths({ ...paths, cert: value })}
          isDisabled={locked}
        >
          <Label>{t('HTTPSSettings.CertPath')}</Label>
          <Input placeholder={t('HTTPSSettings.CertPathExample')} />
        </TextField>
        <TextField value={paths.key} onChange={value => setPaths({ ...paths, key: value })} isDisabled={locked}>
          <Label>{t('HTTPSSettings.KeyPath')}</Label>
          <Input placeholder={t('HTTPSSettings.KeyPathExample')} />
        </TextField>
        <div className='flex items-center gap-2'>
          <Button
            size='sm'
            variant='primary'
            isDisabled={locked || !paths.cert || !paths.key}
            onPress={() =>
              void run(() => axios.post(`${sslHost()}/paths`, paths), t('HTTPSSettings.PathsApplied'))
            }
          >
            {t('HTTPSSettings.UsePaths')}
          </Button>
          {busy ? <Spinner size='sm' /> : null}
        </div>
      </div>

      {message.text ? (
        <Alert status={message.type === 'error' ? 'danger' : 'success'}>
          <Alert.Description>{message.text}</Alert.Description>
        </Alert>
      ) : null}
    </div>
  )
}

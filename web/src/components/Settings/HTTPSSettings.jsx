import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Box, Button, CircularProgress, FormHelperText } from '@material-ui/core'
import axios from 'axios'
import { sslHost } from 'utils/Hosts'

import { SettingsStatusMessage, GstRuntimeStatusList, GstRuntimeStatusItem } from './style'

const expiringSoonMs = 30 * 24 * 60 * 60 * 1000

const Row = ({ label, value, style }) => (
  <div className='gst-status-row'>
    <span className='gst-status-label'>{label}</span>
    <span className='gst-status-value' style={{ whiteSpace: 'normal', textAlign: 'right', ...style }}>
      {value}
    </span>
  </div>
)

// HTTPSSettings shows the HTTPS mode and certificate, and lets the user upload a
// certificate or manage the self-signed one. Changes apply without a restart.
export default function HTTPSSettings({ updateSettings }) {
  const { t } = useTranslation()
  const [status, setStatus] = useState()
  const [loadError, setLoadError] = useState('')
  const [busy, setBusy] = useState(false)
  const [message, setMessage] = useState({ text: '', type: '' })
  const [certFile, setCertFile] = useState(null)
  const [keyFile, setKeyFile] = useState(null)
  const [inputKey, setInputKey] = useState(0)

  const applyStatus = useCallback(
    (data, syncPaths) => {
      setStatus(data)
      // keep the dialog's copy of the paths in sync, or Save would revert the change
      if (syncPaths) updateSettings?.({ SslCert: data.cert.cert_file || '', SslKey: data.cert.key_file || '' })
    },
    [updateSettings],
  )

  useEffect(() => {
    axios
      .get(`${sslHost()}/status`)
      .then(({ data }) => applyStatus(data, false))
      .catch(err => setLoadError(err.response?.data?.error || err.message))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const run = async (request, successText) => {
    setBusy(true)
    setMessage({ text: '', type: '' })
    try {
      const { data } = await request()
      applyStatus(data, true)
      setMessage({ text: successText, type: 'success' })
      return true
    } catch (err) {
      setMessage({ text: err.response?.data?.error || err.message, type: 'error' })
      return false
    } finally {
      setBusy(false)
    }
  }

  const upload = async () => {
    const form = new FormData()
    form.append('cert', certFile)
    form.append('key', keyFile)
    if (await run(() => axios.post(`${sslHost()}/upload`, form), t('HTTPSSettings.Uploaded'))) {
      setCertFile(null)
      setKeyFile(null)
      setInputKey(k => k + 1)
    }
  }

  if (loadError) {
    return <SettingsStatusMessage severity='error'>{loadError}</SettingsStatusMessage>
  }
  if (!status) {
    return <CircularProgress color='secondary' size={24} />
  }

  const { cert } = status
  const selfSigned = cert.source === 'self-signed'
  const expires = cert.not_after ? new Date(cert.not_after) : null
  const expiringSoon = expires && expires - Date.now() < expiringSoonMs
  const names = [...(cert.dns_names || []), ...(cert.ips || [])]
  const healthy = !cert.error && cert.source !== 'none' && !expiringSoon
  const locked = busy || status.read_only || status.cert_from_flags
  const { protocol, hostname } = window.location
  const plainHTTPPage = protocol === 'http:' && !['localhost', '127.0.0.1', '[::1]'].includes(hostname)

  let mode = t('HTTPSSettings.ModeOff')
  if (status.enabled) {
    if (!status.http_enabled) mode = t('HTTPSSettings.ModeHTTPSOnly', { port: status.port })
    else if (status.http_media) mode = t('HTTPSSettings.ModeHTTPMedia', { port: status.port })
    else if (status.force_https) mode = t('HTTPSSettings.ModeForceHTTPS', { port: status.port })
    else mode = t('HTTPSSettings.ModeBoth', { port: status.port, httpPort: status.http_port })
  }

  return (
    <>
      <GstRuntimeStatusList>
        <GstRuntimeStatusItem ok={status.enabled && healthy} warn={status.enabled && !healthy}>
          <Row label={t('HTTPSSettings.Mode')} value={mode} />
          <Row label={t('HTTPSSettings.Certificate')} value={t(`HTTPSSettings.Source.${cert.source}`)} />
          {cert.subject && <Row label={t('HTTPSSettings.Subject')} value={cert.subject} />}
          {cert.issuer && !selfSigned && <Row label={t('HTTPSSettings.Issuer')} value={cert.issuer} />}
          {expires && (
            <Row
              label={t('HTTPSSettings.ValidUntil')}
              value={expires.toLocaleDateString()}
              style={expiringSoon ? { color: '#c82e3f' } : undefined}
            />
          )}
          {cert.source !== 'none' && (
            <Row
              label={t('HTTPSSettings.Trusted')}
              value={cert.trusted ? t('HTTPSSettings.Yes') : t('HTTPSSettings.No')}
            />
          )}
          {names.length > 0 && (
            <div className='gst-status-error'>
              {t('HTTPSSettings.Names')}: {names.join(', ')}
            </div>
          )}
          {cert.error && <div className='gst-status-error'>{cert.error}</div>}
        </GstRuntimeStatusItem>
      </GstRuntimeStatusList>
      {!status.enabled && <FormHelperText>{t('HTTPSSettings.OffHint')}</FormHelperText>}
      {selfSigned && <FormHelperText>{t('HTTPSSettings.SelfSignedHint')}</FormHelperText>}
      {status.cert_from_flags && <FormHelperText>{t('HTTPSSettings.FromFlagsHint')}</FormHelperText>}

      <Box display='flex' flexWrap='wrap' mt={1} style={{ gap: 8 }}>
        {cert.source !== 'none' && (
          <Button variant='outlined' color='secondary' href={`${sslHost()}/cert`} download>
            {t('HTTPSSettings.Download')}
          </Button>
        )}
        {selfSigned && (
          <Button
            variant='outlined'
            color='secondary'
            disabled={locked}
            onClick={() => run(() => axios.post(`${sslHost()}/regenerate`), t('HTTPSSettings.Regenerated'))}
          >
            {t('HTTPSSettings.Regenerate')}
          </Button>
        )}
        {!selfSigned && (
          <Button
            variant='outlined'
            color='secondary'
            disabled={locked}
            onClick={() => run(() => axios.post(`${sslHost()}/selfsigned`), t('HTTPSSettings.SwitchedToSelfSigned'))}
          >
            {t('HTTPSSettings.UseSelfSigned')}
          </Button>
        )}
      </Box>

      <Box mt={2}>
        <FormHelperText>{t('HTTPSSettings.UploadHint')}</FormHelperText>
        {plainHTTPPage && (
          <FormHelperText style={{ color: '#cda184' }}>{t('HTTPSSettings.PlainHTTPWarning')}</FormHelperText>
        )}
        <Box display='flex' flexWrap='wrap' alignItems='center' mt={1} style={{ gap: 8 }} key={inputKey}>
          <Button variant='outlined' component='label' disabled={locked}>
            {certFile ? certFile.name : t('HTTPSSettings.ChooseCert')}
            <input type='file' accept='.pem,.crt,.cer' hidden onChange={e => setCertFile(e.target.files[0] || null)} />
          </Button>
          <Button variant='outlined' component='label' disabled={locked}>
            {keyFile ? keyFile.name : t('HTTPSSettings.ChooseKey')}
            <input type='file' accept='.pem,.key' hidden onChange={e => setKeyFile(e.target.files[0] || null)} />
          </Button>
          <Button variant='contained' color='secondary' disabled={locked || !certFile || !keyFile} onClick={upload}>
            {t('HTTPSSettings.Upload')}
          </Button>
          {busy && <CircularProgress color='secondary' size={20} />}
        </Box>
      </Box>
      {message.text && <SettingsStatusMessage severity={message.type}>{message.text}</SettingsStatusMessage>}
    </>
  )
}

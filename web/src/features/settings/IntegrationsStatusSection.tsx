import { useQuery } from '@tanstack/react-query'
import { Button } from '@heroui/react'
import { Cable } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { getRuntimeStatus, RUNTIME_STATUS_QUERY_KEY } from 'shared/api/runtime'
import { getTorrServerHost } from 'shared/api/hosts'
import { copyToClipboard } from 'shared/lib/clipboard'
import { useOptionalAppToast } from 'shared/ui/Toast'

import SettingsSection from './SettingsSection'

/** Read-only DLNA / Bonjour / WebDAV / FUSE status from GET /runtime/status. */
export default function IntegrationsStatusSection() {
  const { t } = useTranslation()
  const toast = useOptionalAppToast()
  const { data, isLoading, isError } = useQuery({
    queryKey: RUNTIME_STATUS_QUERY_KEY,
    queryFn: ({ signal }) => getRuntimeStatus(signal),
    staleTime: 30_000,
  })

  const row = (label: string, enabled: boolean | undefined, detail?: string) => (
    <div className='flex items-center justify-between gap-3 py-1.5 text-sm'>
      <span className='text-foreground'>{label}</span>
      <span className='text-muted'>
        {enabled ? t('Enabled') : t('Disabled')}
        {detail ? ` · ${detail}` : ''}
      </span>
    </div>
  )

  const webdavUrl = data?.webdav_enabled ? `${getTorrServerHost()}${data.webdav_path || '/dav'}` : null

  return (
    <SettingsSection icon={<Cable />} title={t('IntegrationsStatus')}>
      {isLoading ? <p className='text-sm text-muted'>…</p> : null}
      {isError ? <p className='text-sm text-danger'>{t('Error')}</p> : null}
      {data ? (
        <div className='space-y-1'>
          {row(t('DLNAStatus'), data.dlna_enabled, data.friendly_name || undefined)}
          {row(t('BonjourStatus'), data.bonjour_enabled, data.friendly_name || undefined)}
          {row(t('WebDAVStatus'), data.webdav_enabled, data.webdav_path)}
          {row(t('FuseStatus'), data.fuse_enabled, data.fuse_path || undefined)}
          {webdavUrl ? (
            <Button
              size='sm'
              variant='secondary'
              className='mt-2 min-h-11'
              onPress={() => {
                void copyToClipboard(webdavUrl)
                  .then(() => toast?.showToast({ message: t('Copied'), severity: 'success' }))
                  .catch(() => toast?.showToast({ message: t('Error'), severity: 'error' }))
              }}
            >
              {t('CopyWebDavUrl')}
            </Button>
          ) : null}
        </div>
      ) : null}
    </SettingsSection>
  )
}

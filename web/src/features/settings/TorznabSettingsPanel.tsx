import { useState } from 'react'
import { Plus, Rss, Trash2 } from 'lucide-react'
import { Button, Description, Input, Label, Spinner, TextField } from '@heroui/react'
import axios from 'axios'
import { useTranslation } from 'react-i18next'

import type { BTSets, TorznabUrl } from 'shared/api/types'
import { torznabTestHost } from 'shared/api/hosts'

import SettingsSection from './SettingsSection'

export interface TorznabSettingsPanelProps {
  settings: BTSets
  onUpdate: <K extends keyof BTSets>(key: K, value: BTSets[K]) => void
  footerButtonClassName?: string
}

type TestMessage = { ok: boolean; text: string }

/** Torznab indexer list (Jackett/Prowlarr) with add/remove/test-connection rows. */
export default function TorznabSettingsPanel({ settings, onUpdate, footerButtonClassName }: TorznabSettingsPanelProps) {
  const { t } = useTranslation()
  const [newHost, setNewHost] = useState('')
  const [newKey, setNewKey] = useState('')
  const [newName, setNewName] = useState('')
  const [draftTesting, setDraftTesting] = useState(false)
  const [draftTestMsg, setDraftTestMsg] = useState<TestMessage | null>(null)
  const [rowTestingIndex, setRowTestingIndex] = useState<number | null>(null)
  const [rowTestMsgs, setRowTestMsgs] = useState<Record<number, TestMessage>>({})

  const urls = settings.TorznabUrls || []

  const handleAdd = () => {
    if (!newHost.trim() || !newKey.trim()) return
    const next: TorznabUrl = { Host: newHost.trim(), Key: newKey.trim(), Name: newName.trim() || undefined }
    onUpdate('TorznabUrls', [...urls, next])
    setNewHost('')
    setNewKey('')
    setNewName('')
    setDraftTestMsg(null)
  }

  const handleRemove = (index: number) => {
    const next = [...urls]
    next.splice(index, 1)
    onUpdate('TorznabUrls', next)
    setRowTestMsgs(prev => {
      const updated: Record<number, TestMessage> = {}
      for (const [key, value] of Object.entries(prev)) {
        const rowIndex = Number(key)
        if (rowIndex < index) updated[rowIndex] = value
        else if (rowIndex > index) updated[rowIndex - 1] = value
      }
      return updated
    })
  }

  const runTest = async (host: string, key: string, target: 'draft' | number) => {
    if (target === 'draft') {
      setDraftTesting(true)
      setDraftTestMsg(null)
    } else {
      setRowTestingIndex(target)
      setRowTestMsgs(prev => {
        const next = { ...prev }
        delete next[target]
        return next
      })
    }

    try {
      const { data } = await axios.post(torznabTestHost(), { host, key })
      const message: TestMessage = data.success
        ? { ok: true, text: t('Torznab.ConnectionSuccessful') }
        : { ok: false, text: String(data.error || t('Error')) }
      if (target === 'draft') setDraftTestMsg(message)
      else setRowTestMsgs(prev => ({ ...prev, [target]: message }))
    } catch (err) {
      const message: TestMessage = { ok: false, text: (err as Error).message || t('Torznab.SearchFailed') }
      if (target === 'draft') setDraftTestMsg(message)
      else setRowTestMsgs(prev => ({ ...prev, [target]: message }))
    } finally {
      if (target === 'draft') setDraftTesting(false)
      else setRowTestingIndex(null)
    }
  }

  return (
    <div className='space-y-6'>
      {urls.length > 0 ? (
        <SettingsSection icon={<Rss />} title={t('Torznab.Tab')}>
          <ul className='space-y-2'>
            {urls.map((url, index) => {
              const rowTestMsg = rowTestMsgs[index]
              const rowTesting = rowTestingIndex === index

              return (
                <li
                  key={`${url.Host}-${url.Key}-${index}`}
                  className='rounded-lg border border-border bg-surface px-3 py-2.5'
                >
                  <div className='flex items-center gap-3'>
                    <span className='grid size-9 shrink-0 place-items-center rounded-full bg-accent/10 text-accent'>
                      <Rss className='size-4' />
                    </span>
                    <div className='min-w-0 flex-1'>
                      <p className='truncate font-medium text-foreground'>{url.Name || url.Host}</p>
                      <p className='truncate text-sm text-muted'>
                        {url.Host} &middot; {t('Torznab.APIKey')}: {url.Key.slice(0, 5)}&hellip;
                      </p>
                    </div>
                    <Button
                      variant='outline'
                      size='sm'
                      onPress={() => void runTest(url.Host, url.Key, index)}
                      isDisabled={rowTesting}
                      className='min-h-11 shrink-0'
                    >
                      {rowTesting ? <Spinner size='sm' color='current' /> : t('Torznab.Test')}
                    </Button>
                    <Button
                      isIconOnly
                      variant='ghost'
                      aria-label={t('Delete')}
                      onPress={() => handleRemove(index)}
                      className='min-h-11 min-w-11 shrink-0'
                    >
                      <Trash2 className='size-4' />
                    </Button>
                  </div>
                  {rowTestMsg ? (
                    <p className={`mt-2 text-sm ${rowTestMsg.ok ? 'text-accent' : 'text-danger'}`}>{rowTestMsg.text}</p>
                  ) : null}
                </li>
              )
            })}
          </ul>
        </SettingsSection>
      ) : null}

      <SettingsSection title={t('Torznab.AddServer')}>
        <TextField value={newName} onChange={setNewName}>
          <Label>{t('Torznab.NameOptional')}</Label>
          <Input />
        </TextField>
        <TextField value={newHost} onChange={setNewHost}>
          <Label>{t('Torznab.TorznabHostURL')}</Label>
          <Input placeholder='http://localhost:9696/1' />
          <Description>{t('Torznab.HostHint')}</Description>
        </TextField>
        <TextField value={newKey} onChange={setNewKey}>
          <Label>{t('Torznab.APIKey')}</Label>
          <Input />
        </TextField>

        {draftTestMsg ? (
          <p className={`text-sm ${draftTestMsg.ok ? 'text-accent' : 'text-danger'}`}>{draftTestMsg.text}</p>
        ) : null}

        <div className='flex flex-col gap-2 sm:flex-row'>
          <Button
            variant='outline'
            onPress={() => void runTest(newHost, newKey, 'draft')}
            isDisabled={!newHost || !newKey || draftTesting}
            className={footerButtonClassName}
          >
            {draftTesting ? <Spinner size='sm' color='current' /> : t('Torznab.Test')}
          </Button>
          <Button
            variant='primary'
            onPress={handleAdd}
            isDisabled={!newHost || !newKey}
            className={footerButtonClassName}
          >
            <Plus className='size-4' />
            {t('Torznab.AddServer')}
          </Button>
        </div>
      </SettingsSection>
    </div>
  )
}

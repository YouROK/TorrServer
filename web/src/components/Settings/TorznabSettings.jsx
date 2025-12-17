import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  FormControlLabel,
  FormGroup,
  FormHelperText,
  Switch,
  TextField,
  Button,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  IconButton,
  Typography,
} from '@material-ui/core'
import CircularProgress from '@material-ui/core/CircularProgress'
import DeleteIcon from '@material-ui/icons/Delete'
import axios from 'axios'
import { torznabTestHost } from 'utils/Hosts'

import { SecondarySettingsContent, SettingSectionLabel } from './style'

export default function TorznabSettings({ settings, inputForm, updateSettings }) {
  const { t } = useTranslation()
  const { EnableTorznabSearch, TorznabUrls } = settings || {}
  const [newHost, setNewHost] = useState('')
  const [newKey, setNewKey] = useState('')
  const [newName, setNewName] = useState('')
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState(null)

  const handleAdd = () => {
    if (newHost && newKey) {
      const currentUrls = TorznabUrls || []
      updateSettings({ TorznabUrls: [...currentUrls, { Host: newHost, Key: newKey, Name: newName }] })
      setNewHost('')
      setNewKey('')
      setNewName('')
    }
  }

  const handleDelete = index => {
    const currentUrls = TorznabUrls || []
    const newUrls = [...currentUrls]
    newUrls.splice(index, 1)
    updateSettings({ TorznabUrls: newUrls })
  }

  const handleTest = async () => {
    setTesting(true)
    setTestResult(null)
    try {
      const { data } = await axios.post(torznabTestHost(), {
        host: newHost,
        key: newKey,
      })
      if (data.success) {
        setTestResult({ success: true, msg: t('Torznab.ConnectionSuccessful') })
      } else {
        setTestResult({ success: false, msg: data.error })
      }
    } catch (e) {
      setTestResult({ success: false, msg: e.message })
    }
    setTesting(false)
  }

  return (
    <SecondarySettingsContent>
      <SettingSectionLabel>Torznab</SettingSectionLabel>
      <FormGroup>
        <FormControlLabel
          control={
            <Switch
              checked={EnableTorznabSearch || false}
              onChange={inputForm}
              id='EnableTorznabSearch'
              color='secondary'
            />
          }
          label={t('Torznab.EnableTorznabSearch')}
          labelPlacement='start'
        />
        <FormHelperText margin='none'>{t('Torznab.EnableSearchViaTorznab')}</FormHelperText>
      </FormGroup>

      <div
        style={{
          padding: 20,
          opacity: EnableTorznabSearch ? 1 : 0.5,
          pointerEvents: EnableTorznabSearch ? 'auto' : 'none',
        }}
      >
        <List dense>
          {(TorznabUrls || []).map((url, index) => (
            <ListItem key={`${url.Host}-${url.Key}`}>
              <ListItemText
                primary={url.Name || url.Host}
                secondary={
                  <>
                    {url.Name && (
                      <Typography component='span' variant='body2' display='block' color='textSecondary'>
                        {url.Host}
                      </Typography>
                    )}
                    {`Key: ${url.Key.substring(0, 5)}...`}
                  </>
                }
              />
              <ListItemSecondaryAction>
                <IconButton edge='end' aria-label='delete' onClick={() => handleDelete(index)}>
                  <DeleteIcon />
                </IconButton>
              </ListItemSecondaryAction>
            </ListItem>
          ))}
        </List>

        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            gap: 10,
            marginTop: 10,
          }}
        >
          <TextField
            label={t('Torznab.NameOptional')}
            value={newName}
            onChange={e => setNewName(e.target.value)}
            placeholder='My Tracker'
            variant='outlined'
            size='small'
          />
          <TextField
            label={t('Torznab.TorznabHostURL')}
            value={newHost}
            onChange={e => setNewHost(e.target.value)}
            placeholder='http://localhost:9117'
            variant='outlined'
            size='small'
          />
          <TextField
            label={t('Torznab.APIKey')}
            value={newKey}
            onChange={e => setNewKey(e.target.value)}
            variant='outlined'
            size='small'
          />
          <div style={{ display: 'flex', marginTop: 10 }}>
            <Button
              variant='outlined'
              color='secondary'
              onClick={handleTest}
              disabled={!newHost || !newKey || testing}
              style={{ marginRight: 10 }}
            >
              {testing ? <CircularProgress size={24} color='inherit' /> : t('Torznab.Test')}
            </Button>
            <Button variant='contained' color='secondary' onClick={handleAdd} disabled={!newHost || !newKey}>
              {t('Torznab.AddServer')}
            </Button>
          </div>
          {testResult && (
            <Typography variant='caption' style={{ color: testResult.success ? 'green' : 'red' }}>
              {testResult.msg}
            </Typography>
          )}
        </div>
      </div>
      <br />
    </SecondarySettingsContent>
  )
}

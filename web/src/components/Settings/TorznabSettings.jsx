import React, { useState } from 'react'
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
    Typography
} from '@material-ui/core'
import DeleteIcon from '@material-ui/icons/Delete'
import { SettingSectionLabel } from './style'
import axios from 'axios'
import { torznabTestHost } from 'utils/Hosts'
import CircularProgress from '@material-ui/core/CircularProgress'

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
            updateSettings({
                TorznabUrls: [...currentUrls, { Host: newHost, Key: newKey, Name: newName }]
            })
            setNewHost('')
            setNewKey('')
            setNewName('')
        }
    }

    const handleDelete = (index) => {
        const currentUrls = TorznabUrls || []
        const newUrls = [...currentUrls]
        newUrls.splice(index, 1)
        updateSettings({
            TorznabUrls: newUrls
        })
    }

    const handleTest = async () => {
        setTesting(true)
        setTestResult(null)
        try {
            const { data } = await axios.post(torznabTestHost(), {
                host: newHost,
                key: newKey
            })
            if (data.success) {
                setTestResult({ success: true, msg: 'Connection successful' })
            } else {
                setTestResult({ success: false, msg: data.error })
            }
        } catch (e) {
            setTestResult({ success: false, msg: e.message })
        }
        setTesting(false)
    }

    return (
        <>
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
                    label='Enable Torznab Search'
                    labelPlacement='start'
                />
                <FormHelperText margin='none'>Enable search via Torznab indexers (e.g. Jackett, Prowlarr)</FormHelperText>
            </FormGroup>

            <div style={{ marginTop: 10, paddingLeft: 10, opacity: EnableTorznabSearch ? 1 : 0.5, pointerEvents: EnableTorznabSearch ? 'auto' : 'none' }}>
                <List dense>
                    {(TorznabUrls || []).map((url, index) => (
                        <ListItem key={`${url.Host}-${url.Key}`}>
                            <ListItemText
                                primary={url.Name || url.Host}
                                secondary={
                                    <>
                                        {url.Name && <Typography component="span" variant="body2" display="block" color="textSecondary">{url.Host}</Typography>}
                                        {`Key: ${url.Key.substring(0, 5)}...`}
                                    </>
                                }
                            />
                            <ListItemSecondaryAction>
                                <IconButton edge="end" aria-label="delete" onClick={() => handleDelete(index)}>
                                    <DeleteIcon />
                                </IconButton>
                            </ListItemSecondaryAction>
                        </ListItem>
                    ))}
                </List>

                <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginTop: 10 }}>
                    <TextField
                        label="Name (Optional)"
                        value={newName}
                        onChange={(e) => setNewName(e.target.value)}
                        placeholder="My Tracker"
                        variant="outlined"
                        size="small"
                    />
                    <TextField
                        label="Torznab Host URL"
                        value={newHost}
                        onChange={(e) => setNewHost(e.target.value)}
                        placeholder="http://localhost:9117"
                        variant="outlined"
                        size="small"
                    />
                    <TextField
                        label="API Key"
                        value={newKey}
                        onChange={(e) => setNewKey(e.target.value)}
                        variant="outlined"
                        size="small"
                    />
                    <div style={{ display: 'flex', marginTop: 10 }}>
                        <Button
                            variant="outlined"
                            color="secondary"
                            onClick={handleTest}
                            disabled={!newHost || !newKey || testing}
                            style={{ marginRight: 10 }}
                        >
                            {testing ? <CircularProgress size={24} color="inherit" /> : 'Test'}
                        </Button>
                        <Button variant="contained" color="secondary" onClick={handleAdd} disabled={!newHost || !newKey}>
                            Add Server
                        </Button>
                    </div>
                    {testResult && (
                        <Typography variant="caption" style={{ color: testResult.success ? 'green' : 'red' }}>
                            {testResult.msg}
                        </Typography>
                    )}
                </div>
            </div>
            <br />
        </>
    )
}

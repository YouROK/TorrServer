import React, { useState } from 'react'
import { TextField, Button, List, ListItem, ListItemText, CircularProgress, Typography, Divider, ListItemSecondaryAction, IconButton } from '@material-ui/core'
import { useTranslation } from 'react-i18next'
import axios from 'axios'
import { torznabSearchHost } from 'utils/Hosts'
import { AddCircleOutline as AddIcon } from '@material-ui/icons'

export default function TorznabSearch({ onSelect }) {
  const { t } = useTranslation()
  const [query, setQuery] = useState('')
  const [results, setResults] = useState([])
  const [loading, setLoading] = useState(false)
  const [searched, setSearched] = useState(false)

  const handleSearch = async () => {
    if (!query) return
    setLoading(true)
    setSearched(true)
    try {
      const { data } = await axios.get(torznabSearchHost(), { params: { query } })
      setResults(data || [])
    } catch (error) {
      console.error(error)
      setResults([])
    } finally {
      setLoading(false)
    }
  }

  const handleKeyDown = (e) => {
    if (e.key === 'Enter') {
      handleSearch()
    }
  }

  return (
    <div style={{ marginTop: '1.5em' }}>
      <div style={{ display: 'flex', gap: '8px' }}>
        <TextField
          label={t('Torznab.SearchTorznab')}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          variant="outlined"
          size="small"
          fullWidth
          placeholder={t('Torznab.SearchMoviesShows')}
        />
        <Button variant="contained" color="primary" onClick={handleSearch} disabled={loading} style={{ minWidth: '80px' }}>
          {loading ? <CircularProgress size={24} color="inherit" /> : t('Torznab.SearchTorrents')}
        </Button>
      </div>
      {searched && (
        <div style={{ maxHeight: '200px', overflowY: 'auto', marginTop: '8px', border: '1px solid rgba(0,0,0,0.12)', borderRadius: '4px' }}>
          {results.length === 0 ? (
            <div style={{ padding: '8px', textAlign: 'center' }}>
              <Typography variant="body2">{loading ? t('Torznab.SearchTorrents') : t('Torznab.NoResultsFound')}</Typography>
            </div>
          ) : (
            <List dense>
              {results.map((item, index) => (
                <React.Fragment key={item.Hash || item.Link || index}>
                  <ListItem button onClick={() => onSelect(item.Magnet || item.Link)}>
                    <ListItemText
                      primary={item.Title}
                      secondary={`${item.Size} • S:${item.Seed} P:${item.Peer}`}
                      primaryTypographyProps={{ noWrap: true, style: { fontSize: '0.9rem' } }}
                    />
                    <ListItemSecondaryAction>
                      <IconButton edge="end" aria-label="add" onClick={() => onSelect(item.Magnet || item.Link)}>
                        <AddIcon />
                      </IconButton>
                    </ListItemSecondaryAction>
                  </ListItem>
                  <Divider />
                </React.Fragment>
              ))}
            </List>
          )}
        </div>
      )}
    </div>
  )
}

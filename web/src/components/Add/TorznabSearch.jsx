import React, { useState, useMemo } from 'react'
import {
  TextField,
  Button,
  List,
  ListItem,
  ListItemText,
  CircularProgress,
  Typography,
  Divider,
  ListItemSecondaryAction,
  IconButton,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  useMediaQuery,
} from '@material-ui/core'
import { useTranslation } from 'react-i18next'
import axios from 'axios'
import { torznabSearchHost } from 'utils/Hosts'
import { AddCircleOutline as AddIcon, ArrowUpward, ArrowDownward } from '@material-ui/icons'
import { parseSizeToBytes, formatSizeToClassicUnits } from 'utils/Utils'

export default function TorznabSearch({ onSelect }) {
  const { t } = useTranslation()
  const [query, setQuery] = useState('')
  const [results, setResults] = useState([])
  const [loading, setLoading] = useState(false)
  const [searched, setSearched] = useState(false)
  const [sortField, setSortField] = useState('') // '', 'size', 'seeds', 'peers'
  const [sortDirection, setSortDirection] = useState('desc') // 'asc' or 'desc'
  const isMobile = useMediaQuery('(max-width:600px)')

  const handleSearch = async () => {
    if (!query) return
    setLoading(true)
    setSearched(true)
    try {
      const { data } = await axios.get(torznabSearchHost(), { params: { query } })
      setResults(data || [])
    } catch (error) {
      setResults([])
    } finally {
      setLoading(false)
    }
  }

  const handleKeyDown = e => {
    if (e.key === 'Enter') {
      handleSearch()
    }
  }

  const toggleSortDirection = () => {
    setSortDirection(prev => (prev === 'asc' ? 'desc' : 'asc'))
  }

  const sortedResults = useMemo(() => {
    if (!sortField || results.length === 0) return results

    const sorted = [...results].sort((a, b) => {
      let aVal
      let bVal

      switch (sortField) {
        case 'size':
          aVal = parseSizeToBytes(a.Size || '0')
          bVal = parseSizeToBytes(b.Size || '0')
          break
        case 'seeds':
          aVal = a.Seed || 0
          bVal = b.Seed || 0
          break
        case 'peers':
          aVal = a.Peer || 0
          bVal = b.Peer || 0
          break
        default:
          return 0
      }

      if (aVal === bVal) return 0
      return sortDirection === 'asc' ? (aVal < bVal ? -1 : 1) : aVal > bVal ? -1 : 1
    })

    return sorted
  }, [results, sortField, sortDirection])

  return (
    <div style={{ marginTop: '1.5em' }}>
      <div style={{ display: 'flex', gap: '8px', flexWrap: isMobile ? 'wrap' : 'nowrap' }}>
        <TextField
          label={t('Torznab.SearchTorznab')}
          value={query}
          onChange={e => setQuery(e.target.value)}
          onKeyDown={handleKeyDown}
          variant='outlined'
          size='small'
          fullWidth
          placeholder={t('Torznab.SearchMoviesShows')}
          style={{ flex: isMobile ? '1 1 100%' : '1' }}
        />
        <Button
          variant='contained'
          color='primary'
          onClick={handleSearch}
          disabled={loading}
          style={{
            minWidth: isMobile ? '100%' : '80px',
            flex: isMobile ? '1 1 100%' : '0 0 auto',
          }}
        >
          {loading ? <CircularProgress size={24} color='inherit' /> : t('Torznab.SearchTorrents')}
        </Button>
      </div>
      {searched && (
        <div style={{ marginTop: '8px' }}>
          {results.length > 0 && (
            <div
              style={{
                display: 'flex',
                gap: isMobile ? '8px' : '4px',
                marginBottom: '12px',
                alignItems: 'center',
                padding: isMobile ? '12px 8px' : '8px 12px',
                backgroundColor: 'rgba(0, 0, 0, 0.02)',
                borderRadius: '4px',
                border: '1px solid rgba(0, 0, 0, 0.08)',
                flexWrap: isMobile ? 'wrap' : 'nowrap',
              }}
            >
              <FormControl
                variant='outlined'
                size='small'
                style={{
                  minWidth: isMobile ? '100%' : 140,
                  flexShrink: 0,
                  flex: isMobile ? '1 1 100%' : '0 0 auto',
                }}
              >
                <InputLabel>{t('Torznab.SortBy')}</InputLabel>
                <Select value={sortField} onChange={e => setSortField(e.target.value)} label={t('Torznab.SortBy')}>
                  <MenuItem value=''>{t('Torznab.SortByNone')}</MenuItem>
                  <MenuItem value='size'>{t('Torznab.SortBySize')}</MenuItem>
                  <MenuItem value='seeds'>{t('Torznab.SortBySeeds')}</MenuItem>
                  <MenuItem value='peers'>{t('Torznab.SortByPeers')}</MenuItem>
                </Select>
              </FormControl>
              {sortField && (
                <IconButton
                  size='small'
                  onClick={toggleSortDirection}
                  title={sortDirection === 'asc' ? t('Torznab.SortAscending') : t('Torznab.SortDescending')}
                  style={{
                    marginLeft: isMobile ? 'auto' : '4px',
                    padding: '8px',
                  }}
                >
                  {sortDirection === 'asc' ? <ArrowUpward /> : <ArrowDownward />}
                </IconButton>
              )}
            </div>
          )}
          <div
            style={{
              maxHeight: isMobile ? '300px' : '200px',
              overflowY: 'auto',
              border: '1px solid rgba(0,0,0,0.12)',
              borderRadius: '4px',
            }}
          >
            {results.length === 0 ? (
              <div style={{ padding: '8px', textAlign: 'center' }}>
                <Typography variant='body2'>
                  {loading ? t('Torznab.SearchTorrents') : t('Torznab.NoResultsFound')}
                </Typography>
              </div>
            ) : (
              <List dense>
                {sortedResults.map((item, index) => {
                  const sizeBytes = parseSizeToBytes(item.Size || '0')
                  const formattedSize = formatSizeToClassicUnits(sizeBytes)
                  return (
                    <React.Fragment key={item.Hash || item.Link || index}>
                      <ListItem button onClick={() => onSelect(item.Magnet || item.Link)}>
                        <ListItemText
                          primary={item.Title}
                          secondary={`${formattedSize} • S:${item.Seed || 0} P:${item.Peer || 0}`}
                          primaryTypographyProps={{
                            noWrap: !isMobile,
                            style: {
                              fontSize: isMobile ? '0.85rem' : '0.9rem',
                              whiteSpace: isMobile ? 'normal' : 'nowrap',
                            },
                          }}
                          secondaryTypographyProps={{
                            style: {
                              fontSize: isMobile ? '0.75rem' : '0.8rem',
                            },
                          }}
                        />
                        <ListItemSecondaryAction>
                          <IconButton
                            edge='end'
                            aria-label='add'
                            onClick={() => onSelect(item.Magnet || item.Link)}
                            size={isMobile ? 'small' : 'medium'}
                          >
                            <AddIcon />
                          </IconButton>
                        </ListItemSecondaryAction>
                      </ListItem>
                      <Divider />
                    </React.Fragment>
                  )
                })}
              </List>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

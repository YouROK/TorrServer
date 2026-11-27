import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Box, Button, Collapse, FormHelperText, IconButton, Switch, TextField } from '@material-ui/core'
import ArrowUpwardIcon from '@material-ui/icons/ArrowUpward'
import ArrowDownwardIcon from '@material-ui/icons/ArrowDownward'
import DeleteIcon from '@material-ui/icons/Delete'
import FileCopyIcon from '@material-ui/icons/FileCopy'
import {
  BUILTIN_FILE_NAME_RULES,
  DEFAULT_FILE_NAME_TEMPLATE,
  getFileNameConfig,
  getFileNameRuleError,
} from 'utils/FileNameRules'
import { getFileDisplayName } from 'components/DialogTorrentDetailsContent/helpers'

import { SettingSectionLabel, SettingSubsectionLabel, SettingSubsubsectionLabel } from './style'

const SAMPLE = '[Yousei-raws] Watamote 01 [BDrip 1920x1080 x264 FLAC].mkv'

const NEW_RULE = { Name: '', Enabled: true, Mode: 'replace', Pattern: '', Replacement: '', Flags: 'i' }

const MODE_LABELS = {
  replace: 'FileNameRules.ModeReplace',
  extract: 'FileNameRules.ModeExtract',
  capture: 'FileNameRules.ModeCapture',
}
const MODE_HINTS = {
  replace: 'FileNameRules.ReplaceHint',
  extract: 'FileNameRules.ExtractHint',
  capture: 'FileNameRules.CaptureHint',
}

const fieldsetStyle = { border: 0, margin: 0, padding: 0, minWidth: 0 }

const monospace = { style: { fontFamily: 'monospace' } }
const codeStyle = { fontFamily: 'monospace', fontSize: 12, wordBreak: 'break-all' }

function BuiltinRule({ rule, onCopy }) {
  const { t } = useTranslation()
  const modeLabel = t(MODE_LABELS[rule.Mode] || MODE_LABELS.replace)
  const isReplace = (rule.Mode || 'replace') === 'replace'
  const replacement = isReplace ? `${t('FileNameRules.Replacement')}: "${rule.Replacement || ''}"  ` : ''
  return (
    <Box border={1} borderColor='divider' borderRadius={4} p={1} mb={1}>
      <Box display='flex' alignItems='center' style={{ gap: 8 }}>
        <Box flex={1}>
          <b>{rule.Name}</b> <small>({modeLabel})</small>
        </Box>
        <IconButton size='small' onClick={onCopy} title={t('FileNameRules.CopyOne')}>
          <FileCopyIcon fontSize='small' />
        </IconButton>
      </Box>
      <Box style={codeStyle}>{rule.Pattern}</Box>
      <Box style={codeStyle}>
        {replacement}
        {t('FileNameRules.Flags')}: {rule.Flags || '-'}
      </Box>
    </Box>
  )
}

export default function FileNameRulesSettings({ settings, updateSettings }) {
  const { t } = useTranslation()
  const rules = settings?.FileNameRules || []
  const builtinEnabled = !settings?.FileNameBuiltinRulesDisabled
  const customEnabled = !settings?.FileNameCustomRulesDisabled
  const asFileName = Boolean(settings?.FileNameAsTitle)
  const usePtt = Boolean(settings?.FileNameUsePtt)
  const [showBuiltin, setShowBuiltin] = useState(false)
  const [sample, setSample] = useState(SAMPLE)

  const setRules = next => updateSettings({ FileNameRules: next })
  const updateRule = (index, patch) => setRules(rules.map((rule, i) => (i === index ? { ...rule, ...patch } : rule)))
  const removeRule = index => setRules(rules.filter((_, i) => i !== index))
  const copyToCustom = list => setRules([...rules, ...list.map(rule => ({ ...rule }))])
  const moveRule = (index, delta) => {
    const target = index + delta
    if (target < 0 || target >= rules.length) return
    const next = [...rules]
    const item = next[index]
    next[index] = next[target]
    next[target] = item
    setRules(next)
  }

  return (
    <>
      <SettingSectionLabel style={{ marginTop: '20px' }}>{t('FileNameRules.Title')}</SettingSectionLabel>
      <FormHelperText style={{ marginLeft: 0, marginBottom: '12px' }}>{t('FileNameRules.Hint')}</FormHelperText>

      <Box display='flex' alignItems='center' style={{ gap: 8 }}>
        <Switch
          checked={asFileName}
          onChange={({ target: { checked } }) => updateSettings({ FileNameAsTitle: checked })}
          color='secondary'
        />
        <Box flex={1}>
          <b>{t('FileNameRules.AsFileName')}</b>
        </Box>
      </Box>
      <FormHelperText style={{ marginLeft: 0, marginBottom: '8px' }}>
        {t('FileNameRules.AsFileNameHint')}
      </FormHelperText>

      <Box component='fieldset' disabled={asFileName} style={{ ...fieldsetStyle, opacity: asFileName ? 0.5 : 1 }}>
        <SettingSubsectionLabel>{t('FileNameRules.ParseMethod')}</SettingSubsectionLabel>
        <TextField
          select
          size='small'
          variant='outlined'
          fullWidth
          inputProps={{ 'aria-label': t('FileNameRules.ParseMethod') }}
          SelectProps={{ native: true }}
          value={settings?.FileNameUsePtt ? 'ptt' : 'rules'}
          onChange={({ target: { value } }) => updateSettings({ FileNameUsePtt: value === 'ptt' })}
        >
          <option value='rules'>{t('FileNameRules.ParserRules')}</option>
          <option value='ptt'>{t('FileNameRules.ParserPtt')}</option>
        </TextField>
        <FormHelperText style={{ marginLeft: 0, marginBottom: '12px' }}>{t('FileNameRules.ParserHint')}</FormHelperText>

        <Box component='fieldset' disabled={usePtt} style={{ ...fieldsetStyle, opacity: usePtt ? 0.5 : 1 }}>
          <SettingSubsubsectionLabel>{t('FileNameRules.RulesTitle')}</SettingSubsubsectionLabel>
          <FormHelperText style={{ marginLeft: 0, marginBottom: '12px' }}>
            {t('FileNameRules.RulesHint')}
          </FormHelperText>

          {/* Built-in rules: read-only */}
          <Box display='flex' alignItems='center' style={{ gap: 8 }}>
            <Switch
              checked={builtinEnabled}
              onChange={({ target: { checked } }) => updateSettings({ FileNameBuiltinRulesDisabled: !checked })}
              color='secondary'
            />
            <Box flex={1}>
              <b>{t('FileNameRules.BuiltinTitle')}</b> ({BUILTIN_FILE_NAME_RULES.length})
            </Box>
            <Button size='small' onClick={() => setShowBuiltin(!showBuiltin)}>
              {t(showBuiltin ? 'FileNameRules.Hide' : 'FileNameRules.Show')}
            </Button>
          </Box>
          <FormHelperText style={{ marginLeft: 0, marginBottom: '8px' }}>
            {t('FileNameRules.BuiltinHint')}
          </FormHelperText>
          <Collapse in={showBuiltin} unmountOnExit>
            <Box style={{ opacity: builtinEnabled ? 1 : 0.5 }}>
              {BUILTIN_FILE_NAME_RULES.map(rule => (
                <BuiltinRule key={rule.Name} rule={rule} onCopy={() => copyToCustom([rule])} />
              ))}
            </Box>
            <Button
              variant='outlined'
              color='secondary'
              size='small'
              onClick={() => copyToCustom(BUILTIN_FILE_NAME_RULES)}
            >
              {t('FileNameRules.CopyAll')}
            </Button>
          </Collapse>

          {/* Custom rules: editable */}
          <Box display='flex' alignItems='center' style={{ gap: 8, marginTop: 16 }}>
            <Switch
              checked={customEnabled}
              onChange={({ target: { checked } }) => updateSettings({ FileNameCustomRulesDisabled: !checked })}
              color='secondary'
            />
            <Box flex={1}>
              <b>{t('FileNameRules.CustomTitle')}</b> ({rules.length})
            </Box>
          </Box>
          <FormHelperText style={{ marginLeft: 0, marginBottom: '8px' }}>
            {t('FileNameRules.CustomHint')}
          </FormHelperText>

          <Box style={{ opacity: customEnabled ? 1 : 0.5 }}>
            {rules.map((rule, index) => {
              const error = getFileNameRuleError(rule)
              const withoutReplacement = rule.Mode === 'extract' || rule.Mode === 'capture'
              return (
                // eslint-disable-next-line react/no-array-index-key
                <Box key={index} border={1} borderColor='divider' borderRadius={4} p={1.5} mb={1.5}>
                  <Box display='flex' alignItems='center' flexWrap='wrap' style={{ gap: 8 }}>
                    <Switch
                      checked={Boolean(rule.Enabled)}
                      onChange={({ target: { checked } }) => updateRule(index, { Enabled: checked })}
                      color='secondary'
                    />
                    <TextField
                      size='small'
                      variant='outlined'
                      label={t('FileNameRules.RuleName')}
                      value={rule.Name || ''}
                      onChange={({ target: { value } }) => updateRule(index, { Name: value })}
                      style={{ flex: '1 1 160px' }}
                    />
                    <TextField
                      select
                      size='small'
                      variant='outlined'
                      label={t('FileNameRules.Mode')}
                      SelectProps={{ native: true }}
                      value={rule.Mode || 'replace'}
                      onChange={({ target: { value } }) => updateRule(index, { Mode: value })}
                    >
                      <option value='replace'>{t('FileNameRules.ModeReplace')}</option>
                      <option value='extract'>{t('FileNameRules.ModeExtract')}</option>
                      <option value='capture'>{t('FileNameRules.ModeCapture')}</option>
                    </TextField>
                    <IconButton
                      size='small'
                      disabled={index === 0}
                      onClick={() => moveRule(index, -1)}
                      title={t('FileNameRules.MoveUp')}
                    >
                      <ArrowUpwardIcon fontSize='small' />
                    </IconButton>
                    <IconButton
                      size='small'
                      disabled={index === rules.length - 1}
                      onClick={() => moveRule(index, 1)}
                      title={t('FileNameRules.MoveDown')}
                    >
                      <ArrowDownwardIcon fontSize='small' />
                    </IconButton>
                    <IconButton size='small' onClick={() => removeRule(index)} title={t('FileNameRules.Delete')}>
                      <DeleteIcon fontSize='small' />
                    </IconButton>
                  </Box>

                  <TextField
                    margin='dense'
                    variant='outlined'
                    fullWidth
                    label={t('FileNameRules.Pattern')}
                    value={rule.Pattern || ''}
                    onChange={({ target: { value } }) => updateRule(index, { Pattern: value })}
                    error={Boolean(error)}
                    helperText={error || t(MODE_HINTS[rule.Mode] || MODE_HINTS.replace)}
                    inputProps={monospace}
                  />

                  <Box display='flex' style={{ gap: 8 }}>
                    <TextField
                      margin='dense'
                      variant='outlined'
                      label={t('FileNameRules.Replacement')}
                      value={rule.Replacement || ''}
                      onChange={({ target: { value } }) => updateRule(index, { Replacement: value })}
                      disabled={withoutReplacement}
                      inputProps={monospace}
                      style={{ flex: 1 }}
                    />
                    <TextField
                      margin='dense'
                      variant='outlined'
                      label={t('FileNameRules.Flags')}
                      value={rule.Flags || ''}
                      onChange={({ target: { value } }) => updateRule(index, { Flags: value })}
                      inputProps={monospace}
                      style={{ width: 100 }}
                    />
                  </Box>
                </Box>
              )
            })}
          </Box>

          <Box mb={2}>
            <Button variant='outlined' color='secondary' onClick={() => setRules([...rules, { ...NEW_RULE }])}>
              {t('FileNameRules.Add')}
            </Button>
          </Box>

          <TextField
            margin='dense'
            variant='outlined'
            fullWidth
            label={t('FileNameRules.Template')}
            placeholder={DEFAULT_FILE_NAME_TEMPLATE}
            value={settings?.FileNameTemplate || ''}
            onChange={({ target: { value } }) => updateSettings({ FileNameTemplate: value })}
            inputProps={monospace}
          />
          <FormHelperText style={{ marginLeft: 0, marginBottom: '12px' }}>
            {t('FileNameRules.TemplateHint')}
          </FormHelperText>

          <TextField
            margin='dense'
            variant='outlined'
            fullWidth
            label={t('FileNameRules.Test')}
            value={sample}
            onChange={({ target: { value } }) => setSample(value)}
          />
          <FormHelperText style={{ marginLeft: 0, fontFamily: 'monospace' }}>
            {t('FileNameRules.TestResult')}: {sample ? getFileDisplayName(sample, getFileNameConfig(settings)) : ''}
          </FormHelperText>
        </Box>
      </Box>
    </>
  )
}

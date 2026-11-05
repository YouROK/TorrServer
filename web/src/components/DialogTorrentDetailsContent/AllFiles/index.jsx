import { memo, useMemo, useState } from 'react'
import { Button, IconButton } from '@material-ui/core'
import ChevronRightIcon from '@material-ui/icons/ChevronRight'
import ExpandMoreIcon from '@material-ui/icons/ExpandMore'
import FolderIcon from '@material-ui/icons/Folder'
import FolderOpenIcon from '@material-ui/icons/FolderOpen'
import GetAppIcon from '@material-ui/icons/GetApp'
import InsertDriveFileOutlinedIcon from '@material-ui/icons/InsertDriveFileOutlined'
import { useTranslation } from 'react-i18next'
import { streamHost } from 'utils/Hosts'
import { humanizeSize } from 'utils/Utils'

import { TreeRow, TreeToolbar, TreeWrapper } from './style'

const byName = (a, b) => a.name.localeCompare(b.name, undefined, { numeric: true, sensitivity: 'base' })

const createDir = (name, path) => ({ name, path, dirs: new Map(), files: [], size: 0, count: 0 })

// builds directory tree from the flat file list (file_stats)
const buildTree = files => {
  const root = createDir('', '')

  files.forEach(({ id, path, length }) => {
    const size = length || 0
    const parts = (path || '').split(/[\\/]+/).filter(Boolean)
    const fileName = parts.pop() || String(id)

    let node = root
    node.size += size
    node.count += 1

    parts.forEach(part => {
      if (!node.dirs.has(part)) node.dirs.set(part, createDir(part, `${node.path}/${part}`))
      node = node.dirs.get(part)
      node.size += size
      node.count += 1
    })

    node.files.push({ id, name: fileName, size })
  })

  return root
}

// by default open first level and every chain of directories which contain nothing but one directory
const collectDefaultOpen = (node, parentOpen, depth, out) => {
  node.dirs.forEach(dir => {
    const open = depth === 0 || (parentOpen && node.dirs.size === 1 && node.files.length === 0)
    if (open) out.add(dir.path)
    collectDefaultOpen(dir, open, depth + 1, out)
  })
  return out
}

const collectDirPaths = (node, out = []) => {
  node.dirs.forEach(dir => {
    out.push(dir.path)
    collectDirPaths(dir, out)
  })
  return out
}

const flattenTree = (node, depth, isOpen, out = []) => {
  ;[...node.dirs.values()].sort(byName).forEach(dir => {
    const open = isOpen(dir)
    out.push({ type: 'dir', key: `d:${dir.path}`, depth, dir, open })
    if (open) flattenTree(dir, depth + 1, isOpen, out)
  })
  ;[...node.files].sort(byName).forEach(file => out.push({ type: 'file', key: `f:${file.id}`, depth, file }))
  return out
}

const AllFiles = memo(({ files, hash }) => {
  const { t } = useTranslation()
  // user choices: { [dirPath]: boolean }, everything else is opened/closed by default rules
  const [overrides, setOverrides] = useState({})

  const tree = useMemo(() => buildTree(files || []), [files])
  const defaultOpen = useMemo(() => collectDefaultOpen(tree, true, 0, new Set()), [tree])
  const isOpen = dir => overrides[dir.path] ?? defaultOpen.has(dir.path)
  const rows = flattenTree(tree, 0, isOpen)

  const toggleDir = dir =>
    setOverrides(prev => ({ ...prev, [dir.path]: !(prev[dir.path] ?? defaultOpen.has(dir.path)) }))
  const setAll = value => setOverrides(Object.fromEntries(collectDirPaths(tree).map(path => [path, value])))

  const getFileLink = (name, id) => `${streamHost()}/${encodeURIComponent(name)}?link=${hash}&index=${id}&play`

  const formatSize = size => humanizeSize(size) || `0 ${t('B')}`

  if (!files?.length) return null

  return (
    <>
      <TreeToolbar>
        <div className='tree-summary'>
          {files.length} / {formatSize(tree.size)}
        </div>
        {tree.dirs.size > 0 && (
          <div className='tree-buttons'>
            <Button onClick={() => setAll(true)} variant='outlined' color='primary' size='small'>
              {t('ExpandAll')}
            </Button>
            <Button onClick={() => setAll(false)} variant='outlined' color='primary' size='small'>
              {t('CollapseAll')}
            </Button>
          </div>
        )}
      </TreeToolbar>

      <TreeWrapper role='tree'>
        {rows.map(row => {
          const indent = { paddingLeft: 12 + row.depth * 20 }

          if (row.type === 'dir') {
            const { dir, open } = row
            return (
              <TreeRow
                key={row.key}
                isDir
                role='treeitem'
                aria-expanded={open}
                tabIndex={0}
                style={indent}
                onClick={() => toggleDir(dir)}
                onKeyDown={e => {
                  if (e.key === 'Enter' || e.key === ' ') {
                    e.preventDefault()
                    toggleDir(dir)
                  }
                }}
              >
                {open ? <ExpandMoreIcon /> : <ChevronRightIcon />}
                {open ? <FolderOpenIcon /> : <FolderIcon />}
                <span className='tree-name' title={dir.name}>
                  {dir.name}
                </span>
                <span className='tree-count'>({dir.count})</span>
                <span className='tree-size'>{formatSize(dir.size)}</span>
                <span className='tree-action' />
              </TreeRow>
            )
          }

          const { file } = row
          return (
            <TreeRow key={row.key} role='treeitem' style={{ ...indent, paddingLeft: indent.paddingLeft + 26 }}>
              <InsertDriveFileOutlinedIcon />
              <span className='tree-name' title={file.name}>
                {file.name}
              </span>
              <span className='tree-size'>{formatSize(file.size)}</span>
              <span className='tree-action'>
                <IconButton
                  component='a'
                  href={getFileLink(file.name, file.id)}
                  download={file.name}
                  size='small'
                  title={t('Download')}
                  aria-label={t('Download')}
                >
                  <GetAppIcon />
                </IconButton>
              </span>
            </TreeRow>
          )
        })}
      </TreeWrapper>
    </>
  )
})

export default AllFiles

import { Button, Label, ListBox, Select, useMediaQuery } from '@heroui/react'
import { Folder, FolderOpen } from 'lucide-react'
import { iconMenu } from 'shared/ui/iconProps'
import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import type { PlayableFile, TorrentFileStat } from 'shared/api/types'
import { queryMax } from 'shared/theme/breakpoints'

import FilesDataGrid from './FilesDataGrid'
import {
  buildDirectoryTree,
  directoryTreeIsUseful,
  filesInSelectedFolder,
  flattenFolders,
  type DirectoryNode,
} from './fileTree'

interface FileBrowserProps {
  playableFileList: PlayableFile[]
  viewedFileList?: number[]
  selectedSeason?: number
  seasonAmount?: number[] | null
  hash: string
  allFileStats?: TorrentFileStat[]
  onViewedChange?: () => void
}

function DirectoryTreeList({
  node,
  selectedFolderId,
  onSelect,
  depth = 0,
}: {
  node: DirectoryNode
  selectedFolderId: string
  onSelect: (id: string) => void
  depth?: number
}) {
  return (
    <>
      {[...node.children.values()].map(child => {
        const isSelected = selectedFolderId === child.id
        return (
          <div key={child.id}>
            <Button
              variant={isSelected ? 'primary' : 'ghost'}
              size='sm'
              className='mb-1 w-full justify-start gap-2'
              style={{ paddingLeft: `${8 + depth * 12}px` }}
              onPress={() => onSelect(child.id)}
            >
              {isSelected ? (
                <FolderOpen {...iconMenu} className='shrink-0' aria-hidden />
              ) : (
                <Folder {...iconMenu} className='shrink-0' aria-hidden />
              )}
              <span className='truncate'>{child.label}</span>
            </Button>
            <DirectoryTreeList node={child} selectedFolderId={selectedFolderId} onSelect={onSelect} depth={depth + 1} />
          </div>
        )
      })}
    </>
  )
}

/** Multi-file torrent browser: optional folder tree sidebar plus the file list/table. */
export default function FileBrowser({
  playableFileList,
  viewedFileList,
  selectedSeason,
  seasonAmount,
  hash,
  allFileStats,
  onViewedChange,
}: FileBrowserProps) {
  const { t } = useTranslation()
  const isCompact = useMediaQuery(queryMax('mobile'))
  const tree = useMemo(() => buildDirectoryTree(playableFileList), [playableFileList])
  // A single release folder (common for series packs) isn't worth a sidebar — only show the
  // tree when there are multiple folders or nested structure to navigate.
  const hasUsefulFolders = directoryTreeIsUseful(tree)
  const [selectedFolderId, setSelectedFolderId] = useState('root')
  const folderOptions = useMemo(() => flattenFolders(tree), [tree])

  const visibleFiles = useMemo(
    () => filesInSelectedFolder(playableFileList, tree, selectedFolderId),
    [selectedFolderId, tree, playableFileList],
  )

  return (
    <div className={`grid min-h-[200px] w-full gap-3 ${hasUsefulFolders ? 'md:grid-cols-[200px_1fr]' : 'grid-cols-1'}`}>
      {hasUsefulFolders ? (
        isCompact ? (
          <Select
            selectedKey={selectedFolderId}
            onSelectionChange={key => {
              if (key != null) setSelectedFolderId(String(key))
            }}
            className='w-full'
            aria-label={t('Folders')}
          >
            <Label className='sr-only'>{t('Folders')}</Label>
            <Select.Trigger className='min-h-11 w-full'>
              <Select.Value />
              <Select.Indicator />
            </Select.Trigger>
            <Select.Popover>
              <ListBox>
                <ListBox.Item id='root' textValue={t('AllFiles')}>
                  {t('AllFiles')}
                </ListBox.Item>
                {folderOptions.map(folder => (
                  <ListBox.Item key={folder.id} id={folder.id} textValue={folder.label}>
                    <span style={{ paddingLeft: folder.depth * 12 }}>{folder.label}</span>
                  </ListBox.Item>
                ))}
              </ListBox>
            </Select.Popover>
          </Select>
        ) : (
          <div className='rounded-xl border border-border bg-surface-secondary p-2 md:rounded-none md:border-0 md:border-r md:border-border md:bg-transparent md:p-0 md:pr-3'>
            <p className='mb-1.5 px-1 text-xs font-medium uppercase tracking-wide text-muted'>{t('Folders')}</p>
            <Button
              variant={selectedFolderId === 'root' ? 'primary' : 'ghost'}
              size='sm'
              className='mb-1 w-full justify-start'
              onPress={() => setSelectedFolderId('root')}
            >
              {t('AllFiles')}
            </Button>
            <DirectoryTreeList node={tree} selectedFolderId={selectedFolderId} onSelect={setSelectedFolderId} />
          </div>
        )
      ) : null}

      <FilesDataGrid
        playableFileList={visibleFiles}
        viewedFileList={viewedFileList}
        selectedSeason={selectedSeason}
        seasonAmount={seasonAmount}
        hash={hash}
        allFileStats={allFileStats}
        onViewedChange={onViewedChange}
      />
    </div>
  )
}

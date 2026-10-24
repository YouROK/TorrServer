import type { PlayableFile } from 'shared/api/types'

export interface DirectoryNode {
  id: string
  label: string
  children: Map<string, DirectoryNode>
  files: PlayableFile[]
}

/** Nested folders from torrent paths. A single path segment stays on the root. */
export function buildDirectoryTree(files: PlayableFile[]): DirectoryNode {
  const root: DirectoryNode = { id: 'root', label: '/', children: new Map(), files: [] }
  for (const file of files) {
    const segments = file.path.replace(/\\/g, '/').split('/').filter(Boolean)
    if (segments.length <= 1) {
      root.files.push(file)
      continue
    }
    let node = root
    for (let i = 0; i < segments.length - 1; i++) {
      const segment = segments[i]
      const id = `${node.id}/${segment}`
      if (!node.children.has(segment)) {
        node.children.set(segment, { id, label: segment, children: new Map(), files: [] })
      }
      node = node.children.get(segment)!
    }
    node.files.push(file)
  }
  return root
}

export function collectFilesRecursively(node: DirectoryNode): PlayableFile[] {
  const collected = [...node.files]
  for (const child of node.children.values()) collected.push(...collectFilesRecursively(child))
  return collected
}

export function findNodeById(node: DirectoryNode, id: string): DirectoryNode | null {
  if (node.id === id) return node
  for (const child of node.children.values()) {
    const found = findNodeById(child, id)
    if (found) return found
  }
  return null
}

/** One release folder is not worth a sidebar. */
export function directoryTreeIsUseful(tree: DirectoryNode): boolean {
  return (
    tree.children.size > 1 ||
    [...tree.children.values()].some(child => child.children.size > 0 || child.files.length === 0)
  )
}

export function flattenFolders(node: DirectoryNode, depth = 0): { id: string; label: string; depth: number }[] {
  const rows: { id: string; label: string; depth: number }[] = []
  for (const child of node.children.values()) {
    rows.push({ id: child.id, label: child.label, depth })
    rows.push(...flattenFolders(child, depth + 1))
  }
  return rows
}

/** Files shown for a folder id. Root and a useless tree return the full list. */
export function filesInSelectedFolder(files: PlayableFile[], tree: DirectoryNode, folderId: string): PlayableFile[] {
  if (!directoryTreeIsUseful(tree) || folderId === 'root') return files
  const node = findNodeById(tree, folderId)
  return node ? collectFilesRecursively(node) : files
}

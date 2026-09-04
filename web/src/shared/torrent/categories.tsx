import { Clapperboard, MoreHorizontal, Music, Tv } from 'lucide-react'
import type { ReactNode } from 'react'

export interface TorrentCategoryItem {
  key: string
  name: string
  icon: ReactNode
}

export const TORRENT_CATEGORIES: TorrentCategoryItem[] = [
  { key: 'movie', name: 'Movies', icon: <Clapperboard className='size-5' aria-hidden /> },
  { key: 'tv', name: 'Series', icon: <Tv className='size-5' aria-hidden /> },
  { key: 'music', name: 'Music', icon: <Music className='size-5' aria-hidden /> },
  { key: 'other', name: 'Other', icon: <MoreHorizontal className='size-5' aria-hidden /> },
]

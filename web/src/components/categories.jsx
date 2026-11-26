import MovieCreationIcon from '@material-ui/icons/MovieCreation'
import LiveTvIcon from '@material-ui/icons/LiveTv'
import MusicNoteIcon from '@material-ui/icons/MusicNote'
import CategoryIcon from '@material-ui/icons/Category'

export const TORRENT_CATEGORIES = [
  { key: 'movie', name: 'Movies', icon: <MovieCreationIcon /> },
  { key: 'tv', name: 'Series', icon: <LiveTvIcon /> },
  { key: 'music', name: 'Music', icon: <MusicNoteIcon /> },
  { key: 'other', name: 'Other', icon: <CategoryIcon /> },
]

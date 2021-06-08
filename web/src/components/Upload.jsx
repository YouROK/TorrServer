import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import ListItem from '@material-ui/core/ListItem'
import PublishIcon from '@material-ui/icons/Publish'
import { torrentUploadHost } from 'utils/Hosts'
import axios from 'axios'
import { useTranslation } from 'react-i18next'

export default function UploadDialog() {
  const { t } = useTranslation()
  const handleCapture = ({ target: { files } }) => {
    const [file] = files
    const data = new FormData()
    data.append('save', 'true')
    data.append('file', file)
    axios.post(torrentUploadHost(), data)
  }
  return (
    <div>
      <label htmlFor='upload-file'>
        <input onChange={handleCapture} accept='.torrent' type='file' style={{ display: 'none' }} id='upload-file' />

        <ListItem button variant='raised' type='submit' component='span' key={t('UploadFile')}>
          <ListItemIcon>
            <PublishIcon />
          </ListItemIcon>

          <ListItemText primary={t('UploadFile')} />
        </ListItem>
      </label>
    </div>
  )
}

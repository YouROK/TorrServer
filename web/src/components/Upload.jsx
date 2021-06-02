import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import ListItem from '@material-ui/core/ListItem'
import PublishIcon from '@material-ui/icons/Publish'
import { torrentUploadHost } from 'utils/Hosts'
import axios from 'axios'

export default function UploadDialog() {
  const handleCapture = ({ target: { files } }) => {
    const [file] = files
    const data = new FormData()
    data.append('save', 'true')
    data.append('file', file)
    axios.post(torrentUploadHost(), data)
  }

  return (
    <div>
      <label htmlFor='raised-button-file'>
        <input onChange={handleCapture} accept='*/*' type='file' style={{ display: 'none' }} id='raised-button-file' />

        <ListItem button variant='raised' type='submit' component='span' key='Upload file'>
          <ListItemIcon>
            <PublishIcon />
          </ListItemIcon>

          <ListItemText primary='Upload file' />
        </ListItem>
      </label>
    </div>
  )
}

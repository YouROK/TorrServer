import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'
import { useTranslation } from 'react-i18next'

export default function FilterByCategory({ categoryKey, categoryName, setGlobalFilterCategory, icon }) {
  const onClick = () => {
    setGlobalFilterCategory(categoryKey)
    // if (process.env.NODE_ENV !== 'production') {
      // eslint-disable-next-line no-console
      // console.log('FilterByCategory categoryKey: %s categoryName: %s', categoryKey, categoryName)
    // }
  }
  const { t } = useTranslation()

  return (
    <>
      <ListItem button key={categoryKey} onClick={onClick}>
        <ListItemIcon>{icon}</ListItemIcon>
        <ListItemText primary={t(categoryName)} />
      </ListItem>
    </>
  )
}

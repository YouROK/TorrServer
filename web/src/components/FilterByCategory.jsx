import ListItem from '@material-ui/core/ListItem'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import ListItemText from '@material-ui/core/ListItemText'

export default function FilterByCategory({ categoryName, setGlobalFilterCategory, icon }) {
  const onClick = () => {
    setGlobalFilterCategory(categoryName)
  }

  return (
    <>
      <ListItem button key={categoryName} onClick={onClick}>
        <ListItemIcon>{icon}</ListItemIcon>

        <ListItemText primary={categoryName} />
      </ListItem>
    </>
  )
}

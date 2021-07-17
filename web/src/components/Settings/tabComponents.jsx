export const a11yProps = index => ({
  id: `full-width-tab-${index}`,
  'aria-controls': `full-width-tabpanel-${index}`,
})

export const TabPanel = ({ children, value, index, ...other }) => (
  <div role='tabpanel' hidden={value !== index} id={`full-width-tabpanel-${index}`} {...other}>
    {value === index && <>{children}</>}
  </div>
)

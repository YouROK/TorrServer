import { ListItem } from '@material-ui/core'
import styled from 'styled-components'

export default styled(ListItem).attrs({ button: true })`
  @media screen and (display-mode: standalone) {
    width: 100%;
    height: 60px;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    font-size: 10px;
  }
`

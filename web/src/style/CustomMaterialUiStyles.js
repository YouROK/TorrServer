import { ListItem } from '@material-ui/core'
import Dialog from '@material-ui/core/Dialog'
import { pwaFooterHeight } from 'components/App/PWAFooter/style'
import styled from 'styled-components'
import { Header } from 'style/DialogStyles'

export const StyledMenuButtonWrapper = styled(ListItem).attrs({ button: true })`
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

export const StyledDialog = styled(Dialog)`
  @media screen and (display-mode: standalone) {
    margin-bottom: ${pwaFooterHeight}px;

    .MuiDialog-container .MuiPaper-root {
      box-shadow: none;
    }
  }
`

export const StyledHeader = styled(Header)`
  @media screen and (display-mode: standalone) {
    padding-top: 47px;
  }
`

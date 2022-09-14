import { css } from 'styled-components'

export const standaloneMedia = styles => css`
  @media screen and (display-mode: standalone) {
    ${styles};
  }
`

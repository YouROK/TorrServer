import { standaloneMedia } from 'style/standaloneMedia'
import styled, { css } from 'styled-components'

export const pwaFooterHeight = 90

export default styled.div`
  background: #575757;
  color: #fff;
  position: fixed;
  bottom: 0;
  width: 100%;
  height: ${pwaFooterHeight}px;

  display: none;

  ${standaloneMedia(css`
    display: grid;
    grid-template-columns: repeat(5, calc(100% / 5));
    justify-items: center;
  `)}
`

import styled from 'styled-components'

export const pwaFooterHeight = 90

export default styled.div`
  background: #575757;
  color: #fff;
  position: fixed;
  bottom: 0;
  width: 100%;
  height: ${pwaFooterHeight}px;

  display: none;

  @media screen and (display-mode: standalone) {
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    justify-items: center;
  }
`

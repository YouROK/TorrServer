import styled, { css } from 'styled-components'

export default styled.div`
  ${({ isButton }) => css`
    display: grid;
    place-items: center;
    padding: 20px 40px;
    border-radius: 5px;

    ${isButton &&
    css`
      background: #88cdaa;
      transition: 0.2s;
      cursor: pointer;

      :hover {
        background: #74c39c;
      }
    `}

    lord-icon {
      width: 200px;
      height: 200px;
    }

    .icon-label {
      font-size: 20px;
    }
  `}
`

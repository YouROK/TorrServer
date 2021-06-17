import styled, { css } from 'styled-components'

export default styled.div`
  ${({ isButton }) => css`
    display: grid;
    place-items: center;
    padding: 20px 40px;
    border-radius: 5px;

    ${isButton &&
    css`
      background: #93d7b4;
      transition: 0.2s;

      cursor: pointer;
      :hover {
        background: #71cc9d;
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

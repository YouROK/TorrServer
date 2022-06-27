import styled, { css } from 'styled-components'

export const StyledWrapper = styled.div`
  ${({ isOpen }) => css`
    position: absolute;
    bottom: 10px;
    left: 50%;
    background: #eeeef0;
    width: calc(100% - 20px);
    z-index: 9999;
    border-radius: 10px;
    transition: all 0.3s;
    color: #000;

    ${isOpen
      ? css`
          opacity: 1;
          transform: translate(-50%, 0);
        `
      : css`
          transform: translate(-50%, 150%);
          opacity: 0;
          pointer-events: none;
        `}

    > :not(:last-child) {
      border-bottom: 1px solid #dadadc;
    }

    > * {
      padding: 20px;
    }
  `}
`

export const StyledHeader = styled.div`
  display: grid;
  grid-auto-flow: column;
  grid-template-columns: min-content 1fr;
  gap: 20px;
  align-items: center;
  font-weight: 700;

  img {
    border-radius: 5px;
  }
`

export const StyledContent = styled.div`
  > :not(:last-child) {
    margin-bottom: 25px;
  }

  span {
    background: #fefcfd;
    padding: 5px;
    border-radius: 5px;
  }
`

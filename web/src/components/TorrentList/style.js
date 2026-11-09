import styled, { css } from 'styled-components'

export const SearchFieldWrapper = styled.div`
  position: relative;
  grid-column: 1 / -1;
  margin-top: 3px;

  > .MuiSvgIcon-root {
    position: absolute;
    left: 14px;
    top: 50%;
    transform: translateY(-50%);
    z-index: 1;
    pointer-events: none;
  }

  .MuiOutlinedInput-input {
    padding-left: 46px;
  }

  .MuiInputLabel-outlined:not(.MuiInputLabel-shrink) {
    transform: translate(46px, 12px) scale(1);
  }
`

export default styled.div`
  ${({
    isButton,
    theme: {
      addDialog: { notificationSuccessBGColor, languageSwitchBGColor },
    },
  }) => css`
    display: grid;
    place-items: center;
    padding: 20px 40px;
    border-radius: 5px;

    ${isButton &&
    css`
      background: ${notificationSuccessBGColor};
      transition: 0.2s;
      cursor: pointer;

      :hover {
        background: ${languageSwitchBGColor};
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

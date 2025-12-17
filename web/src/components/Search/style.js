import styled, { css } from 'styled-components'

export const Content = styled.div`
  ${({
    isLoading,
    theme: {
      settingsDialog: { contentBG },
    },
  }) => css`
    background: ${contentBG};
    overflow: auto;
    flex: 1;

    ${isLoading &&
    css`
      min-height: 500px;
      display: grid;
      place-items: center;
    `}
  `}
`

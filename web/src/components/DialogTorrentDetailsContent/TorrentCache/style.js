import styled, { css } from 'styled-components'

import { miniCacheMaxHeight } from './snakeSettings'

export const ScrollNotification = styled.div`
  margin-top: 10px;
  text-transform: uppercase;
  color: rgba(0, 0, 0, 0.5);
  align-self: center;
`

export const SnakeWrapper = styled.div`
  ${({ isMini }) => css`
    ${isMini &&
    css`
      display: grid;
      justify-content: center;
      max-height: ${miniCacheMaxHeight}px;
      overflow: auto;
    `}

    canvas {
      display: block;
    }
  `}
`

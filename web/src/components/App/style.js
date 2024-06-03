import { IconButton } from '@material-ui/core'
import { rgba } from 'polished'
import { standaloneMedia } from 'style/standaloneMedia'
import styled, { css } from 'styled-components'

import { pwaFooterHeight } from './PWAFooter/style'

export const AppWrapper = styled.div`
  ${({
    theme: {
      app: { appSecondaryColor },
    },
  }) => css`
    height: 100%;
    background: ${rgba(appSecondaryColor, 0.8)};
    display: grid;
    grid-template-columns: 60px 1fr;
    grid-template-rows: 60px 1fr;
    grid-template-areas:
      'head head'
      'side content';

    ${standaloneMedia(css`
      grid-template-columns: 0 1fr;
      grid-template-rows: ${pwaFooterHeight}px 1fr ${pwaFooterHeight}px;
      height: 100vh;
    `)}
  `}
`

export const CenteredGrid = styled.div`
  display: grid;
  place-items: center;

  ${standaloneMedia(css`
    height: 100vh;
    width: 100vw;
  `)}
`

export const AppHeader = styled.div`
  ${({ theme: { primary } }) => css`
    background: ${primary};
    color: #fff;
    grid-area: head;
    display: grid;
    grid-auto-flow: column;
    align-items: center;
    grid-template-columns: repeat(2, max-content) 1fr;
    box-shadow: 0px 2px 4px -1px rgb(0 0 0 / 20%), 0px 4px 5px 0px rgb(0 0 0 / 14%), 0px 1px 10px 0px rgb(0 0 0 / 12%);
    padding: 0 16px;
    z-index: 3;

    ${standaloneMedia(css`
      grid-template-columns: max-content 1fr;
      align-items: end;
      padding: 7px 16px;
      position: fixed;
      width: 100%;
      height: ${pwaFooterHeight}px;
    `)}
  `}
`
export const AppSidebarStyle = styled.div`
  ${({
    isDrawerOpen,
    theme: {
      app: { appSecondaryColor, sidebarBGColor, sidebarFillColor },
    },
  }) => css`
    grid-area: side;
    width: ${isDrawerOpen ? '400%' : '100%'};
    z-index: 2;
    overflow-x: hidden;
    transition: width 195ms cubic-bezier(0.4, 0, 0.6, 1) 0ms;
    border-right: 1px solid ${rgba(appSecondaryColor, 0.12)};
    background: ${sidebarBGColor};
    color: ${sidebarFillColor};
    white-space: nowrap;
    /* hide scrollbars */
    overflow-y: scroll;
    scrollbar-width: none; /* Firefox */
    -ms-overflow-style: none; /* Internet Explorer 10+ */
    ::-webkit-scrollbar {
      display: none; /* Safari and Chrome */
      width: 0; /* Remove scrollbar space */
      background: transparent;
    }

    svg {
      fill: ${sidebarFillColor};
    }

    ${standaloneMedia(css`
      display: none;
    `)}
  `}
`
export const TorrentListWrapper = styled.div`
  grid-area: content;
  padding: 20px;
  overflow: auto;

  display: grid;
  place-content: start;
  grid-template-columns: repeat(auto-fit, minmax(max-content, 570px));
  gap: 20px;

  @media (max-width: 1260px), (max-height: 500px) {
    padding: 10px;
    gap: 15px;
    grid-template-columns: repeat(3, 1fr);
  }

  @media (max-width: 1100px) {
    grid-template-columns: repeat(2, 1fr);
  }

  @media (max-width: 700px) {
    grid-template-columns: 1fr;
  }

  ${standaloneMedia(css`
    height: calc(100vh - ${pwaFooterHeight}px);
    padding-bottom: 105px;
  `)}
`

export const HeaderToggle = styled.div`
  ${({
    theme: {
      app: { headerToggleColor },
    },
  }) => css`
    cursor: pointer;
    border-radius: 50%;
    background: ${headerToggleColor};
    height: 35px;
    width: 35px;
    transition: all 0.2s;
    font-weight: 600;
    display: grid;
    place-items: center;
    color: #fff;

    :hover {
      background: ${rgba(headerToggleColor, 0.7)};
    }

    @media (max-width: 700px) {
      height: 28px;
      width: 28px;
      font-size: 12px;

      svg {
        width: 17px;
      }
    }
  `}
`

export const StyledIconButton = styled(IconButton)`
  margin-right: 6px;

  ${standaloneMedia(css`
    display: none;
  `)}
`

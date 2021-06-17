import styled, { css } from 'styled-components'

export const AppWrapper = styled.div`
  height: 100%;
  background: #cbe8d9;
  display: grid;
  grid-template-columns: 60px 1fr;
  grid-template-rows: 60px 1fr;
  grid-template-areas:
    'head head'
    'side content';
`

export const CenteredGrid = styled.div`
  height: 100%;
  display: grid;
  place-items: center;
`

export const AppHeader = styled.div`
  background: #00a572;
  color: #eee;
  grid-area: head;
  display: grid;
  grid-auto-flow: column;
  align-items: center;
  grid-template-columns: repeat(2, max-content) 1fr;
  box-shadow: 0px 2px 4px -1px rgb(0 0 0 / 20%), 0px 4px 5px 0px rgb(0 0 0 / 14%), 0px 1px 10px 0px rgb(0 0 0 / 12%);
  padding: 0 24px;
  z-index: 3;
`
export const AppSidebarStyle = styled.div`
  ${({ isDrawerOpen }) => css`
    grid-area: side;
    width: ${isDrawerOpen ? '400%' : '100%'};
    z-index: 2;
    overflow-x: hidden;
    transition: width 195ms cubic-bezier(0.4, 0, 0.6, 1) 0ms;
    border-right: 1px solid rgba(0, 0, 0, 0.12);
    background: #575757;
    color: #eee;
    white-space: nowrap;
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
`

export const LanguageSwitch = styled.div`
  cursor: pointer;
  border-radius: 50%;
  background: #56b887;
  height: 35px;
  width: 35px;
  transition: all 0.2s;
  font-weight: 600;
  display: grid;
  place-items: center;
  color: #eee;

  :hover {
    background: #7ec9a3;
  }

  @media (max-width: 700px) {
    height: 28px;
    width: 28px;
  }
`

import styled, { css } from 'styled-components'

export const Header = styled.div`
  ${({ theme: { primary } }) => css`
    background: ${primary};
    color: rgba(0, 0, 0, 0.87);
    font-size: 20px;
    color: #fff;
    font-weight: 600;
    box-shadow: 0px 2px 4px -1px rgb(0 0 0 / 20%), 0px 4px 5px 0px rgb(0 0 0 / 14%), 0px 1px 10px 0px rgb(0 0 0 / 12%);
    padding: 15px 24px;
    position: relative;
  `}
`

export const ButtonWrapper = styled.div`
  padding: 20px;
  display: flex;
  justify-content: flex-end;

  > :not(:last-child) {
    margin-right: 10px;
  }
`

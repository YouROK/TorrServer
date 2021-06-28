import styled, { css } from 'styled-components'

export const DialogWrapper = styled.div`
  background: #f1eff3;
  height: 100%;
  display: grid;
  grid-template-rows: max-content 1fr max-content;
`

export const HeaderSection = styled.section`
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 40px;
  font-weight: 300;
  padding: 20px;
  color: #323637;

  img {
    width: 80px;
  }

  @media (max-width: 930px) {
    font-size: 30px;
    padding: 10px 20px;

    img {
      width: 60px;
    }
  }
`

export const ThanksSection = styled.section`
  background: #545a5e;
  color: #f1eff3;
  padding: 40px 20px;
  text-align: center;
  font-size: 30px;
  font-weight: 300;

  @media (max-width: 930px) {
    font-size: 20px;
    padding: 30px 20px;
  }
`

export const SpecialThanksSection = styled.section`
  padding: 40px 20px;
  color: #323637;

  > span {
    font-size: 20px;
    display: block;
    margin-bottom: 15px;
  }

  > div {
    display: grid;
    gap: 10px;
    grid-template-columns: repeat(4, max-content);

    @media (max-width: 930px) {
      grid-template-columns: repeat(3, 1fr);
    }

    @media (max-width: 780px) {
      grid-template-columns: repeat(2, 1fr);
    }

    @media (max-width: 550px) {
      grid-template-columns: 1fr;
    }
  }
`

export const FooterSection = styled.div`
  padding: 20px;
  display: grid;
  grid-auto-flow: column;
  grid-template-columns: repeat(2, max-content);
  justify-content: end;
  gap: 15px;
  align-self: end;
  background: #e8e5eb;

  a {
    text-decoration: none;
  }
`

export const NameWrapper = styled.a`
  ${({ isLink }) => css`
    display: inline-flex;
    align-items: center;
    justify-content: start;
    border: 1px solid;
    padding: 10px;
    border-radius: 5px;
    text-transform: uppercase;
    text-decoration: none;
    background: #545a5e;
    color: #f1eff3;
    transition: 0.2s;

    > * {
      transition: 0.2s;
    }

    ${isLink
      ? css`
          :hover {
            filter: brightness(1.1);

            > * {
              transform: translateY(-1px);
            }
          }
        `
      : css`
          cursor: default;
        `}
  `}
`

export const NameIcon = styled.div`
  display: grid;
  margin-right: 10px;
`

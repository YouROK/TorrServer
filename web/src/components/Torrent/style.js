import styled, { css } from 'styled-components'

export const TorrentCard = styled.div`
  border: 1px solid;
  border-radius: 5px;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  grid-template-rows: 175px minmax(min-content, 1fr);
  grid-template-areas:
    'poster buttons'
    'description description';
  gap: 10px;
  padding: 10px;
  background: #3fb57a;
  box-shadow: 0px 2px 4px -1px rgb(0 0 0 / 20%), 0px 4px 5px 0px rgb(0 0 0 / 14%), 0px 1px 10px 0px rgb(0 0 0 / 12%);

  @media (max-width: 600px), (max-height: 500px) {
    grid-template-areas:
      'poster description'
      'buttons buttons';
    grid-template-columns: 25% 1fr;
    grid-template-rows: 100px min-content;
  }
`

export const TorrentCardPoster = styled.div`
  grid-area: poster;
  border-radius: 5px;
  overflow: hidden;
  text-align: center;

  ${({ isPoster }) =>
    isPoster
      ? css`
          img {
            height: 100%;
            border-radius: 5px;
          }
        `
      : css`
          display: grid;
          place-items: center;
          background: #74c39c;
          border: 1px solid;

          svg {
            transform: translateY(-3px);
          }
        `};

  @media (max-width: 600px), (max-height: 500px) {
    svg {
      width: 50%;
    }
  }
`
export const TorrentCardButtons = styled.div`
  grid-area: buttons;
  display: grid;
  gap: 5px;

  @media (max-width: 600px), (max-height: 500px) {
    grid-template-columns: repeat(4, 1fr);
  }
`
export const TorrentCardDescription = styled.div`
  grid-area: description;
  background: #74c39c;
  border-radius: 5px;
  padding: 5px;
  word-break: break-word;

  @media (max-width: 600px), (max-height: 500px) {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
  }
`

export const TorrentCardDescriptionLabel = styled.div`
  text-transform: uppercase;
  font-size: 10px;
  font-weight: 500;
  letter-spacing: 0.4px;
  color: #216e47;
`

export const TorrentCardDescriptionContent = styled.div`
  margin-left: 5px;
  margin-bottom: 10px;
  word-break: break-all;

  @media (max-width: 600px), (max-height: 500px) {
    font-size: 11px;
    margin-bottom: 3px;
    margin-left: 0;

    ${({ isTitle }) =>
      isTitle &&
      css`
        overflow: auto;
        height: 45px;
      `}
  }

  @media (max-width: 410px) {
    font-size: 10px;
  }
`

export const StyledButton = styled.button`
  border-radius: 5px;
  border: none;
  cursor: pointer;
  transition: 0.2s;
  display: flex;
  align-items: center;
  text-transform: uppercase;
  background: #216e47;
  color: #fff;
  font-size: 1rem;
  font-family: 'Roboto', 'Helvetica', 'Arial', sans-serif;
  letter-spacing: 0.009em;

  > :first-child {
    margin-right: 10px;
  }

  @media (max-width: 600px), (max-height: 500px) {
    padding: 5px 0;
    font-size: 0.8rem;
    justify-content: center;

    span {
      display: none;
    }

    svg {
      width: 20px;
    }

    > :first-child {
      margin-right: 0;
    }
  }

  @media (max-width: 500px) {
    font-size: 0.7rem;
  }

  :hover {
    background: #2a7e54;
  }
`

export const TorrentCardDetails = styled.div`
  @media (max-width: 600px), (max-height: 500px) {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
  }

  /* @media (max-width: 410px) {
    display: none;
  } */
`

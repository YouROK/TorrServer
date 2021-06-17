import styled, { css } from 'styled-components'

export const TorrentCard = styled.div`
  border-radius: 5px;
  display: grid;
  grid-template-columns: 120px 260px 1fr;
  grid-template-rows: 180px;
  grid-template-areas: 'poster description buttons';
  gap: 10px;
  padding: 10px;
  background: #00a572;
  box-shadow: 0px 2px 4px -1px rgb(0 0 0 / 20%), 0px 4px 5px 0px rgb(0 0 0 / 14%), 0px 1px 10px 0px rgb(0 0 0 / 12%);

  @media (max-width: 1260px), (max-height: 500px) {
    grid-template-areas:
      'poster description'
      'buttons buttons';

    grid-template-columns: 70px 1fr;
    grid-template-rows: 110px max-content;
  }

  @media (max-width: 770px) {
    grid-template-columns: 60px 1fr;
    grid-template-rows: 90px max-content;
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
            width: 100%;
            height: 100%;
            object-fit: cover;
            border-radius: 5px;
          }
        `
      : css`
          display: grid;
          place-items: center;
          background: #74c39c;
          border: 1px solid #337a57;

          svg {
            transform: translateY(-3px);
          }
        `};

  @media (max-width: 1260px), (max-height: 500px) {
    svg {
      width: 50%;
    }
  }
`

export const TorrentCardButtons = styled.div`
  grid-area: buttons;
  display: grid;
  gap: 10px;

  @media (max-width: 1260px), (max-height: 500px) {
    grid-template-columns: repeat(4, 1fr);
  }

  @media (max-width: 340px) {
    gap: 5px;
  }
`
export const TorrentCardDescription = styled.div`
  grid-area: description;
  background: #74c39c;
  border-radius: 5px;
  padding: 5px;
  display: grid;
  grid-template-rows: 55% 1fr;
  gap: 10px;

  @media (max-width: 770px) {
    grid-template-rows: 60% 1fr;
    gap: 3px;
  }

  .description-title-wrapper {
    display: flex;
    flex-direction: column;
  }

  .description-section-name {
    text-transform: uppercase;
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.4px;
    color: #216e47;

    @media (max-width: 770px) {
      font-size: 0.4rem;
    }
  }

  .description-torrent-title {
    overflow: auto;
    word-break: break-all;
  }

  .description-statistics-wrapper {
    display: grid;
    grid-template-columns: 80px 80px 1fr;
    align-self: end;

    @media (max-width: 1260px), (max-height: 500px) {
      grid-template-columns: 70px 70px 1fr;
    }

    @media (max-width: 770px) {
      grid-template-columns: 65px 65px 1fr;
    }

    @media (max-width: 700px) {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
    }
  }

  .description-statistics-element-wrapper {
  }

  .description-statistics-element-value {
    margin-left: 5px;
    margin-bottom: 10px;
    word-break: break-all;

    @media (max-width: 1260px), (max-height: 500px) {
      font-size: 0.7rem;
      margin-bottom: 0;
      margin-left: 0;
    }
  }

  .description-torrent-title,
  .description-statistics-element-value {
    @media (max-width: 770px) {
      font-size: 0.6rem;
    }

    @media (max-width: 410px) {
      font-size: 10px;
    }
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
  background: #268757;
  color: #fff;
  font-size: 0.9rem;
  letter-spacing: 0.009em;
  padding: 0 12px;
  svg {
    width: 20px;
  }

  :hover {
    background: #2a7e54;
  }

  > :first-child {
    margin-right: 10px;
  }

  @media (max-width: 1260px), (max-height: 500px) {
    padding: 7px 10px;
    justify-content: center;
    font-size: 0.8rem;

    svg {
      display: none;
    }
  }

  @media (max-width: 770px) {
    font-size: 0.7rem;
  }

  @media (max-width: 420px) {
    font-size: 0.6rem;
    padding: 7px 5px;
  }
`

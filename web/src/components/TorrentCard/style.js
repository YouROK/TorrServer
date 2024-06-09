import styled, { css } from 'styled-components'

export const TorrentCard = styled.div`
  ${({
    theme: {
      torrentCard: { cardPrimaryColor },
    },
  }) => css`
    border-radius: 5px;
    display: grid;
    grid-template-columns: 120px 260px 1fr;
    grid-template-rows: 180px;
    grid-template-areas: 'poster description buttons';
    gap: 10px;
    padding: 10px;
    background: ${cardPrimaryColor};
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
  `}
`

export const TorrentCardPoster = styled.div`
  grid-area: poster;
  border-radius: 5px;
  overflow: hidden;
  text-align: center;
  cursor: pointer;
  transition: 0.2s;
  position: relative;

  :hover {
    filter: brightness(0.7);
  }

  ${({
    isPoster,
    theme: {
      torrentCard: { cardSecondaryColor, accentCardColor },
    },
  }) =>
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
          background: ${cardSecondaryColor};
          border: 1px solid ${accentCardColor};

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
  ${({
    theme: {
      torrentCard: { cardSecondaryColor, accentCardColor },
    },
  }) => css`
    grid-area: description;
    background: ${cardSecondaryColor};
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

    // .description-title-wrapper > .description-section-name {
    //   display: flex;
    //   flex-wrap: nowrap;
    //   justify-content: space-between;
    //   self-align: end;
    // }

    // .description-category-wrapper {
    //   display: inline-flex;
    //   color: #1a1a1a;
    // }

    .description-section-name {
      text-transform: uppercase;
      font-size: 10px;
      font-weight: 600;
      letter-spacing: 0.4px;
      color: ${accentCardColor};

      @media (max-width: 770px) {
        font-size: 0.5rem;
        line-height: 10px;
      }
    }

    .description-status-wrapper {
      display: inline-block;
      height: 8px;
      margin-inline-end: 4px;
      vertical-align: baseline;
    }

    .description-torrent-title {
      overflow: hidden;
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
      margin-bottom: 10px;
      margin-left: 0;

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
        font-size: 9px;
      }
    }
  `}
`

export const StyledButton = styled.button`
  ${({
    theme: {
      torrentCard: { buttonBGColor, accentCardColor },
    },
  }) => css`
    border-radius: 5px;
    border: none;
    cursor: pointer;
    transition: 0.2s;
    display: flex;
    align-items: center;
    text-transform: uppercase;
    background: ${buttonBGColor};
    color: #fff;
    font-size: 0.9rem;
    letter-spacing: 0.009em;
    padding: 0 12px;
    svg {
      width: 20px;
    }

    :hover {
      background: ${accentCardColor};
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
  `}
`

export const StatusIndicators = styled.div`
  ${({ color }) => css`
    height: 8px;
    width: 8px;
    background-color: ${color};
    border-radius: 50%;
    position: relative;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    box-shadow: 1px 1px 2px rgba(0, 0, 0, 0.3);
  `}
`

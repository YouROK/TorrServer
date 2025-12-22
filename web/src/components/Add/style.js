import { Button } from '@material-ui/core'
import styled, { css } from 'styled-components'

export const Content = styled.div`
  ${({
    isEditMode,
    theme: {
      addDialog: { gradientStartColor, gradientEndColor, fontColor },
    },
  }) => css`
    height: 550px;
    background: linear-gradient(145deg, ${gradientStartColor}, ${gradientEndColor});
    flex: 1;
    display: grid;
    grid-template-columns: repeat(${isEditMode ? '1' : '2'}, 1fr);
    border-bottom: 1px solid rgba(0, 0, 0, 0.12);
    overflow: auto;
    color: ${fontColor};

    @media (max-width: 540px) {
      ${'' /* Just for bug fixing on small screens */}
      overflow: scroll;
    }

    @media (max-width: 930px) {
      grid-template-columns: 1fr;
    }

    @media (max-width: 500px) {
      align-content: start;
    }
  `}
`

export const RightSide = styled.div`
  padding: 0 20px 20px 20px;
`

export const RightSideContainer = styled.div`
  ${({
    isHidden,
    notificationMessage,
    isError,
    theme: {
      addDialog: { notificationErrorBGColor, notificationSuccessBGColor },
    },
  }) => css`
    height: 530px;

    ${notificationMessage &&
    css`
      position: relative;
      white-space: nowrap;

      :before {
        font-size: 20px;
        font-weight: 300;
        content: '${notificationMessage}';
        display: grid;
        place-items: center;
        background: ${isError ? notificationErrorBGColor : notificationSuccessBGColor};
        padding: 10px 15px;
        position: absolute;
        top: 52%;
        left: 50%;
        transform: translate(-50%, -50%);
        border-radius: 5px;
      }
    `};

    ${isHidden &&
    css`
      display: none;
    `};

    @media (max-width: 500px) {
      height: 170px;
    }
  `}
`
export const LeftSide = styled.div`
  display: flex;
  flex-direction: column;
  border-right: 1px solid rgba(0, 0, 0, 0.12);
`

export const LeftSideBottomSectionBasicStyles = css`
  transition: transform 0.3s;
  padding: 20px;
  height: 100%;
  display: grid;
`

export const LeftSideBottomSectionNoFile = styled.div`
  ${LeftSideBottomSectionBasicStyles}
  border: 4px dashed rgba(0,0,0,0.1);
  text-align: center;
  outline: none;

  ${({ isDragActive }) => isDragActive && `border: 4px dashed green`};

  justify-items: center;
  grid-template-rows: 130px 1fr;
  cursor: pointer;

  :hover {
    background-color: rgba(0, 0, 0, 0.04);
    svg {
      transform: translateY(-4%);
    }
  }

  @media (max-width: 930px) {
    border: 4px dashed transparent;
    height: 400px;
    place-items: center;
    grid-template-rows: 40% 1fr;
  }

  @media (max-width: 500px) {
    height: 170px;
    grid-template-rows: 1fr;

    > div:first-of-type {
      display: none;
    }
  }
`

export const LeftSideBottomSectionFileSelected = styled.div`
  ${LeftSideBottomSectionBasicStyles}
  place-items: center;

  @media (max-width: 930px) {
    height: 400px;
  }

  @media (max-width: 500px) {
    height: 170px;
  }
`

export const TorrentIconWrapper = styled.div`
  position: relative;
`

export const CancelIconWrapper = styled.div`
  position: absolute;
  top: -9px;
  left: 10px;
  cursor: pointer;

  > svg {
    transition: all 0.3s;
    fill: rgba(0, 0, 0, 0.7);

    :hover {
      fill: rgba(0, 0, 0, 0.6);
    }
  }
`

export const IconWrapper = styled.div`
  display: grid;
  justify-items: center;
  align-content: start;
  gap: 10px;
  align-self: start;

  svg {
    transition: all 0.3s;
  }
`

export const LeftSideTopSection = styled.div`
  ${({
    active,
    theme: {
      addDialog: { gradientStartColor },
    },
  }) => css`
    background: ${gradientStartColor};
    padding: 0 20px 20px 20px;
    transition: all 0.3s;

    ${active && 'box-shadow: 0 8px 10px -9px rgba(0, 0, 0, 0.5)'};
  `}
`

export const PosterWrapper = styled.div`
  margin-top: 20px;
  display: grid;
  grid-template-columns: max-content 1fr;
  grid-template-rows: 300px max-content;
  column-gap: 5px;
  position: relative;
  margin-bottom: 20px;

  grid-template-areas:
    'poster suggestions'
    'clear empty';

  @media (max-width: 540px) {
    grid-template-columns: 1fr;
    gap: 5px 0;
    justify-items: center;
    grid-template-areas:
      'poster'
      'clear'
      'suggestions';
  }
`

export const PosterSuggestions = styled.div`
  display: grid;
  grid-area: suggestions;
  grid-auto-flow: column;
  grid-template-columns: repeat(3, max-content);
  grid-template-rows: repeat(4, max-content);
  gap: 5px;

  @media (max-width: 540px) {
    grid-auto-flow: row;
    grid-template-columns: repeat(5, max-content);
  }
  @media (max-width: 375px) {
    grid-template-columns: repeat(4, max-content);
  }
`

export const PosterSuggestionsItem = styled.div`
  cursor: pointer;
  width: 71px;
  height: 71px;

  @media (max-width: 430px) {
    width: 60px;
    height: 60px;
  }

  @media (max-width: 375px) {
    width: 71px;
    height: 71px;
  }

  @media (max-width: 355px) {
    width: 60px;
    height: 60px;
  }

  img {
    transition: all 0.3s;
    border-radius: 5px;
    width: 100%;
    height: 100%;
    object-fit: cover;

    :hover {
      filter: brightness(130%);
    }
  }
`

export const Poster = styled.div`
  ${({
    poster,
    theme: {
      addDialog: { posterBGColor },
    },
  }) => css`
    border-radius: 5px;
    overflow: hidden;
    width: 200px;
    grid-area: poster;

    ${poster
      ? css`
          img {
            width: 200px;
            object-fit: cover;
            border-radius: 5px;
            height: 100%;
          }
        `
      : css`
          display: grid;
          place-items: center;
          background: ${posterBGColor};

          svg {
            transform: scale(1.5) translateY(-3px);
          }
        `}
  `}
`

export const ClearPosterButton = styled(Button)`
  grid-area: clear;
  justify-self: flex-start;
  transform: translateY(-50%);
  position: absolute;
  ${({ showbutton }) => !showbutton && 'display: none'};

  @media (max-width: 540px) {
    transform: translateY(-140%);
  }
`

export const UpdatePosterButton = styled(Button)`
  grid-area: clear;
  justify-self: flex-end;
  transform: translateY(-50%);
  position: absolute;

  @media (max-width: 540px) {
    transform: translateY(-140%);
  }
`

export const PosterLanguageSwitch = styled.div`
  ${({
    showbutton,
    theme: {
      addDialog: { languageSwitchBGColor, languageSwitchFontColor },
    },
  }) => css`
    grid-area: poster;
    z-index: 5;
    position: absolute;
    top: 0;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 30px;
    height: 30px;
    background: ${languageSwitchBGColor};
    border-radius: 50%;
    display: grid;
    place-items: center;
    color: ${languageSwitchFontColor};
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s;

    ${!showbutton && 'display: none'};

    :hover {
      filter: brightness(1.1);
    }
  `}
`

export const StyledPWAAddButton = styled.div`
  border: 2px solid white;
  border-radius: 50%;
  height: 45px;
  width: 45px;
  position: relative;

  :before,
  :after {
    content: '';
    background: white;
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
  }

  :before {
    width: 2px;
    height: 25px;
  }
  :after {
    width: 25px;
    height: 2px;
  }
`

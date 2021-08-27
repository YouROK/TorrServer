import { rgba } from 'polished'
import styled, { css } from 'styled-components'

export const DialogContentGrid = styled.div`
  display: grid;
  grid-template-columns: 70% 1fr;
  grid-template-rows: repeat(2, min-content);
  grid-template-areas:
    'main cache'
    'file-list file-list';

  @media (max-width: 1450px) {
    grid-template-columns: 1fr;
    grid-template-rows: repeat(3, min-content);
    grid-template-areas:
      'main'
      'cache'
      'file-list';
  }
`
export const Poster = styled.div`
  ${({
    poster,
    theme: {
      dialogTorrentDetailsContent: { posterBGColor },
    },
  }) => css`
    height: 400px;
    border-radius: 5px;
    overflow: hidden;
    align-self: center;

    ${poster
      ? css`
          img {
            border-radius: 5px;
            height: 100%;
          }
        `
      : css`
          width: 300px;
          display: grid;
          place-items: center;
          background: ${posterBGColor};

          svg {
            transform: scale(2.5) translateY(-3px);
          }
        `}

    @media (max-width: 1280px) {
      align-self: start;
    }

    @media (max-width: 840px) {
      ${poster
        ? css`
            height: 200px;
          `
        : css`
            display: none;
          `}
    }
  `}
`
export const MainSection = styled.section`
  ${({
    theme: {
      dialogTorrentDetailsContent: { gradientStartColor, gradientEndColor },
    },
  }) => css`
    grid-area: main;
    padding: 40px;
    display: grid;
    grid-template-columns: min-content 1fr;
    gap: 30px;
    background: linear-gradient(145deg, ${gradientStartColor}, ${gradientEndColor});

    @media (max-width: 840px) {
      grid-template-columns: 1fr;
    }

    @media (max-width: 800px) {
      padding: 20px;
    }
  `}
`

export const CacheSection = styled.section`
  ${({
    theme: {
      dialogTorrentDetailsContent: { chacheSectionBGColor },
    },
  }) => css`
    grid-area: cache;
    padding: 40px;
    display: grid;
    align-content: start;
    grid-template-rows: min-content 1fr min-content;
    background: ${chacheSectionBGColor};

    @media (max-width: 800px) {
      padding: 20px;
    }
  `}
`

export const TorrentFilesSection = styled.section`
  ${({
    theme: {
      dialogTorrentDetailsContent: { torrentFilesSectionBGColor },
    },
  }) => css`
    grid-area: file-list;
    padding: 40px;
    box-shadow: inset 3px 25px 8px -25px rgba(0, 0, 0, 0.5);
    background: ${torrentFilesSectionBGColor};

    @media (max-width: 800px) {
      padding: 20px;
    }
  `}
`

export const SectionSubName = styled.div`
  ${({
    theme: {
      dialogTorrentDetailsContent: { subNameFontColor },
    },
  }) => css`
    ${({ mb }) => css`
      ${mb && `margin-top: ${mb / 3}px`};
      ${mb && `margin-bottom: ${mb}px`};
      line-height: 1.2;
      color: ${subNameFontColor};

      @media (max-width: 800px) {
        ${mb && `margin-top: ${mb / 4}px`};
        ${mb && `margin-bottom: ${mb / 2}px`};
        font-size: 14px;
      }
    `}
  `}
`

export const SectionTitle = styled.div`
  ${({
    color,
    theme: {
      dialogTorrentDetailsContent: { titleFontColor },
    },
  }) => css`
    ${({ mb }) => css`
      ${mb && `margin-bottom: ${mb}px`};
      font-size: 34px;
      font-weight: 300;
      line-height: 1;
      word-break: break-word;
      color: ${color || titleFontColor};

      @media (max-width: 800px) {
        font-size: 24px;
        line-height: 1.1;
        ${mb && `margin-bottom: ${mb / 2}px`};
      }
    `}
  `}
`

export const SectionHeader = styled.div`
  margin-bottom: 20px;
`

export const WidgetWrapper = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(max-content, 220px));
  gap: 20px;

  @media (max-width: 800px) {
    gap: 15px;
  }
  @media (max-width: 410px) {
    gap: 10px;
  }

  ${({ detailedView }) =>
    detailedView
      ? css`
          @media (max-width: 800px) {
            grid-template-columns: repeat(2, 1fr);
          }
          @media (max-width: 410px) {
            grid-template-columns: 1fr;
          }
        `
      : css`
          @media (max-width: 800px) {
            grid-template-columns: repeat(auto-fit, minmax(max-content, 185px));
          }
          @media (max-width: 480px) {
            grid-template-columns: 1fr 1fr;
          }
          @media (max-width: 390px) {
            grid-template-columns: 1fr;
          }
        `}
`

export const WidgetFieldWrapper = styled.div`
  display: grid;
  grid-template-columns: 40px 1fr;
  grid-template-rows: min-content 50px;
  grid-template-areas:
    'title title'
    'icon value';

  > * {
    display: grid;
    place-items: center;
  }

  @media (max-width: 800px) {
    grid-template-columns: 30px 1fr;
    grid-template-rows: min-content 40px;
  }
`
export const WidgetFieldTitle = styled.div`
  ${({
    theme: {
      dialogTorrentDetailsContent: { titleFontColor },
    },
  }) => css`
    grid-area: title;
    justify-self: start;
    text-transform: uppercase;
    font-size: 11px;
    margin-bottom: 2px;
    font-weight: 600;
    color: ${titleFontColor};
  `}
`

export const WidgetFieldIcon = styled.div`
  ${({ bgColor }) => css`
    grid-area: icon;
    color: ${rgba('#fff', 0.8)};
    background: ${bgColor};
    border-radius: 5px 0 0 5px;

    @media (max-width: 800px) {
      > svg {
        width: 50%;
      }
    }
  `}
`
export const WidgetFieldValue = styled.div`
  ${({
    bgColor,
    theme: {
      dialogTorrentDetailsContent: { widgetFontColor },
    },
  }) => css`
    grid-area: value;
    font-size: 24px;
    padding: 0 20px 0 0;
    color: ${widgetFontColor};
    background: ${bgColor};
    border-radius: 0 5px 5px 0;
    white-space: nowrap;

    @media (max-width: 800px) {
      font-size: 18px;
      padding: 0 16px 0 0;
    }
  `}
`

export const LoadingProgress = styled.div.attrs(
  ({
    value,
    fullAmount,
    theme: {
      dialogTorrentDetailsContent: { gradientStartColor, gradientEndColor },
    },
  }) => {
    const percentage = Math.min(100, (value * 100) / fullAmount)

    return {
      // this block is here according to styled-components recomendation about fast changable components
      style: {
        background: `linear-gradient(to right, ${gradientStartColor} 0%, ${gradientEndColor} ${percentage}%, #eee ${percentage}%, #fff 100%)`,
      },
    }
  },
)`
  ${({ label }) => css`
    border: 1px solid;
    padding: 10px 20px;
    border-radius: 5px;
    color: #000;

    :before {
      content: '${label}';
      display: grid;
      place-items: center;
      font-size: 20px;
    }
  `}
`

export const Divider = styled.div`
  height: 1px;
  background-color: rgba(0, 0, 0, 0.12);
  margin: 30px 0;
`

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
  ${({ poster }) => css`
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
          background: #74c39c;

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
  grid-area: main;
  padding: 40px;
  display: grid;
  grid-template-columns: min-content 1fr;
  gap: 30px;
  background: linear-gradient(145deg, #e4f6ed, #b5dec9);

  @media (max-width: 840px) {
    grid-template-columns: 1fr;
  }

  @media (max-width: 800px) {
    padding: 20px;
  }
`

export const CacheSection = styled.section`
  grid-area: cache;
  padding: 40px;
  display: grid;
  align-content: start;
  grid-template-rows: min-content 1fr min-content;
  background: #88cdaa;

  @media (max-width: 800px) {
    padding: 20px;
  }
`

export const TorrentFilesSection = styled.section`
  grid-area: file-list;
  padding: 40px;
  box-shadow: inset 3px 25px 8px -25px rgba(0, 0, 0, 0.5);

  @media (max-width: 800px) {
    padding: 20px;
  }
`

export const SectionSubName = styled.div`
  ${({ mb }) => css`
    ${mb && `margin-bottom: ${mb}px`};
    color: #7c7b7c;

    @media (max-width: 800px) {
      ${mb && `margin-bottom: ${mb / 2}px`};
      font-size: 11px;
    }
  `}
`

export const SectionTitle = styled.div`
  ${({ mb }) => css`
    ${mb && `margin-bottom: ${mb}px`};
    font-size: 35px;
    font-weight: 300;
    line-height: 1;
    word-break: break-word;

    @media (max-width: 800px) {
      font-size: 25px;
      ${mb && `margin-bottom: ${mb / 2}px`};
    }
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
  grid-area: title;
  justify-self: start;
  text-transform: uppercase;
  font-size: 11px;
  margin-bottom: 2px;
  font-weight: 600;
`

export const WidgetFieldIcon = styled.div`
  ${({ bgColor }) => css`
    grid-area: icon;
    color: rgba(255, 255, 255, 0.8);
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
  ${({ bgColor }) => css`
    grid-area: value;
    padding: 0 20px;
    color: #fff;
    font-size: 25px;
    background: ${bgColor};
    border-radius: 0 5px 5px 0;

    @media (max-width: 800px) {
      font-size: 18px;
      padding: 0 4px;
    }
  `}
`

export const LoadingProgress = styled.div.attrs(({ value, fullAmount }) => {
  const percentage = Math.min(100, (value * 100) / fullAmount)

  return {
    // this block is here according to styled-components recomendation about fast changable components
    style: {
      background: `linear-gradient(to right, #b5dec9 0%, #b5dec9 ${percentage}%, #fff ${percentage}%, #fff 100%)`,
    },
  }
})`
  ${({ label }) => css`
    border: 1px solid;
    padding: 10px 20px;
    border-radius: 5px;

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

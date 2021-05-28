import styled, { css } from 'styled-components'

export const DialogContentGrid = styled.div`
  display: grid;
  grid-template-columns: 70% 1fr;
  grid-template-rows: min-content 80px min-content;
  grid-template-areas:
    'main cache'
    'buttons buttons'
    'file-list file-list';
`
export const Poster = styled.div`
  ${({ poster }) => css`
    height: 400px;
    border-radius: 5px;
    overflow: hidden;

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
  `}
`
export const TorrentMainSection = styled.section`
  grid-area: main;
  padding: 40px;
  display: grid;
  grid-template-columns: min-content 1fr;
  gap: 30px;
  background: linear-gradient(145deg, #e4f6ed, #b5dec9);
`

export const CacheSection = styled.section`
  grid-area: cache;
  padding: 40px;
  display: grid;
  align-content: start;
  grid-template-rows: min-content 1fr min-content;
`

export const ButtonSection = styled.section`
  grid-area: buttons;
  box-shadow: 0px 4px 4px -1px rgb(0 0 0 / 30%);
  display: flex;
  justify-content: space-evenly;
  align-items: center;
  text-transform: uppercase;
`

export const ButtonSectionButton = styled.div`
  background: lightblue;
  height: 100%;
  flex: 1;
  display: grid;
  place-items: center;
  cursor: pointer;
  font-size: 15px;

  :not(:last-child) {
    border-right: 1px solid blue;
  }

  :hover {
    background: red;
  }

  .hash-group {
    display: grid;
    place-items: center;
  }

  .hash-text {
    font-size: 10px;
    color: #7c7b7c;
  }
`

export const TorrentFilesSection = styled.div`
  grid-area: file-list;
  padding: 40px;
`

export const SectionSubName = styled.div`
  ${({ mb }) => css`
    ${mb && `margin-bottom: ${mb}px`};
    color: #7c7b7c;
  `}
`

export const SectionTitle = styled.div`
  ${({ mb }) => css`
    ${mb && `margin-bottom: ${mb}px`};
    font-size: 35px;
    font-weight: 200;
    line-height: 1;
    word-break: break-word;
  `}
`

export const SectionHeader = styled.div`
  margin-bottom: 20px;
`

export const DetailedTorrentCacheViewWrapper = styled.div`
  padding-top: 50px;
`

export const StatisticsWrapper = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(190px, min-content));
  gap: 20px;
`

export const StatisticsFieldWrapper = styled.div`
  display: grid;
  grid-template-columns: 40px max-content;
  grid-template-rows: min-content 50px;
  grid-template-areas:
    'title title'
    'icon value';

  > * {
    display: grid;
    place-items: center;
  }
`
export const StatisticsFieldTitle = styled.div`
  grid-area: title;
  justify-self: start;
  text-transform: uppercase;
  font-size: 11px;
  margin-bottom: 2px;
  font-weight: 500;
`

export const StatisticsFieldIcon = styled.div`
  ${({ bgColor }) => css`
    grid-area: icon;
    color: rgba(255, 255, 255, 0.8);
    background: ${bgColor};
    border-radius: 5px 0 0 5px;
  `}
`
export const StatisticsFieldValue = styled.div`
  ${({ bgColor }) => css`
    grid-area: value;
    min-width: 150px;
    padding: 0 20px;
    color: #fff;
    font-size: 25px;
    background: ${bgColor};
    border-radius: 0 5px 5px 0;
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

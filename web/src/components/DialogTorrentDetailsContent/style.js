import styled, { css } from 'styled-components'

export const DialogContentGrid = styled.div`
  display: grid;
  overflow: auto;
  grid-template-columns: 70% 1fr;
  grid-template-rows: repeat(2, min-content);
  grid-template-areas:
    'main cache'
    'file-list file-list';
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
  `}
`
export const MainSection = styled.section`
  grid-area: main;
  padding: 40px;
  display: grid;
  grid-template-columns: min-content 1fr;
  gap: 30px;
  background: linear-gradient(145deg, #e4f6ed, #b5dec9);
`

export const MainSectionButtonGroup = styled.div`
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;

  :not(:last-child) {
    margin-bottom: 30px;
  }
`

export const CacheSection = styled.section`
  grid-area: cache;
  padding: 40px;
  display: grid;
  align-content: start;
  grid-template-rows: min-content 1fr min-content;
  background: #88cdaa;
`

export const TorrentFilesSection = styled.section`
  grid-area: file-list;
  padding: 40px;
  box-shadow: inset 3px 25px 8px -25px rgba(0, 0, 0, 0.5);
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
  overflow: auto;
  min-height: 80vh;
`

export const StatisticsWrapper = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(210px, min-content));
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
    min-width: 170px;
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

export const Divider = styled.div`
  height: 1px;
  background-color: rgba(0, 0, 0, 0.12);
  margin: 30px 0;
`

export const SmallLabel = styled.div`
  ${({ mb }) => css`
    ${mb && `margin-bottom: ${mb}px`};
    font-size: 20px;
    font-weight: 300;
    line-height: 1;
  `}
`

export const Table = styled.table`
  border-collapse: collapse;
  margin: 25px 0;
  font-size: 0.9em;
  width: 100%;
  border-radius: 5px 5px 0 0;
  overflow: hidden;
  box-shadow: 0 0 20px rgba(0, 0, 0, 0.15);

  thead tr {
    background: #009879;
    color: #fff;
    text-align: left;
    text-transform: uppercase;
  }

  th,
  td {
    padding: 12px 15px;
  }

  tbody tr {
    border-bottom: 1px solid #ddd;

    :last-of-type {
      border-bottom: 2px solid #009879;
    }

    &.viewed-file-row {
      background: #f3f3f3;
    }
  }

  td {
    &.viewed-file-indicator {
      position: relative;
      :before {
        content: '';
        width: 10px;
        height: 10px;
        background: #15d5af;
        border-radius: 50%;
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
      }
    }

    &.button-cell {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 10px;
    }
  }
`

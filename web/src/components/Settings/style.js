import styled, { css } from 'styled-components'
import { mainColors } from 'style/colors'

export const cacheBeforeReaderColor = '#b3dfc9'
export const cacheAfterReaderColor = mainColors.light.primary

export const FooterSection = styled.div`
  padding: 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #e8e5eb;

  > :last-child > :not(:last-child) {
    margin-right: 10px;
  }
`
export const Divider = styled.div`
  height: 1px;
  background-color: rgba(0, 0, 0, 0.12);
  margin: 30px 0;
`

export const Content = styled.div`
  ${({ isLoading }) => css`
    background: #f1eff3;
    min-height: 500px;
    overflow: auto;

    ${isLoading &&
    css`
      display: grid;
      place-items: center;
    `}
  `}
`

export const PreloadCacheValue = styled.div`
  ${({ color }) => css`
    display: grid;
    grid-template-columns: max-content 100px 1fr;
    gap: 10px;
    align-items: center;

    :not(:last-child) {
      margin-bottom: 5px;
    }

    :before {
      content: '';
      background: ${color};
      width: 15px;
      height: 15px;
      border-radius: 50%;
    }
  `}
`

export const MainSettingsContent = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 40px;
  padding: 20px;

  @media (max-width: 930px) {
    grid-template-columns: 1fr;
  }
`
export const SecondarySettingsContent = styled.div`
  padding: 20px;
`

export const StorageButton = styled.div`
  ${({ small, selected }) => css`
    transition: 0.2s;
    cursor: default;

    ${!selected &&
    css`
      cursor: pointer;

      :hover {
        filter: brightness(0.8);
      }
    `}

    ${small
      ? css`
          display: grid;
          grid-template-columns: max-content 1fr;
          gap: 20px;
          align-items: center;
          margin-bottom: 20px;
        `
      : css`
          display: grid;
          place-items: center;
          gap: 10px;
        `}
  `}
`

export const StorageIconWrapper = styled.div`
  ${({ selected, small }) => css`
    width: ${small ? '60px' : '150px'};
    height: ${small ? '60px' : '150px'};
    border-radius: 50%;
    background: ${selected ? '#323637' : '#dee3e5'};

    svg {
      transform: rotate(-45deg) scale(0.75);
    }
  `}
`

export const CacheStorageSelector = styled.div`
  display: grid;
  grid-template-rows: max-content 1fr;
  grid-template-columns: 1fr 1fr;
  grid-template-areas: 'label label';
  place-items: center;

  @media (max-width: 930px) {
    grid-template-columns: repeat(2, max-content);
    column-gap: 30px;
  }
`

export const SettingSectionLabel = styled.div`
  font-size: 25px;
  padding-bottom: 20px;

  small {
    display: block;
    font-size: 11px;
  }
`

export const PreloadCachePercentage = styled.div.attrs(({ value }) => ({
  // this block is here according to styled-components recomendation about fast changable components
  style: {
    background: `linear-gradient(to right, ${cacheBeforeReaderColor} 0%, ${cacheBeforeReaderColor} ${value}%, ${cacheAfterReaderColor} ${value}%, ${cacheAfterReaderColor} 100%)`,
  },
}))`
  ${({ label, isPreloadEnabled }) => css`
    border: 1px solid #323637;
    padding: 10px 20px;
    border-radius: 5px;
    color: #000;
    margin-bottom: 10px;
    position: relative;

    :before {
      content: '${label}';
      display: grid;
      place-items: center;
      font-size: 20px;
    }

    ${isPreloadEnabled &&
    css`
      :after {
        content: '';
        width: 100%;
        height: 2px;
        background: #323637;
        position: absolute;
        bottom: 0;
        left: 0;
      }
    `}
  `}
`

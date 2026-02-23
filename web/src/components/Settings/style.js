import styled, { css } from 'styled-components'
import Tabs from '@material-ui/core/Tabs'
import Tab from '@material-ui/core/Tab'
import { mainColors } from 'style/colors'
import { StyledHeader } from 'style/CustomMaterialUiStyles'

export const cacheBeforeReaderColor = '#b3dfc9'

export const StyledTabs = styled(Tabs)`
  .MuiTabs-flexContainer {
    @media (max-width: 600px) {
      gap: 0;
    }
  }

  .MuiTabs-scrollButtons {
    @media (max-width: 600px) {
      width: 24px;
    }
  }
`

export const StyledTab = styled(Tab)`
  min-width: auto;
  padding: 6px 16px;
  font-size: 14px;
  white-space: nowrap;
  flex-shrink: 0;

  @media (max-width: 600px) {
    padding: 6px 12px;
    font-size: 12px;
    min-height: 48px;
  }

  @media (max-width: 400px) {
    padding: 6px 8px;
    font-size: 11px;
  }

  .MuiTab-wrapper {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
  }

  .disabled-hint {
    font-size: 9px;
    opacity: 0.7;
    display: block;

    @media (max-width: 600px) {
      font-size: 8px;
    }
  }
`
export const cacheAfterReaderColor = mainColors.light.primary

export const SettingsHeader = styled(StyledHeader)`
  display: grid;
  grid-auto-flow: column;
  align-items: center;
  justify-content: space-between;

  @media (max-width: 340px) {
    grid-auto-flow: row;
  }
`

export const FooterSection = styled.div`
  ${({
    theme: {
      settingsDialog: { footerBG },
    },
  }) => css`
    padding: 20px;
    display: grid;
    grid-auto-flow: column;
    justify-content: end;
    gap: 10px;
    align-items: center;
    background: ${footerBG};

    @media (max-width: 500px) {
      grid-auto-flow: row;
      justify-content: stretch;
    }
  `}
`
export const Divider = styled.div`
  height: 1px;
  background-color: rgba(0, 0, 0, 0.12);
  margin: 30px 0;
`

export const Content = styled.div`
  ${({
    isLoading,
    theme: {
      settingsDialog: { contentBG },
    },
  }) => css`
    background: ${contentBG};
    overflow: auto;
    flex: 1;

    ${isLoading &&
    css`
      min-height: 500px;
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
    align-items: flex-start;

    :not(:last-child) {
      margin-bottom: 5px;
    }

    :before {
      content: '';
      background: ${color};
      width: 16px;
      height: 16px;
      border-radius: 50%;
      margin-top: 2px;
    }

    @media (max-width: 600px) {
      grid-template-columns: max-content 80px 1fr;
      gap: 8px;
      font-size: 13px;

      :before {
        width: 12px;
        height: 12px;
      }
    }

    @media (max-width: 400px) {
      grid-template-columns: max-content 1fr;
      font-size: 12px;
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

  @media (max-width: 600px) {
    padding: 16px 12px;
    gap: 24px;
  }

  .MuiFormControlLabel-root {
    margin-left: 0;
    margin-right: 0;
    justify-content: space-between;
    width: 100%;

    @media (max-width: 600px) {
      flex-wrap: wrap;
      gap: 4px;
    }
  }

  .MuiFormControlLabel-label {
    @media (max-width: 600px) {
      font-size: 14px;
      flex: 1;
      min-width: 0;
      word-break: break-word;
    }
  }

  .MuiFormHelperText-root {
    @media (max-width: 600px) {
      font-size: 11px;
      margin-top: 2px;
    }
  }
`
export const SecondarySettingsContent = styled.div`
  padding: 20px;

  @media (max-width: 600px) {
    padding: 16px 12px;
  }

  .MuiFormControlLabel-root {
    margin-left: 0;
    margin-right: 0;
    justify-content: space-between;
    width: 100%;

    @media (max-width: 600px) {
      flex-wrap: wrap;
      gap: 4px;
    }
  }

  .MuiFormControlLabel-label {
    @media (max-width: 600px) {
      font-size: 14px;
      flex: 1;
      min-width: 0;
      word-break: break-word;
    }
  }

  .MuiFormHelperText-root {
    @media (max-width: 600px) {
      font-size: 11px;
      margin-top: 2px;
    }
  }

  .MuiInputLabel-root {
    @media (max-width: 600px) {
      font-size: 14px;
    }
  }

  .MuiOutlinedInput-root {
    @media (max-width: 600px) {
      font-size: 14px;
    }
  }
`
export const StorageButton = styled.div`
  ${({ small, selected }) => css`
    transition: 0.2s;
    cursor: default;
    text-align: center;

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
          justify-items: start;
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

    @media (max-width: 930px) {
      width: ${small ? '50px' : '90px'};
      height: ${small ? '50px' : '90px'};
    }
  `}
`

export const CacheStorageSelector = styled.div`
  display: grid;
  grid-template-rows: max-content 1fr;
  grid-template-areas: 'label label';
  place-items: center;

  @media (max-width: 930px) {
    justify-content: start;
    column-gap: 30px;
  }
`

export const SettingSectionLabel = styled.div`
  font-size: 25px;
  padding-bottom: 20px;

  @media (max-width: 600px) {
    font-size: 20px;
    padding-bottom: 16px;
  }

  @media (max-width: 400px) {
    font-size: 18px;
    padding-bottom: 12px;
  }

  small {
    display: block;
    font-size: 11px;

    @media (max-width: 600px) {
      font-size: 10px;
    }
  }
`

export const PreloadCachePercentage = styled.div.attrs(({ value }) => ({
  // this block is here according to styled-components recomendation about fast changable components
  style: {
    background: `linear-gradient(to right, ${cacheBeforeReaderColor} 0%, ${cacheBeforeReaderColor} ${value}%, ${cacheAfterReaderColor} ${value}%, ${cacheAfterReaderColor} 100%)`,
  },
}))`
  ${({ label, preloadCachePercentage }) => css`
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

    :after {
      content: '';
      width: ${preloadCachePercentage}%;
      height: 100%;
      background: #323637;
      position: absolute;
      bottom: 0;
      left: 0;
      border-radius: 4px;
      filter: opacity(0.15);
    }
  `}
`

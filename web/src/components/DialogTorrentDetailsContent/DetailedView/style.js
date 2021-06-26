import styled, { css } from 'styled-components'

export const DetailedViewWidgetSection = styled.section`
  ${({
    theme: {
      detailedView: { gradientStartColor, gradientEndColor },
    },
  }) => css`
    padding: 40px;
    background: linear-gradient(145deg, ${gradientStartColor}, ${gradientEndColor});

    @media (max-width: 800px) {
      padding: 20px;
    }
  `}
`

export const DetailedViewCacheSection = styled.section`
  ${({
    theme: {
      detailedView: { cacheSectionBGColor },
    },
  }) => css`
    padding: 40px;
    box-shadow: inset 3px 25px 8px -25px rgba(0, 0, 0, 0.5);
    background: ${cacheSectionBGColor};
    flex: 1;

    @media (max-width: 800px) {
      padding: 20px;
    }
  `}
`

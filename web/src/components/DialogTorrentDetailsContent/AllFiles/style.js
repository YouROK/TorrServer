import styled from 'styled-components'

const borderColor = 'rgba(128, 128, 128, 0.3)'
const hoverColor = 'rgba(128, 128, 128, 0.12)'

export const TreeToolbar = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin: 25px 0 15px;

  .tree-summary {
    font-size: 0.9em;
    opacity: 0.7;
  }

  .tree-buttons {
    display: flex;
    gap: 10px;
  }
`

export const TreeWrapper = styled.div`
  border: 1px solid ${borderColor};
  border-radius: 5px;
  overflow: hidden;
  font-size: 0.9em;
`

export const TreeRow = styled.div`
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-top: 1px solid ${borderColor};
  cursor: ${({ isDir }) => (isDir ? 'pointer' : 'default')};

  :first-child {
    border-top: none;
  }

  :hover {
    background: ${hoverColor};
  }

  :focus-visible {
    outline: 2px solid currentColor;
    outline-offset: -2px;
  }

  svg {
    flex: none;
    font-size: 20px;
    opacity: ${({ isDir }) => (isDir ? 0.85 : 0.6)};
  }

  .tree-name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-weight: ${({ isDir }) => (isDir ? 500 : 400)};
  }

  .tree-count {
    flex: none;
    opacity: 0.6;
    font-size: 0.9em;
  }

  .tree-size {
    flex: none;
    min-width: 80px;
    margin-left: 16px;
    text-align: right;
    white-space: nowrap;
    font-variant-numeric: tabular-nums;
  }

  @media (max-width: 600px) {
    padding: 8px;

    .tree-size {
      min-width: 0;
      margin-left: 8px;
    }
  }
`

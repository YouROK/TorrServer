import { FixedSizeGrid as Grid } from 'react-window'
import AutoSizer from 'react-virtualized-auto-sizer'
import { memo } from 'react'

import { getLargeSnakeColors } from './colors'

const Cell = memo(({ columnIndex, rowIndex, style, data }) => {
  const { columnCount, cacheMap, gutterSize, borderSize, pieces } = data
  const itemIndex = rowIndex * columnCount + columnIndex

  const { borderColor, backgroundColor } = getLargeSnakeColors(cacheMap[itemIndex] || {})

  const newStyle = {
    ...style,
    left: style.left + gutterSize,
    top: style.top + gutterSize,
    width: style.width - gutterSize,
    height: style.height - gutterSize,
    border: `${borderSize}px solid ${borderColor}`,
    display: itemIndex >= pieces ? 'none' : null,
    background: backgroundColor,
  }

  return <div style={newStyle} />
})

const gutterSize = 2
const borderSize = 1
const pieceSize = 12
const pieceSizeWithSpacing = pieceSize + gutterSize

export default function LargeSnake({ cacheMap }) {
  const pieces = cacheMap.length

  return (
    <div style={{ height: '60vh', overflow: 'hidden' }}>
      <AutoSizer>
        {({ height, width }) => {
          const columnCount = Math.floor(width / (gutterSize + pieceSize)) - 1
          const rowCount = pieces / columnCount + 1

          return (
            <Grid
              columnCount={columnCount}
              rowCount={rowCount}
              columnWidth={pieceSizeWithSpacing}
              rowHeight={pieceSizeWithSpacing}
              height={height}
              width={width}
              itemData={{ columnCount, cacheMap, gutterSize, borderSize, pieces }}
            >
              {Cell}
            </Grid>
          )
        }}
      </AutoSizer>
    </div>
  )
}

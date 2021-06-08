import { Rect } from 'react-konva'

import { activeColor, completeColor, defaultBorderColor, progressColor, rangeColor } from './colors'

export default function SingleBlock({
  x,
  y,
  percentage,
  isActive = false,
  inProgress = false,
  isReaderRange = false,
  isComplete = false,
  boxHeight,
  strokeWidth,
}) {
  const strokeColor = isActive
    ? activeColor
    : isComplete
    ? completeColor
    : inProgress
    ? progressColor
    : isReaderRange
    ? rangeColor
    : defaultBorderColor
  const backgroundColor = inProgress ? progressColor : defaultBorderColor
  const percentageProgressColor = completeColor
  const processCompletedColor = completeColor

  return (
    <Rect
      x={x}
      y={y}
      stroke={strokeColor}
      strokeWidth={strokeWidth}
      height={boxHeight}
      width={boxHeight}
      fillAfterStrokeEnabled
      preventDefault={false}
      {...(isComplete
        ? { fill: processCompletedColor }
        : inProgress && {
            fillLinearGradientStartPointY: boxHeight,
            fillLinearGradientEndPointY: 0,
            fillLinearGradientColorStops: [
              0,
              percentageProgressColor,
              percentage,
              percentageProgressColor,
              percentage,
              backgroundColor,
            ],
          })}
    />
  )
}

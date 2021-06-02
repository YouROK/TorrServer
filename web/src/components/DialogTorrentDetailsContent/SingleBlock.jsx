import { Rect } from 'react-konva'

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
    ? '#000'
    : isComplete
    ? '#3fb57a'
    : inProgress
    ? '#00d0d0'
    : isReaderRange
    ? '#9a9aff'
    : '#eef2f4'
  const backgroundColor = inProgress ? '#00d0d0' : '#eef2f4'
  const percentageProgressColor = '#3fb57a'
  const processCompletedColor = '#3fb57a'

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

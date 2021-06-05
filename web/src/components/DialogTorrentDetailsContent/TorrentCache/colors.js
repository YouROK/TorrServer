export const defaultBorderColor = '#eef2f4'
export const defaultBackgroundColor = '#fff'
export const completeColor = '#3fb57a'
export const progressColor = '#00d0d0'
export const activeColor = '#000'
export const rangeColor = '#9a9aff'

export const getLargeSnakeColors = ({ isActive, isComplete, inProgress, isReaderRange, percentage }) => {
  const gradientBackgroundColor = inProgress ? progressColor : defaultBackgroundColor
  const gradient = `linear-gradient(to top, ${completeColor} 0%, ${completeColor} ${
    percentage * 100
  }%, ${gradientBackgroundColor} ${percentage * 100}%, ${gradientBackgroundColor} 100%)`

  const borderColor = isActive
    ? activeColor
    : isComplete
    ? completeColor
    : inProgress
    ? progressColor
    : isReaderRange
    ? rangeColor
    : defaultBorderColor
  const backgroundColor = isComplete ? completeColor : inProgress ? gradient : defaultBackgroundColor

  return { borderColor, backgroundColor }
}

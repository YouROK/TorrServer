import { Grid, Input, Slider } from '@material-ui/core'

export default function SliderInput({
  isProMode,
  title,
  value,
  setValue,
  sliderMin,
  sliderMax,
  inputMin,
  inputMax,
  step = 1,
  onBlurCallback,
}) {
  const onBlur = ({ target: { value } }) => {
    if (value < inputMin) return setValue(inputMin)
    if (value > inputMax) return setValue(inputMax)

    onBlurCallback && onBlurCallback(value)
  }

  const onInputChange = ({ target: { value } }) => setValue(value === '' ? '' : Number(value))
  const onSliderChange = (_, newValue) => setValue(newValue)

  return (
    <>
      <div>{title}</div>

      <Grid container spacing={2} alignItems='center'>
        <Grid item xs>
          <Slider min={sliderMin} max={sliderMax} value={value} onChange={onSliderChange} step={step} />
        </Grid>

        {isProMode && (
          <Grid item>
            <Input
              value={value}
              margin='dense'
              onChange={onInputChange}
              onBlur={onBlur}
              style={{ width: '65px' }}
              inputProps={{ step, min: inputMin, max: inputMax, type: 'number' }}
            />
          </Grid>
        )}
      </Grid>
    </>
  )
}

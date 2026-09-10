import styled from 'styled-components'

export const ReadoutWrapper = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(90px, 1fr));
  gap: 10px 16px;
  margin-top: 12px;
`

export const ReadoutField = styled.div`
  display: grid;
  align-content: start;
  gap: 2px;
`

export const ReadoutTitle = styled.div`
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  opacity: 0.6;
`

export const ReadoutValue = styled.div`
  font-size: 1rem;
  font-variant-numeric: tabular-nums;
`

export const ReadoutNote = styled.div`
  font-size: 0.7rem;
  opacity: 0.6;
`

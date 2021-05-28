import { StatisticsFieldWrapper, StatisticsFieldIcon, StatisticsFieldValue, StatisticsFieldTitle } from './style'

export default function StatisticsField({ icon: Icon, title, value, iconBg, valueBg }) {
  return (
    <StatisticsFieldWrapper>
      <StatisticsFieldTitle>{title}</StatisticsFieldTitle>
      <StatisticsFieldIcon bgColor={iconBg}>
        <Icon />
      </StatisticsFieldIcon>

      <StatisticsFieldValue bgColor={valueBg}>{value}</StatisticsFieldValue>
    </StatisticsFieldWrapper>
  )
}

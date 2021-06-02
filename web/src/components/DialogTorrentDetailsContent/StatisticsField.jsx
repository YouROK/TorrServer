import { WidgetFieldWrapper, WidgetFieldIcon, WidgetFieldValue, WidgetFieldTitle } from './style'

export default function StatisticsField({ icon: Icon, title, value, iconBg, valueBg }) {
  return (
    <WidgetFieldWrapper>
      <WidgetFieldTitle>{title}</WidgetFieldTitle>
      <WidgetFieldIcon bgColor={iconBg}>
        <Icon />
      </WidgetFieldIcon>

      <WidgetFieldValue bgColor={valueBg}>{value}</WidgetFieldValue>
    </WidgetFieldWrapper>
  )
}

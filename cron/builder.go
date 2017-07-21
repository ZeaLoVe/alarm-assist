package cron

import (
	"fmt"

	"github.com/open-falcon/common/model"
	"github.com/open-falcon/common/utils"
)

func BuildCommonSMSContent(event *model.Event) string {
	return fmt.Sprintf(
		"[Alarm-assist][%s][%s][Endpoint:%s][%s %s %s %s%s%s][step:%d][time:%s]",
		event.Status,
		event.Note(),
		event.Endpoint,
		event.Func(),
		event.Metric(),
		utils.SortedTags(event.PushedTags),
		utils.ReadableFloat(event.LeftValue),
		event.Operator(),
		utils.ReadableFloat(event.RightValue()),
		event.CurrentStep,
		event.FormattedTime(),
	)
}

func BuildCommonIMSmsContent(event *model.Event) string {
	return fmt.Sprintf(
		"[Alarm-assist][%s][%s][Endpoint:%s][%s %s %s %s%s%s][累计出现: %d 次][发生时间: %s]",
		event.Status,
		event.Note(),
		event.Endpoint,
		event.Func(),
		event.Metric(),
		utils.SortedTags(event.PushedTags),
		utils.ReadableFloat(event.LeftValue),
		event.Operator(),
		utils.ReadableFloat(event.RightValue()),
		event.CurrentStep,
		event.FormattedTime(),
	)
}

func BuildCommonPhoneContent(event *model.Event) string {
	if event.Status == "OK" {
		return fmt.Sprintf(
			"您所关注的事件%s, 指标Endpoint:%s,Metrics:%s,Tags:%s 恢复正常,该指标值为%s,报警阈值为%s",
			event.Note(),
			event.Endpoint,
			event.Metric(),
			utils.SortedTags(event.PushedTags),
			utils.ReadableFloat(event.LeftValue),
			utils.ReadableFloat(event.RightValue()),
		)
	}
	return fmt.Sprintf(
		"您所关注的事件%s, 指标Endpoint:%s,Metrics:%s,Tags:%s 出现异常,该指标值为%s,报警阈值为%s.该报警累计出现%d次",
		event.Note(),
		event.Endpoint,
		event.Metric(),
		utils.SortedTags(event.PushedTags),
		utils.ReadableFloat(event.LeftValue),
		utils.ReadableFloat(event.RightValue()),
		event.CurrentStep,
	)
}

func BuildCommonMailContent(event *model.Event) string {
	return fmt.Sprintf(
		"%s\r\n%s\r\nP%d\r\nEndpoint:%s\r\nMetric:%s\r\nTags:%s\r\n%s: %s%s%s\r\nMax:%d, Current:%d\r\nTimestamp:%s\r\n%s\r\n",
		event.Note(),
		event.Status,
		event.Priority(),
		event.Endpoint,
		event.Metric(),
		utils.SortedTags(event.PushedTags),
		event.Func(),
		utils.ReadableFloat(event.LeftValue),
		event.Operator(),
		utils.ReadableFloat(event.RightValue()),
		event.MaxStep(),
		event.CurrentStep,
		event.FormattedTime(),
	)
}

func GenerateSmsContent(event *model.Event) string {
	return BuildCommonSMSContent(event)
}

func GenerateMailContent(event *model.Event) string {
	return BuildCommonMailContent(event)
}

func GenerateIMSmsContent(event *model.Event) string {
	return BuildCommonIMSmsContent(event)
}

func GeneratePhoneContent(event *model.Event) string {
	return BuildCommonPhoneContent(event)
}

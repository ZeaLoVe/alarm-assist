package cron

import (
	"strings"

	"github.com/astaxie/beego"

	"github.com/ZeaLoVe/alarm-assist/cache"
	"github.com/ZeaLoVe/alarm-assist/sender"
	"github.com/open-falcon/common/model"
	"github.com/toolkits/container/set"
)

func consume(event *model.Event) {

	subscribles := cache.Subscribles.GetSubscribles(event.ExpressionId())

	for _, sub := range subscribles {
		event.Expression.Note = sub.Note
		consumeEvent(event, &sub.Pattern, sub.Users)
	}

}

// 处理事件
func consumeEvent(event *model.Event, pattern *cache.Pattern, users []*cache.User) {
	if len(users) == 0 {
		return
	}

	if pattern == nil {
		return
	}

	channals := strings.Split(pattern.Channal, ",")
	if len(channals) == 0 {
		return
	}

	//获取联系方式
	phones, mails, ims, wechats := paseUsers(users)

	smsContent := GenerateSmsContent(event)
	mailContent := GenerateMailContent(event)
	imContent := GenerateIMSmsContent(event)
	phoneContent := GeneratePhoneContent(event)

	for _, channal := range channals {
		switch channal {
		case "im":
			beego.Debug("im send", ims)
			sender.WriteIMSms(ims, imContent)
		case "mail":
			beego.Debug("mail send", mails)
			sender.WriteMail(mails, smsContent, mailContent)
		case "phone":
			beego.Debug("phone send", phones)
			sender.WritePhone(phones, phoneContent)
		case "wechat":
			beego.Debug("wechat send", wechats)
			sender.WriteWechat(wechats, imContent)
		default:
			beego.Debug("<-----Not such channal defined: ", channal)
			beego.Debug(smsContent, " is not sent.------>")
		}
	}
}

//处理联系方式，获取 电话列表，邮件列表，IM列表
func paseUsers(users []*cache.User) ([]string, []string, []string, []string) {
	if len(users) == 0 {
		return []string{}, []string{}, []string{}, []string{}
	}
	imSet := set.NewStringSet()
	phoneSet := set.NewStringSet()
	mailSet := set.NewStringSet()
	wechatSet := set.NewStringSet()
	for _, user := range users {
		imSet.Add(user.IM)
		phoneSet.Add(user.Phone)
		mailSet.Add(user.Email)
		wechatSet.Add(user.Wechat)
	}
	return phoneSet.ToSlice(), mailSet.ToSlice(), imSet.ToSlice(), wechatSet.ToSlice()
}

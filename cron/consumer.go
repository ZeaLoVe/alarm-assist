package cron

import (
	"strings"

	"github.com/astaxie/beego"

	"github.com/ZeaLoVe/alarm-assist/db"
	"github.com/ZeaLoVe/alarm-assist/sender"
	"github.com/open-falcon/common/model"
	"github.com/toolkits/container/set"
)

func consume(event *model.Event) {

	subscrible_items := db.Subscribles.GetSubscrible(event.ExpressionId())

	for _, item := range subscrible_items {
		consumeEvent(event, item.Pattern, item.Users)
	}

}

// 处理事件
func consumeEvent(event *model.Event, pattern *db.Pattern, users []*db.User) {
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
	phones, mails, ims := paseUsers(users)

	smsContent := GenerateSmsContent(event)
	mailContent := GenerateMailContent(event)
	imContent := GenerateIMSmsContent(event)
	phoneContent := GeneratePhoneContent(event)

	for _, channal := range channals {
		switch channal {
		case "im":
			beego.Debug("im sender", ims)
			sender.WriteIMSms(ims, imContent)
		case "mail":
			beego.Debug("mail sender", mails)
			sender.WriteMail(mails, smsContent, mailContent)
		case "phone":
			beego.Debug("phone sender", phones)
			sender.WritePhone(phones, phoneContent)
		default:
			beego.Debug("<-----Not such channal defined: ", channal)
			beego.Debug(smsContent, " is not sent.------>")
		}
	}
}

//处理联系方式，获取 电话列表，邮件列表，IM列表
func paseUsers(users []*db.User) ([]string, []string, []string) {
	if len(users) == 0 {
		return []string{}, []string{}, []string{}
	}
	imSet := set.NewStringSet()
	phoneSet := set.NewStringSet()
	mailSet := set.NewStringSet()
	for _, user := range users {
		imSet.Add(user.IM)
		phoneSet.Add(user.Phone)
		mailSet.Add(user.Email)
	}
	return phoneSet.ToSlice(), mailSet.ToSlice(), imSet.ToSlice()
}

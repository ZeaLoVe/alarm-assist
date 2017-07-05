package sender

import (
	"encoding/json"
	"strings"

	"github.com/astaxie/beego"

	"github.com/ZeaLoVe/alarm-assist/g"
	"github.com/ZeaLoVe/go-utils/model"
)

//sms, short message by phone
//imsms, im message by im(it's depends)
//phone, phone message by phone call(api of nexmo)
//mail, mail message by smtp

func LPUSH(queue, message string) {
	rc := g.RedisConnPool.Get()
	defer rc.Close()
	_, err := rc.Do("LPUSH", queue, message)
	if err != nil {
		beego.Warn("LPUSH redis", queue, "fail:", err, "message:", message)
	}
}

func WriteIMSmsModel(imsms *model.IMSms) {
	if imsms == nil {
		return
	}

	bs, err := json.Marshal(imsms)
	if err != nil {
		beego.Warn(err)
		return
	}

	LPUSH(g.Config().Queue.IMSms, string(bs))
}

func WriteSmsModel(sms *model.Sms) {
	if sms == nil {
		return
	}

	bs, err := json.Marshal(sms)
	if err != nil {
		beego.Warn(err)
		return
	}

	LPUSH(g.Config().Queue.Sms, string(bs))
}

func WriteMailModel(mail *model.Mail) {
	if mail == nil {
		return
	}

	bs, err := json.Marshal(mail)
	if err != nil {
		beego.Warn(err)
		return
	}

	LPUSH(g.Config().Queue.Mail, string(bs))
}

func WritePhoneModel(phone *model.Phone) {
	if phone == nil {
		return
	}

	bs, err := json.Marshal(phone)
	if err != nil {
		beego.Warn(err)
		return
	}

	LPUSH(g.Config().Queue.Phone, string(bs))
}

func WriteIMSms(tos []string, content string) {
	if len(tos) == 0 {
		return
	}

	imsms := &model.IMSms{Tos: strings.Join(tos, ","), Content: content}
	WriteIMSmsModel(imsms)
}

func WriteSms(tos []string, content string) {
	if len(tos) == 0 {
		return
	}

	sms := &model.Sms{Tos: strings.Join(tos, ","), Content: content}
	WriteSmsModel(sms)
}

func WriteMail(tos []string, subject, content string) {
	if len(tos) == 0 {
		return
	}

	mail := &model.Mail{Tos: strings.Join(tos, ","), Subject: subject, Content: content}
	WriteMailModel(mail)
}

func WritePhone(tos []string, content string) {
	if len(tos) == 0 {
		return
	}

	phone := &model.Phone{Tos: strings.Join(tos, ","), Content: content}
	WritePhoneModel(phone)
}

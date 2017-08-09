package controllers

import (
	"encoding/json"

	"github.com/ZeaLoVe/alarm-assist/metrics"
	"github.com/ZeaLoVe/alarm-assist/sender"
)

type AlarmRequestBody struct {
	Type      string   `json:"type"`
	Subject   string   `json:"subject"`
	Content   string   `json:"content"`
	Recievers []string `json:"recievers"`
}

type AlarmApiController struct {
	ApiController
}

func getUniqueRecivers(recievers []string) (ret []string) {
	filter := make(map[string]bool)
	for _, reciever := range recievers {
		filter[reciever] = true
	}
	for reciever, _ := range filter {
		ret = append(ret, reciever)
	}
	return ret
}

func (c *AlarmApiController) Alarms() {
	var body AlarmRequestBody
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &body)
	if err != nil {
		c.RenderError(err.Error())
		return
	} else {
		if body.Type == "" {
			c.RenderError("no alarm type")
			return
		}
		if body.Content == "" {
			c.RenderError("no alarm content")
			return
		}
		if len(body.Recievers) == 0 {
			c.RenderError("no recievers")
			return
		}
		recievers := getUniqueRecivers(body.Recievers)
		switch body.Type {
		case "im":
			sender.WriteIMSms(recievers, body.Content)
			metrics.ReportRequestCount(metrics.Alarm_api_im)
		case "mail":
			if body.Subject == "" {
				c.RenderError("no subject when type is mail")
				return
			}
			sender.WriteMail(recievers, body.Subject, body.Content)
			metrics.ReportRequestCount(metrics.Alarm_api_mail)
		case "phone":
			sender.WritePhone(recievers, body.Content)
			metrics.ReportRequestCount(metrics.Alarm_api_phone)
		case "wechat":
			sender.WriteWechat(recievers, body.Content)
			metrics.ReportRequestCount(metrics.Alarm_api_wechat)
		default:
			c.RenderError("alarm type no support")
			return
		}
		c.RenderSuccess()
	}
}

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
		switch body.Type {
		case "im":
			sender.WriteIMSms(body.Recievers, body.Content)
			metrics.ReportRequestCount(metrics.Alarm_api_im)
		case "mail":
			if body.Subject == "" {
				c.RenderError("no subject when type is mail")
				return
			}
			sender.WriteMail(body.Recievers, body.Subject, body.Content)
			metrics.ReportRequestCount(metrics.Alarm_api_mail)
		case "phone":
			sender.WritePhone(body.Recievers, body.Content)
			metrics.ReportRequestCount(metrics.Alarm_api_phone)
		default:
			c.RenderError("alarm type no support")
			return
		}
		c.RenderSuccess()
	}
}

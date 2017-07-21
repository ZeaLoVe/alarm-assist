package controllers

import (
	"os"
	"time"

	"github.com/ZeaLoVe/alarm-assist/g"
	"github.com/astaxie/beego"
)

const REFLESHINTERVAL = 15

type MainController struct {
	beego.Controller
}

func (this *MainController) Version() {
	this.Ctx.WriteString(g.VERSION)
}

func (this *MainController) Health() {
	this.Ctx.WriteString("ok")
}

func (c *MainController) Home() {
	c.Ctx.ResponseWriter.Write([]byte("welcome to alarm-assist\n"))
	c.Ctx.ResponseWriter.Write([]byte("more details :http://wiki.sdp.nd/index.php?title=Alarm-assist%E6%8E%A5%E5%8F%A3%E6%96%87%E6%A1%A3"))
}

type OKDto struct {
	Code string `json:"code"`
}

type StatusDto struct {
	Status bool `json:"status"`
}

type Dto struct {
	Code        string `json:"code"`
	Msg         string `json:"message"`
	Host_id     string `json:"host_id"`
	Server_time string `json:"server_time"`
}

func NewStatusDto(flag bool) *StatusDto {
	return &StatusDto{
		Status: flag,
	}
}

func NewDto() *Dto {
	host_name, _ := os.Hostname()
	return &Dto{
		Host_id:     host_name,
		Server_time: time.Now().Format(time.RFC3339),
	}
}

type ApiController struct {
	beego.Controller
}

func (c *ApiController) SetHeaders() {
	c.Ctx.Output.Header("Content-Type", "application/json; charset=UTF-8")
	c.Ctx.Output.Header("Access-Control-Allow-Headers", "Authorization,DNT,X-Mx-ReqToken,Keep-Alive,User-Agen,x-requested-with,Content-Type,Content-Length")
	c.Ctx.Output.Header("Access-Control-Allow-Methods", "POST,GET,OPTIONS,PUT")
	c.Ctx.Output.Header("Access-Control-Allow-Origin", "*")
	c.Ctx.Output.Header("Access-Control-Allow-Credentials", "true")
}

func (c *ApiController) RenderError(msg string) {
	resp := NewDto()
	resp.Code = "fail"
	resp.Msg = msg
	c.Data["json"] = resp
	c.SetHeaders()
	c.Ctx.Output.SetStatus(400)
	c.ServeJSON()
}

func (c *ApiController) RenderSuccess() {
	var resp OKDto
	resp.Code = "ok"
	c.Data["json"] = resp
	c.SetHeaders()
	c.ServeJSON()
}

func (c *ApiController) RenderJson(json interface{}) {
	c.Data["json"] = json
	c.SetHeaders()
	c.ServeJSON()
}

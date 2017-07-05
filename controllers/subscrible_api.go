package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ZeaLoVe/alarm-assist/db"
	"github.com/ZeaLoVe/alarm-assist/metrics"
	"github.com/astaxie/beego"
)

//tags like "txxx=vxxx,txxx2=vxxx2"
type ExpressionRequest struct {
	Metric     string `json:"metric"`
	Endpoint   string `json:"endpoint"`
	Tags       string `json:"tags"`
	Func       string `json:"func"`
	Operator   string `json:"op"`
	RightValue string `json:"right_value"`
	MaxStep    int    `json:"max_step"`
}

type SubscribleRequest struct {
	Expression ExpressionRequest `json:"expression"`
	Pattern_id int               `json:"pattern_id"`
	Recievers  []string          `json:"recievers"`
	Cancels    []string          `json:"cancels"`
}

type SubscribleApiController struct {
	ApiController
}

//计算请求的表达式结构指纹,如果不存在该表达式则自动添加，返回指纹字符串和结果
func (this *ExpressionRequest) FingerPrint() (string, error) {
	tags := make(map[string]string)
	if this.Tags != "" {
		arr := strings.Split(strings.TrimSpace(this.Tags), ",")
		for _, item := range arr {
			kv := strings.Split(item, "=")
			if len(kv) != 2 {
				return "", fmt.Errorf("parse expression request fail in calulate fingerprint")
			}
			tags[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	expStr := db.GetExpString(this.Metric, this.Endpoint, tags)
	if expStr == "" {
		return "", fmt.Errorf("can't get available expression string")
	}

	expression := db.Expression{
		Expression: expStr,
		Func:       this.Func,
		Operator:   this.Operator,
		RightValue: this.RightValue,
		MaxStep:    this.MaxStep,
	}
	fp := expression.FingerPrint()
	if db.Expressions.Get(fp) == 0 {
		//TODO 支持不同的通道，或者通道可以配置
		err := expression.Insert(7)
		if err == nil {
			db.Expressions.Set(fp, expression.Id)
		}
		return fp, err
	}
	return fp, nil
}

func (c *SubscribleApiController) AddSubscrible() {
	var body SubscribleRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &body)
	if err != nil {
		c.RenderError(err.Error())
		return
	} else {
		if body.Pattern_id == 0 {
			c.RenderError("need pattern_id, use defualt pattern or create one ")
			return
		}
		fp, err := body.Expression.FingerPrint()
		if err != nil {
			errMsg := fmt.Sprintf("get fingerprint with error: %v", err.Error())
			c.RenderError(errMsg)
			return
		}
		if len(body.Recievers) == 0 && len(body.Cancels) == 0 {
			c.RenderError("you need set recievers or cancels")
			return
		}

		var sub db.Subscrible
		sub.Expression_id = db.Expressions.Get(fp)
		sub.Pattern_id = body.Pattern_id
		sub.Id = db.Subscribles.GetSubscribleId(sub.Expression_id, sub.Pattern_id)
		if sub.Id == 0 {
			if sub.Expression_id != 0 {
				if db.Patterns.Get(sub.Pattern_id) != nil {
					sub.Name = "auto"
					sub.Creator = "alarm-assist"
					err = sub.Insert()
					if err != nil {
						c.RenderError("can't auto build subscrible")
						return
					}
				} else {
					c.RenderError("pattern_id not exist, check it with patterns api")
					return
				}
			} else {
				c.RenderError("expresion_id is 0, create expression error")
				return
			}
		}
		follow_sum := 0
		for _, reciever := range body.Recievers {
			if id := db.Users.GetId(reciever); id != 0 {
				err = sub.Follow(id)
				if err != nil {
					beego.Debug("follow with error ", err.Error())
				} else {
					follow_sum = follow_sum + 1
				}
			}
		}
		cancel_sum := 0
		for _, canceler := range body.Cancels {
			if id := db.Users.GetId(canceler); id != 0 {
				err = sub.CancelFollow(id)
				if err != nil {
					beego.Debug("cancel follow with error ", err.Error())
				} else {
					cancel_sum = cancel_sum + 1
				}
			}
		}

		if (follow_sum + cancel_sum) > 0 {
			c.RenderSuccess()
			metrics.ReportRequestCount(metrics.Alarm_api_subscrible)
			return
		} else {
			c.RenderError("No subscrible operations")
			return
		}
	}
}

func (c *SubscribleApiController) PauseSubscrible() {
	var body SubscribleRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &body)
	if err != nil {
		c.RenderError(err.Error())
		return
	} else {
		if body.Pattern_id == 0 {
			c.RenderError("need pattern_id ")
			return
		}
		fp, err := body.Expression.FingerPrint()
		if err != nil {
			errMsg := fmt.Sprintf("get fingerprint with error: %v", err.Error())
			c.RenderError(errMsg)
			return
		}
		var sub db.Subscrible
		sub.Expression_id = db.Expressions.Get(fp)
		sub.Pattern_id = body.Pattern_id
		sub.Id = db.Subscribles.GetSubscribleId(sub.Expression_id, sub.Pattern_id)
		if sub.Id == 0 {
			c.RenderError("no such subscrible")
			return
		} else {
			if err = sub.PauseSub(); err != nil {
				c.RenderError("subscrible pause fail")
				return
			}
			c.RenderSuccess()
		}
	}
}

func (c *SubscribleApiController) ResumeSubscrible() {
	var body SubscribleRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &body)
	if err != nil {
		c.RenderError(err.Error())
		return
	} else {
		if body.Pattern_id == 0 {
			c.RenderError("need pattern_id ")
			return
		}
		fp, err := body.Expression.FingerPrint()
		if err != nil {
			errMsg := fmt.Sprintf("get fingerprint with error: %v", err.Error())
			c.RenderError(errMsg)
			return
		}
		var sub db.Subscrible
		sub.Expression_id = db.Expressions.Get(fp)
		sub.Pattern_id = body.Pattern_id
		sub.Id = db.Subscribles.GetSubscribleId(sub.Expression_id, sub.Pattern_id)
		if sub.Id == 0 {
			c.RenderError("no such subscrible")
			return
		} else {
			if err = sub.Resume(); err != nil {
				c.RenderError("subscrible resume fail")
				return
			}
			c.RenderSuccess()
		}
	}
}

func (c *SubscribleApiController) UsersSubscrible() {
	var body SubscribleRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &body)
	if err != nil {
		c.RenderError(err.Error())
		return
	} else {
		if body.Pattern_id == 0 {
			c.RenderError("need pattern_id")
			return
		}
		fp, err := body.Expression.FingerPrint()
		if err != nil {
			errMsg := fmt.Sprintf("get fingerprint with error: %v", err.Error())
			c.RenderError(errMsg)
			return
		}
		users := db.Subscribles.GetSubscribleUsers(db.Expressions.Get(fp), body.Pattern_id)
		c.RenderJson(users)
	}
}

func (c *SubscribleApiController) StatusSubscrible() {
	var body SubscribleRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &body)
	if err != nil {
		c.RenderError(err.Error())
		return
	} else {
		if body.Pattern_id == 0 {
			c.RenderError("need pattern_id ")
			return
		}
		fp, err := body.Expression.FingerPrint()
		if err != nil {
			errMsg := fmt.Sprintf("get fingerprint with error: %v", err.Error())
			c.RenderError(errMsg)
			return
		}
		flag := db.Subscribles.GetSubscribleStatus(db.Expressions.Get(fp), body.Pattern_id)
		resp := NewStatusDto(flag)
		c.RenderJson(resp)
	}
}

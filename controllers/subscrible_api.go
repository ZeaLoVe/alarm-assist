package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ZeaLoVe/alarm-assist/cache"
	"github.com/astaxie/beego"
)

//TODO 分页算法写成函数
type SubscribleApiResponse struct {
	TotalElements int                `json:"totalElements"`
	TotalPages    int                `json:"totalPages"`
	Itenms        []cache.Subscrible `json:"items"`
}

func GetSubscribleArray(creator string) []cache.Subscrible {

	var tmpArray []cache.Subscrible
	tmpCache, _ := cache.QuerySubscribleByCreator(creator)
	for _, sub := range tmpCache {
		tmpArray = append(tmpArray, *sub)
	}
	return tmpArray
}

func subscribleFilter(subs []cache.Subscrible, endpoint string, counter string) (ret []cache.Subscrible) {
	if endpoint == "" && counter == "" {
		return subs
	} else {
		for _, sub := range subs {
			comp_endpoint, comp_counter := cache.GetEndpointCounter(sub.Expression.Expression)
			if endpoint == "" && counter != "" {
				if comp_counter == counter {
					ret = append(ret, sub)
				}
			} else if endpoint != "" && counter == "" {
				if comp_endpoint == endpoint {
					ret = append(ret, sub)
				}
			} else if endpoint != "" && counter != "" {
				if comp_endpoint == endpoint && comp_counter == counter {
					ret = append(ret, sub)
				}
			}
		}
		return ret
	}
}

type ExpressionRequest struct {
	Counter    string `json:"counter"` // make by metric and tags
	Endpoint   string `json:"endpoint"`
	FuncCount  int    `json:"count"`    // default 3, max 11
	FuncStr    string `json:"func_str"` // default all,suports all,avg,sum....
	Operator   string `json:"op"`
	RightValue string `json:"right_value"`
	MaxStep    int    `json:"max_step"`
}

type SubscribleRequest struct {
	Id          int               `json:"id"`
	Expression  ExpressionRequest `json:"expression"`
	Pattern_id  int               `json:"pattern_id"`
	Pause       int               `json:"pause"`
	Sub_Note    string            `json:"note"`
	Sub_Name    string            `json:"name"`
	Sub_Creator string            `json:"creator"`
}

type FollowRequest struct {
	Recievers []string `json:"recievers"`
}

type SubscribleApiController struct {
	ApiController
}

//计算请求的表达式结构指纹,如果不存在该表达式则自动添加，返回表达式ID和结果
func (this *ExpressionRequest) GetFitExpression() (int, error) {
	expStr, _ := cache.GetExpStringFromCounter(this.Endpoint, this.Counter)
	if expStr == "" {
		return 0, fmt.Errorf("can't get available expression string")
	}

	if this.FuncStr == "" {
		this.FuncStr = "all"
	}
	if this.FuncCount < 0 {
		this.FuncCount = 3
	}

	expFunc := fmt.Sprintf("%v(#%v)", this.FuncStr, this.FuncCount)

	expression := cache.Expression{
		Expression: expStr,
		Func:       expFunc,
		Operator:   this.Operator,
		RightValue: this.RightValue,
		MaxStep:    this.MaxStep,
	}
	fp := expression.FingerPrint()
	id := cache.Expressions.Get(fp)
	if id == 0 {
		//TODO 支持不同的通道，或者通道可以配置
		err := expression.Insert(7)
		if err == nil {
			cache.Expressions.Set(fp, expression.Id, &expression)
		}
		return expression.Id, err
	}
	return id, nil
}

func (c *SubscribleApiController) UpdateSubscrible() {
	sub_id := c.Ctx.Input.Param(":splat")
	if sub_id == "" {
		c.RenderError("need subscrible id")
		return
	}
	var sub cache.Subscrible
	id, err := strconv.Atoi(sub_id)
	if err != nil {
		c.RenderError("parse subscrible id error")
		return
	}
	sub.Id = id
	var body SubscribleRequest
	err = json.Unmarshal(c.Ctx.Input.RequestBody, &body)
	if err != nil {
		c.RenderError(err.Error())
		return
	} else {
		if body.Expression.Endpoint == "" || body.Expression.Counter == "" {
			c.RenderError("need endpoint and counter")
			return
		}
		if body.Expression.Operator == "" {
			c.RenderError("need operator")
			return
		}
		if body.Expression.RightValue == "" {
			c.RenderError("need right value")
			return
		}
		if body.Expression.FuncCount == 0 {
			c.RenderError("need count")
			return
		}
		if body.Expression.MaxStep < 0 {
			body.Expression.MaxStep = 3
		}
		exp_id, err := body.Expression.GetFitExpression()
		if err != nil {
			errMsg := fmt.Sprintf("fit expression with error: %v", err.Error())
			c.RenderError(errMsg)
			return
		}
		err = sub.Read()
		if err != nil {
			errMsg := fmt.Sprintf("can't get subscrible by id", sub.Id)
			c.RenderError(errMsg)
			return
		}
		//检查expression
		if sub.Expression.Id != exp_id {
			sub.Expression.Id = exp_id
		}
		//检查pattern
		if body.Pattern_id != 0 && sub.Pattern.Id != body.Pattern_id {
			sub.Pattern.Id = body.Pattern_id
		}
		if body.Sub_Name != "" {
			sub.Name = body.Sub_Name
		}
		if body.Sub_Note != "" {
			sub.Note = body.Sub_Note
		}
		sub.Pause = body.Pause //0是默认值，开启  1为关闭
		err = sub.Update()
		if err != nil {
			c.RenderError("update subscrible error")
			return
		}
	}
	c.RenderSuccess()
}

func (c *SubscribleApiController) AddSubscrible() {
	var body SubscribleRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &body)
	if err != nil {
		c.RenderError(err.Error())
		return
	} else {
		if body.Sub_Creator == "" {
			c.RenderError("need subscrible creator")
			return
		}
		if body.Expression.Endpoint == "" || body.Expression.Counter == "" {
			c.RenderError("need endpoint and counter")
			return
		}
		if body.Pattern_id == 0 {
			c.RenderError("need pattern_id, use defualt pattern or create one ")
			return
		}
		if body.Expression.Operator == "" {
			c.RenderError("need operator")
			return
		}
		if body.Expression.RightValue == "" {
			c.RenderError("need right value")
			return
		}
		if body.Expression.MaxStep < 0 {
			body.Expression.MaxStep = 3
		}
		exp_id, err := body.Expression.GetFitExpression()
		if err != nil {
			errMsg := fmt.Sprintf("fit expression with error: %v", err.Error())
			c.RenderError(errMsg)
			return
		}
		beego.Debug("expression_id : ", exp_id)
		var sub cache.Subscrible

		sub.Expression.Id = exp_id
		sub.Pattern.Id = body.Pattern_id
		sub.Id = cache.Subscribles.GetSubscribleId(sub.Expression.Id, sub.Pattern.Id)
		if sub.Id == 0 {
			if sub.Expression.Id != 0 {
				if cache.Patterns.Get(sub.Pattern.Id) != nil {
					if body.Sub_Name == "" {
						sub.Name = "auto"
					} else {
						sub.Name = body.Sub_Name
					}
					sub.Creator = body.Sub_Creator
					sub.Note = body.Sub_Note
					err = sub.Insert()
					if err != nil {
						c.RenderError("can't build subscrible")
						return
					}
				} else {
					c.RenderError("Pattern.Id not exist, check it with patterns api")
					return
				}
			} else {
				c.RenderError("expresion_id is 0, create expression error")
				return
			}
		} else {
			c.RenderError("subscrible already set, please set id in body.")
			return
		}
		sub.Read()
		c.RenderJson(sub)
	}
}

func (c *SubscribleApiController) DeleteSubscrible() {
	sub_id := c.Ctx.Input.Param(":splat")
	if sub_id == "" {
		c.RenderError("need subscrible id")
		return
	}
	var sub cache.Subscrible
	id, err := strconv.Atoi(sub_id)
	if err != nil {
		c.RenderError("parse subscrible id error")
		return
	}
	sub.Id = id

	err = sub.Delete()
	if err != nil {
		c.RenderError("delete subscrible error")
		return
	}
	c.RenderSuccess()
}

func (c *SubscribleApiController) GetSubscribles() {
	page, err := c.GetInt("page")
	if err != nil || page < 1 {
		page = 1
	}

	size, err := c.GetInt("size")
	if err != nil || size < 1 {
		size = 20
	}

	creator := c.GetString("creator")
	if creator == "" {
		c.RenderError("need creator to search subscribles")
		return
	}

	endpoint := c.GetString("endpoint")
	counter := c.GetString("counter")

	var resp SubscribleApiResponse
	subs := subscribleFilter(GetSubscribleArray(creator), endpoint, counter)
	count := len(subs)

	resp.TotalElements = count

	if count == 0 {
		c.RenderJson(resp)
		return
	}

	begin := (page - 1) * size
	if begin >= count {
		c.RenderError("page out of range")
		return
	}
	end := page * size
	if end >= count {
		end = count
	}

	if count%size != 0 {
		resp.TotalPages = (count / size) + 1
	} else {
		resp.TotalPages = (count / size)
	}

	resp.Itenms = subs[begin:end]
	c.RenderJson(resp)
}

func (c *SubscribleApiController) FollowSubscrible() {
	sub_id := c.Ctx.Input.Param(":splat")
	if sub_id == "" {
		c.RenderError("need subscrible id")
		return
	}

	var body FollowRequest
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &body)
	if err != nil {
		c.RenderError(err.Error())
		return
	} else {
		var sub cache.Subscrible
		sub.Id, err = strconv.Atoi(sub_id)
		if err != nil {
			c.RenderError("can't parse id to int")
			return
		}
		//recievers 如果是空，则清空订阅者
		var user_list []int
		for _, im := range body.Recievers {
			id := cache.Users.GetByIm(im)
			if id != 0 {
				user_list = append(user_list, id)
			}
		}
		err = sub.BatchFollow(user_list)
		if err != nil {
			c.RenderError(err.Error())
			return
		}
	}
	c.RenderSuccess()

}

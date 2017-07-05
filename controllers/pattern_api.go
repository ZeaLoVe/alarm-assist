package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ZeaLoVe/alarm-assist/db"
)

type PatternsApiResponse struct {
	TotalElements int          `json:"totalElements"`
	TotalPages    int          `json:"totalPages"`
	Itenms        []db.Pattern `json:"items"`
}

type PatternApiController struct {
	ApiController
}

var pattern_lock sync.Mutex
var patternArray []db.Pattern
var patternLastUpdate int64

func GetPatternsArray() []db.Pattern {
	pattern_lock.Lock()
	defer pattern_lock.Unlock()
	if time.Now().Unix() < patternLastUpdate+REFLESHINTERVAL {
		return patternArray
	} else {
		var tmpArray []db.Pattern
		tmpCache := db.Patterns.M
		for _, pattern := range tmpCache {
			tmpArray = append(tmpArray, *pattern)
		}
		patternLastUpdate = time.Now().Unix()
		patternArray = tmpArray
	}
	return patternArray
}

func (c *PatternApiController) GetPatterns() {
	page, err := c.GetInt("page")
	if err != nil || page < 1 {
		page = 1
	}

	size, err := c.GetInt("size")
	if err != nil || size < 1 {
		size = 20
	}
	var resp PatternsApiResponse
	patterns := GetPatternsArray()
	count := len(patterns)

	if count == 0 {
		c.RenderSuccess()
		return
	}
	resp.TotalElements = count

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

	resp.Itenms = patterns[begin:end]
	c.RenderJson(resp)
}

func (c *PatternApiController) GetPattern() {
	pattern_id := c.Ctx.Input.Param(":splat")
	if pattern_id == "" {
		c.RenderError("need pattern id")
		return
	}
	id, err := strconv.Atoi(pattern_id)
	if err != nil {
		c.RenderError("cant parse id to int")
		return
	}
	var resp db.Pattern
	pattern := db.Patterns.Get(id)

	if pattern == nil {
		c.RenderError("No such pattern")
		return
	}
	resp = *pattern
	c.RenderJson(resp)

}

func (c *PatternApiController) AddPattern() {
	var body db.Pattern
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &body)
	if err != nil {
		c.RenderError(err.Error())
	} else {
		//TODO 添加channal验证
		err := body.Insert()
		if err != nil {
			errorMsg := fmt.Sprintf("insert pattern with err:%v", err.Error())
			c.RenderError(errorMsg)
			return
		}
		c.RenderJson(body)
	}
}

func (c *PatternApiController) DeletePattern() {
	pattern_id := c.Ctx.Input.Param(":splat")
	if pattern_id == "" {
		c.RenderError("need pattern id")
		return
	}
	id, err := strconv.Atoi(pattern_id)
	if err != nil {
		c.RenderError("cant parse id to int")
		return
	}
	pattern := db.Patterns.Get(id)
	if pattern == nil {
		c.RenderError("no such pattern id")
		return
	}
	err = pattern.Delete()
	if err != nil {
		errorMsg := fmt.Sprintf("delete pattern with err:%v", err.Error())
		c.RenderError(errorMsg)
		return
	}
	c.RenderSuccess()
}

func (c *PatternApiController) UpdatePattern() {
	pattern_id := c.Ctx.Input.Param(":splat")
	if pattern_id == "" {
		c.RenderError("need pattern id")
		return
	}
	id, err := strconv.Atoi(pattern_id)
	if err != nil {
		c.RenderError("cant parse id to int")
		return
	}
	var body db.Pattern
	err = json.Unmarshal(c.Ctx.Input.RequestBody, &body)
	if err != nil {
		c.RenderError(err.Error())
		return
	} else {
		body.Id = id
		if body.Id == 0 {
			c.RenderError("pattern id not given")
			return
		}
		err := body.Update()
		if err != nil {
			errorMsg := fmt.Sprintf("update pattern with err:%v", err.Error())
			c.RenderError(errorMsg)
			return
		}
		c.RenderSuccess()
	}
}

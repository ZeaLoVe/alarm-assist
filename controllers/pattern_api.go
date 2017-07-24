package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ZeaLoVe/alarm-assist/cache"
)

type PatternsApiResponse struct {
	TotalElements int             `json:"totalElements"`
	TotalPages    int             `json:"totalPages"`
	Itenms        []cache.Pattern `json:"items"`
}

type PatternApiController struct {
	ApiController
}

func GetPatternsArray() []cache.Pattern {
	var tmpArray []cache.Pattern
	cache.Patterns.RLock()
	tmpCache := cache.Patterns.M
	cache.Patterns.RUnlock()
	for _, pattern := range tmpCache {
		tmpArray = append(tmpArray, *pattern)
	}
	return tmpArray

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
	var resp cache.Pattern
	pattern := cache.Patterns.Get(id)

	if pattern == nil {
		c.RenderError("No such pattern")
		return
	}
	resp = *pattern
	c.RenderJson(resp)

}

func (c *PatternApiController) AddPattern() {
	var body cache.Pattern
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &body)
	if err != nil {
		c.RenderError(err.Error())
	} else {
		if body.Channal == "" {
			c.RenderError("empty channals")
			return
		}
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
	pattern := cache.Patterns.Get(id)
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
	var body cache.Pattern
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

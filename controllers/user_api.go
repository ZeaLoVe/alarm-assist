package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ZeaLoVe/alarm-assist/db"
)

type UsersApiResponse struct {
	TotalElements int       `json:"totalElements"`
	TotalPages    int       `json:"totalPages"`
	Itenms        []db.User `json:"items"`
}

type UserApiController struct {
	ApiController
}

var user_lock sync.Mutex
var userArray []db.User
var userLastUpdate int64

func GetUsersArray() []db.User {
	if time.Now().Unix() < userLastUpdate+REFLESHINTERVAL {
		return userArray
	} else {
		var tmpArray []db.User
		tmpCache := db.Users.M
		for _, user := range tmpCache {
			tmpArray = append(tmpArray, *user)
		}
		userLastUpdate = time.Now().Unix()
		userArray = tmpArray
	}
	return userArray
}

func (c *UserApiController) GetUsers() {
	page, err := c.GetInt("page")
	if err != nil || page < 1 {
		page = 1
	}

	size, err := c.GetInt("size")
	if err != nil || size < 1 {
		size = 20
	}
	var resp UsersApiResponse
	users := GetUsersArray()
	count := len(users)

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

	resp.Itenms = users[begin:end]
	c.RenderJson(resp)
}

func (c *UserApiController) GetUser() {
	user_id := c.Ctx.Input.Param(":splat")
	if user_id == "" {
		c.RenderError("need user id")
		return
	}
	id, err := strconv.Atoi(user_id)
	if err != nil {
		c.RenderError("cant parse id to int")
		return
	}
	var resp db.User
	user := db.Users.Get(id)

	if user == nil {
		c.RenderError("No such user")
		return
	}

	resp = *user

	c.RenderJson(resp)

}

func (c *UserApiController) AddUser() {
	var body db.User
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &body)
	if err != nil {
		c.RenderError(err.Error())
	} else {
		//强制要求输入IM号
		if body.IM == "" {
			c.RenderError("need im info")
			return
		}
		//默认用IM号做名字
		if body.Name == "" {
			body.Name = body.IM
		}
		err := body.Insert()
		if err != nil {
			errorMsg := fmt.Sprintf("insert user with err:%v", err.Error())
			c.RenderError(errorMsg)
			return
		}
		c.RenderJson(body)
	}
}

func (c *UserApiController) DeleteUser() {
	user_id := c.Ctx.Input.Param(":splat")
	if user_id == "" {
		c.RenderError("need user id")
		return
	}
	id, err := strconv.Atoi(user_id)
	if err != nil {
		c.RenderError("cant parse id to int")
		return
	}
	user := db.Users.Get(id)
	err = user.Delete()
	if err != nil {
		errorMsg := fmt.Sprintf("delete user with err:%v", err.Error())
		c.RenderError(errorMsg)
		return
	}
	c.RenderSuccess()
}

func (c *UserApiController) UpdateUser() {
	user_id := c.Ctx.Input.Param(":splat")
	if user_id == "" {
		c.RenderError("need user id")
		return
	}
	id, err := strconv.Atoi(user_id)
	if err != nil {
		c.RenderError("cant parse id to int")
		return
	}
	var body db.User
	err = json.Unmarshal(c.Ctx.Input.RequestBody, &body)
	if err != nil {
		c.RenderError(err.Error())
		return
	} else {
		body.Id = id
		if body.Id == 0 {
			c.RenderError("user id not given")
			return
		}
		err := body.Update()
		if err != nil {
			errorMsg := fmt.Sprintf("update user with err:%v", err.Error())
			c.RenderError(errorMsg)
			return
		}
		c.RenderSuccess()
	}
}

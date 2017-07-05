package db

import (
	"log"
	"testing"
)

func init() {
	InitPortalDB()
	InitUicDB()
}

func TestDBQuerySubscribleUser(t *testing.T) {
	buildUsersCache()
	ret, err := QuerySubscribleUsers()
	if err != nil {
		t.Fatal(err.Error())
	}

	for i, _ := range ret {
		for _, val := range ret[i] {
			log.Println(val.Name, " ", val.IM)
		}
	}
}

func TestDBQueryLevel(t *testing.T) {
	ret, err := QueryPattern()
	if err != nil {
		t.Fatal(err.Error())
	}

	for i, _ := range ret {
		log.Println(ret[i].Channal)
	}
}

func TestDBSubscrible(t *testing.T) {
	buildSubscribleCache()
	for key, _ := range Subscribles.M {
		for _, item := range Subscribles.M[key] {
			for _, val := range item.Users {
				log.Println(val.Name, val.Email, val.QQ, val.Phone)
			}
		}
	}
}

func TestUserApi(t *testing.T) {
	user := &User{
		Name:   "test_insert",
		Cnname: "测试",
		Email:  "xxx.xxx@129.com",
		Phone:  "111132424",
	}
	user.Insert()
	user.Phone = "182324234"
	user.Email = "update.1@1.com"
	user.Update(9)
	user.Delete(7)
}

func TestPatternApi(t *testing.T) {
	pattern := &Pattern{
		Channal: "im,wechat",
		Name:    "测试2",
		Note:    "微信还没支持",
	}
	pattern.Insert()
	pattern.Channal = "im"
	pattern.Note = "测试更新"
	pattern.Update(4)
	pattern.Delete(1)
}

func TestSubscribleApi(t *testing.T) {
	sub := &Subscrible{
		Name:          "api test",
		Creator:       "apiweb",
		Note:          "只是用来测试",
		Expression_id: 3,
		Pattern_id:    1,
	}
	sub.Insert()
	res, err := QuerySubscrible()
	if err != nil {
		t.Fatalf("test subscrible db error", err.Error())
	}
	for _, val := range res {
		log.Println(val.Name, val.Creator, val.Note, val.Expression_id, val.Pattern_id)
	}
	sub.Note = "测试下修改"
	sub.Creator = "测试修改"
	sub.Pattern_id = 3
	sub.Update(5)
	sub.Delete(4)
}

func TestFollow(t *testing.T) {
	sub := &Subscrible{
		Id: 2,
	}
	sub.Follow(1)
	sub.Follow(2)
	sub.Id = 1
	sub.CancelFollow(1)
	sub.CancelFollow(2)
}

package cache

import (
	"fmt"
	"strings"
	"sync"

	"github.com/astaxie/beego"
)

//M 是 id -> *User
//M2是 im -> id
type UsersCache struct {
	sync.RWMutex
	M  map[int]*User
	M2 map[string]int
}

var Users = &UsersCache{M: make(map[int]*User), M2: make(map[string]int)}

func (this *UsersCache) GetUserList(id_list []int) (ret []*User) {
	this.RLock()
	defer this.RUnlock()
	for _, id := range id_list {
		if user, exist := this.M[id]; exist {
			ret = append(ret, user)
		}
	}
	return ret
}

func (this *UsersCache) Get(id int) *User {
	this.RLock()
	defer this.RUnlock()
	val, exists := this.M[id]
	if !exists {
		return nil
	}

	return val
}

func (this *UsersCache) Set(user *User) {
	this.Lock()
	defer this.Unlock()
	this.M[user.Id] = user
	this.M2[user.IM] = user.Id
}

func (this *UsersCache) GetByIm(im string) int {
	this.RLock()
	defer this.RUnlock()
	id, exists := this.M2[im]
	if !exists {
		return 0
	}
	return id
}

func (this *UsersCache) CheckUsers(im_list []string) (ok_list []string, fail_list []string) {
	for _, im := range im_list {
		if _, exist := Users.M2[im]; exist {
			ok_list = append(ok_list, im)
		} else {
			fail_list = append(fail_list, im)
		}
	}
	return ok_list, fail_list
}

func (this *UsersCache) QueryByIM(content string) []*User {
	this.RLock()
	defer this.RUnlock()
	var ret []*User
	for _, user := range this.M {
		if strings.Contains(user.IM, content) {
			ret = append(ret, user)
		}
	}
	return ret
}

func (this *User) updateCache() {
	//更新cache
	err := this.Read()
	if err != nil {
		Users.Set(this)
	}
}

func (this *User) Insert() error {
	sql := fmt.Sprintf("insert into user Set name='%v',passwd='cantlogin',cnname='%v',email='%v',phone='%v',im='%v',qq='%v'",
		this.Name,
		this.Cnname,
		this.Email,
		this.Phone,
		this.IM,
		this.QQ,
	)
	beego.Debug(sql)
	res, err := UicDB.Exec(sql)
	if err != nil {
		return err
	}
	go this.updateCache()
	id64, _ := res.LastInsertId()
	this.Id = int(id64)
	return nil
}

func (this *User) getUpdateSets() string {
	var names, values []string
	if this.Name != "" {
		names = append(names, "name")
		values = append(values, this.Name)
	}
	if this.Cnname != "" {
		names = append(names, "cnname")
		values = append(values, this.Cnname)
	}
	if this.Email != "" {
		names = append(names, "email")
		values = append(values, this.Email)
	}
	if this.IM != "" {
		names = append(names, "phone")
		values = append(values, this.Phone)
	}
	if this.Phone != "" {
		names = append(names, "im")
		values = append(values, this.IM)
	}
	if this.QQ != "" {
		names = append(names, "qq")
		values = append(values, this.QQ)
	}
	return genUpdateSQL(names, values)
}

func (this *User) Update() error {

	sql := fmt.Sprintf("update user Set %v where id=%v",
		this.getUpdateSets(),
		this.Id,
	)
	beego.Debug(sql)
	_, err := UicDB.Exec(sql)
	if err != nil {
		return err
	}
	go this.updateCache()
	return nil

}

func (this *User) Delete() error {
	sql := fmt.Sprintf("delete from user where id=%v",
		this.Id,
	)
	beego.Debug(sql)
	_, err := UicDB.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func (this *User) Read() error {
	sql := fmt.Sprintf("select id, name, cnname, email, phone, im, wechat from user where id=%v",
		this.Id,
	)
	rows, err := UicDB.Query(sql)
	if err != nil {
		return err
	}

	defer rows.Close()
	var count int
	for rows.Next() {
		err = rows.Scan(
			&this.Id,
			&this.Name,
			&this.Cnname,
			&this.Email,
			&this.Phone,
			&this.IM,
			&this.Wechat,
		)
		if err != nil {
			return err
		}
		count = count + 1
	}
	if count == 0 {
		return fmt.Errorf("no such user id:", this.Id)
	}
	return nil
}

func buildUsersCache() {

	sql := "select id, name, cnname, email, phone, im, wechat from user"
	rows, err := UicDB.Query(sql)
	if err != nil {
		beego.Debug("ERROR:", err)
		return
	}
	users := make(map[int]*User)
	ims := make(map[string]int)
	defer rows.Close()
	for rows.Next() {
		user := User{}
		err = rows.Scan(
			&user.Id,
			&user.Name,
			&user.Cnname,
			&user.Email,
			&user.Phone,
			&user.IM,
			&user.Wechat,
		)

		if err != nil {
			beego.Debug("WARN:", err)
			continue
		}

		users[user.Id] = &user
		ims[user.IM] = user.Id
	}

	Users.Lock()
	defer Users.Unlock()

	Users.M = users
	Users.M2 = ims

}

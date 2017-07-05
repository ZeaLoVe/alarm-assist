package db

import (
	"fmt"
	"sync"

	"github.com/astaxie/beego"
)

type User struct {
	Id     int    `json:"id"`
	Cnname string `json:"cnname"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	IM     string `json:"im"`
	QQ     string `json:"qq"`
}

//M 是 id -> *User
//M2是 im -> id
type UsersCache struct {
	sync.RWMutex
	M  map[int]*User
	M2 map[string]int
}

var Users = &UsersCache{M: make(map[int]*User), M2: make(map[string]int)}

func (this *UsersCache) Get(id int) *User {
	this.RLock()
	defer this.RUnlock()
	val, exists := this.M[id]
	if !exists {
		return nil
	}

	return val
}

func (this *UsersCache) Set(id int, user *User) {
	this.Lock()
	defer this.Unlock()
	this.M[id] = user
}

func (this *UsersCache) GetId(im string) int {
	this.RLock()
	defer this.RUnlock()
	id, exists := this.M2[im]
	if !exists {
		return 0
	}

	return id
}

func (this *UsersCache) SetId(im string, id int) {
	this.Lock()
	defer this.Unlock()
	this.M2[im] = id
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
	id64, _ := res.LastInsertId()
	this.Id = int(id64)
	return nil
}

func (this *User) perepareUpdate(curUser *User) {
	if this.Name == "" && curUser.Name != "" {
		this.Name = curUser.Name
	}
	if this.Cnname == "" && curUser.Cnname != "" {
		this.Cnname = curUser.Cnname
	}
	if this.Email == "" && curUser.Email != "" {
		this.Email = curUser.Email
	}
	if this.IM == "" && curUser.IM != "" {
		this.IM = curUser.IM
	}
	if this.Phone == "" && curUser.Phone != "" {
		this.Phone = curUser.Phone
	}
	if this.QQ == "" && curUser.QQ != "" {
		this.QQ = curUser.QQ
	}
}

func (this *User) Update() error {
	tmpUser := *this
	err := tmpUser.Read()
	if err != nil {
		return fmt.Errorf("read user before update error")
	}
	this.perepareUpdate(&tmpUser)
	sql := fmt.Sprintf("update user Set name='%v',cnname='%v',email='%v',phone='%v',im='%v',qq='%v' where id=%v",
		this.Name,
		this.Cnname,
		this.Email,
		this.Phone,
		this.IM,
		this.QQ,
		this.Id,
	)
	beego.Debug(sql)
	_, err = UicDB.Exec(sql)
	if err != nil {
		return err
	}
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
	sql := fmt.Sprintf("select id, name, cnname, email, phone, im from user where id=%v",
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

	sql := "select id, name, cnname, email, phone, im from user"
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

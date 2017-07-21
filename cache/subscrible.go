package cache

import (
	"fmt"
	"strconv"

	"github.com/astaxie/beego"
)

func (this *Subscrible) updateCache() {
	//更新cache
	err := this.Read()
	if err != nil {
		Subscribles.SetSubscrible(this)
	}
}

func (this *Subscrible) Insert() error {
	sql := fmt.Sprintf("insert into subscrible Set name='%v',creator='%v',note='%v',expression_id=%v, pattern_id=%v",
		this.Name,
		this.Creator,
		this.Note,
		this.Expression.Id,
		this.Pattern.Id,
	)
	beego.Debug(sql)
	res, err := PortalDB.Exec(sql)
	if err != nil {
		return err
	}
	go this.updateCache()
	id64, _ := res.LastInsertId()
	this.Id = int(id64)
	return nil
}

func (this *Subscrible) getUpdateSets() string {
	var names, values []string
	if this.Name != "" {
		names = append(names, "name")
		values = append(values, this.Name)
	}
	if this.Note != "" {
		names = append(names, "note")
		values = append(values, this.Note)
	}
	if this.Expression.Id != 0 {
		names = append(names, "expression_id")
		values = append(values, strconv.Itoa(this.Expression.Id))
	}
	if this.Pattern.Id != 0 {
		names = append(names, "pattern_id")
		values = append(values, strconv.Itoa(this.Pattern.Id))
	}
	if this.Pause != 0 {
		names = append(names, "pause")
		values = append(values, strconv.Itoa(this.Pause))
	}
	return genUpdateSQL(names, values)
}

func (this *Subscrible) Update() error {
	sql := fmt.Sprintf("update subscrible Set %v where id=%v",
		this.getUpdateSets(),
		this.Id,
	)

	beego.Debug(sql)
	_, err := PortalDB.Exec(sql)
	if err != nil {
		return err
	}

	go this.updateCache()

	return nil

}

func (this *Subscrible) Delete() error {
	tr, err := PortalDB.Begin()
	if err != nil {
		return err
	}
	//删除订阅事件
	sql := fmt.Sprintf("delete from subscrible where id=%v ",
		this.Id,
	)
	beego.Debug(sql)
	_, err = PortalDB.Exec(sql)
	tr.Rollback()
	if err != nil {
		return err
	}
	//删除订阅用户
	sql = fmt.Sprintf("delete from subscrible_user where subscrible_id=%v",
		this.Id,
	)
	_, err = PortalDB.Exec(sql)
	if err != nil {
		return err
	}
	tr.Commit()

	return nil
}

//func (this *Subscrible) PauseSub() error {
//	if this.Id == 0 {
//		return fmt.Errorf("Subscrible need id in pause")
//	}
//	sql := fmt.Sprintf("update subscrible Set pause=1 where id=%v", this.Id)

//	_, err := PortalDB.Exec(sql)
//	if err != nil {
//		return err
//	}
//	return nil
//}

//func (this *Subscrible) Resume() error {
//	if this.Id == 0 {
//		return fmt.Errorf("Subscrible need id in resume")
//	}
//	sql := fmt.Sprintf("update subscrible Set pause=0 where id=%v", this.Id)

//	_, err := PortalDB.Exec(sql)
//	if err != nil {
//		return err
//	}
//	return nil
//}

//func (this *Subscrible) Follow(user_id int) error {
//	if this.Id == 0 {
//		return fmt.Errorf("Subscrible need id")
//	}
//	sql := fmt.Sprintf("insert into subscrible_user Set subscrible_id=%v, user_id=%v",
//		this.Id,
//		user_id,
//	)
//	beego.Debug(sql)
//	_, err := PortalDB.Exec(sql)
//	if err != nil {
//		return err
//	}
//	return nil
//}

//func (this *Subscrible) CancelFollow(user_id int) error {
//	if this.Id == 0 {
//		return fmt.Errorf("Subscrible need id")
//	}
//	sql := fmt.Sprintf("delete from subscrible_user where subscrible_id=%v and user_id=%v",
//		this.Id,
//		user_id,
//	)
//	beego.Debug(sql)
//	_, err := PortalDB.Exec(sql)
//	if err != nil {
//		return err
//	}
//	return nil
//}

func (this *Subscrible) BatchFollow(users []int) error {

	tr, err := PortalDB.Begin() //开始事务
	if err != nil {
		return err
	}

	values_statment := ""
	for _, user_id := range users {
		if values_statment != "" {
			values_statment = fmt.Sprintf("%v,('%v','%v')", values_statment, this.Id, user_id)
		} else {
			values_statment = fmt.Sprintf("('%v','%v')", this.Id, user_id)
		}
	}
	delete_sql := fmt.Sprintf("delete from subscrible_user where subscrible_id = '%v'",
		this.Id,
	)

	sql := fmt.Sprintf("insert into subscrible_user (subscrible_id, user_id ) values %v",
		values_statment,
	)

	beego.Debug(delete_sql)

	_, err = PortalDB.Exec(delete_sql)
	if len(users) == 0 { //如果参数为空，则清空完直接提交
		tr.Commit()
		return nil
	}

	beego.Debug(sql)
	defer tr.Rollback() //失败回滚

	if err != nil {
		return err
	}

	_, err = PortalDB.Exec(sql)
	if err != nil {
		return err
	}

	tr.Commit() //提交
	subscrible2users.SetSubscribleUsers(this.Id, Users.GetUserList(users))
	go this.updateCache()
	return nil
}

func (this *Subscrible) Read() error {
	sql := fmt.Sprintf("select id, name, creator, note, expression_id, pattern_id, pause from subscrible where id=%v",
		this.Id,
	)
	rows, err := PortalDB.Query(sql)
	if err != nil {
		return err
	}

	defer rows.Close()
	var count int
	for rows.Next() {
		err = rows.Scan(
			&this.Id,
			&this.Name,
			&this.Note,
			&this.Expression.Id,
			&this.Pattern.Id,
			&this.Pause,
		)
		if exp := Expressions.GetById(this.Expression.Id); exp != nil {
			this.Expression = *exp
		}

		if pat := Patterns.Get(this.Pattern.Id); pat != nil {
			this.Pattern = *pat
		}

		subscrible2users.RLock()
		this.Users = subscrible2users.M[this.Id]
		subscrible2users.RUnlock()

		count = count + 1
	}
	if count == 0 {
		return fmt.Errorf("no such subscrible id:", this.Id)
	}
	return nil
}

//creator为空则输出所有
func QuerySubscribleByCreator(creator string) (ret map[int]*Subscrible, err error) {
	var sql string
	if creator != "" {
		sql = fmt.Sprintf("select id, name, creator, note, expression_id, pattern_id, pause from subscrible where creator='%v'", creator)
	} else {
		sql = "select id, name, creator, note, expression_id, pattern_id, pause from subscrible"
	}
	rows, err := PortalDB.Query(sql)
	if err != nil {
		beego.Debug("ERROR:", err)
		return ret, err
	}
	ret = make(map[int]*Subscrible)
	defer rows.Close()
	for rows.Next() {
		s := Subscrible{}
		err = rows.Scan(
			&s.Id,
			&s.Name,
			&s.Creator,
			&s.Note,
			&s.Expression.Id,
			&s.Pattern.Id,
			&s.Pause,
		)

		if exp := Expressions.GetById(s.Expression.Id); exp != nil {
			s.Expression = *exp
		}

		if pat := Patterns.Get(s.Pattern.Id); pat != nil {
			s.Pattern = *pat
		}

		subscrible2users.RLock()
		s.Users = subscrible2users.M[s.Id]
		subscrible2users.RUnlock()

		if err != nil {
			beego.Debug("WARN:", err)
			continue
		}

		ret[s.Id] = &s
	}

	return ret, nil
}

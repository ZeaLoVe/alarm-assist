package db

import (
	"fmt"

	"github.com/astaxie/beego"
)

type Subscrible struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Creator       string `json:"creator"`
	Note          string `json:"note"`
	Expression_id int
	Pattern_id    int
	Pause         int
}

func (this *Subscrible) Insert() error {
	sql := fmt.Sprintf("insert into subscrible Set name='%v',creator='%v',note='%v',expression_id=%v, pattern_id=%v",
		this.Name,
		this.Creator,
		this.Note,
		this.Expression_id,
		this.Pattern_id,
	)
	beego.Debug(sql)
	res, err := PortalDB.Exec(sql)
	if err != nil {
		return err
	}
	id64, _ := res.LastInsertId()
	this.Id = int(id64)
	return nil
}

func (this *Subscrible) perepareUpdate(curSubscrible *Subscrible) {
	if this.Name == "" && curSubscrible.Name != "" {
		this.Name = curSubscrible.Name
	}
	if this.Note == "" && curSubscrible.Note != "" {
		this.Note = curSubscrible.Note
	}
	if this.Expression_id == 0 && curSubscrible.Expression_id != 0 {
		this.Expression_id = curSubscrible.Expression_id
	}
	if this.Pattern_id == 0 && curSubscrible.Pattern_id != 0 {
		this.Pattern_id = curSubscrible.Pattern_id
	}
	if this.Pause == 0 && curSubscrible.Pause != 0 {
		this.Pause = curSubscrible.Pause
	}
}

func (this *Subscrible) PauseSub() error {
	if this.Id == 0 {
		return fmt.Errorf("Subscrible need id in pause")
	}
	sql := fmt.Sprintf("update subscrible Set pause=1 where id=%v", this.Id)

	_, err := PortalDB.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func (this *Subscrible) Resume() error {
	if this.Id == 0 {
		return fmt.Errorf("Subscrible need id in resume")
	}
	sql := fmt.Sprintf("update subscrible Set pause=0 where id=%v", this.Id)

	_, err := PortalDB.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func (this *Subscrible) Update() error {
	tmpSubscrible := *this
	err := tmpSubscrible.Read()
	if err != nil {
		return fmt.Errorf("read subscrible before update error")
	}
	this.perepareUpdate(&tmpSubscrible)

	sql := fmt.Sprintf("update subscrible Set name='%v',note='%v',expression_id=%v, pattern_id=%v where id=%v",
		this.Name,
		this.Note,
		this.Expression_id,
		this.Pattern_id,
		this.Id,
	)
	beego.Debug(sql)
	_, err = PortalDB.Exec(sql)
	if err != nil {
		return err
	}
	return nil

}

func (this *Subscrible) Delete() error {
	sql := fmt.Sprintf("delete from subscrible where id=%v",
		this.Id,
	)
	beego.Debug(sql)
	_, err := PortalDB.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func (this *Subscrible) Follow(user_id int) error {
	if this.Id == 0 {
		return fmt.Errorf("Subscrible need id")
	}
	sql := fmt.Sprintf("insert into subscrible_user Set subscrible_id=%v, user_id=%v",
		this.Id,
		user_id,
	)
	beego.Debug(sql)
	_, err := PortalDB.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func (this *Subscrible) CancelFollow(user_id int) error {
	if this.Id == 0 {
		return fmt.Errorf("Subscrible need id")
	}
	sql := fmt.Sprintf("delete from subscrible_user where subscrible_id=%v and user_id=%v",
		this.Id,
		user_id,
	)
	beego.Debug(sql)
	_, err := PortalDB.Exec(sql)
	if err != nil {
		return err
	}
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
			&this.Expression_id,
			&this.Pattern_id,
			&this.Pause,
		)
		count = count + 1
	}
	if count == 0 {
		return fmt.Errorf("no such subscrible id:", this.Id)
	}
	return nil
}

func QuerySubscrible() (ret map[int]*Subscrible, err error) {
	sql := "select id, name, creator, note, expression_id, pattern_id, pause from subscrible"
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
			&s.Expression_id,
			&s.Pattern_id,
			&s.Pause,
		)

		if err != nil {
			beego.Debug("WARN:", err)
			continue
		}

		ret[s.Id] = &s
	}

	return ret, nil
}

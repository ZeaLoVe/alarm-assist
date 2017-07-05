package db

import (
	"fmt"

	"github.com/astaxie/beego"
)

type Pattern struct {
	Id      int    `json:"id"`
	Channal string `json:"channal"`
	Name    string `json:"name"`
	Note    string `json:"note"`
}

func (this *Pattern) Insert() error {
	sql := fmt.Sprintf("insert into pattern Set channal='%v',name='%v',note='%v'",
		this.Channal,
		this.Name,
		this.Note,
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

func (this *Pattern) perepareUpdate(curPattern *Pattern) {
	if this.Name == "" && curPattern.Name != "" {
		this.Name = curPattern.Name
	}
	if this.Channal == "" && curPattern.Channal != "" {
		this.Channal = curPattern.Channal
	}
	if this.Note == "" && curPattern.Note != "" {
		this.Note = curPattern.Note
	}
}

func (this *Pattern) Update() error {
	tmpPattern := *this
	err := tmpPattern.Read()
	if err != nil {
		return fmt.Errorf("read Pattern before update error")
	}
	this.perepareUpdate(&tmpPattern)

	sql := fmt.Sprintf("update pattern Set channal='%v',name='%v',note='%v' where id=%v",
		this.Channal,
		this.Name,
		this.Note,
		this.Id,
	)
	beego.Debug(sql)
	_, err = PortalDB.Exec(sql)
	if err != nil {
		return err
	}
	return nil

}

func (this *Pattern) Delete() error {
	sql := fmt.Sprintf("delete from pattern where id=%v",
		this.Id,
	)
	beego.Debug(sql)
	_, err := PortalDB.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func (this *Pattern) Read() error {
	sql := fmt.Sprintf("select id, channal, name, note from pattern where id=%v",
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
			&this.Channal,
			&this.Note,
		)
		count = count + 1
	}
	if count == 0 {
		return fmt.Errorf("no such pattern id:", this.Id)
	}
	return nil
}

func QueryPattern() (ret map[int]*Pattern, err error) {
	sql := "select id, channal, name, note from pattern"
	rows, err := PortalDB.Query(sql)
	if err != nil {
		beego.Debug("ERROR:", err)
		return ret, err
	}
	ret = make(map[int]*Pattern)
	defer rows.Close()
	for rows.Next() {
		l := Pattern{}
		err = rows.Scan(
			&l.Id,
			&l.Channal,
			&l.Name,
			&l.Note,
		)

		if err != nil {
			beego.Debug("WARN:", err)
			continue
		}

		ret[l.Id] = &l
	}

	return ret, nil
}

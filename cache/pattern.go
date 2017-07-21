package cache

import (
	"fmt"

	"github.com/astaxie/beego"
)

func (this *Pattern) updateCache() {
	//更新cache
	err := this.Read()
	if err != nil {
		Patterns.Set(this)
	}
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
	go this.updateCache()
	id64, _ := res.LastInsertId()
	this.Id = int(id64)
	return nil
}

func (this *Pattern) getUpdateSets() string {
	var names, values []string
	if this.Name != "" {
		names = append(names, "name")
		values = append(values, this.Name)
	}
	if this.Channal != "" {
		names = append(names, "channal")
		values = append(values, this.Channal)
	}
	if this.Note != "" {
		names = append(names, "note")
		values = append(values, this.Note)
	}
	return genUpdateSQL(names, values)
}

func (this *Pattern) Update() error {
	sql := fmt.Sprintf("update pattern Set %v where id=%v",
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

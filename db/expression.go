package db

import (
	"crypto/md5"
	"fmt"
	"io"

	"github.com/astaxie/beego"
)

type Expression struct {
	Id         int    `json:"id"`
	Expression string `json:"expression"`
	Func       string `json:"func"`
	Operator   string `json:"op"`
	RightValue string `json:"right_value"`
	MaxStep    int    `json:"max_step"`
}

//metric 必须，且expression需要两个以上的选项
func GetExpString(metric string, endpoint string, tags map[string]string) string {
	var exp string
	valid_num := 0
	if metric != "" {
		valid_num = valid_num + 1
		exp = fmt.Sprintf("metric=%v", metric)
	}
	if endpoint != "" {
		valid_num = valid_num + 1
		exp = fmt.Sprintf("%v endpoint=%v", exp, endpoint)
	}
	if len(tags) != 0 {
		for tag_name, tag_value := range tags {
			valid_num = valid_num + 1
			exp = fmt.Sprintf("%v %v=%v", exp, tag_name, tag_value)
		}
	}
	if valid_num < 2 {
		return ""
	}
	return fmt.Sprintf("each(%v)", exp)
}

func (this *Expression) FingerPrint() string {
	before_hash := fmt.Sprintf("%v_%v_%v_%v_%v", this.Expression, this.Func, this.Operator, this.RightValue, this.MaxStep)
	t := md5.New()
	io.WriteString(t, before_hash)
	return fmt.Sprintf("%x", t.Sum(nil))
}

//priority falcon正常报警通道用了1-6，priority写入的时候要大于这个数，具体看alarm-assist的配置项queryQueues
//由于不使用falcon自带的action报警，id统一写1，create_user统一root
func (this *Expression) Insert(priority int) error {
	sql := fmt.Sprintf("insert into expression Set expression='%v',func='%v',op='%v',right_value='%v',max_step=%v,priority=%v,action_id=1,create_user='root'",
		this.Expression,
		this.Func,
		this.Operator,
		this.RightValue,
		this.MaxStep,
		priority,
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

//expression fingerprint->  expression_id
func QueryExpressions() (ret map[string]int, err error) {
	sql := "select id, expression, func, op, right_value, max_step from expression where priority > 6"
	rows, err := PortalDB.Query(sql)
	if err != nil {
		beego.Debug("ERROR:", err)
		return ret, err
	}

	ret = make(map[string]int)
	defer rows.Close()
	for rows.Next() {
		e := Expression{}
		err = rows.Scan(
			&e.Id,
			&e.Expression,
			&e.Func,
			&e.Operator,
			&e.RightValue,
			&e.MaxStep,
		)

		if err != nil {
			beego.Debug("WARN:", err)
			continue
		}

		ret[e.FingerPrint()] = e.Id
	}

	return ret, nil
}

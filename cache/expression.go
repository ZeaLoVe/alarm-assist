package cache

import (
	"crypto/md5"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
)

//从func获得 count(次数)、funcStr（all/avg...）
func GetCountAndFuncStr(str string) (int, string, error) {
	idx := strings.Index(str, "#")
	limit, err := strconv.ParseInt(str[idx+1:len(str)-1], 10, 64)
	if err != nil {
		return 0, "", err
	}
	return int(limit), str[:idx-1], nil
}

func GetEndpointCounter(each string) (endpoint string, counter string) {
	idx := strings.Index(each, "each(")

	//values[0] = metric=XXX
	//values[1] = endpoint=XX
	//values[2...] = tags
	values := strings.Split(each[idx+5:len(each)-1], " ")

	for i, _ := range values {
		if i == 0 {
			counter = strings.Split(values[0], "=")[1] //metric先赋值
		} else if i == 1 {
			endpoint = strings.Split(values[1], "=")[1]
		} else if i == 2 {
			counter = fmt.Sprintf("%v/%v", counter, values[2])
		} else {
			counter = fmt.Sprintf("%v,%v", counter, values[i])
		}
	}
	return endpoint, counter
}

func GetExpStringFromCounter(endpoint string, counter string) (string, error) {
	var metric string
	var tagsStr []string
	idx := strings.Index(counter, "/")
	if idx == -1 {
		metric = counter
	} else {
		metric = counter[0 : idx-1]
		tagsStr = strings.Split(counter[idx+1:], ",")
	}

	expStr := GetExpString(metric, endpoint, tagsStr)
	if expStr == "" {
		return expStr, fmt.Errorf("empty expression")
	}
	return expStr, nil
}

//metric 必须，且expression需要两个以上的选项
//tags 为 key=val 格式的数组
func GetExpString(metric string, endpoint string, tags []string) string {
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
		for _, tag := range tags {
			valid_num = valid_num + 1
			exp = fmt.Sprintf("%v %v", exp, tag)
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
//add: expression id-> *Expression
func QueryExpressions() (ret map[string]int, ret2 map[int]*Expression, err error) {
	sql := "select id, expression, func, op, right_value, max_step from expression where priority > 6"
	rows, err := PortalDB.Query(sql)
	if err != nil {
		beego.Debug("ERROR:", err)
		return ret, ret2, err
	}

	ret = make(map[string]int)
	ret2 = make(map[int]*Expression)
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

		e.FuncCount, e.FuncStr, _ = GetCountAndFuncStr(e.Func)
		e.Endpoint, e.Counter = GetEndpointCounter(e.Expression)

		if err != nil {
			beego.Debug("WARN:", err)
			continue
		}
		ret[e.FingerPrint()] = e.Id
		ret2[e.Id] = &e
	}

	return ret, ret2, nil
}

package cache

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/astaxie/beego"

	"github.com/ZeaLoVe/alarm-assist/g"
	_ "github.com/go-sql-driver/mysql"
)

type Expression struct {
	Id         int    `json:"id"`
	Expression string `json:"-"`
	Func       string `json:"-"`
	Counter    string `json:"counter"` // make by metric and tags
	Endpoint   string `json:"endpoint"`
	FuncCount  int    `json:"count"`    // default 3, max 11
	FuncStr    string `json:"func_str"` // default all,suports all,avg,sum....
	Operator   string `json:"op"`
	RightValue string `json:"right_value"`
	MaxStep    int    `json:"max_step"`
}

type Pattern struct {
	Id      int    `json:"id"`
	Channal string `json:"channal"`
	Name    string `json:"name,omitempty"`
	Note    string `json:"note,omitempty"`
}

type User struct {
	Id     int    `json:"id"`
	Cnname string `json:"cnname"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	IM     string `json:"im"`
	QQ     string `json:"-"`
	Wechat string `json:"wechat"`
}

type Subscrible struct {
	Id      int    `json:"id"`
	Name    string `json:"name,omitempty"`
	Creator string `json:"creator,omitempty"`
	Note    string `json:"note"`
	//	Expression_id int
	//	Pattern_id    int
	Expression Expression
	Pattern    Pattern
	Users      []*User
	Pause      int `json:"pause"`
}

func genUpdateSQL(name []string, value []string) string {
	var ret string
	for i, _ := range name {
		ret = fmt.Sprintf("%v%v='%v',", ret, name[i], value[i])
	}
	//	log.Printf(ret)
	return ret[0 : len(ret)-1]
}

var PortalDB *sql.DB
var UicDB *sql.DB

func InitPortalDB() {
	var err error
	PortalDB, err = sql.Open("mysql", g.Config().Portal)
	//	PortalDB, err = sql.Open("mysql", "root:root@tcp(172.24.133.109:3360)/falcon_portal?loc=Local&parseTime=true")
	if err != nil {
		beego.Debug("open db fail:", err)
	}

	PortalDB.SetMaxIdleConns(100)

	err = PortalDB.Ping()
	if err != nil {
		beego.Debug("ping db fail:", err)
	}
}

func InitUicDB() {
	var err error
	UicDB, err = sql.Open("mysql", g.Config().Uic)
	//	UicDB, err = sql.Open("mysql", "root:root@tcp(172.24.133.109:3360)/uic?loc=Local&parseTime=true")
	if err != nil {
		beego.Debug("open db fail:", err)
	}

	UicDB.SetMaxIdleConns(100)

	err = UicDB.Ping()
	if err != nil {
		beego.Debug("ping db fail:", err)
	}
}

func Init() {
	InitPortalDB()
	InitUicDB()

	go LoopInit()
}

func LoopInit() {
	for {
		buildUsersCache()
		buildExpressionCache()
		buildSubscribleCache()
		time.Sleep(time.Minute)
	}
}

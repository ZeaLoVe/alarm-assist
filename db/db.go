package db

import (
	"database/sql"
	"time"

	"github.com/astaxie/beego"

	_ "github.com/go-sql-driver/mysql"
)

var PortalDB *sql.DB
var UicDB *sql.DB

func InitPortalDB() {
	var err error
	//	PortalDB, err = sql.Open("mysql", g.Config().Portal)
	PortalDB, err = sql.Open("mysql", "root:root@tcp(172.24.133.109:3360)/falcon_portal?loc=Local&parseTime=true")
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
	//	UicDB, err = sql.Open("mysql", g.Config().Uic)
	UicDB, err = sql.Open("mysql", "root:root@tcp(172.24.133.109:3360)/uic?loc=Local&parseTime=true")
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
		time.Sleep(15 * time.Second)
	}
}

package http

import (
	"log"

	"github.com/ZeaLoVe/alarm-assist/controllers"
	"github.com/ZeaLoVe/alarm-assist/g"
	"github.com/ZeaLoVe/alarm-assist/metrics"
	"github.com/astaxie/beego"
)

func configRoutes() {
	beego.Router("/", &controllers.MainController{}, "get:Home")
	beego.Router("/version", &controllers.MainController{}, "get:Version")
	beego.Router("/health", &controllers.MainController{}, "get:Health")
	beego.Router("/v1/api/alarms", &controllers.AlarmApiController{}, "post:Alarms")

	//user api
	beego.Router("/v1/api/uic/users", &controllers.UserApiController{}, "get:GetUsers")
	beego.Router("/v1/api/uic/users/*", &controllers.UserApiController{}, "get:GetUser")
	beego.Router("/v1/api/uic/users/*", &controllers.UserApiController{}, "post:UpdateUser")
	beego.Router("/v1/api/uic/users/*", &controllers.UserApiController{}, "delete:DeleteUser")
	beego.Router("/v1/api/uic/users", &controllers.UserApiController{}, "post:AddUser")

	//pattern api
	beego.Router("/v1/api/portal/patterns", &controllers.PatternApiController{}, "get:GetPatterns")
	beego.Router("/v1/api/portal/patterns/*", &controllers.PatternApiController{}, "get:GetPattern")
	beego.Router("/v1/api/portal/patterns/*", &controllers.PatternApiController{}, "post:UpdatePattern")
	beego.Router("/v1/api/portal/patterns/*", &controllers.PatternApiController{}, "delete:DeletePattern")
	beego.Router("/v1/api/portal/patterns", &controllers.PatternApiController{}, "post:AddPattern")

	//subscrible api
	beego.Router("/v1/api/portal/subscribles", &controllers.SubscribleApiController{}, "post:AddSubscrible")
	beego.Router("/v1/api/portal/subscribles/*", &controllers.SubscribleApiController{}, "delete:DeleteSubscrible")      // * is subscrible id
	beego.Router("/v1/api/portal/subscribles/*", &controllers.SubscribleApiController{}, "post:UpdateSubscrible")        // * is subscrible id
	beego.Router("/v1/api/portal/subscribles/follow/*", &controllers.SubscribleApiController{}, "post:FollowSubscrible") // * is subscrible id
	beego.Router("/v1/api/portal/subscribles", &controllers.SubscribleApiController{}, "get:GetSubscribles")

}

func init() {
	configRoutes()
}

func Start() {
	if !g.Config().Http.Enabled {
		return
	}

	addr := g.Config().Http.Listen
	if addr == "" {
		return
	}

	metrics.Port = beego.AppConfig.String("prometheus_port")
	metrics.Metrics()

	if g.Config().Debug {
		beego.SetLevel(beego.LevelDebug)
	} else {
		beego.SetLevel(beego.LevelInformational)
	}

	beego.Run(addr)

	log.Println("http listening", addr)
}

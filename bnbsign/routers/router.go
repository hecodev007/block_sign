package routers

import (
	"bnbsign/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/genaddr", &controllers.GenAddressController{})
	beego.Router("/sign", &controllers.MainController{})
	beego.Router("/push", &controllers.SendController{})
}

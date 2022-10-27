package routers

import (
	"veservice/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/vet", &controllers.MainController{})
	beego.Router("/v1/vet/transfer", &controllers.TransferController{})
	beego.Router("/v1/vet/repay", &controllers.RepayController{})
}

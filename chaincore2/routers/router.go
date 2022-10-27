package routers

import (
	"github.com/astaxie/beego"
	"github.com/group-coldwallet/chaincore2/controllers"
	"github.com/group-coldwallet/chaincore2/controllers/addrManager"
	"github.com/group-coldwallet/chaincore2/controllers/agent"
	"github.com/group-coldwallet/chaincore2/controllers/ar"
	"github.com/group-coldwallet/chaincore2/controllers/auth"
	"github.com/group-coldwallet/chaincore2/controllers/bnb"
	"github.com/group-coldwallet/chaincore2/controllers/chainx"
	"github.com/group-coldwallet/chaincore2/controllers/ckb"
	"github.com/group-coldwallet/chaincore2/controllers/cocos"
	"github.com/group-coldwallet/chaincore2/controllers/dash"
	"github.com/group-coldwallet/chaincore2/controllers/dcr"
	"github.com/group-coldwallet/chaincore2/controllers/dhx"
	"github.com/group-coldwallet/chaincore2/controllers/eos"
	"github.com/group-coldwallet/chaincore2/controllers/fibos"
	"github.com/group-coldwallet/chaincore2/controllers/hc"
	"github.com/group-coldwallet/chaincore2/controllers/kakao"

	//"github.com/group-coldwallet/chaincore2/controllers/kakao"
	"github.com/group-coldwallet/chaincore2/controllers/ksm"
	"github.com/group-coldwallet/chaincore2/controllers/mdu"
	"github.com/group-coldwallet/chaincore2/controllers/neo"
	"github.com/group-coldwallet/chaincore2/controllers/ont"
	"github.com/group-coldwallet/chaincore2/controllers/qtum"
	"github.com/group-coldwallet/chaincore2/controllers/rsk"
	"github.com/group-coldwallet/chaincore2/controllers/ruby"
	"github.com/group-coldwallet/chaincore2/controllers/seek"
	"github.com/group-coldwallet/chaincore2/controllers/stacks"
	"github.com/group-coldwallet/chaincore2/controllers/ve"
	"github.com/group-coldwallet/chaincore2/controllers/ycash"
	"github.com/group-coldwallet/chaincore2/controllers/zvc"
)

// common, 所有服务都加载
func CommonInit() {
	beego.Router("/info", &controllers.InfoController{})
}

// ycash yec
func YcashInit() {
	beego.Router("/ycash/rpc", &ycash.RpcController{})
	beego.Router("/ycash/insert", &ycash.InsertController{})
	beego.Router("/ycash/remove", &ycash.RemoveController{})
	beego.Router("/ycash/update", &ycash.UpdateController{})
	beego.Router("/ycash/repush", &ycash.RepushTxController{})
	beego.Router("/ycash/test", &ycash.TestController{})
}

// ruby
func RubyInit() {
	beego.Router("/ruby/rpc", &ruby.RpcController{})
	beego.Router("/ruby/insert", &ruby.InsertController{})
	beego.Router("/ruby/remove", &ruby.RemoveController{})
	beego.Router("/ruby/update", &ruby.UpdateController{})
	beego.Router("/ruby/test", &ruby.TestController{})
	beego.Router("/ruby/repush", &ruby.RepushTxController{})
	beego.Router("/ruby/reback", &ruby.RebackController{})
}

// cocos
func CocosInit() {
	beego.Router("/cocos/rpc", &cocos.RpcController{})
	beego.Router("/cocos/insert", &cocos.InsertController{})
	beego.Router("/cocos/remove", &cocos.RemoveController{})
	beego.Router("/cocos/update", &cocos.UpdateController{})
	beego.Router("/cocos/test", &cocos.TestController{})
	beego.Router("/cocos/repush", &cocos.RepushTxController{})
}

// qtum
func QtumInit() {
	beego.Router("/qtum/rpc", &qtum.RpcController{})
	beego.Router("/qtum/insert", &qtum.InsertController{})
	beego.Router("/qtum/remove", &qtum.RemoveController{})
	beego.Router("/qtum/update", &qtum.UpdateController{})
	beego.Router("/qtum/test", &qtum.TestController{})
	beego.Router("/qtum/repush", &qtum.RepushTxController{})
}

// ont
func OntInit() {
	beego.Router("/ont/rpc", &ont.RpcController{})
	beego.Router("/ont/insert", &ont.InsertController{})
	beego.Router("/ont/remove", &ont.RemoveController{})
	beego.Router("/ont/update", &ont.UpdateController{})
	beego.Router("/ont/repush", &ont.RepushTxController{})
}

// hc
func HcInit() {
	beego.Router("/hc/rpc", &hc.RpcController{})
	beego.Router("/hc/insert", &hc.InsertController{})
	beego.Router("/hc/remove", &hc.RemoveController{})
	beego.Router("/hc/update", &hc.UpdateController{})
	beego.Router("/hc/repush", &hc.RepushTxController{})
	beego.Router("/hc/push", &hc.SendTXController{})
	beego.Router("/hc/checkutxo", &hc.CheckTXController{})
}

// hc
func ChainXInit() {
	beego.Router("/chainx/rpc", &chainx.RpcController{})
	beego.Router("/chainx/insert", &chainx.InsertController{})
	beego.Router("/chainx/remove", &chainx.RemoveController{})
	beego.Router("/chainx/update", &chainx.UpdateController{})
	beego.Router("/chainx/repush", &chainx.RepushTxController{})
}

// stacks
func StacksInit() {
	beego.Router("/stacks/rpc", &stacks.RpcController{})
	beego.Router("/stacks/insert", &stacks.InsertController{})
	beego.Router("/stacks/remove", &stacks.RemoveController{})
	beego.Router("/stacks/update", &stacks.UpdateController{})
	beego.Router("/stacks/repush", &stacks.RepushTxController{})
}

// bnb
func BnbInit() {
	beego.Router("/bnb/rpc", &bnb.RpcController{})
	beego.Router("/bnb/insert", &bnb.InsertController{})
	beego.Router("/bnb/remove", &bnb.RemoveController{})
	beego.Router("/bnb/update", &bnb.UpdateController{})
	beego.Router("/bnb/repush", &bnb.RepushTxController{})
	beego.Router("/bnb/repush2", &bnb.RePushController{})
	beego.Router("/bnb/repush3", &bnb.RePushController3{})
}

// mdu
func MduInit() {
	beego.Router("/mdu/rpc", &mdu.RpcController{})
	beego.Router("/mdu/insert", &mdu.InsertController{})
	beego.Router("/mdu/remove", &mdu.RemoveController{})
	beego.Router("/mdu/update", &mdu.UpdateController{})
	beego.Router("/mdu/repush", &mdu.RepushTxController{})
}

//ar
func ArInit() {
	beego.Router("/ar/rpc", &ar.RpcController{})
	beego.Router("/ar/insert", &ar.InsertController{})
	beego.Router("/ar/remove", &ar.RemoveController{})
	beego.Router("/ar/update", &ar.UpdateController{})
	beego.Router("/ar/repush", &ar.RepushTxController{})
	beego.Router("/ar/repush/height", &ar.RepushTxWithHeightController{}) //重推[从节点去拿数据，需要传入高度值和节点ID]
	//beego.Router("/ar/reback", &ar.RebackController{})

	beego.Router("/ar/check", &ar.CheckController{})
}

//ve
func VeInit() {
	beego.Router("/vet/rpc", &ve.RpcController{})
	beego.Router("/vet/insert", &ve.InsertController{})
	beego.Router("/vet/remove", &ve.RemoveController{})
	beego.Router("/vet/update", &ve.UpdateController{})
	beego.Router("/vet/repush", &ve.RepushTxController{})

	//beego.Router("/ar/repush/height", &ar.RepushTxWithHeightController{}) //重推[从节点去拿数据，需要传入高度值和节点ID]
	//beego.Router("/ar/reback", &ar.RebackController{})

	beego.Router("/vet/check", &ve.CheckController{})
}

//rsk
func RifInit() {
	beego.Router("/rsk/rpc", &rsk.RpcController{})
	beego.Router("/rsk/insert", &rsk.InsertController{})
	beego.Router("/rsk/remove", &rsk.RemoveController{})
	beego.Router("/rsk/update", &rsk.UpdateController{})
	beego.Router("/rsk/repush", &rsk.RepushTxController{})
	//beego.Router("/rsk/reback", &rsk.RebackController{})
}

// fo
func FibosInit() {
	beego.Router("/fibos/rpc", &fibos.RpcController{})
	beego.Router("/fibos/insert", &fibos.InsertController{})
	beego.Router("/fibos/remove", &fibos.RemoveController{})
	beego.Router("/fibos/update", &fibos.UpdateController{})
	beego.Router("/fibos/repush", &fibos.RepushTxController{})

	beego.Router("/fibos/insertContract", &fibos.InsertContractController{})
	beego.Router("/fibos/removeContract", &fibos.RemoveContractController{})
	beego.Router("/fibos/updateContract", &fibos.UpdateContractController{})
}

// eos
func EosInit() {
	beego.Router("/eos/rpc", &eos.RpcController{})
	beego.Router("/eos/insert", &eos.InsertController{})
	beego.Router("/eos/remove", &eos.RemoveController{})
	beego.Router("/eos/update", &eos.UpdateController{})
	beego.Router("/eos/repush", &eos.RepushTxController{})

	beego.Router("/eos/insertContract", &eos.InsertContractController{})
	beego.Router("/eos/removeContract", &eos.RemoveContractController{})
	beego.Router("/eos/updateContract", &eos.UpdateContractController{})
}

// qtum
func DashInit() {
	beego.Router("/dash/rpc", &dash.RpcController{})
	beego.Router("/dash/insert", &dash.InsertController{})
	beego.Router("/dash/remove", &dash.RemoveController{})
	beego.Router("/dash/update", &dash.UpdateController{})
	beego.Router("/dash/repush", &dash.RepushTxController{})
}

// kakao
func KakaoInit() {
	beego.Router("/kakao/rpc", &kakao.RpcController{})
	beego.Router("/kakao/insert", &kakao.InsertController{})
	beego.Router("/kakao/remove", &kakao.RemoveController{})
	beego.Router("/kakao/update", &kakao.UpdateController{})
	beego.Router("/kakao/repush", &kakao.RepushTxController{})
}

// dcr
func DcrInit() {
	beego.Router("/dcr/rpc", &dcr.RpcController{})
	beego.Router("/dcr/insert", &dcr.InsertController{})
	beego.Router("/dcr/remove", &dcr.RemoveController{})
	beego.Router("/dcr/update", &dcr.UpdateController{})
	beego.Router("/dcr/repush", &dcr.RepushTxController{})
	beego.Router("/dcr/push", &dcr.SendTXController{})
	beego.Router("/dcr/checkutxo", &dcr.CheckTXController{})
}

// zvc
func ZvcInit() {
	beego.Router("/zvc/rpc", &zvc.RpcController{})
	beego.Router("/zvc/insert", &zvc.InsertController{})
	beego.Router("/zvc/remove", &zvc.RemoveController{})
	beego.Router("/zvc/update", &zvc.UpdateController{})
	beego.Router("/zvc/repush", &zvc.RepushTxController{})
}

//neo
func NeoInit() {
	beego.Router("/neo/rpc", &neo.RpcController{})
	beego.Router("/neo/insert", &neo.InsertController{})
	beego.Router("/neo/remove", &neo.RemoveController{})
	beego.Router("/neo/update", &neo.UpdateController{})
	beego.Router("/neo/test", &neo.TestController{})
	beego.Router("/neo/repush", &neo.RepushTxController{})
	beego.Router("/neo/reback", &neo.RebackController{})
}

// ckb
func CkbInit() {
	beego.Router("/ckb/rpc", &ckb.RpcController{})
	beego.Router("/ckb/insert", &ckb.InsertController{})
	beego.Router("/ckb/remove", &ckb.RemoveController{})
	beego.Router("/ckb/update", &ckb.UpdateController{})
	beego.Router("/ckb/repush", &ckb.RepushTxController{})
}

// ksm
func KsmInit() {
	beego.Router("/ksm/rpc", &ksm.RpcController{})
	beego.Router("/ksm/insert", &ksm.InsertController{})
	beego.Router("/ksm/remove", &ksm.RemoveController{})
	beego.Router("/ksm/update", &ksm.UpdateController{})
	beego.Router("/ksm/repush", &ksm.RepushTxController{})
	beego.Router("/ksm/repush2", &ksm.KsmRepushTx2Controller{})
	beego.Router("/ksm/repush/height", &ksm.RepushTxWithHeightController{}) //重推[从节点去拿数据，需要传入高度值和节点ID]

	beego.Router("/ksm/check", &ksm.CheckController{})
}

// dhx
func DhxInit() {
	beego.Router("/dhx/rpc", &dhx.RpcController{})
	beego.Router("/dhx/insert", &dhx.InsertController{})
	beego.Router("/dhx/remove", &dhx.RemoveController{})
	beego.Router("/dhx/update", &dhx.UpdateController{})
	beego.Router("/dhx/repush", &dhx.RepushTxController{})
	beego.Router("/dhx/repush2", &dhx.DhxRepushTx2Controller{})
	beego.Router("/dhx/repush/height", &dhx.RepushTxWithHeightController{}) //重推[从节点去拿数据，需要传入高度值和节点ID]

	beego.Router("/dhx/check", &dhx.CheckController{})
}

// seek
func SeekInit() {
	beego.Router("/seek/rpc", &seek.RpcController{})
	beego.Router("/seek/insert", &seek.InsertController{})
	beego.Router("/seek/remove", &seek.RemoveController{})
	beego.Router("/seek/update", &seek.UpdateController{})
	beego.Router("/seek/repush", &seek.RepushTxController{})
}

// agent
func AgentInit() {
	beego.Router("/agent/index", &agent.IndexController{})
	beego.Router("/agent/repush", &agent.RepushController{})
	beego.Router("/agent/:coin", &agent.RpcController{})
	beego.Router("/agent/gotask", &agent.GoTaskController{})
}

// addr manager
func AddrManagerInit() {
	beego.Router("/contract/insert", &addrManager.InsertController{})
	beego.Router("/contract/update", &addrManager.UpdateController{})
	beego.Router("/contract/remove", &addrManager.RemoveController{})
}

// 验证路由
func AuthInit() {
	beego.Router("/auth/test", &auth.TestController{})
}

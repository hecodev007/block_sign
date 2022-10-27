package log

import (
	"encoding/json"
	"go.uber.org/zap"
	"testing"
)

func Test_log(t *testing.T){
	InitLogger("dev", "info", "json", "tmplogs/info", "tmplogs/err")
	Info("a")
	Debug("debug,最灵繁的人也看不见自己的背脊")
	Info("info,最困难的事情就是认识自己。")
	Warn("warn,有勇气承担命运这才是英雄好汉")
	Error("error,与肝胆人共事，无字句处读书。")
	DPanic("dpanic,阅读使人充实，会谈使人敏捷，写作使人精确。")
	//Panic("panic,最大的骄傲于最大的自卑都表示心灵的最软弱无力。")
	//Fatal("fatal,自知之明是最难得的知识。")
	Debugf("debugf,勇气通往天堂，怯懦通往地狱。")
	Infof("infof,有时候读书是%s一种巧%s妙地避开思考%s的方法。","test","demo","done")
	Warnf("warnf,阅读%s一切好书%s如同和过去%s最杰出的人谈话。","test","demo","done")
	Errorf("errorf,越是%s没有本领%s的就越加%s自命不凡。","test","demo","done")
	DPanicf("dpanicf,越是%s无能的人，%s越喜欢挑剔%s别人的错儿。","test","demo","done")
	//logger, _ := zap.NewDevelopment()
	////defer logger.Sync()
	//logger.Info("无法获取网址")


}


func Test_Sample(t *testing.T) {
	//logger,_ = zap.NewProduction(zap.AddCaller())
	logger,_ = zap.NewDevelopment(zap.AddCaller())
	//defer _log.Sync()
	//InitLog()
	Debug("debug,最灵繁的人也看不见自己的背脊")
	Info("info,最困难的事情就是认识自己。")
	Warn("warn,有勇气承担命运这才是英雄好汉")
	Error("error,与肝胆人共事，无字句处读书。")
	DPanic("dpanic,阅读使人充实，会谈使人敏捷，写作使人精确。")
	//Panic("panic,最大的骄傲于最大的自卑都表示心灵的最软弱无力。")
	//Fatal("fatal,自知之明是最难得的知识。")
	Debugf("debugf,勇气通往天堂，怯懦通往地狱。")
	Infof("infof,有时候读书是%s一种巧%s妙地避开思考%s的方法。","test","demo","done")
	Warnf("warnf,阅读%s一切好书%s如同和过去%s最杰出的人谈话。","test","demo","done")
	Errorf("errorf,越是%s没有本领%s的就越加%s自命不凡。","test","demo","done")
	DPanicf("dpanicf,越是%s无能的人，%s越喜欢挑剔%s别人的错儿。","test","demo","done")
	//Panicf("panicf,知人者智%s，自知者明%s。胜人者有力%s，自胜者强。","test","demo","done")
	//Fatalf("fatalf,意志坚强%s的人能把%s世界放在手中像泥块%s一样任意揉捏。","test","demo","done")
}
func Json(d interface{})string{
	str,_ :=json.Marshal(d)
	return string(str)
}

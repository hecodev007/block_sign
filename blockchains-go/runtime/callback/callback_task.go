package callback

import (
	"github.com/group-coldwallet/blockchains-go/pkg/timer"
)

//初始化推送任务，暂时不使用
var CallBackTask *timer.TaskScheduler

func InitCallBackTask() {
	cron := timer.GetTaskScheduler()
	go cron.Start()
}

package global

/*
   在使用go的robfig/con定时任务时,大家很容易忽略的一个问题
   cronTab.AddFunc方法每次间隔时间内都会执行,如果在间隔时间
   内你要处理的任务没有完成

   定时任务会再次发布一个go的channel,你的之前的
   任务的定时任务执行的channel并不会停止也会执行
   完成才会停止

   这就会出现同一任务多次执行同样操作的的问题,如果是
   数据的插入或者唯一性有要求的时候,就会出现问题.
*/
//全局变量标识
var G_TASK_MARK = false

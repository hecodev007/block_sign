package server

import (
	"context"
	"github.com/group-coldwalle/coinsign/qieusdtserver/api"
	"github.com/group-coldwalle/coinsign/qieusdtserver/config"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Run(cfg *config.GlobalConfig) {
	srv := &http.Server{
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		Addr:         cfg.HttpPort,
		Handler:      makeRouter(),
	}
	go func() {
		err := srv.ListenAndServe()
		logrus.Info(err)
	}()
	logrus.Info("use Ctrl + c stop server.")

	signalChan := make(chan os.Signal)
	// 监听指定信号
	signal.Notify(signalChan, syscall.SIGUSR1, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signalChan
	//关闭上传多线程--------
	api.StopUploadJob()
	//--------------------
	logrus.Info("get signal: ", sig, ". sever will stop and showdown after 30 second!")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel() //源码内超时会自动调用这个cancel,这里其实也可以不用执行这个
	srv.Shutdown(ctx)
	logrus.Info("server showdown")
}

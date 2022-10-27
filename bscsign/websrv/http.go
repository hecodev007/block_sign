package websrv

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/bsc-sign/conf"
	"github.com/bsc-sign/routers"
	"github.com/bsc-sign/util/log"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

type WebSrv struct {
	httpSrv *http.Server
}

func NewWebSrv(ctx context.Context) *WebSrv {
	r := gin.Default()
	path := fmt.Sprintf("%s/%s", strings.ToLower(conf.Config.Version), strings.ToLower(conf.Config.CoinType))
	group := r.Group(path)
	// 初始化路由
	routers.InitRouters(ctx, group)

	srv := &http.Server{
		Addr:         ":" + conf.Config.Port,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		Handler:      r,
	}

	return &WebSrv{
		httpSrv: srv,
	}
}

func (w *WebSrv) StartAsync() {
	go func() {
		if err := w.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s\n", err)
		}
	}()
	logrus.Info("Web服务已启动")
}

func (w *WebSrv) Stop() {
	logrus.Infof("准备停止web服务...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.httpSrv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
		logrus.Info("Web服务已停止")
	}
}

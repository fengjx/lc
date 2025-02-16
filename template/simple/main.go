package main

import (
	"time"

	"github.com/fengjx/go-halo/halo"
	"github.com/fengjx/luchen"
	"github.com/fengjx/luchen/log"
	"go.uber.org/zap"

	"github.com/fengjx/lc/simple/endpoint"
	"github.com/fengjx/lc/simple/middleware"
	"github.com/fengjx/lc/simple/transport/http"
)

func main() {
	luchen.UseGlobalHTTPMiddleware(
		luchen.LogMiddleware,
		middleware.AccessMiddleware,
	)
	hs := http.GetServer()
	endpoint.Init(hs)
	luchen.Start(hs)

	// 监听kill信号，收到信号后最多等待30秒处理程序退出逻辑
	if err := halo.Wait(time.Second * 30); err != nil {
		log.Info("server shutdown err", zap.Error(err))
	}
}

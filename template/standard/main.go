package main

import (
	"time"

	"github.com/fengjx/go-halo/halo"
	"github.com/fengjx/luchen"

	"github.com/fengjx/lc/standard/logic"
	"github.com/fengjx/lc/standard/transport/grpc"
	"github.com/fengjx/lc/standard/transport/http"
)

func main() {
	hs := http.GetServer()
	gs := grpc.GetServer()

	logic.Init(hs, gs)

	// 启动服务
	luchen.Start(hs, gs)

	// 注册服务到 etcd 并启动服务，如果启用，可以把 luchen.Start(hs, gs) 删除并去掉下面注释
	//registrar := luchen.NewEtcdV3Registrar(
	//	hs,
	//	gs,
	//)
	//// 把服务注册到 etcd 并启动
	//registrar.Register()

	// 阻塞服务并监听 kill 信号，收到 kill 信号后退出（最长等待10秒）
	halo.Wait(10 * time.Second)
}

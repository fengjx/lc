package grpc

import (
	"github.com/fengjx/luchen"
	"github.com/fengjx/go-halo/halo"
	"github.com/fengjx/lc/standard/common/config"
)

var s = halo.NewSingleton[luchen.GRPCServer](func() *luchen.GRPCServer {
	serverConfig := config.GetConfig().Server.GRPC
	gs := luchen.NewGRPCServer(
		luchen.WithServiceName(serverConfig.ServerName),
	)
	return gs
})

func GetServer() *luchen.GRPCServer {
	return s.Get()
}

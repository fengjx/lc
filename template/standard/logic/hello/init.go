package hello

import (
	"github.com/fengjx/lc/standard/logic/hello/internal/endpoint"
	"github.com/fengjx/luchen"
)

func Init(hs *luchen.HTTPServer, gs *luchen.GRPCServer) {
	endpoint.Init(hs, gs)
}

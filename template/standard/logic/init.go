package logic

import (
	"github.com/fengjx/luchen"

	"github.com/fengjx/lc/standard/logic/hello"
	"github.com/fengjx/lc/standard/pkg/lifecycle"
)

func Init(hs *luchen.HTTPServer, gs *luchen.GRPCServer) {
	hello.Init(hs, gs)

	lifecycle.DoHooks()
}

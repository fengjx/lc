package endpoint

import (
	"github.com/fengjx/lc/standard/logic/hello/internal/endpoint/greet"
	"github.com/fengjx/luchen"
)

func Init(hs *luchen.HTTPServer, gs *luchen.GRPCServer) {
	greet.RegisterGreeterHTTPHandler(hs)
	greet.RegisterGreeterGRPCHandler(gs)
}

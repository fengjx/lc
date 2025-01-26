package endpoint

import (
	"github.com/fengjx/lc/simple/endpoint/hello"
	"github.com/fengjx/luchen"
)

func Init(hs *luchen.HTTPServer) {
	hello.RegisterHelloHTTPHandler(hs)
}

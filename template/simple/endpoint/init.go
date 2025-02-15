package endpoint

import (
	"github.com/fengjx/luchen"

	"github.com/fengjx/lc/simple/endpoint/hello"
)

func Init(hs *luchen.HTTPServer) {
	hello.RegisterGreeterHTTPHandler(hs)
}

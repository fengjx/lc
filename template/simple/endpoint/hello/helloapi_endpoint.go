package hello

import (
	"context"
	"reflect"

	"github.com/fengjx/lc/simple/proto"
	"github.com/fengjx/luchen"
)

func RegisterHelloHTTPHandler(hs *luchen.HTTPServer) {
	e := &helloEndpoint{}
	hs.Handle(&luchen.EndpointDefine{
		Endpoint: e.makeSayHelloEndpoint(),
		Name:     "Hello.SayHello",
		Path:     "/hello/say-hello",
		ReqType:  reflect.TypeOf(&proto.SayHelloReq{}),
		RspType:  reflect.TypeOf(&proto.SayHelloRsp{}),
	})
}

type helloEndpoint struct {
}

func (e *helloEndpoint) makeSayHelloEndpoint() luchen.Endpoint {
	return func(ctx context.Context, request any) (response any, err error) {
		req := request.(*proto.SayHelloReq)
		response = &proto.SayHelloRsp{
			Msg: "hello " + req.Name,
		}
		return
	}
}

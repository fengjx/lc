package pbgreet

import (
	"context"
	"fmt"

	"github.com/fengjx/lc/test/proto/pbgreet"
	"github.com/fengjx/luchen"
)

// RegisterGreeterGRPCHandler 注册 GRPC 服务处理器
func RegisterGreeterGRPCHandler(gs *luchen.GRPCServer) {
	pbgreet.RegisterGreeterGRPCHandler(gs, GreeterEndpointImpl)
}

// RegisterGreeterHTTPHandler 注册 HTTP 服务处理器
func RegisterGreeterHTTPHandler(hs *luchen.HTTPServer) {
	pbgreet.RegisterGreeterHTTPHandler(hs, GreeterEndpointImpl)
}

// GreeterEndpointImpl 默认的服务实现
var GreeterEndpointImpl = &GreeterEndpoint{
	handler: &GreeterHandlerImpl{},
}

// GreeterHandlerImpl 服务处理器实现
type GreeterHandlerImpl struct {
}

// GreeterEndpoint 服务 Endpoint 定义
type GreeterEndpoint struct {
	handler pbgreet.GreeterHandler
}


// SayHello Sends a greeting
// http.path=/greeter/say-hello
func (e *GreeterEndpoint) SayHelloEndpoint() luchen.Endpoint {
	fn := func(ctx context.Context, request any) (any, error) {
		req, ok := request.(*pbgreet.HelloReq)
		if !ok {
			return nil, fmt.Errorf("invalid request type: %T", request)
		}
		return e.handler.SayHello(ctx, req)
	}
	return fn
}

// Ping ping service
// http.path=/greeter/ping
func (e *GreeterEndpoint) PingEndpoint() luchen.Endpoint {
	fn := func(ctx context.Context, request any) (any, error) {
		req, ok := request.(*pbgreet.PingReq)
		if !ok {
			return nil, fmt.Errorf("invalid request type: %T", request)
		}
		return e.handler.Ping(ctx, req)
	}
	return fn
}
 
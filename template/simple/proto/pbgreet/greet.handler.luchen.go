// Code generated by lc pbgen. DO NOT EDIT.
package pbgreet

import (
	"context"

	grpctransport "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"reflect"
	"sync"

	"github.com/fengjx/luchen"
)

// NewGreeterService 返回一个 GreeterClient
func NewGreeterService(serverName string) GreeterClient {
	cli := luchen.GetGRPCClient(
		serverName,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	return NewGreeterClient(cli)
}

// GreeterHandler 定义服务处理器接口
type GreeterHandler interface {
	SayHello(ctx context.Context, in *HelloReq) (*HelloResp, error)
}

// GreeterEndpoint 定义服务 Endpoint 接口
type GreeterEndpoint interface {
	SayHelloEndpoint() luchen.Endpoint
}

// GreeterServiceImpl 服务实现
type GreeterServiceImpl struct {
	UnimplementedGreeterServer
	middlewares    []luchen.Middleware
	endpoint       GreeterEndpoint
	sayhelloDefine *luchen.EndpointDefine
	sayhello       grpctransport.Handler
}

var (
	greeterServiceImplOnce = sync.Once{}
	greeterServiceImpl     *GreeterServiceImpl
)

// GetGreeterServiceImpl 获取服务实现的单例
func GetGreeterServiceImpl(e GreeterEndpoint, middlewares ...luchen.Middleware) *GreeterServiceImpl {
	greeterServiceImplOnce.Do(func() {
		greeterServiceImpl = newGreeterServiceImpl(e, middlewares...)
	})
	return greeterServiceImpl
}

// newGreeterServiceImpl 创建新的服务实现
func newGreeterServiceImpl(e GreeterEndpoint, middlewares ...luchen.Middleware) *GreeterServiceImpl {
	sayhelloDefine := &luchen.EndpointDefine{
		Name:        "Greeter.SayHello",
		Path:        "/say-hello",
		ReqType:     reflect.TypeOf(&HelloReq{}),
		RspType:     reflect.TypeOf(&HelloResp{}),
		Endpoint:    e.SayHelloEndpoint(),
		Middlewares: middlewares,
	}
	impl := &GreeterServiceImpl{
		endpoint:       e,
		sayhelloDefine: sayhelloDefine,
	}
	impl.sayhello = luchen.NewGRPCTransportServer(sayhelloDefine)
	return impl
}

// SayHello Sends a greeting
// http.path=/say-hello
func (s *GreeterServiceImpl) SayHello(ctx context.Context, req *HelloReq) (*HelloResp, error) {
	_, resp, err := s.sayhello.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*HelloResp), nil
}

// RegisterGreeterGRPCHandler 注册 GRPC 接口实现
func RegisterGreeterGRPCHandler(gs *luchen.GRPCServer, e GreeterEndpoint, middlewares ...luchen.Middleware) {
	impl := GetGreeterServiceImpl(e, middlewares...)
	RegisterGreeterServer(gs, impl)
}

// RegisterGreeterHTTPHandler 注册 HTTP 请求路由
func RegisterGreeterHTTPHandler(hs *luchen.HTTPServer, e GreeterEndpoint, middlewares ...luchen.Middleware) {
	impl := GetGreeterServiceImpl(e, middlewares...)
	if impl.sayhelloDefine.Path != "" {
		hs.Handle(impl.sayhelloDefine)
	}
}

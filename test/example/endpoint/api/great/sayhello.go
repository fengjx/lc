package pbgreet

import (
	"context"
	"fmt"

	"github.com/fengjx/lc/test/proto/pbgreet"
	"github.com/fengjx/luchen"
)

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

// SayHello Sends a greeting
// http.path=/greeter/say-hello
func (h *GreeterHandlerImpl) SayHello(ctx context.Context, req *pbgreet.HelloReq) (*pbgreet.HelloResp, error) {
	// TODO: implement me
	return &pbgreet.HelloResp{}, nil
}

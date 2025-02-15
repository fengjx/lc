package greet

import (
	"context"
	"fmt"

	"github.com/fengjx/lc/standard/logic/hello/internal/proto/pbgreet"
	"github.com/fengjx/luchen"
)

func (e *GreeterEndpoint) SayHelloEndpoint() luchen.Endpoint {
	fn := func(ctx context.Context, request any) (any, error) {
		req, ok := request.(*pbgreet.HelloReq)
		if !ok {
			msg := fmt.Sprintf("invalid request type: %T", request)
			return nil, luchen.ErrBadRequest.WithMsg(msg)
		}
		return e.handler.SayHello(ctx, req)
	}
	return fn
}

// SayHello Sends a greeting
// http.path=/hello/say-hello
func (h *GreeterHandlerImpl) SayHello(ctx context.Context, req *pbgreet.HelloReq) (*pbgreet.HelloResp, error) {
	return &pbgreet.HelloResp{
		Message: fmt.Sprintf("hi: %s", req.Name),
	}, nil
}

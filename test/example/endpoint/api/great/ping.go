package pbgreet

import (
	"context"
	"fmt"

	"github.com/fengjx/lc/test/proto/pbgreet"
	"github.com/fengjx/luchen"
)

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

// Ping ping service
// http.path=/greeter/ping
func (h *GreeterHandlerImpl) Ping(ctx context.Context, req *pbgreet.PingReq) (*pbgreet.PingResp, error) {
	// TODO: implement me
	return &pbgreet.PingResp{}, nil
}

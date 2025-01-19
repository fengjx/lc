package great

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fengjx/lctest/proto/pbgreet"
	"github.com/fengjx/luchen"
)

func (e *GreeterEndpoint) PingEndpoint() luchen.Endpoint {
	fn := func(ctx context.Context, request any) (any, error) {
		req, ok := request.(*pbgreet.PingReq)
		if !ok {
			msg := fmt.Sprintf("invalid request type: %T", request)
			return nil, luchen.NewErrno(http.StatusBadRequest, msg)
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

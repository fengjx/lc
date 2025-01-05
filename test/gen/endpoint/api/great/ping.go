package pbgreet

import (
	"context"

	"github.com/fengjx/lc/test/proto/pbgreet"
)
// Ping ping service
// http.path=/greeter/ping
func (h *GreeterHandlerImpl) Ping(ctx context.Context, req *pbgreet.PingReq) (*pbgreet.PingResp, error) {
	// TODO: implement me
	return &pbgreet.PingResp{}, nil
}
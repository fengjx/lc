package pbgreet

import (
	"context"

	"github.com/fengjx/lc/test/proto/pbgreet"
)
// SayHello Sends a greeting
// http.path=/greeter/say-hello
func (h *GreeterHandlerImpl) SayHello(ctx context.Context, req *pbgreet.HelloReq) (*pbgreet.HelloResp, error) {
	// TODO: implement me
	return &pbgreet.HelloResp{}, nil
}
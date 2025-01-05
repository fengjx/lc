package greet

import (
	"context"

	"github.com/fengjx/lc/test/pb/pbgreet"
)

func (h *GreeterHandlerImpl) SayHello(ctx context.Context, req *pbgreet.HelloReq) (*pbgreet.HelloResp, error) {
	msg := "hello: " + req.Name
	return &pbgreet.HelloResp{
		Message: msg,
	}, nil
}

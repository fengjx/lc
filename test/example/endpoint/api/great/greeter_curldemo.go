package great

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fengjx/lctest/proto/pbgreet"
	"github.com/fengjx/luchen"
)

func (e *GreeterEndpoint) CurlDemoEndpoint() luchen.Endpoint {
	fn := func(ctx context.Context, request any) (any, error) {
		req, ok := request.(*pbgreet.CurlDemoReq)
		if !ok {
			msg := fmt.Sprintf("invalid request type: %T", request)
			return nil, luchen.NewErrno(http.StatusBadRequest, msg)
		}
		return e.handler.CurlDemo(ctx, req)
	}
	return fn
}

// CurlDemo 示例
// http.path=/greeter/curl-demo
func (h *GreeterHandlerImpl) CurlDemo(ctx context.Context, req *pbgreet.CurlDemoReq) (*pbgreet.CurlDemoRsp, error) {
	// TODO: implement me
	return &pbgreet.CurlDemoRsp{}, nil
}

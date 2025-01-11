package great

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fengjx/lctest/proto/pbgreet"
	"github.com/fengjx/luchen"
)

func (e *GreeterEndpoint) GetUserEndpoint() luchen.Endpoint {
	fn := func(ctx context.Context, request any) (any, error) {
		req, ok := request.(*pbgreet.GetUserReq)
		if !ok {
			msg := fmt.Sprintf("invalid request type: %T", request)
			return nil, luchen.NewErrno(http.StatusBadRequest, msg)
		}
		return e.handler.GetUser(ctx, req)
	}
	return fn
}

// GetUser 实现 GreeterHandler 接口中的 GetUser 方法
func (h *GreeterHandlerImpl) GetUser(ctx context.Context, req *pbgreet.GetUserReq) (*pbgreet.GetUserResp, error) {
	// TODO: implement me
	return &pbgreet.GetUserResp{}, nil
}

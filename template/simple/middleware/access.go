package middleware

import (
	"github.com/fengjx/luchen"
)

// AccessMiddleware 请求日志
var AccessMiddleware = luchen.AccessMiddleware(&luchen.AccessLogOpt{
	MaxDay: 15,
})

package http

import (
	"net/http"

	"github.com/fengjx/go-halo/halo"
	"github.com/fengjx/lc/simple/config"
	"github.com/fengjx/luchen"
	"github.com/fengjx/xin/middleware"
	"github.com/fengjx/xin/pprof"
)

var serverSingle = halo.NewSingleton[luchen.HTTPServer](func() *luchen.HTTPServer {
	serverConfig := config.GetConfig().Server.HTTP
	hs := luchen.NewHTTPServer(
		luchen.WithServiceName(serverConfig.ServerName),
		luchen.WithServerAddr(serverConfig.Listen),
	)
	mux := hs.Mux()
	mux.Use(
		middleware.Recoverer,
		middleware.RequestID,
		middleware.RealIP,
		middleware.CorsHandler(middleware.CorsOptions{
			AllowedOrigins: serverConfig.Cors.AllowOrigins,
			AllowedMethods: []string{
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
			},
			AllowedHeaders: []string{"*"},
			ExposedHeaders: []string{
				"Content-Disposition",
				"Content-Type",
			},
			AllowCredentials: true,
		}),
	)
	// 开启 pprof，使用basic认证，用户名和密码为foo/bar
	mux.Handle(pprof.DefaultPrefix, pprof.Profiler(map[string]string{
		"fengjx": "hello1024",
	}))
	return hs
})

func GetServer() *luchen.HTTPServer {
	return serverSingle.Get()
}

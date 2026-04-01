package global

import "github.com/sky-night-net/snet/web"

var (
	webServer *web.Server
)

func SetWebServer(s *web.Server) {
	webServer = s
}

func GetWebServer() *web.Server {
	return webServer
}

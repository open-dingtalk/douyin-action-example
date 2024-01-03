package actions

import (
	"douyin-action-example/internal/actions/controllers"
	"github.com/chzealot/gobase/logger"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net"
	"net/http"
)

type HttpServer struct {
}

func NewHttpServer() *HttpServer {
	return &HttpServer{}
}

func (s *HttpServer) Run(address string) error {
	logger.Infof("run http server on %s", address)
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	asset := controllers.NewAssetHandler()
	r.GET("/openapi.yaml", asset.OpenApiSpecYaml)

	ac := controllers.NewAuthController()
	r.GET("/auth/authorize", ac.Authorize)
	r.POST("/auth/token", ac.Token)
	r.GET("/auth/callback", ac.Callback)

	bc := controllers.NewBizController()
	r.POST("/userInfo", bc.UserInfo)
	r.GET("/videoList", bc.GetVideoList)
	r.GET("/videoBase", bc.GetVideoBase)

	//ph := controllers.NewProfileHandler()
	//r.GET("/profile/:unionId", ph.GetProfile)
	//
	//ch := controllers.NewCalendarHandler()
	//r.GET("/calendar/events", ch.GetEvents)
	//r.GET("/calendars", ch.GetCalendars)
	//
	//r.NoRoute(controllers.NewNotFoundHandler().Process)

	server := &http.Server{Addr: address, Handler: r}
	ln, err := net.Listen("tcp4", address)
	if err != nil {
		return err
	}
	type tcpKeepAliveListener struct {
		*net.TCPListener
	}
	return errors.WithStack(server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)}))
}

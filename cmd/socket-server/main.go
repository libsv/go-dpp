package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	echoProm "github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/libsv/go-p4/cmd/internal"
	"github.com/libsv/go-p4/config"
	"github.com/libsv/go-p4/data/sockets"
	"github.com/libsv/go-p4/docs"
	"github.com/libsv/go-p4/log"
	"github.com/libsv/go-p4/service"
	p4Handlers "github.com/libsv/go-p4/transports/http"
	p4Middleware "github.com/libsv/go-p4/transports/http/middleware"
	p4soc "github.com/libsv/go-p4/transports/sockets"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	echoSwagger "github.com/swaggo/echo-swagger"
	smw "github.com/theflyingcodr/sockets/middleware"
	"github.com/theflyingcodr/sockets/server"
)

const appname = "payment-protocol-socket-server"
const banner = `
====================================================================
         _         _       _            _            _     
        /\ \      /\ \    /\ \        /\ \          _\ \   
       /  \ \    /  \ \   \_\ \      /  \ \        /\__ \  
      / /\ \ \  / /\ \ \  /\__ \    / /\ \ \      / /_ \_\ 
     / / /\ \_\/ / /\ \_\/ /_ \ \  / / /\ \ \    / / /\/_/ 
    / / /_/ / / / /_/ / / / /\ \ \/ / /  \ \_\  / / /      
   / / /__\/ / / /__\/ / / /  \/_/ / /    \/_/ / / /       
  / / /_____/ / /_____/ / /     / / /         / / / ____   
 / / /     / / /     / / /     / / /________ / /_/_/ ___/\ 
/ / /     / / /     /_/ /     / / /_________/_______/\__\/ 
\/_/      \/_/      \_\/      \/____________\_______\/     

====================================================================
`

// main is the entry point of the application.
// @title Payment Protocol Socket Server
// @version 0.0.1
// @description Payment Protocol Server is an implementation of a Bip-270 payment flow.
// @termsOfService https://github.com/libsv/go-payment_protocol/blob/master/CODE_STANDARDS.md
// @license.name ISC
// @license.url https://github.com/libsv/go-payment_protocol/blob/master/LICENSE
// @host localhost:8445
// @schemes:
//	- http
//	- https
func main() {
	println("\033[32m" + banner + "\033[0m")
	config.SetupDefaults()
	cfg := config.NewViperConfig(appname).
		WithServer().
		WithDeployment(appname).
		WithLog().
		WithPayD().
		WithSockets().
		Load()
	log := log.NewZero(cfg.Logging)
	log.Infof("\n------Environment: %#v -----\n", cfg.Server)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	p := echoProm.NewPrometheus("p4", nil)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.HTTPErrorHandler = p4Middleware.ErrorHandler(log)
	if cfg.Server.SwaggerEnabled {
		docs.SwaggerInfo.Host = cfg.Server.SwaggerHost
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}
	p.Use(e)

	deps := internal.SetupDeps(*cfg)
	g := e.Group("/")

	p4Handlers.NewPaymentRequestHandler(deps.PaymentRequestService).RegisterRoutes(g)
	p4Handlers.NewPaymentHandler(deps.PaymentService).RegisterRoutes(g)

	// create socket server
	s := server.New(
		server.WithMaxMessageSize(10000),
		server.WithChannelTimeout(time.Duration(cfg.Sockets.ChannelTimeoutSeconds)*time.Second))
	defer s.Close()
	p4Handlers.NewProofs(service.NewProof(sockets.NewPayd(s))).RegisterRoutes(g)
	// add middleware, with panic going first
	s.WithMiddleware(smw.PanicHandler, smw.Timeout(smw.NewTimeoutConfig()), smw.Metrics())

	// simple metrics
	gCo := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "sockets",
		Subsystem: "server",
		Name:      "gauge_total_connections",
	})

	s.OnClientJoin(func(clientID, channelID string) {
		gCo.Inc()
	})

	s.OnClientLeave(func(clientID, channelID string) {
		gCo.Dec()
	})

	gCh := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "sockets",
		Subsystem: "server",
		Name:      "gauge_total_channels",
	})

	s.OnChannelCreate(func(channelID string) {
		gCh.Inc()
	})

	s.OnChannelClose(func(channelID string) {
		gCh.Dec()
	})
	registerHandlers(s)

	// this is our websocket endpoint, clients will hit this with the channelID they wish to connect to
	e.GET("/ws/:channelID", WsHandler(s))

	if cfg.Deployment.IsDev() {
		internal.PrintDev(e)
	}

	go func() {
		log.Error(e.Start(cfg.Server.Port), "echo server failed")
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Error(err, "")
	}
}

// WsHandler will upgrade connections to a websocket and then wait for messages.
func WsHandler(svr *server.SocketServer) echo.HandlerFunc {
	upgrader := websocket.Upgrader{}
	return func(c echo.Context) error {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		defer func() {
			_ = ws.Close()
		}()
		if err := svr.Listen(ws, c.Param("channelID")); err != nil {
			return err
		}
		return nil
	}
}

func registerHandlers(s *server.SocketServer) {
	p4soc.NewPaymentRequest().Register(s)
	p4soc.NewPayment().Register(s)
}

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/libsv/go-p4/cmd/internal"
	"github.com/libsv/go-p4/config"
	p4soc "github.com/libsv/go-p4/transports/sockets"
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

var (
	upgrader = websocket.Upgrader{}
)

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
		Load()
	config.SetupLog(cfg.Logging)
	log.Infof("\n------Environment: %s -----\n", cfg.Server)

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	p := prometheus.NewPrometheus("echo", nil)
	p.Use(e)

	// create socket server
	s := server.New(server.WithMaxMessageSize(10000))
	defer s.Close()

	// add middleware, with panic going first
	s.WithMiddleware(smw.PanicHandler, smw.Timeout(smw.NewTimeoutConfig()), smw.Metrics())

	registerHandlers(s, internal.SetupDeps(*cfg))

	// this is our websocket endpoint, clients will hit this with the channelID they wish to connect to
	e.GET("/ws/:channelID", WsHandler(s))

	if cfg.Deployment.IsDev() {
		internal.PrintDev(e)
	}

	go func() {
		log.Error(e.Start(cfg.Server.Port))
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Error(err)
	}
}

func WsHandler(svr *server.SocketServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		defer ws.Close()
		if err := svr.Listen(ws, c.Param("channelID")); err != nil {
			return err
		}

		fmt.Println("exiting listener")
		return nil
	}
}

func registerHandlers(s *server.SocketServer, d *internal.Deps) {
	p4soc.NewPaymentRequest().Register(s)
	p4soc.NewPayment().Register(s)
}

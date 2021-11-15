package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/libsv/go-p4/cmd/internal"
	"github.com/libsv/go-p4/config"
	"github.com/libsv/go-p4/docs"
	"github.com/libsv/go-p4/log"
	p4Handlers "github.com/libsv/go-p4/transports/http"

	p4Middleware "github.com/libsv/go-p4/transports/http/middleware"
)

const appname = "payment-protocol-rest-server"
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
// @title Payment Protocol Server
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
	log := log.NewZero(cfg.Logging)
	log.Infof("\n------Environment: %s -----\n", cfg.Server)

	e := echo.New()
	e.HideBanner = true
	g := e.Group("/")
	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.HTTPErrorHandler = p4Middleware.ErrorHandler(log)
	if cfg.Server.SwaggerEnabled {
		docs.SwaggerInfo.Host = cfg.Server.SwaggerHost
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	if cfg.Deployment.IsDev() {
		internal.PrintDev(e)
	}

	// setup services
	deps := internal.SetupDeps(*cfg)

	// handlers
	p4Handlers.NewPaymentHandler(deps.PaymentService).RegisterRoutes(g)
	p4Handlers.NewPaymentRequestHandler(deps.PaymentRequestService).RegisterRoutes(g)

	e.Logger.Fatal(e.Start(cfg.Server.Port))
}

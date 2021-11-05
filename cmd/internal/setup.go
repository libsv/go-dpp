package internal

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"

	"github.com/libsv/go-p4"
	"github.com/libsv/go-p4/config"
	"github.com/libsv/go-p4/data"
	"github.com/libsv/go-p4/data/noop"
	"github.com/libsv/go-p4/data/payd"
	"github.com/libsv/go-p4/service"
)

type Deps struct {
	PaymentService        p4.PaymentService
	PaymentRequestService p4.PaymentRequestService
}

func SetupDeps(cfg config.Config) *Deps {
	httpClient := &http.Client{Timeout: 5 * time.Second}
	if !cfg.PayD.Secure { // for testing, don't validate server cert
		// #nosec
		httpClient.Transport = &http.Transport{
			// #nosec
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	// stores
	paydStore := payd.NewPayD(cfg.PayD, data.NewClient(httpClient))

	// services
	paymentSvc := service.NewPayment(paydStore)
	paymentReqSvc := service.NewPaymentRequest(cfg.Server, paydStore, paydStore)
	if cfg.PayD.Noop {
		noopStore := noop.NewNoOp()
		paymentSvc = service.NewPayment(noopStore)
		paymentReqSvc = service.NewPaymentRequest(cfg.Server, noopStore, noopStore)
	}

	return &Deps{
		PaymentService:        paymentSvc,
		PaymentRequestService: paymentReqSvc,
	}
}

// PrintDev outputs some useful dev information such as http routes
// and current settings being used.
func PrintDev(e *echo.Echo) {
	fmt.Println("==================================")
	fmt.Println("DEV mode, printing http routes:")
	for _, r := range e.Routes() {
		fmt.Printf("%s: %s\n", r.Method, r.Path)
	}
	fmt.Println("==================================")
	fmt.Println("DEV mode, printing settings:")
	for _, v := range viper.AllKeys() {
		fmt.Printf("%s: %v\n", v, viper.Get(v))
	}
	fmt.Println("==================================")
}

package pptcl

import (
	"context"
	"time"

	"github.com/libsv/go-bt/v2"
)

// PaymentRequest message used in BIP270.
// See https://github.com/moneybutton/bips/blob/master/bip-0270.mediawiki#paymentrequest
type PaymentRequest struct {
	// Network  Always set to "bitcoin" (but seems to be set to 'bitcoin-sv'
	// outside bip270 spec, see https://handcash.github.io/handcash-merchant-integration/#/merchant-payments)
	// {enum: bitcoin, bitcoin-sv, test}
	// Required.
	Network string `json:"network" example:"mainnet" enums:"mainnet,testnet,stn,regtest"`
	// Outputs Is an array of outputs. required, but can have zero elements.
	// Required.
	Outputs []Output `json:"outputs"`
	// CreationTimestamp Unix timestamp (seconds since 1-Jan-1970 UTC) when the PaymentRequest was created.
	// Required.
	CreationTimestamp time.Time `json:"creationTimestamp" swaggertype:"primitive,string" example:"2019-10-12T07:20:50.52Z"`
	// ExpirationTimestamp Unix timestamp (UTC) after which the PaymentRequest should be considered invalid.
	// Optional.
	ExpirationTimestamp time.Time `json:"expirationTimestamp" swaggertype:"primitive,string" example:"2019-10-12T07:20:50.52Z"`
	// PaymentURL secure HTTPS location where a Payment message (see below) will be sent to obtain a PaymentACK.
	// Maximum length is 4000 characters
	PaymentURL string `json:"paymentUrl" example:"https://localhost:3443/api/v1/payment/123456"`
	// Memo Optional note that should be displayed to the customer, explaining what this PaymentRequest is for.
	// Maximum length is 50 characters.
	Memo string `json:"memo" example:"invoice number 123456"`
	// MerchantData contains arbitrary data that may be used by the payment host to identify the PaymentRequest.
	// May be omitted if the payment host does not need to associate Payments with PaymentRequest
	// or if they associate each PaymentRequest with a separate payment address.
	// Maximum length is 10000 characters.
	MerchantData *MerchantData `json:"merchantData,omitempty"`
	// FeeRate defines the amount of fees a users wallet should add to the payment
	// when submitting their final payments.
	FeeRate *bt.FeeQuote `json:"fee"`
}

// PaymentRequestArgs are request arguments that can be passed to the service.
type PaymentRequestArgs struct {
	// PaymentID is an identifier for an invoice.
	PaymentID string `param:"paymentID"`
}

// PaymentRequestService can be implemented to enforce business rules
// and process in order to fulfil a PaymentRequest.
type PaymentRequestService interface {
	CreatePaymentRequest(ctx context.Context, args PaymentRequestArgs) (*PaymentRequest, error)
}

// FeeReader can be implemented to read fees from a data store for returning as
// part of the Payment Request.
type FeeReader interface {
	// Fees will return the current fees from a data store.
	Fees(ctx context.Context, args PaymentRequestArgs) (*bt.FeeQuote, error)
}

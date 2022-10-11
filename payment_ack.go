package dpp

import (
	"github.com/libsv/go-dpp/modes/hybridmode"
)

// These structures are defined in the TSC spec:
// See https://tsc.bitcoinassociation.net/standards/direct_payment_protocol

// PaymentACK message used in the TSC DPP spec.
// See https://tsc.bitcoinassociation.net/standards/direct_payment_protocol/#PaymentModes
type PaymentACK struct {
	// ModeID the chosen mode.
	ModeID string `json:"modeId" binding:"required" example:"ef63d9775da5"`
	// Mode data required by specific payment mode
	Mode        *hybridmode.PaymentACK      `json:"mode"`
	PeerChannel *hybridmode.PeerChannelData `json:"peerChannel"`
	RedirectURL string                      `json:"redirectUrl"`
}

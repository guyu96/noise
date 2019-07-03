package noise

import (
	"time"

	"github.com/guyu96/noise/identity"
	"github.com/guyu96/noise/nat"
	"github.com/guyu96/noise/transport"
)

type parameters struct {
	Host               string
	Port, ExternalPort uint16

	NAT       nat.Provider
	Keys      identity.Keypair
	Transport transport.Layer

	Metadata map[string]interface{}

	MaxMessageSize uint64

	SendMessageTimeout    time.Duration
	ReceiveMessageTimeout time.Duration

	SendWorkerBusyTimeout time.Duration
}

func DefaultParams() parameters {
	return parameters{
		Host:           "127.0.0.1",
		Transport:      transport.NewTCP(),
		Metadata:       map[string]interface{}{},
		MaxMessageSize: 1048576,

		SendMessageTimeout:    3 * time.Second,
		ReceiveMessageTimeout: 3 * time.Second,

		SendWorkerBusyTimeout: 3 * time.Second,
	}
}

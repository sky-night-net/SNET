package adapters

import (
	"github.com/sky-night-net/snet/database/model"
)

type Traffic struct {
	Up   int64 `json:"up"`
	Down int64 `json:"down"`
}

type KeyPair struct {
	PrivateKey   string `json:"private_key"`
	PublicKey    string `json:"public_key"`
	PresharedKey string `json:"preshared_key,omitempty"`
}

type VPNAdapter interface {
	// Protocol name identifying this adapter
	Protocol() model.Protocol

	// Process management
	Start(inbound *model.Inbound) error
	Stop(inbound *model.Inbound) error
	IsRunning(inbound *model.Inbound) bool

	// Client management
	AddClient(inbound *model.Inbound, client *model.Client) error
	RemoveClient(inbound *model.Inbound, client *model.Client) error

	// Key generation
	GenerateKeypair() (KeyPair, error)

	// Config generation
	GenerateServerConfig(inbound *model.Inbound) (string, error)
	GenerateClientConfig(inbound *model.Inbound, client *model.Client, host string) (string, error)

	// Traffic stats (returns map of client email/id to traffic)
	GetTraffic(inbound *model.Inbound) (map[string]Traffic, error)

	// Prerequisites check
	CheckPrerequisites() error
}

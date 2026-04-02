package adapters

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/sky-night-net/snet/database/model"
	"golang.org/x/crypto/curve25519"
)

type XrayAdapter struct {
	protocol model.Protocol
}

func NewXrayAdapter(p model.Protocol) *XrayAdapter {
	return &XrayAdapter{protocol: p}
}

func (a *XrayAdapter) Protocol() model.Protocol {
	return a.protocol
}

func (a *XrayAdapter) Start(inbound *model.Inbound) error {
	// XrayService handles starting the core process
	return nil
}

func (a *XrayAdapter) Stop(inbound *model.Inbound) error {
	// XrayService handles stopping the core process
	return nil
}

func (a *XrayAdapter) IsRunning(inbound *model.Inbound) bool {
	// XrayService handles status
	return true
}

func (a *XrayAdapter) AddClient(inbound *model.Inbound, client *model.Client) error {
	return nil
}

func (a *XrayAdapter) RemoveClient(inbound *model.Inbound, client *model.Client) error {
	return nil
}

func (a *XrayAdapter) GenerateKeypair() (KeyPair, error) {
	var priv [32]byte
	if _, err := rand.Read(priv[:]); err != nil {
		return KeyPair{}, err
	}

	// Clamp the private key for X25519
	priv[0] &= 248
	priv[31] &= 127
	priv[31] |= 64

	var pub [32]byte
	curve25519.ScalarBaseMult(&pub, &priv)

	return KeyPair{
		PrivateKey: base64.RawURLEncoding.EncodeToString(priv[:]),
		PublicKey:  base64.RawURLEncoding.EncodeToString(pub[:]),
	}, nil
}

func (a *XrayAdapter) GenerateServerConfig(inbound *model.Inbound) (string, error) {
	return "", nil
}

func (a *XrayAdapter) GenerateClientConfig(inbound *model.Inbound, client *model.Client) (string, error) {
	return "", nil
}

func (a *XrayAdapter) GetTraffic(inbound *model.Inbound) (map[string]Traffic, error) {
	return nil, nil
}

func (a *XrayAdapter) CheckPrerequisites() error {
	return nil
}

func init() {
	// Register for common Xray protocols to enable Keygen and node management
	protocols := []model.Protocol{
		model.VLESS,
		model.VMESS,
		model.Trojan,
		model.Shadowsocks,
	}

	for _, p := range protocols {
		Register(p, NewXrayAdapter(p))
	}
}

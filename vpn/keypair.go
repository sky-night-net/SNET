package vpn

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sky-night-net/snet/util/sys"
	"github.com/sky-night-net/snet/vpn/adapters"
)

func GenerateWGKeypair() (adapters.KeyPair, error) {
	// Private Key
	priv, stderr, err := sys.RunCommand("wg", "genkey")
	if err != nil {
		return adapters.KeyPair{}, fmt.Errorf("wg genkey failed: %v: %s", err, stderr)
	}
	priv = strings.TrimSpace(priv)

	// Public Key
	cmd := exec.Command("wg", "pubkey")
	cmd.Stdin = strings.NewReader(priv)
	var pubOut, pubErr bytes.Buffer
	cmd.Stdout = &pubOut
	cmd.Stderr = &pubErr
	if err := cmd.Run(); err != nil {
		return adapters.KeyPair{}, fmt.Errorf("wg pubkey failed: %v: %s", err, pubErr.String())
	}
	pub := strings.TrimSpace(pubOut.String())

	// Preshared Key
	psk, pskErr, err := sys.RunCommand("wg", "genpsk")
	if err != nil {
		return adapters.KeyPair{}, fmt.Errorf("wg genpsk failed: %v: %s", err, pskErr)
	}
	psk = strings.TrimSpace(psk)

	return adapters.KeyPair{
		PrivateKey:   priv,
		PublicKey:    pub,
		PresharedKey: psk,
	}, nil
}

package adapters

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/util/sys"
)

type OpenVPNXORAdapter struct {
	protocol model.Protocol
}

func NewOpenVPNXORAdapter() *OpenVPNXORAdapter {
	return &OpenVPNXORAdapter{
		protocol: model.OpenVPNXOR,
	}
}

func (a *OpenVPNXORAdapter) Protocol() model.Protocol {
	return a.protocol
}

const (
	OpenVPNConfigDir = "/etc/openvpn"
	EasyRSADir       = "/etc/openvpn/easy-rsa"
	OpenVPNStatusLog = "/var/log/openvpn/status.log"
)

func (a *OpenVPNXORAdapter) Start(inbound *model.Inbound) error {
	iface := fmt.Sprintf("tun_snet%d", inbound.Id)
	pidPath := fmt.Sprintf("/var/run/snet_openvpn_%d.pid", inbound.Id)

	// Proactive cleanup
	a.Stop(inbound)
	sys.Execute(fmt.Sprintf("ip link delete %s 2>/dev/null || true", iface))

	// Generate server config
	conf, err := a.GenerateServerConfig(inbound)
	if err != nil {
		return err
	}
	confPath := fmt.Sprintf("%s/server_%d.conf", OpenVPNConfigDir, inbound.Id)
	os.MkdirAll(OpenVPNConfigDir, 0755)
	os.MkdirAll("/var/log/openvpn", 0755)
	os.WriteFile(confPath, []byte(conf), 0600)

	// Start native openvpn-xor daemon
	binPath := os.Getenv("OPENVPN_XOR_PATH")
	if binPath == "" {
		binPath = "/usr/local/snet/bin/openvpn-xor"
	}
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		binPath = "openvpn"
	}

	_, err = sys.Execute(fmt.Sprintf("%s --config %s --dev %s --dev-type tun --daemon --writepid %s", binPath, confPath, iface, pidPath))
	return err
}

func (a *OpenVPNXORAdapter) Stop(inbound *model.Inbound) error {
	iface := fmt.Sprintf("tun_snet%d", inbound.Id)
	pidPath := fmt.Sprintf("/var/run/snet_openvpn_%d.pid", inbound.Id)

	// Kill process if pid file exists
	sys.Execute(fmt.Sprintf("if [ -f %s ]; then kill -9 \"$(cat %s)\" 2>/dev/null && rm -f %s; fi", pidPath, pidPath, pidPath))
	// Cleanup interface
	sys.Execute(fmt.Sprintf("ip link delete %s 2>/dev/null || true", iface))

	return nil
}

func (a *OpenVPNXORAdapter) IsRunning(inbound *model.Inbound) bool {
	pidPath := fmt.Sprintf("/var/run/snet_openvpn_%d.pid", inbound.Id)
	res, err := sys.Execute(fmt.Sprintf("if [ -f %s ]; then ps -p \"$(cat %s)\" > /dev/null && echo 'true'; fi", pidPath, pidPath))
	return err == nil && strings.TrimSpace(res) == "true"
}

func (a *OpenVPNXORAdapter) AddClient(inbound *model.Inbound, client *model.Client) error {
	const easyrsa = "/usr/share/easy-rsa/easyrsa"
	// Check if cert already exists
	certPath := filepath.Join(EasyRSADir, "pki/issued", client.Email+".crt")
	if _, err := os.Stat(certPath); err == nil {
		return nil
	}

	cmd := fmt.Sprintf("bash -c 'cd %s && %s --batch build-client-full %s nopass'", EasyRSADir, easyrsa, client.Email)
	_, err := sys.Execute(cmd)
	return err
}

func (a *OpenVPNXORAdapter) RemoveClient(inbound *model.Inbound, client *model.Client) error {
	const easyrsa = "/usr/share/easy-rsa/easyrsa"
	cmd := fmt.Sprintf("bash -c 'cd %s && %s --batch revoke %s && %s gen-crl'", EasyRSADir, easyrsa, client.Email, easyrsa)
	_, err := sys.Execute(cmd)
	return err
}

func (a *OpenVPNXORAdapter) GenerateKeypair() (KeyPair, error) {
	if _, err := os.Stat(filepath.Join(EasyRSADir, "pki")); os.IsNotExist(err) {
		os.RemoveAll(EasyRSADir)
		os.MkdirAll(filepath.Dir(EasyRSADir), 0755)

		const easyrsa = "/usr/share/easy-rsa/easyrsa"
		sys.Execute(fmt.Sprintf("make-cadir %s", EasyRSADir))
		// Fix for easy-rsa 3 compatibility
		sys.Execute(fmt.Sprintf("bash -c 'cd %s && %s init-pki && EASYRSA_BATCH=1 %s build-ca nopass && EASYRSA_BATCH=1 %s build-server-full server nopass && %s gen-dh'", EasyRSADir, easyrsa, easyrsa, easyrsa, easyrsa))
		
		binPath := os.Getenv("OPENVPN_XOR_PATH")
		if binPath == "" {
			binPath = "/usr/local/snet/bin/openvpn-xor"
		}
		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			binPath = "openvpn"
		}
		sys.Execute(fmt.Sprintf("%s --genkey --secret %s/pki/ta.key", binPath, EasyRSADir))
	}

	return KeyPair{
		PublicKey: "ca_initialized",
	}, nil
}

func (a *OpenVPNXORAdapter) GenerateServerConfig(inbound *model.Inbound) (string, error) {
	var settings map[string]interface{}
	json.Unmarshal([]byte(inbound.Settings), &settings)

	var obfs map[string]interface{}
	json.Unmarshal([]byte(inbound.StreamSettings), &obfs)

	port := inbound.Port
	scramblePassword := ""
	if sp, ok := obfs["scramble_password"].(string); ok {
		scramblePassword = sp
	}

	proto := "udp"
	if p, ok := settings["proto"].(string); ok {
		proto = p
	}

	cipher := "AES-256-GCM"
	if c, ok := settings["cipher"].(string); ok {
		cipher = c
	}

	addressFull := "10.8.0.0/24"
	if addr, ok := settings["address"].(string); ok {
		addressFull = addr
	}

	parts := strings.Split(addressFull, "/")
	network := parts[0]
	mask := "255.255.255.0"

	scrambleLine := ""
	if scramblePassword != "" {
		scrambleLine = fmt.Sprintf("scramble obfuscate %s", scramblePassword)
	}

	statusLog := fmt.Sprintf("/var/log/openvpn/status_%d.log", inbound.Id)

	config := fmt.Sprintf(`port %d
proto %s4
dev tun
ca %s/pki/ca.crt
cert %s/pki/issued/server.crt
key %s/pki/private/server.key
dh %s/pki/dh.pem
tls-auth %s/pki/ta.key 0
server %s %s
ifconfig-pool-persist ipp_%d.txt
push "redirect-gateway def1 bypass-dhcp"
push "dhcp-option DNS 1.1.1.1"
push "dhcp-option DNS 8.8.8.8"
keepalive 10 120
topology subnet
data-ciphers %s:AES-256-GCM:AES-128-GCM
cipher %s
auth SHA256
explicit-exit-notify 1
tls-version-min 1.2
persist-key
persist-tun
status %s 10
verb 3
mssfix 1350
%s
`, port, proto, EasyRSADir, EasyRSADir, EasyRSADir, EasyRSADir, EasyRSADir, network, mask, inbound.Id, cipher, cipher, statusLog, scrambleLine)

	return config, nil
}

func (a *OpenVPNXORAdapter) GenerateClientConfig(inbound *model.Inbound, client *model.Client, host string) (string, error) {
	var settings map[string]interface{}
	json.Unmarshal([]byte(inbound.Settings), &settings)

	var obfs map[string]interface{}
	json.Unmarshal([]byte(inbound.StreamSettings), &obfs)

	port := inbound.Port
	serverIP := host

	scramblePassword := ""
	if sp, ok := obfs["scramble_password"].(string); ok {
		scramblePassword = sp
	}

	proto := "udp"
	if p, ok := settings["proto"].(string); ok {
		proto = p
	}

	cipher := "AES-256-GCM"
	if c, ok := settings["cipher"].(string); ok {
		cipher = c
	}

	readCert := func(path string) string {
		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Failed to read cert %s: %v", path, err)
			return ""
		}
		s := string(content)
		if strings.Contains(s, "-----BEGIN CERTIFICATE-----") {
			return s[strings.Index(s, "-----BEGIN CERTIFICATE-----"):]
		}
		if strings.Contains(s, "-----BEGIN PRIVATE KEY-----") {
			return s[strings.Index(s, "-----BEGIN PRIVATE KEY-----"):]
		}
		if strings.Contains(s, "-----BEGIN OpenVPN Static key V1-----") {
			return s[strings.Index(s, "-----BEGIN OpenVPN Static key V1-----"):]
		}
		return strings.TrimSpace(s)
	}

	scrambleLine := ""
	if scramblePassword != "" {
		scrambleLine = fmt.Sprintf("scramble obfuscate %s", scramblePassword)
	}

	config := fmt.Sprintf(`client
dev tun
proto %s
remote %s %d
resolv-retry infinite
nobind
persist-key
persist-tun
remote-cert-tls server
data-ciphers %s:AES-256-GCM:AES-128-GCM
cipher %s
auth SHA256
auth-nocache
tls-client
tls-version-min 1.2
key-direction 1
explicit-exit-notify
ignore-unknown-option block-outside-dns
setenv opt block-outside-dns
verb 3
mssfix 1350
%s
<ca>
%s
</ca>
<cert>
%s
</cert>
<key>
%s
</key>
<tls-auth>
%s
</tls-auth>
`, proto, serverIP, port, cipher, cipher, scrambleLine,
		readCert(filepath.Join(EasyRSADir, "pki/ca.crt")),
		readCert(filepath.Join(EasyRSADir, "pki/issued", client.Email+".crt")),
		readCert(filepath.Join(EasyRSADir, "pki/private", client.Email+".key")),
		readCert(filepath.Join(EasyRSADir, "pki/ta.key")))

	return config, nil
}

func (a *OpenVPNXORAdapter) GetTraffic(inbound *model.Inbound) (map[string]Traffic, error) {
	statusLog := fmt.Sprintf("/var/log/openvpn/status_%d.log", inbound.Id)
	content, err := os.ReadFile(statusLog)
	if err != nil {
		return nil, err
	}

	trafficMap := make(map[string]Traffic)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, ",") && strings.Contains(line, ".") {
			parts := strings.Split(line, ",")
			if len(parts) >= 5 && parts[0] != "Common Name" && !strings.Contains(parts[0], "UNDEF") {
				// parts[0] is Common Name (client email)
				// parts[2] is Bytes Received
				// parts[3] is Bytes Sent
				var rx, tx int64
				fmt.Sscanf(parts[2], "%d", &rx)
				fmt.Sscanf(parts[3], "%d", &tx)
				trafficMap[parts[0]] = Traffic{
					Up:   tx,
					Down: rx,
				}
			}
		}
	}
	return trafficMap, nil
}

func (a *OpenVPNXORAdapter) CheckPrerequisites() error {
	_, err := sys.Execute("openvpn --version")
	if err != nil {
		return fmt.Errorf("openvpn is not installed: %v", err)
	}
	_, err = sys.Execute("easyrsa --version")
	if err != nil {
		// Try full path if not in PATH
		_, err = sys.Execute("/usr/share/easy-rsa/easyrsa --version")
		if err != nil {
			return fmt.Errorf("easy-rsa is not installed: %v", err)
		}
	}
	return nil
}

func init() {
	Register(model.OpenVPNXOR, NewOpenVPNXORAdapter())
}

package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/logger"
	"github.com/sky-night-net/snet/util/sys"
)

type OpenVPNXORAdapter struct {
	DockerBaseAdapter
	protocol model.Protocol
}

func NewOpenVPNXORAdapter() *OpenVPNXORAdapter {
	base, _ := NewDockerBaseAdapter()
	return &OpenVPNXORAdapter{
		DockerBaseAdapter: *base,
		protocol:          model.OpenVPNXOR,
	}
}

func (a *OpenVPNXORAdapter) Protocol() model.Protocol {
	return a.protocol
}

const (
	OpenVPNConfigDir = "/etc/openvpn"
	EasyRSADir      = "/etc/openvpn/easy-rsa"
	OpenVPNStatusLog = "/var/log/openvpn/status.log"
)

func (a *OpenVPNXORAdapter) Start(inbound *model.Inbound) error {
	ctx := context.Background()
	iface := fmt.Sprintf("tun_snet%d", inbound.Id)
	containerName := fmt.Sprintf("snet_openvpn_xor_%d", inbound.Id)
	
	// Generate server config
	conf, _ := a.GenerateServerConfig(inbound)
	confPath := fmt.Sprintf("%s/server_%d.conf", OpenVPNConfigDir, inbound.Id)
	os.MkdirAll(OpenVPNConfigDir, 0755)
	os.WriteFile(confPath, []byte(conf), 0600)

	// Proactive cleanup
	a.Stop(inbound)

	// Start container using Docker SDK
	resp, err := a.cli.ContainerCreate(ctx, &container.Config{
		Image: "snet-local/openvpn-xor:latest",
		Cmd:   []string{"openvpn", "--config", confPath, "--dev", iface},
	}, &container.HostConfig{
		NetworkMode: "host",
		Privileged:  true,
		CapAdd:      []string{"NET_ADMIN", "SYS_PTRACE"},
		Binds:       []string{
			fmt.Sprintf("%s:/etc/openvpn", OpenVPNConfigDir),
			"/var/log/openvpn:/var/log/openvpn",
		},
		Resources: container.Resources{
			Devices: []container.DeviceMapping{
				{PathOnHost: "/dev/net/tun", PathInContainer: "/dev/net/tun", CgroupPermissions: "rwm"},
			},
		},
	}, nil, nil, containerName)

	if err != nil {
		return err
	}

	return a.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
}

func (a *OpenVPNXORAdapter) Stop(inbound *model.Inbound) error {
	ctx := context.Background()
	containerName := fmt.Sprintf("snet_openvpn_xor_%d", inbound.Id)
	iface := fmt.Sprintf("tun_snet%d", inbound.Id)
	
	a.StopAndRemove(ctx, containerName)
	sys.Execute(fmt.Sprintf("ip link delete %s", iface))
	
	return nil
}

func (a *OpenVPNXORAdapter) IsRunning(inbound *model.Inbound) bool {
	ctx := context.Background()
	containerName := fmt.Sprintf("snet_openvpn_xor_%d", inbound.Id)
	running, _ := a.IsContainerRunning(ctx, containerName)
	return running
}

func (a *OpenVPNXORAdapter) AddClient(inbound *model.Inbound, client *model.Client) error {
	// For OpenVPN, "AddClient" means generating certificates via Easy-RSA
	cmd := fmt.Sprintf("bash -c 'cd %s && ./easyrsa --batch build-client-full %s nopass'", EasyRSADir, client.Email)
	_, err := sys.Execute(cmd)
	return err
}

func (a *OpenVPNXORAdapter) RemoveClient(inbound *model.Inbound, client *model.Client) error {
	cmd := fmt.Sprintf("bash -c 'cd %s && ./easyrsa --batch revoke %s && ./easyrsa gen-crl'", EasyRSADir, client.Email)
	_, err := sys.Execute(cmd)
	return err
}

func (a *OpenVPNXORAdapter) GenerateKeypair() (KeyPair, error) {
	// For OpenVPN, this initializes the CA/PKI
	if _, err := os.Stat(filepath.Join(EasyRSADir, "pki")); os.IsNotExist(err) {
		os.RemoveAll(EasyRSADir)
		os.MkdirAll(filepath.Dir(EasyRSADir), 0755)
		
		sys.Execute(fmt.Sprintf("make-cadir %s", EasyRSADir))
		sys.Execute(fmt.Sprintf("bash -c 'cd %s && ./easyrsa init-pki && EASYRSA_BATCH=1 ./easyrsa build-ca nopass && EASYRSA_BATCH=1 ./easyrsa build-server-full server nopass && ./easyrsa gen-dh'", EasyRSADir))
		sys.Execute(fmt.Sprintf("openvpn --genkey --secret %s/pki/ta.key", EasyRSADir))
	}
	
	return KeyPair{
		PublicKey: "ca_initialized", // Dummy for interface
	}, nil
}

func (a *OpenVPNXORAdapter) GenerateServerConfig(inbound *model.Inbound) (string, error) {
	var settings map[string]interface{}
	json.Unmarshal([]byte(inbound.Settings), &settings)
	
	var obfs map[string]interface{}
	json.Unmarshal([]byte(inbound.Obfuscation), &obfs)
	
	port := inbound.Port
	scramblePassword := ""
	if sp, ok := obfs["scramble_password"].(string); ok {
		scramblePassword = sp
	}
	
	proto := "udp"
	if p, ok := settings["proto"].(string); ok {
		proto = p
	}
	
	cipher := "AES-256-CBC"
	if c, ok := settings["cipher"].(string); ok {
		cipher = c
	}

	addressFull := "10.9.0.0/24"
	if a, ok := settings["address"].(string); ok {
		addressFull = a
	}
	
	// Simplify CIDR to network/mask
	parts := strings.Split(addressFull, "/")
	network := parts[0]
	mask := "255.255.255.0" // Default for /24
	
	scrambleLine := ""
	if scramblePassword != "" {
		scrambleLine = fmt.Sprintf("scramble obfuscate %s", scramblePassword)
	}

	config := fmt.Sprintf(`port %d
proto %s4
dev tun
ca %s/pki/ca.crt
cert %s/pki/issued/server.crt
key %s/pki/private/server.key
dh %s/pki/dh.pem
tls-auth %s/pki/ta.key 0
server %s %s
ifconfig-pool-persist ipp.txt
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
`, port, proto, EasyRSADir, EasyRSADir, EasyRSADir, EasyRSADir, EasyRSADir, network, mask, cipher, cipher, OpenVPNStatusLog, scrambleLine)

	return config, nil
}

func (a *OpenVPNXORAdapter) GenerateClientConfig(inbound *model.Inbound, client *model.Client) (string, error) {
	var settings map[string]interface{}
	json.Unmarshal([]byte(inbound.Settings), &settings)
	
	var obfs map[string]interface{}
	json.Unmarshal([]byte(inbound.Obfuscation), &obfs)
	
	port := inbound.Port
	serverIP := ""
	if sip, ok := settings["server_ip"].(string); ok {
		serverIP = sip
	}
	
	scramblePassword := ""
	if sp, ok := obfs["scramble_password"].(string); ok {
		scramblePassword = sp
	}
	
	proto := "udp"
	if p, ok := settings["proto"].(string); ok {
		proto = p
	}
	
	cipher := "AES-256-CBC"
	if c, ok := settings["cipher"].(string); ok {
		cipher = c
	}

	readCert := func(path string) string {
		content, err := os.ReadFile(path)
		if err != nil {
			return ""
		}
		s := string(content)
		if strings.Contains(s, "-----BEGIN CERTIFICATE-----") {
			return s[strings.Index(s, "-----BEGIN CERTIFICATE-----"):]
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
	content, err := os.ReadFile(OpenVPNStatusLog)
	if err != nil {
		return nil, err
	}
	
	trafficMap := make(map[string]Traffic)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, ",") && strings.Contains(line, ".") {
			parts := strings.Split(line, ",")
			if len(parts) >= 5 && parts[0] != "Common Name" {
				// parts[0] is Common Name (client email)
				// parts[2] is Bytes Received
				// parts[3] is Bytes Sent
				var rx, tx int64
				fmt.Sscanf(parts[2], "%d", &rx)
				fmt.Sscanf(parts[3], "%d", &tx)
				trafficMap[parts[0]] = Traffic{
					Up: tx,
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

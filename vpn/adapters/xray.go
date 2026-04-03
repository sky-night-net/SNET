package adapters

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

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

func (a *XrayAdapter) GenerateClientConfig(inbound *model.Inbound, client *model.Client, host string) (string, error) {
	port := inbound.Port
	protocol := strings.ToLower(string(inbound.Protocol))
	remark := inbound.Remark
	if remark == "" {
		remark = fmt.Sprintf("%s-%d", protocol, port)
	}

	var stream map[string]interface{}
	json.Unmarshal([]byte(inbound.StreamSettings), &stream)

	var settings map[string]interface{}
	json.Unmarshal([]byte(inbound.Settings), &settings)

	// Determine UID/Password
	uid := client.ID
	if uid == "" {
		uid = client.Password
	}

	if protocol == "vmess" {
		// VMess uses a JSON config base64-encoded
		vmessConfig := map[string]interface{}{
			"v":    "2",
			"ps":   remark,
			"add":  host,
			"port": port,
			"id":   uid,
			"aid":  "0",
			"scy":  "auto",
			"net":  stream["network"],
			"type": "none",
			"host": "",
			"path": "",
			"tls":  "",
			"sni":  "",
			"alpn": "",
			"fp":   "",
		}
		// Extract network specific settings
		if nw, ok := stream["network"].(string); ok {
			switch nw {
			case "ws":
				if ws, ok := stream["wsSettings"].(map[string]interface{}); ok {
					vmessConfig["path"] = ws["path"]
					if headers, ok := ws["headers"].(map[string]interface{}); ok {
						vmessConfig["host"] = headers["Host"]
					}
				}
			case "grpc":
				if grpc, ok := stream["grpcSettings"].(map[string]interface{}); ok {
					vmessConfig["path"] = grpc["serviceName"]
				}
			}
		}
		if sec, ok := stream["security"].(string); ok && sec != "none" {
			vmessConfig["tls"] = sec
		}

		jb, _ := json.Marshal(vmessConfig)
		return "vmess://" + base64.StdEncoding.EncodeToString(jb), nil
	}

	// ── Shadowsocks — ss://BASE64(method:password)@host:port#remark ─────────
	if protocol == "shadowsocks" {
		method := "chacha20-ietf-poly1305"
		if m, ok := settings["method"].(string); ok && m != "" {
			method = m
		}
		password := uid // uid is client.Password for SS
		if password == "" {
			password = client.Password
		}
		creds := base64.StdEncoding.EncodeToString([]byte(method + ":" + password))
		return fmt.Sprintf("ss://%s@%s:%d#%s", creds, host, port, url.QueryEscape(remark)), nil
	}

	// ── Trojan — trojan://PASSWORD@host:port?security=tls#remark ─────────────
	if protocol == "trojan" {
		password := uid
		if password == "" {
			password = client.Password
		}
		security := "tls"
		if sec, ok := stream["security"].(string); ok && sec != "" && sec != "none" {
			security = sec
		}
		q := url.Values{}
		q.Set("security", security)
		q.Set("type", func() string {
			if nw, ok := stream["network"].(string); ok && nw != "" {
				return nw
			}
			return "tcp"
		}())
		if tls, ok := stream["tlsSettings"].(map[string]interface{}); ok {
			if sni, ok := tls["serverName"].(string); ok && sni != "" {
				q.Set("sni", sni)
			}
		}
		return fmt.Sprintf("trojan://%s@%s:%d?%s#%s",
			url.QueryEscape(password), host, port, q.Encode(), url.QueryEscape(remark)), nil
	}

	// ── VLESS (and other Xray URI protocols) ──────────────────────────────────
	u := &url.URL{
		Scheme: protocol,
		User:   url.User(uid),
		Host:   fmt.Sprintf("%s:%d", host, port),
	}
	q := u.Query()

	// Network
	if nw, ok := stream["network"].(string); ok && nw != "tcp" {
		q.Set("type", nw)
		switch nw {
		case "ws":
			if ws, ok := stream["wsSettings"].(map[string]interface{}); ok {
				q.Set("path", fmt.Sprintf("%v", ws["path"]))
				if headers, ok := ws["headers"].(map[string]interface{}); ok {
					q.Set("host", fmt.Sprintf("%v", headers["Host"]))
				}
			}
		case "grpc":
			if grpc, ok := stream["grpcSettings"].(map[string]interface{}); ok {
				q.Set("serviceName", fmt.Sprintf("%v", grpc["serviceName"]))
			}
		}
	}

	// Security
	security := "none"
	if sec, ok := stream["security"].(string); ok {
		security = sec
	}
	q.Set("security", security)

	if security == "tls" {
		if tls, ok := stream["tlsSettings"].(map[string]interface{}); ok {
			q.Set("sni", fmt.Sprintf("%v", tls["serverName"]))
			if fp, ok := tls["fingerprint"].(string); ok && fp != "" {
				q.Set("fp", fp)
			}
		}
	} else if security == "reality" {
		if reality, ok := stream["realitySettings"].(map[string]interface{}); ok {
			if sNames, ok := reality["serverNames"].([]interface{}); ok && len(sNames) > 0 {
				q.Set("sni", fmt.Sprintf("%v", sNames[0]))
			}
			q.Set("pbk", fmt.Sprintf("%v", reality["publicKey"]))
			if sid, ok := reality["shortIds"].([]interface{}); ok && len(sid) > 0 {
				q.Set("sid", fmt.Sprintf("%v", sid[0]))
			}
			q.Set("fp", "chrome")
		}
	}

	// Flow (for VLESS)
	if protocol == "vless" {
		if clients, ok := settings["clients"].([]interface{}); ok {
			for _, c := range clients {
				if cm, ok := c.(map[string]interface{}); ok && cm["id"] == uid {
					if flow, ok := cm["flow"].(string); ok && flow != "" {
						q.Set("flow", flow)
					}
				}
			}
		}
	}

	u.RawQuery = q.Encode()
	u.Fragment = remark
	return u.String(), nil
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

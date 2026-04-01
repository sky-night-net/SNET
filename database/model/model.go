// Package model defines the database models for SNET.
package model

import (
	"fmt"
	"github.com/sky-night-net/snet/util/json_util"
)

type Protocol string

const (
	VMESS       Protocol = "vmess"
	VLESS       Protocol = "vless"
	Trojan      Protocol = "trojan"
	Shadowsocks Protocol = "shadowsocks"
	WireGuard   Protocol = "wireguard"
	AmneziaWGv1 Protocol = "amneziawg_v1"
	AmneziaWGv2 Protocol = "amneziawg_v2"
	OpenVPNXOR  Protocol = "openvpn_xor"
)

type User struct {
	Id       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	Username string `json:"username" gorm:"unique"`
	Password string `json:"password"`
}

type Inbound struct {
	Id                   int      `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Up                   int64    `json:"up" form:"up"`
	Down                 int64    `json:"down" form:"down"`
	Total                int64    `json:"total" form:"total"`
	Remark               string   `json:"remark" form:"remark"`
	Enable               bool     `json:"enable" form:"enable" gorm:"index"`
	ExpiryTime           int64    `json:"expiryTime" form:"expiryTime"`
	Listen               string   `json:"listen" form:"listen"`
	Port                 int      `json:"port" form:"port"`
	Protocol             Protocol `json:"protocol" form:"protocol"`
	Settings             string   `json:"settings" form:"settings"`             // JSON for Xray or VPN server keys/config
	StreamSettings       string   `json:"streamSettings" form:"streamSettings"` // Only for Xray
	Tag                  string   `json:"tag" form:"tag" gorm:"unique"`
	Sniffing             string   `json:"sniffing" form:"sniffing"`
	
	// Relation to clients
	Clients              []Client `gorm:"foreignKey:InboundId;references:Id;constraint:OnDelete:CASCADE" json:"clients"`
}

type Client struct {
	Id           int    `json:"id" gorm:"primaryKey;autoIncrement"`
	InboundId    int    `json:"inboundId" gorm:"index"`
	Email        string `json:"email" gorm:"index"`
	Enable       bool   `json:"enable" gorm:"default:true"`
	Up           int64  `json:"up"`
	Down         int64  `json:"down"`
	Total        int64  `json:"total"`
	ExpiryTime   int64  `json:"expiryTime"`
	LastOnline   int64  `json:"lastOnline"`
	
	// VPN specific (AWG / WG / OpenVPN)
	PublicKey    string `json:"publicKey"`
	PrivateKey   string `json:"privateKey"`
	PresharedKey string `json:"presharedKey"`
	AllowedIPs   string `json:"allowedIPs"`
	
	// Xray specific
	UUID         string `json:"uuid"`
	Password     string `json:"password"`
	Flow         string `json:"flow"`
	SubID        string `json:"subId" gorm:"uniqueIndex"`
	
	CreatedAt    int64  `json:"created_at"`
}

type Setting struct {
	Id    int    `json:"id" form:"id" gorm:"primaryKey;autoIncrement"`
	Key   string `json:"key" form:"key" gorm:"unique"`
	Value string `json:"value" form:"value"`
}

// XrayInboundConfig temporary struct for Xray compatibility
type XrayInboundConfig struct {
	Listen         json_util.RawMessage `json:"listen"`
	Port           int                  `json:"port"`
	Protocol       string               `json:"protocol"`
	Settings       json_util.RawMessage `json:"settings"`
	StreamSettings json_util.RawMessage `json:"streamSettings"`
	Tag            string               `json:"tag"`
	Sniffing       json_util.RawMessage `json:"sniffing"`
}

func (i *Inbound) GenXrayInboundConfig() *XrayInboundConfig {
	listen := i.Listen
	if listen == "" {
		listen = "0.0.0.0"
	}
	listen = fmt.Sprintf("\"%v\"", listen)
	return &XrayInboundConfig{
		Listen:         json_util.RawMessage(listen),
		Port:           i.Port,
		Protocol:       string(i.Protocol),
		Settings:       json_util.RawMessage(i.Settings),
		StreamSettings: json_util.RawMessage(i.StreamSettings),
		Tag:            i.Tag,
		Sniffing:       json_util.RawMessage(i.Sniffing),
	}
}

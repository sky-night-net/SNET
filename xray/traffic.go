package xray

// Traffic represents network traffic statistics for Xray connections.
type Traffic struct {
	IsInbound  bool   `json:"isInbound"`
	IsOutbound bool   `json:"isOutbound"`
	Tag        string `json:"tag"`
	Up         int64  `json:"up"`
	Down       int64  `json:"down"`
}

type ClientTraffic struct {
	Id        int    `json:"id" gorm:"primaryKey;autoIncrement"`
	InboundId int    `json:"inboundId" gorm:"index"`
	Email     string `json:"email" gorm:"index"`
	Up        int64  `json:"up"`
	Down      int64  `json:"down"`
	ExpiryTime int64 `json:"expiryTime"`
	Enable    bool   `json:"enable"`
}

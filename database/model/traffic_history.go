package model

import (
	"time"
)

type TrafficHistory struct {
	Id         int   `json:"id" gorm:"primaryKey;autoIncrement"`
	Timestamp  int64 `json:"timestamp" gorm:"index"`
	InboundId  int   `json:"inboundId" gorm:"index"`
	Up         int64 `json:"up"`
	Down       int64 `json:"down"`
}

func NewTrafficHistory(ibId int, up, down int64) *TrafficHistory {
	return &TrafficHistory{
		Timestamp: time.Now().Truncate(time.Hour).Unix(),
		InboundId: ibId,
		Up:        up,
		Down:      down,
	}
}

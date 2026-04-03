package service

import (
	"sync"
	"time"

	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/logger"
	"github.com/sky-night-net/snet/xray"
)

type TrafficService struct {
	ticker *time.Ticker
	done   chan bool
	lock   sync.Mutex
}

var (
	trafficInstance *TrafficService
	trafficOnce     sync.Once
)

func GetTrafficService() *TrafficService {
	trafficOnce.Do(func() {
		trafficInstance = &TrafficService{
			done: make(chan bool),
		}
	})
	return trafficInstance
}

func (s *TrafficService) Start() {
	s.ticker = time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-s.done:
				return
			case <-s.ticker.C:
				s.sync()
			}
		}
	}()
	logger.Info("Traffic synchronization service started")
}

func (s *TrafficService) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.done <- true
}

func (s *TrafficService) sync() {
	s.lock.Lock()
	defer s.lock.Unlock()

	db := database.GetDB()
	var inbounds []model.Inbound
	if err := db.Find(&inbounds).Error; err != nil {
		return
	}

	xraySvc := GetXrayService()
	vpnSvc := GetVpnService()
	manager := vpnSvc.GetManager()

	// 1. Sync Xray Traffic (Incremental)
	if xraySvc.IsRunning() && xraySvc.GetAPI() != nil {
		tags, clients, err := xraySvc.GetAPI().GetTraffic(true) // Reset=true gets delta
		if err == nil {
			for _, t := range tags {
				db.Model(&model.Inbound{}).Where("tag = ?", t.Tag).Updates(map[string]interface{}{
					"up":   database.GetDB().Raw("up + ?", t.Up),
					"down": database.GetDB().Raw("down + ?", t.Down),
				})
			}
			for _, c := range clients {
				db.Model(&xray.ClientTraffic{}).Where("email = ?", c.Email).Updates(map[string]interface{}{
					"up":   database.GetDB().Raw("up + ?", c.Up),
					"down": database.GetDB().Raw("down + ?", c.Down),
				})
			}
		}
	}

	// 2. Sync VPN Traffic (Cumulative from status/logs)
	for _, ib := range inbounds {
		if ib.Protocol == "amneziawg" || ib.Protocol == "amneziawg-v1" || ib.Protocol == "openvpn" {
			trafficMap, err := manager.GetTraffic(&ib)
			if err != nil {
				continue
			}

			for email, t := range trafficMap {
				var stat xray.ClientTraffic
				if err := db.Where("inbound_id = ? AND email = ?", ib.Id, email).First(&stat).Error; err == nil {
					if t.Up > stat.Up || t.Down > stat.Down {
						db.Model(&stat).Updates(map[string]interface{}{
							"up":   t.Up,
							"down": t.Down,
						})
					}
				} else {
					// Create new entry
					db.Create(&xray.ClientTraffic{
						InboundId: ib.Id,
						Email:     email,
						Up:        t.Up,
						Down:      t.Down,
						Enable:    true,
					})
				}
			}
		}
	}
}

package job

import (
	"time"

	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/database/model"
	"github.com/sky-night-net/snet/logger"
	"github.com/sky-night-net/snet/web/service"
	"gorm.io/gorm"
)

type TrafficJob struct {
	inboundService *service.InboundService
	xrayService    *service.XrayService
	vpnService     *service.VPNService
}

func NewTrafficJob(ib *service.InboundService, xray *service.XrayService, vpn *service.VPNService) *TrafficJob {
	return &TrafficJob{
		inboundService: ib,
		xrayService:    xray,
		vpnService:     vpn,
	}
}

func (j *TrafficJob) Run() {
	db := database.GetDB()
	inbounds, err := j.inboundService.GetAllInbounds()
	if err != nil {
		logger.Error("TrafficJob: failed to get inbounds:", err)
		return
	}

	// 1. Process Xray traffic
	xrayTraffic, clientTraffic, err := j.xrayService.GetXrayTraffic()
	if err == nil {
		for _, t := range xrayTraffic {
			db.Model(&model.Inbound{}).Where("tag = ?", t.Tag).Updates(map[string]interface{}{
				"up":   gorm.Expr("up + ?", t.Up),
				"down": gorm.Expr("down + ?", t.Down),
			})
		}
		for _, ct := range clientTraffic {
			db.Model(&model.Client{}).Where("email = ?", ct.Email).Updates(map[string]interface{}{
				"up":   gorm.Expr("up + ?", ct.Up),
				"down": gorm.Expr("down + ?", ct.Down),
			})
		}
	}

	// 2. Process VPN traffic
	for _, ib := range inbounds {
		if j.inboundService.IsVPN(ib.Protocol) && ib.Enable {
			stats, err := j.vpnService.GetTraffic(ib)
			if err != nil {
				continue
			}
			
			var totalUp, totalDown int64
			for email, t := range stats {
				totalUp += t.Up
				totalDown += t.Down
				
				// Update client
				db.Model(&model.Client{}).Where("inbound_id = ? AND email = ?", ib.Id, email).Updates(map[string]interface{}{
					"up":   gorm.Expr("up + ?", t.Up),
					"down": gorm.Expr("down + ?", t.Down),
				})
			}
			
			// Update inbound total
			db.Model(&model.Inbound{}).Where("id = ?", ib.Id).Updates(map[string]interface{}{
				"up":   gorm.Expr("up + ?", totalUp),
				"down": gorm.Expr("down + ?", totalDown),
			})

			// Record History Snapshot
			history := model.NewTrafficHistory(ib.Id, totalUp, totalDown)
			db.Save(history)
		}
	}
}

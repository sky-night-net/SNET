package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sky-night-net/snet/api"
	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/service"
)

func main() {
	log.Println("SNET 3.0 Native VPN Panel (BBR Optimized)")

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = os.Getenv("SNET_DB_PATH") // Fallback
	}
	if dbPath == "" {
		dbPath = "snet.db"
	}

	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Start background telemetry collection
	statsSvc := service.GetStatsService()
	statsSvc.Start()

	// Initial Xray start
	xraySvc := service.GetXrayService()
	_ = xraySvc.ApplyConfig()

	// Start VPN Protocol Manager (AmneziaWG, OpenVPN XOR)
	vpnSvc := service.GetVpnService()
	vpnSvc.Start()
	
	// Sync firewall rules with system
	go func() {
		if err := service.GetFirewallService().Sync(); err != nil {
			fmt.Printf("Firewall sync error: %v\n", err)
		} else {
			fmt.Println("Firewall rules synchronized successfully")
		}
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := api.NewServer()

	if err := server.Start(":" + port); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}

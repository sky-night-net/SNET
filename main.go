package main

import (
	"log"
	"os"

	"github.com/sky-night-net/snet/api"
	"github.com/sky-night-net/snet/database"
)

func main() {
	log.Println("SNET 3.0 Backend Starting...")

	dbPath := os.Getenv("SNET_DB_PATH")
	if dbPath == "" {
		dbPath = "snet.db"
	}

	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := api.NewServer()
	if err := server.Start(":" + port); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}

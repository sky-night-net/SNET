package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/op/go-logging"
	"github.com/sky-night-net/snet/config"
	"github.com/sky-night-net/snet/database"
	"github.com/sky-night-net/snet/logger"
	"github.com/sky-night-net/snet/sub"
	"github.com/sky-night-net/snet/web"
)

func runWebServer() {
	fmt.Printf("Starting %v %v\n", config.GetName(), config.GetVersion())

	switch config.GetLogLevel() {
	case config.Debug:
		logger.InitLogger(logging.DEBUG)
	case config.Info:
		logger.InitLogger(logging.INFO)
	case config.Notice:
		logger.InitLogger(logging.NOTICE)
	case config.Warning:
		logger.InitLogger(logging.WARNING)
	case config.Error:
		logger.InitLogger(logging.ERROR)
	default:
		log.Fatalf("Unknown log level: %v", config.GetLogLevel())
	}

	err := database.InitDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	server := web.NewServer()
	err = server.Start()
	if err != nil {
		log.Fatalf("Error starting web server: %v", err)
	}

	subServer := sub.NewServer(server.GetInboundService(), server.GetSettingService())
	err = subServer.Start()
	if err != nil {
		log.Fatalf("Error starting sub server: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	sig := <-sigCh
	logger.Infof("Received signal %v, shutting down...", sig)
	
	server.Stop()
	subServer.Stop()
	logger.Info("SNET v2 stopped.")
}

func main() {
	if len(os.Args) < 2 {
		runWebServer()
		return
	}

	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.Parse()

	if showVersion {
		fmt.Println(config.GetVersion())
		return
	}

	if os.Args[1] == "run" {
		runWebServer()
	} else if os.Args[1] == "migrate" {
		// Migration logic
		log.Fatal("Migration not implemented yet")
	} else {
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

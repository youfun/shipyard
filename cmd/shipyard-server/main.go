package main

import (
	"flag"
	"log"
	"youfun/shipyard/internal/api"
	"youfun/shipyard/internal/api/handlers"
	"youfun/shipyard/internal/config"
	"youfun/shipyard/internal/database"
)

// Version is set by build flags
var Version = "dev"

func main() {
	port := flag.String("port", "", "Server port")
	configPath := flag.String("config", "shipyard.toml", "Path to configuration file")
	flag.Parse()

	// Set version for API handlers
	handlers.SetVersion(Version)

	// Set global config path
	config.ConfigPath = *configPath

	// Initialize Database
	log.Println("Initializing database...")
	database.InitDB()

	// Initialize and start Server
	server := api.NewServer(*port)
	if err := server.Run(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

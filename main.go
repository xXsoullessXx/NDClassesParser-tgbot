package main

import (
	"flag"
	"log"
	"os"

	"NDClasses/clients/checker"
	"NDClasses/clients/database"
	"NDClasses/clients/logger"
	"NDClasses/clients/telegram"
)

func main() {
	// Define command-line flags
	debugMode := flag.Bool("debug", false, "Enable debug mode to see all parser actions")
	flag.Parse()

	// Create logger based on debug mode flag
	logger := logger.New(*debugMode)

	// Log startup message
	logger.Info("Starting ND Classes Parser Bot")
	if *debugMode {
		logger.Info("Debug mode enabled")
	}

	// Get bot token from environment variables
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		log.Fatal("BOT_TOKEN not set in environment variables")
	}

	// Create database connection
	db, err := database.New()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Create Telegram client
	TGclient := telegram.New("api.telegram.org", botToken)

	// Create message processor
	processor := telegram.NewMessageProcessor(&TGclient, db, logger)

	// Create and start checker service
	checker := checker.New(db, TGclient, logger)
	checker.Start()

	// Start polling for updates
	logger.Info("Starting bot polling...")
	if err := TGclient.PollUpdates(processor); err != nil {
		log.Fatalf("Error polling updates: %v", err)
	}
}

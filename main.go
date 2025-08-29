package main

import (
	"fmt"
	"log"
	"os"

	"NDClasses/clients/checker"
	"NDClasses/clients/database"
	"NDClasses/clients/telegram"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
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
	processor := telegram.NewMessageProcessor(&TGclient, db)

	// Create and start checker service
	checker := checker.New(db, TGclient)
	checker.Start()

	// Start polling for updates
	fmt.Println("Starting bot polling...")
	if err := TGclient.PollUpdates(processor); err != nil {
		log.Fatalf("Error polling updates: %v", err)
	}
}

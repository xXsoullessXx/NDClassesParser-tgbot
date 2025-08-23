package main

import (
	"fmt"
	"log"
	"os"

	"NDClasses/clients/ndparser"

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

	// Create Telegram client
	//TGclient := telegram.New("api.telegram.org", botToken)

	//chatID := int64(779703230)

	// Example of using the ND parser
	// Create ND parser
	ndParser := ndparser.New()

	// Search for a class by CRN
	class, err := ndParser.SearchClass("12345")
	if err != nil {
		log.Printf("Error searching for class: %v", err)
	} else {
		fmt.Println(*class)
	}
}

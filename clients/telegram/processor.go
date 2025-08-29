package telegram

import (
	"fmt"
	"strings"

	"NDClasses/clients/database"
	"NDClasses/clients/ndparser"
)

// MessageProcessor handles processing of Telegram messages and commands
type MessageProcessor struct {
	client *Client
	parser ndparser.Parser
	db     *database.Database
}

// NewMessageProcessor creates a new message processor
func NewMessageProcessor(client *Client, db *database.Database) *MessageProcessor {
	return &MessageProcessor{
		client: client,
		parser: ndparser.New(),
		db:     db,
	}
}

// ProcessUpdate processes a single update
func (p *MessageProcessor) ProcessUpdate(update Update) error {
	// Check if the update contains a message
	if update.Message.Text == "" {
		return nil // No text message to process
	}

	// Get chat ID and message text
	chatID := update.Message.Chat.ID
	text := strings.TrimSpace(update.Message.Text)

	// Check if it's a command (starts with /)
	if strings.HasPrefix(text, "/") {
		return p.processCommand(chatID, text)
	}

	// Process regular message
	return p.processMessage(chatID, text)
}

// processCommand processes a command message
func (p *MessageProcessor) processCommand(chatID int64, command string) error {
	// Create or get user in database
	user, err := p.db.CreateUser(chatID, "")
	if err != nil {
		fmt.Printf("Error creating/getting user: %v\n", err)
	}

	switch command {
	case "/start":
		return p.client.SendMessage(chatID, "Hello! I'm the ND Classes bot. I can help you track class availability.\n\nUse /add CRN to add a class to track\nUse /remove CRN to stop tracking a class\nUse /list to see all classes you're tracking\nUse /check CRN to check a class availability now")
	case "/help":
		return p.client.SendMessage(chatID, "Available commands:\n/start - Start the bot\n/help - Show this help message\n/add CRN - Add a class to track\n/remove CRN - Stop tracking a class\n/list - List all tracked classes\n/check CRN - Check class availability now")
	case "/list":
		return p.listTrackedCRNs(chatID, user.ID)
	default:
		// Check if it's a command with arguments
		if strings.HasPrefix(command, "/add ") {
			crn := strings.TrimSpace(strings.TrimPrefix(command, "/add "))
			return p.addTrackedCRN(chatID, user.ID, crn)
		} else if strings.HasPrefix(command, "/remove ") {
			crn := strings.TrimSpace(strings.TrimPrefix(command, "/remove "))
			return p.removeTrackedCRN(chatID, user.ID, crn)
		} else if strings.HasPrefix(command, "/check ") {
			crn := strings.TrimSpace(strings.TrimPrefix(command, "/check "))
			p.client.SendMessage(chatID, "Checking...")
			return p.checkClassAvailability(chatID, crn)
		} else if strings.HasPrefix(command, "/check_") {
			crn := strings.TrimPrefix(command, "/check_")
			p.client.SendMessage(chatID, "Checking...")
			return p.checkClassAvailability(chatID, crn)
		}
		return p.client.SendMessage(chatID, "Unknown command. Type /help for available commands.")
	}
}

// addTrackedCRN adds a CRN to the user's tracking list
func (p *MessageProcessor) addTrackedCRN(chatID int64, userID int64, crn string) error {
	// Check class availability to get the title
	class, err := p.parser.SearchClass(crn)
	if err != nil {
		return p.client.SendMessage(chatID, fmt.Sprintf("Error checking class: %v", err))
	}

	// Add CRN to database
	trackedCRN, err := p.db.AddTrackedCRN(userID, crn, class.Title)
	if err != nil {
		return p.client.SendMessage(chatID, fmt.Sprintf("Error adding CRN to tracking list: %v", err))
	}

	// Update the title in case it changed
	if trackedCRN.Title != class.Title {
		p.db.UpdateCRNTitle(userID, crn, class.Title)
	}

	return p.client.SendMessage(chatID, fmt.Sprintf("Added CRN %s (%s) to your tracking list.", crn, class.Title))
}

// removeTrackedCRN removes a CRN from the user's tracking list
func (p *MessageProcessor) removeTrackedCRN(chatID int64, userID int64, crn string) error {
	// Remove CRN from database
	err := p.db.RemoveTrackedCRN(userID, crn)
	if err != nil {
		return p.client.SendMessage(chatID, fmt.Sprintf("Error removing CRN from tracking list: %v", err))
	}

	return p.client.SendMessage(chatID, fmt.Sprintf("Removed CRN %s from your tracking list.", crn))
}

// listTrackedCRNs lists all CRNs tracked by the user
func (p *MessageProcessor) listTrackedCRNs(chatID int64, userID int64) error {
	// Get all tracked CRNs for the user
	crns, err := p.db.GetUserTrackedCRNs(userID)
	if err != nil {
		return p.client.SendMessage(chatID, fmt.Sprintf("Error retrieving tracked CRNs: %v", err))
	}

	if len(crns) == 0 {
		return p.client.SendMessage(chatID, "You are not tracking any classes.")
	}

	// Format the response
	response := "You are tracking the following classes:\n"
	for _, crn := range crns {
		response += fmt.Sprintf("- %s (%s)\n", crn.CRN, crn.Title)
	}

	return p.client.SendMessage(chatID, response)
}

// processMessage processes a regular message
func (p *MessageProcessor) processMessage(chatID int64, text string) error {
	// For now, just echo the message
	return p.client.SendMessage(chatID, fmt.Sprintf("You said: %s", text))
}

// checkClassAvailability checks the availability of a class by CRN
func (p *MessageProcessor) checkClassAvailability(chatID int64, crn string) error {
	// Use the ND parser to check class availability
	class, err := p.parser.SearchClass(crn)
	if err != nil {
		return p.client.SendMessage(chatID, fmt.Sprintf("Error checking class availability: %v", err))
	}

	// Format the response
	response := fmt.Sprintf("Class CRN %s:\nTitle: %s\nSeats Available: %d",
		class.CRN, class.Title, class.Seats)

	return p.client.SendMessage(chatID, response)
}

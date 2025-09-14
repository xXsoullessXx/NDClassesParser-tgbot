package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"NDClasses/clients/database"
	"NDClasses/clients/logger"
	"NDClasses/clients/ndparser"
)

// MessageProcessor handles processing of Telegram messages and commands
type MessageProcessor struct {
	client *Client
	parser ndparser.Parser
	db     *database.Database
	logger *logger.Logger
}

// NewMessageProcessor creates a new message processor
func NewMessageProcessor(client *Client, db *database.Database, logger *logger.Logger) *MessageProcessor {
	return &MessageProcessor{
		client: client,
		parser: ndparser.New(logger),
		db:     db,
		logger: logger,
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
			go p.addTrackedCRN(chatID, user.ID, crn)
			return nil
		} else if strings.HasPrefix(command, "/remove ") {
			crn := strings.TrimSpace(strings.TrimPrefix(command, "/remove "))
			return p.removeTrackedCRN(chatID, user.ID, crn)
		} else if strings.HasPrefix(command, "/check ") {
			crn := strings.TrimSpace(strings.TrimPrefix(command, "/check "))
			p.logger.Info("Processing /check command for CRN: %s, ChatID: %d", crn, chatID)
			p.client.SendMessage(chatID, "Checking...")
			go p.checkClassAvailability(chatID, crn)
			return nil
		} else if strings.HasPrefix(command, "/check_") {
			crn := strings.TrimPrefix(command, "/check_")
			p.logger.Info("Processing /check_ command for CRN: %s, ChatID: %d", crn, chatID)
			p.client.SendMessage(chatID, "Checking...")
			go p.checkClassAvailability(chatID, crn)
			return nil
		}
		return p.client.SendMessage(chatID, "Unknown command. Type /help for available commands.")
	}
}

// addTrackedCRN adds a CRN to the user's tracking list
func (p *MessageProcessor) addTrackedCRN(chatID int64, userID int64, crn string) error {
	p.logger.Info("Starting addTrackedCRN for CRN: %s, UserID: %d, ChatID: %d", crn, userID, chatID)

	// Validate CRN input
	if crn == "" {
		p.logger.Error("Empty CRN provided for addTrackedCRN")
		return p.client.SendMessage(chatID, "Error: CRN cannot be empty")
	}
	err := p.client.SendMessage(chatID, "Checking class and adding to list...")
	p.logger.Debug("CRN validation passed: %s", crn)

	// Check class availability to get the title
	p.logger.Info("Calling parser.SearchClass for CRN: %s (addTrackedCRN)", crn)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	class, err := p.parser.SearchClass(ctx, crn)
	if err != nil {
		p.logger.Error("SearchClass failed in addTrackedCRN for CRN %s: %v", crn, err)
		p.logger.Error("Error type: %T", err)
		p.logger.Error("Error details: %+v", err)

		errorMsg := fmt.Sprintf("Error checking class for CRN %s:\n\nError: %v\n\nType: %T\n\nPlease verify the CRN is correct and try again.", crn, err, err)
		return p.client.SendMessage(chatID, errorMsg)
	}

	p.logger.Info("SearchClass successful in addTrackedCRN for CRN %s", crn)
	p.logger.Debug("Class data retrieved - CRN: %s, Title: %s, Seats: %d", class.CRN, class.Title, class.Seats)

	// Validate class data
	if class == nil {
		p.logger.Error("SearchClass returned nil class in addTrackedCRN for CRN: %s", crn)
		return p.client.SendMessage(chatID, fmt.Sprintf("Error: No class data returned for CRN %s", crn))
	}

	if class.Title == "" {
		p.logger.Info("Class title is empty for CRN: %s, using fallback", crn)
		class.Title = "Unknown Class"
	}

	// Add CRN to database
	p.logger.Info("Adding CRN %s to database for UserID: %d", crn, userID)
	trackedCRN, isNew, err := p.db.AddTrackedCRN(userID, crn, class.Title)
	if err != nil {
		p.logger.Error("Failed to add CRN %s to database for UserID %d: %v", crn, userID, err)
		return p.client.SendMessage(chatID, fmt.Sprintf("Error adding CRN to tracking list: %v", err))
	}

	p.logger.Info("Successfully processed CRN %s in database (new: %v)", crn, isNew)

	// Update the title in case it changed
	if trackedCRN.Title != class.Title {
		p.logger.Info("Updating CRN title from '%s' to '%s' for CRN: %s", trackedCRN.Title, class.Title, crn)
		p.db.UpdateCRNTitle(userID, crn, class.Title)
	}

	// Prepare appropriate message based on whether it was newly added or already existed
	var successMsg string
	if isNew {
		successMsg = fmt.Sprintf("Added CRN %s (%s) to your tracking list.", crn, class.Title)
	} else {
		successMsg = fmt.Sprintf("CRN %s (%s) is already in your tracking list.", crn, class.Title)
	}
	p.logger.Info("Sending message for CRN %s: %s", crn, successMsg)

	return p.client.SendMessage(chatID, successMsg)
}

// removeTrackedCRN removes a CRN from the user's tracking list
func (p *MessageProcessor) removeTrackedCRN(chatID int64, userID int64, crn string) error {
	// Remove CRN from database
	wasRemoved, err := p.db.RemoveTrackedCRN(userID, crn)
	if err != nil {
		return p.client.SendMessage(chatID, fmt.Sprintf("Error removing CRN from tracking list: %v", err))
	}

	// Send appropriate message based on whether the CRN was actually removed
	if wasRemoved {
		return p.client.SendMessage(chatID, fmt.Sprintf("Removed CRN %s from your tracking list.", crn))
	} else {
		return p.client.SendMessage(chatID, fmt.Sprintf("CRN %s is not in your tracking list.", crn))
	}
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
	p.logger.Info("Starting class availability check for CRN: %s, ChatID: %d", crn, chatID)

	// Validate CRN input
	if crn == "" {
		p.logger.Error("Empty CRN provided for class availability check")
		return p.client.SendMessage(chatID, "Error: CRN cannot be empty")
	}

	p.logger.Debug("CRN validation passed: %s", crn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	p.logger.Debug("Created context with 90s timeout for SearchClass operation")

	// Use the ND parser to check class availability
	p.logger.Info("Calling parser.SearchClass for CRN: %s", crn)
	class, err := p.parser.SearchClass(ctx, crn)

	if err != nil {
		p.logger.Error("SearchClass failed for CRN %s: %v", crn, err)
		p.logger.Error("Error type: %T", err)
		p.logger.Error("Error details: %+v", err)

		// Send detailed error message to user
		errorMsg := fmt.Sprintf("Error checking class availability for CRN %s:\n\nError: %v\n\nType: %T\n\nPlease try again or contact support if the issue persists.", crn, err, err)
		return p.client.SendMessage(chatID, errorMsg)
	}

	p.logger.Info("SearchClass successful for CRN %s", crn)
	p.logger.Debug("Class data retrieved - CRN: %s, Title: %s, Seats: %d", class.CRN, class.Title, class.Seats)

	// Validate class data
	if class == nil {
		p.logger.Error("SearchClass returned nil class for CRN: %s", crn)
		return p.client.SendMessage(chatID, fmt.Sprintf("Error: No class data returned for CRN %s", crn))
	}

	if class.CRN == "" {
		p.logger.Info("Class CRN is empty for input CRN: %s", crn)
		class.CRN = crn // Use input CRN as fallback
	}

	if class.Title == "" {
		p.logger.Info("Class title is empty for CRN: %s", crn)
		class.Title = "Unknown Class"
	}

	p.logger.Debug("Class data validation completed successfully")

	// Format the response
	response := fmt.Sprintf("Class CRN %s:\nTitle: %s\nSeats Available: %d",
		class.CRN, class.Title, class.Seats)

	p.logger.Info("Sending response to user for CRN %s: %s", crn, response)

	err = p.client.SendMessage(chatID, response)
	if err != nil {
		p.logger.Error("Failed to send response message for CRN %s: %v", crn, err)
		return err
	}

	p.logger.Info("Successfully completed class availability check for CRN: %s", crn)
	return nil
}

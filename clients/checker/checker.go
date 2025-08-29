package checker

import (
	"fmt"
	"log"
	"time"

	"NDClasses/clients/database"
	"NDClasses/clients/ndparser"
	"NDClasses/clients/telegram"
)

// Checker periodically checks class availability for all tracked CRNs
type Checker struct {
	db     *database.Database
	parser ndparser.Parser
	client telegram.Client
}

// New creates a new checker
func New(db *database.Database, client telegram.Client) *Checker {
	return &Checker{
		db:     db,
		parser: ndparser.New(),
		client: client,
	}
}

// Start begins the periodic checking process
func (c *Checker) Start() {
	go func() {
		for {
			// Check all tracked CRNs
			if err := c.checkAllTrackedCRNs(); err != nil {
				log.Printf("Error checking tracked CRNs: %v", err)
			}

			// Wait for 5 minutes before checking again
			time.Sleep(5 * time.Minute)
		}
	}()
}

// checkAllTrackedCRNs checks availability for all tracked CRNs
func (c *Checker) checkAllTrackedCRNs() error {
	// Get all tracked CRNs
	trackedCRNs, err := c.db.GetAllTrackedCRNs()
	if err != nil {
		return fmt.Errorf("failed to get tracked CRNs: %w", err)
	}

	// Keep track of users we've already notified to avoid duplicate messages
	notifiedUsers := make(map[int64]bool)

	// Check each CRN
	for _, crn := range trackedCRNs {
		// Check class availability
		class, err := c.parser.SearchClass(crn.CRN)
		if err != nil {
			log.Printf("Error checking class %s: %v", crn.CRN, err)
			continue
		}

		// If seats are available and we haven't notified this user yet
		if class.Seats > 0 && !notifiedUsers[crn.UserID] {
			// Get user by ID
			user, err := c.db.GetUserByID(crn.UserID)
			if err != nil {
				log.Printf("Error getting user %d: %v", crn.UserID, err)
				continue
			}

			// Send notification
			message := fmt.Sprintf("Good news! Class %s (%s) now has %d seat(s) available.",
				crn.CRN, crn.Title, class.Seats)
			if err := c.client.SendMessage(user.TelegramID, message); err != nil {
				log.Printf("Error sending message to user %d: %v", user.TelegramID, err)
			}

			// Mark user as notified
			notifiedUsers[crn.UserID] = true
		}
	}

	return nil
}

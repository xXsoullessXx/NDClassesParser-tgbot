package checker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"NDClasses/clients/database"
	"NDClasses/clients/logger"
	"NDClasses/clients/ndparser"
	"NDClasses/clients/telegram"
)

// Checker periodically checks class availability for all tracked CRNs
type Checker struct {
	db     *database.Database
	parser ndparser.Parser
	client telegram.Client
	logger *logger.Logger
}

// New creates a new checker
func New(db *database.Database, client telegram.Client, logger *logger.Logger) *Checker {
	return &Checker{
		db:     db,
		parser: ndparser.New(logger),
		client: client,
		logger: logger,
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

			// Wait for 3 minutes before checking again
			time.Sleep(3 * time.Minute)
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
	var mu sync.Mutex

	// Check each CRN
	wg := sync.WaitGroup{}
	for _, crn := range trackedCRNs {
		wg.Add(1)
		// Check class availability
		go func(crn database.TrackedCRN) {
			defer wg.Done()
			class, err := c.parser.SearchClass(context.Background(), crn.CRN)
			if err != nil {
				log.Printf("Error checking class %s: %v", crn.CRN, err)
				return
			}

			// If seats are available and we haven't notified this user yet
			mu.Lock()
			alreadyNotified := notifiedUsers[crn.UserID]
			mu.Unlock()

			if class.Seats > 0 && !alreadyNotified {
				// Get user by ID
				user, err := c.db.GetUserByID(crn.UserID)
				if err != nil {
					log.Printf("Error getting user %d: %v", crn.UserID, err)
					return
				}

				// Send notification
				message := fmt.Sprintf("Good news! Class %s (%s) now has %d seat(s) available.",
					crn.CRN, crn.Title, class.Seats)
				if err := c.client.SendMessage(user.TelegramID, message); err != nil {
					log.Printf("Error sending message to user %d: %v", user.TelegramID, err)
				}

				// Mark user as notified
				mu.Lock()
				notifiedUsers[crn.UserID] = true
				mu.Unlock()
			}
		}(crn)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	return nil
}

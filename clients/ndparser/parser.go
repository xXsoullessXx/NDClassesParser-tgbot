package ndparser

import (
	"context"
	"fmt"
	"time"

	"NDClasses/clients/web"

	"github.com/chromedp/chromedp"
)

// Parser represents a parser for ND class information
type Parser struct {
	client  web.Client
	timeout time.Duration
}

// New creates a new ND class parser
func New() Parser {
	return Parser{
		client:  web.New(),
		timeout: 20 * time.Second, // Default timeout of 60 seconds
	}
}

// SearchClass searches for a class by CRN
func (p *Parser) SearchClass(crn string) (*Class, error) {
	// Create a new chromedp context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Adding timeout to context
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Navigate to the term selection page
	termURL := "https://bxeregprod.oit.nd.edu/StudentRegistration/ssb/term/termSelection?mode=search"

	// Execute the chromedp tasks
	var class Class
	var seatsStr string

	err := chromedp.Run(ctx,
		// Navigate to the term selection page
		chromedp.Navigate(termURL),

		// Wait for page to load
		chromedp.Sleep(14*time.Second),

		// Click on the term selection input
		chromedp.Click(`#s2id_txt_term`, chromedp.ByID),

		// Wait for the term selection input to be ready
		chromedp.Sleep(2*time.Second),

		// Fill in "Fall Semester 2025" in the term selection field
		chromedp.SendKeys(`#s2id_autogen1_search`, "Fall Semester 2025", chromedp.ByID),
		chromedp.Sleep(1*time.Second),
		chromedp.SendKeys(`#s2id_autogen1_search`, "\n", chromedp.ByID),

		// Click the search button
		chromedp.Click(`#term-go`, chromedp.ByID),

		// Wait for the keyword input to be ready
		chromedp.WaitVisible(`#txt_keywordlike`, chromedp.ByID),

		// Fill in the CRN in the keyword field
		chromedp.SendKeys(`#txt_keywordlike`, crn, chromedp.ByID),

		// Wait a bit and click the search button again
		chromedp.Sleep(1*time.Second),
		chromedp.Click(`#search-go`, chromedp.ByID),

		// Wait for results to load
		chromedp.Sleep(3*time.Second),

		// Extract class information
		chromedp.Text(`[data-content="Title"]`, &class.Title, chromedp.ByQuery),
		chromedp.Text(`[data-content="Status"]`, &seatsStr, chromedp.ByQuery),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to parse class information: %w", err)
	}

	// Convert string values to integers
	fmt.Sscanf(seatsStr, "%d", &class.Seats)

	class.CRN = crn

	return &class, nil
}

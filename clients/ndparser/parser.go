package ndparser

import (
	"context"
	"fmt"
	"time"

	"NDClasses/clients/logger"
	"NDClasses/clients/web"

	"github.com/chromedp/chromedp"
)

// Parser represents a parser for ND class information
type Parser struct {
	client  web.Client
	timeout time.Duration
	logger  *logger.Logger
}

// New creates a new ND class parser
func New(logger *logger.Logger) Parser {
	return Parser{
		client:  web.New(),
		timeout: 30 * time.Second, // Default timeout of 30 seconds
		logger:  logger,
	}
}

// SearchClass searches for a class by CRN
func (p *Parser) SearchClass(ctx context.Context, crn string) (*Class, error) {
	headless := !p.logger.IsDebugMode() // Headless in normal mode, visible in debug mode
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.Flag("no-sandbox", true), // Show browser in debug mode
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("remote-debugging-port", "0"),
		chromedp.Flag("disable-dev-shm-usage", true),                                       // Включить GPU (если нужно)
		chromedp.Flag("ignore-certificate-errors", true),                                   // Игнорировать ошибки сертификатов
		chromedp.Flag("window-size", "1200,800"),                                           // Размер окна
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"), // Кастомный User-Agent
		chromedp.Flag("disable-crash-reporter", true),                                      // Disable crash reporting to avoid database requirement
		chromedp.Flag("crash-dumps-dir", "/tmp"),                                           // Set crash dumps directory
	)

	p.logger.Debug("Launching Chrome with headless=%v", headless)

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()
	// Create a new chromedp context
	var newCtx context.Context
	if p.logger.IsDebugMode() {
		newCtx, cancel = chromedp.NewContext(allocCtx, chromedp.WithLogf(func(s string, i ...interface{}) {
			p.logger.Debug("ChromeDP: "+s, i...)
		}))
	} else {
		newCtx, cancel = chromedp.NewContext(allocCtx)
	}
	ctx = newCtx
	defer cancel()

	// Adding timeout to context
	p.logger.Debug("Before WithTimeout, ctx is done: %v", ctx.Err() != nil)
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	p.logger.Debug("After WithTimeout on ctx, ctx is done: %v", ctx.Err() != nil)
	defer cancel()

	// Navigate to the term selection page
	termURL := "https://bxeregprod.oit.nd.edu/StudentRegistration/ssb/term/termSelection?mode=search"

	// Execute the chromedp tasks
	var class Class
	var seatsStr string

	p.logger.Debug("Before chromedp.Run, ctx is done: %v", ctx.Err() != nil)
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
	p.logger.Debug("After chromedp.Run, ctx is done: %v, err: %v", ctx.Err() != nil, err)

	if err != nil {
		return nil, fmt.Errorf("failed to parse class information: %w", err)
	}

	// Convert string values to integers
	fmt.Sscanf(seatsStr, "%d", &class.Seats)

	class.CRN = crn

	return &class, nil
}

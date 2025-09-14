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
		timeout: 60 * time.Second, // Default timeout of 60 seconds for Railway
		logger:  logger,
	}
}

// SearchClass searches for a class by CRN
func (p *Parser) SearchClass(ctx context.Context, crn string) (*Class, error) {
	// Simplified Chrome configuration for Railway
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.Flag("remote-debugging-port", "0"),
		chromedp.Flag("disable-logging", true),
		chromedp.Flag("disable-crash-reporter", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-crashpad", true),
		chromedp.Flag("no-crash-upload", true),
		chromedp.Flag("user-data-dir", "/tmp/chrome-user-data"),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-translate", true),
		chromedp.Flag("disable-component-update", true),
	)

	// Try direct Chrome path first, fallback to wrapper if needed
	opts = append(opts, chromedp.ExecPath("/usr/bin/chromium-browser"))

	// Add essential environment variables
	opts = append(opts, chromedp.Env("DISPLAY", ":99"))

	p.logger.Info("Creating Chrome allocator context for CRN: %s", crn)
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	p.logger.Info("Creating Chrome context for CRN: %s", crn)
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

	p.logger.Info("Chrome context created successfully for CRN: %s", crn)

	// Test Chrome startup with a simple operation first
	p.logger.Info("Testing Chrome startup for CRN: %s", crn)
	testCtx, testCancel := context.WithTimeout(ctx, 20*time.Second)
	defer testCancel()

	testErr := chromedp.Run(testCtx,
		chromedp.Navigate("about:blank"),
		chromedp.Sleep(2*time.Second),
	)
	if testErr != nil {
		p.logger.Error("Chrome startup test failed for CRN %s: %v", crn, testErr)
		p.logger.Error("Chrome startup test error type: %T", testErr)
		return nil, fmt.Errorf("Chrome failed to start: %w", testErr)
	}
	p.logger.Info("Chrome startup test successful for CRN: %s", crn)

	// Test navigation to a simple page first
	p.logger.Info("Testing navigation to Google for CRN: %s", crn)
	navCtx, navCancel := context.WithTimeout(ctx, 15*time.Second)
	defer navCancel()

	navErr := chromedp.Run(navCtx, chromedp.Navigate("https://www.google.com"))
	if navErr != nil {
		p.logger.Error("Navigation test failed for CRN %s: %v", crn, navErr)
		return nil, fmt.Errorf("Chrome navigation test failed: %w", navErr)
	}
	p.logger.Info("Navigation test successful for CRN: %s", crn)

	// Adding timeout to context - reduced for Railway environment
	p.logger.Debug("Before WithTimeout, ctx is done: %v", ctx.Err() != nil)
	ctx, cancel = context.WithTimeout(ctx, 90*time.Second) // Reduced to 90 seconds for Railway
	p.logger.Debug("After WithTimeout on ctx, ctx is done: %v", ctx.Err() != nil)
	defer cancel()

	// Navigate to the term selection page
	termURL := "https://bxeregprod.oit.nd.edu/StudentRegistration/ssb/term/termSelection?mode=search"

	// Execute the chromedp tasks
	var class Class
	var seatsStr string

	p.logger.Debug("Before chromedp.Run, ctx is done: %v", ctx.Err() != nil)
	p.logger.Info("Starting Chrome automation for CRN: %s", crn)

	p.logger.Debug("Starting Chrome automation steps for CRN: %s", crn)

	// Step 1: Navigate to the page
	p.logger.Info("Step 1: Navigating to term selection page for CRN: %s", crn)
	p.logger.Debug("Context status before Step 1: %v", ctx.Err())
	err := chromedp.Run(ctx, chromedp.Navigate(termURL))
	if err != nil {
		p.logger.Error("Step 1 failed for CRN %s: %v", crn, err)
		return nil, fmt.Errorf("failed to navigate to page: %w", err)
	}
	p.logger.Info("Step 1 completed for CRN: %s", crn)

	// Step 2: Wait for page load
	p.logger.Info("Step 2: Waiting for page load for CRN: %s", crn)
	p.logger.Debug("Context status before Step 2: %v", ctx.Err())
	err = chromedp.Run(ctx, chromedp.Sleep(8*time.Second)) // Reduced from 10 to 8 seconds
	if err != nil {
		p.logger.Error("Step 2 failed for CRN %s: %v", crn, err)
		return nil, fmt.Errorf("failed to wait for page load: %w", err)
	}
	p.logger.Info("Step 2 completed for CRN: %s", crn)

	// Step 3: Click term selection
	p.logger.Info("Step 3: Clicking term selection for CRN: %s", crn)
	err = chromedp.Run(ctx, chromedp.Click(`#s2id_txt_term`, chromedp.ByID))
	if err != nil {
		p.logger.Error("Step 3 failed for CRN %s: %v", crn, err)
		return nil, fmt.Errorf("failed to click term selection: %w", err)
	}
	p.logger.Info("Step 3 completed for CRN: %s", crn)

	// Step 4: Wait and fill term
	p.logger.Info("Step 4: Filling term selection for CRN: %s", crn)
	err = chromedp.Run(ctx,
		chromedp.Sleep(2*time.Second),
		chromedp.SendKeys(`#s2id_autogen1_search`, "Fall Semester 2025", chromedp.ByID),
		chromedp.Sleep(1*time.Second),
		chromedp.SendKeys(`#s2id_autogen1_search`, "\n", chromedp.ByID),
	)
	if err != nil {
		p.logger.Error("Step 4 failed for CRN %s: %v", crn, err)
		return nil, fmt.Errorf("failed to fill term selection: %w", err)
	}
	p.logger.Info("Step 4 completed for CRN: %s", crn)

	// Step 5: Click search button
	p.logger.Info("Step 5: Clicking search button for CRN: %s", crn)
	err = chromedp.Run(ctx, chromedp.Click(`#term-go`, chromedp.ByID))
	if err != nil {
		p.logger.Error("Step 5 failed for CRN %s: %v", crn, err)
		return nil, fmt.Errorf("failed to click search button: %w", err)
	}
	p.logger.Info("Step 5 completed for CRN: %s", crn)

	// Wait a bit longer for the page to load the keyword input field
	p.logger.Debug("Waiting for page to load keyword input field for CRN: %s", crn)
	p.logger.Debug("Context status before wait: %v", ctx.Err())

	// Use shorter wait time to avoid context cancellation
	err = chromedp.Run(ctx, chromedp.Sleep(2*time.Second))
	if err != nil {
		p.logger.Error("Wait after Step 5 failed for CRN %s: %v", crn, err)
		p.logger.Error("Context status after wait failure: %v", ctx.Err())
		return nil, fmt.Errorf("failed to wait after search: %w", err)
	}

	p.logger.Debug("Context status after wait: %v", ctx.Err())

	// Step 6: Wait for keyword input and fill CRN
	p.logger.Info("Step 6: Filling CRN field for CRN: %s", crn)

	// Create a shorter timeout for this step
	step6Ctx, step6Cancel := context.WithTimeout(ctx, 30*time.Second)
	defer step6Cancel()

	// Try to wait for the element with a timeout
	p.logger.Debug("Waiting for keyword input field to be visible for CRN: %s", crn)
	err = chromedp.Run(step6Ctx, chromedp.WaitVisible(`#txt_keywordlike`, chromedp.ByID))
	if err != nil {
		p.logger.Error("Keyword input field not visible for CRN %s: %v", crn, err)
		// Try alternative selector
		p.logger.Info("Trying alternative selector for CRN: %s", crn)
		err = chromedp.Run(step6Ctx, chromedp.WaitVisible(`input[name="keywordlike"]`, chromedp.ByQuery))
		if err != nil {
			p.logger.Error("Alternative selector also failed for CRN %s: %v", crn, err)
			return nil, fmt.Errorf("keyword input field not found: %w", err)
		}
		p.logger.Info("Alternative selector worked for CRN: %s", crn)
	}

	// Fill CRN field
	p.logger.Debug("Filling CRN field for CRN: %s", crn)
	err = chromedp.Run(step6Ctx, chromedp.SendKeys(`#txt_keywordlike`, crn, chromedp.ByID))
	if err != nil {
		p.logger.Error("Failed to fill CRN field for CRN %s: %v", crn, err)
		return nil, fmt.Errorf("failed to fill CRN field: %w", err)
	}

	// Click search button
	p.logger.Debug("Clicking search button for CRN: %s", crn)
	err = chromedp.Run(step6Ctx,
		chromedp.Sleep(1*time.Second),
		chromedp.Click(`#search-go`, chromedp.ByID),
	)
	if err != nil {
		p.logger.Error("Failed to click search button for CRN %s: %v", crn, err)
		return nil, fmt.Errorf("failed to click search button: %w", err)
	}

	p.logger.Info("Step 6 completed for CRN: %s", crn)

	// Step 7: Wait for results and extract data
	p.logger.Info("Step 7: Extracting class data for CRN: %s", crn)
	err = chromedp.Run(ctx,
		chromedp.Sleep(2*time.Second),
		chromedp.Text(`[data-content="Title"]`, &class.Title, chromedp.ByQuery),
		chromedp.Text(`[data-content="Status"]`, &seatsStr, chromedp.ByQuery),
	)
	if err != nil {
		p.logger.Error("Step 7 failed for CRN %s: %v", crn, err)
		return nil, fmt.Errorf("failed to extract class data: %w", err)
	}
	p.logger.Info("Step 7 completed for CRN: %s", crn)

	p.logger.Debug("All Chrome automation steps completed for CRN: %s", crn)

	// Check if context was canceled
	if ctx.Err() == context.Canceled {
		p.logger.Error("Context was canceled during Chrome automation for CRN: %s", crn)
		return nil, fmt.Errorf("operation was canceled: %w", ctx.Err())
	}

	if ctx.Err() == context.DeadlineExceeded {
		p.logger.Error("Context deadline exceeded during Chrome automation for CRN: %s", crn)
		return nil, fmt.Errorf("operation timed out: %w", ctx.Err())
	}

	// Convert string values to integers
	fmt.Sscanf(seatsStr, "%d", &class.Seats)

	class.CRN = crn

	return &class, nil
}

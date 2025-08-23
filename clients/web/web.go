package web

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client represents a web client for parsing information from sites
type Client struct {
	client  http.Client
	timeout time.Duration
}

// New creates a new web client
func New() Client {
	return Client{
		client:  http.Client{},
		timeout: 30 * time.Second, // Default timeout of 30 seconds
	}
}

// Updates retrieves updates from a web source
func (c *Client) Updates(targetURL string) ([]Update, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	data, err := c.doRequest(ctx, targetURL, "GET", nil)
	if err != nil {
		return nil, err
	}

	// For simplicity, we're returning a single update with the content
	// In a real implementation, this would parse the data and extract updates
	updates := []Update{
		{
			ID:      1,
			Content: string(data),
		},
	}

	return updates, nil
}

// SendMessage sends data to a web endpoint
func (c *Client) SendMessage(targetURL string, data string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Create form values for POST request
	formData := url.Values{}
	formData.Add("data", data)

	_, err := c.doRequest(ctx, targetURL, "POST", formData)
	if err != nil {
		return fmt.Errorf("can't send message: %w", err)
	}

	return nil
}

// doRequest performs an HTTP request with context
func (c *Client) doRequest(ctx context.Context, targetURL string, method string, data url.Values) ([]byte, error) {
	var req *http.Request
	var err error

	if method == "POST" {
		req, err = http.NewRequestWithContext(ctx, method, targetURL, nil)
		if err != nil {
			return nil, fmt.Errorf("can't create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.URL.RawQuery = data.Encode()
	} else {
		req, err = http.NewRequestWithContext(ctx, method, targetURL, nil)
		if err != nil {
			return nil, fmt.Errorf("can't create request: %w", err)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't do request: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("can't read response body: %w", err)
	}

	return body, nil
}

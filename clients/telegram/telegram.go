package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

type Client struct {
	host     string
	basePath string
	client   http.Client
	timeout  time.Duration
}

func New(host string, token string) Client {
	return Client{
		host:     host,
		basePath: "bot" + token,
		client:   http.Client{},
		timeout:  5 * time.Second, // Default timeout of 30 seconds
	}
}

func (c *Client) Updates(offset int, limit int) ([]Update, error) {
	q := url.Values{}
	q.Add("offset", strconv.Itoa(offset))
	q.Add("limit", strconv.Itoa(limit))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	data, err := c.doRequest(ctx, "getUpdates", q)
	if err != nil {
		return nil, err
	}
	var resp UpdatesResponse

	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("can't parse json: %w", err)
	}

	return resp.Result, nil
}

func (c *Client) SendMessage(chatID int64, text string) error {
	q := url.Values{}
	q.Add("chat_id", strconv.FormatInt(chatID, 10))
	q.Add("text", text)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	_, err := c.doRequest(ctx, "sendMessage", q)
	if err != nil {
		return fmt.Errorf("can't send message: %w", err)
	}

	return nil
}

func (c *Client) doRequest(ctx context.Context, method string, query url.Values) ([]byte, error) {
	url := url.URL{
		Scheme: "https",
		Host:   c.host,
		Path:   path.Join(c.basePath, method),
	}
	const errMsg = "can't do request: %w"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	req.URL.RawQuery = query.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(errMsg, err)
	}
	return body, nil
}

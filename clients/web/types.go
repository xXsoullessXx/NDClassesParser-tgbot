package web

// WebResponse represents a generic response from a web request
type WebResponse struct {
	Ok     bool   `json:"ok"`
	Result string `json:"result"`
}

// Update represents a generic update from a web source
type Update struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

// UpdatesResponse represents a response containing multiple updates
type UpdatesResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	URL  string `json:"url"`
	Data string `json:"data"`
}

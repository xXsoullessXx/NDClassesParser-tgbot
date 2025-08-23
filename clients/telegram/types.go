package telegram

type UpdatesResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	ID      int    `json:"update_id"`
	Message string `json:"message"`
}

type SendMessageRequest struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

type Message struct {
	MessageID int `json:"message_id"`
	// Add other fields as needed
}

type SendMessageResponse struct {
	Ok     bool    `json:"ok"`
	Result Message `json:"result"`
}

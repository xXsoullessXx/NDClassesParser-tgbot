package database

// User represents a Telegram user
type User struct {
	ID         int64  `json:"id" gorm:"primaryKey"`
	TelegramID int64  `json:"telegram_id" gorm:"uniqueIndex"`
	Username   string `json:"username"`
	CreatedAt  int64  `json:"created_at"`
}

// TrackedCRN represents a CRN that a user wants to track
type TrackedCRN struct {
	ID        int64  `json:"id" gorm:"primaryKey"`
	UserID    int64  `json:"user_id" gorm:"index"`
	CRN       string `json:"crn" gorm:"index"`
	Title     string `json:"title"`
	Active    bool   `json:"active" gorm:"default:true"`
	CreatedAt int64  `json:"created_at"`
}

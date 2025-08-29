# ND Classes Parser Telegram Bot

A Telegram bot that helps students track class availability at the University of Notre Dame by monitoring CRNs (Course Reference Numbers).

## Features

- Track class availability by CRN
- Receive notifications when tracked classes have open seats
- Add/remove CRNs from your tracking list
- Check class availability on demand

## Commands

- `/start` - Start the bot and see available commands
- `/help` - Show help message with available commands
- `/add CRN` - Add a class to track by CRN
- `/remove CRN` - Stop tracking a class by CRN
- `/list` - List all classes you're currently tracking
- `/check CRN` - Check class availability now

## Setup

1. Create a Telegram bot using BotFather and get your bot token
2. Set up a PostgreSQL database
3. Copy `.env.example` to `.env` and fill in your configuration:
   ```
   BOT_TOKEN=your_telegram_bot_token_here
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=your_database_user
   DB_PASSWORD=your_database_password
   DB_NAME=ndclasses
   ```
4. Run `go mod tidy` to install dependencies
5. Run the bot with `go run main.go`

## Database Schema

The bot uses two tables:

### Users
- `id` - Primary key
- `telegram_id` - Unique Telegram user ID
- `username` - Telegram username (optional)
- `created_at` - Unix timestamp of when the user was created

### TrackedCRNs
- `id` - Primary key
- `user_id` - Foreign key to Users table
- `crn` - Course Reference Number
- `title` - Class title
- `active` - Whether the CRN is actively being tracked
- `created_at` - Unix timestamp of when the CRN was added

## How It Works

1. Users interact with the bot through Telegram commands
2. The bot stores user information and their tracked CRNs in a PostgreSQL database
3. A background service checks all tracked CRNs every 5 minutes
4. When a class has available seats, the bot notifies the user via Telegram

## Dependencies

- Go 1.19+
- PostgreSQL
- Telegram Bot API
- GORM for database ORM
- chromedp for web scraping

## License

This project is licensed under the MIT License.
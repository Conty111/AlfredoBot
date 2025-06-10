package configs

type Configuration struct {
	App      *App
	DB       *DatabaseConfig
	Telegram *TelegramConfig
}

type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     int
	SSLMode  string
	DSN      string
}

// TelegramConfig holds configuration for the Telegram bot
type TelegramConfig struct {
	Token     string
	Debug     bool
	Timeout   int
	WebhookURL string
	UseWebhook bool
}

type App struct {
	Name        string
	Version     string
	Environment string
}

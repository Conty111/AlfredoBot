package configs

// Configuration contains all application configurations
type Configuration struct {
	App      *App            `mapstructure:"app"`
	DB       *DatabaseConfig `mapstructure:"db"`
	Telegram *TelegramConfig `mapstructure:"telegram"`
	S3       *S3Config       `mapstructure:"s3"`
}

// App contains application configuration
type App struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
	LogLevel    string `mapstructure:"log_level"`
	LogFile     string `mapstructure:"log_file"`
	JSONLogs    bool   `mapstructure:"enable_json_logs"`
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
	DSN      string `mapstructure:"-"`
}

// TelegramConfig contains Telegram bot configuration
type TelegramConfig struct {
	Token      string `mapstructure:"token"`
	TokenFile  string `mapstructure:"token_file"`
	Timeout    int    `mapstructure:"timeout"`
	WebhookURL string `mapstructure:"webhook_url"`
	UseWebhook bool   `mapstructure:"use_webhook"`
	Debug      bool   `mapstructure:"debug"`
}

// S3Config contains S3 storage configuration
type S3Config struct {
	Endpoint            string `mapstructure:"endpoint"`
	Region              string `mapstructure:"region"`
	AccessKeyID         string `mapstructure:"access_key_id"`
	AccessKeyIDFile     string `mapstructure:"access_key_id_file"`
	SecretAccessKey     string `mapstructure:"secret_access_key"`
	SecretAccessKeyFile string `mapstructure:"secret_access_key_file"`
	Bucket              string `mapstructure:"bucket"`
	UseSSL              bool   `mapstructure:"use_ssl"`
}

// GetConfig loads configuration using default path
func GetConfig() (*Configuration, error) {
	return LoadConfig("")
}

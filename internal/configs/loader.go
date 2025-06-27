package configs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Configuration, error) {
	v := viper.New()

	setDefaults(v)

	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}
	v.AutomaticEnv()

	// Bind environment variables to config fields
	if err := bindEnv(v); err != nil {
		return nil, fmt.Errorf("failed to bind environment variables: %w", err)
	}

	// Load from file if path provided
	if configPath != "" {
		// Convert to absolute path
		if !filepath.IsAbs(configPath) {
			absPath, err := filepath.Abs(configPath)
			if err != nil {
				return nil, fmt.Errorf("failed to get absolute path: %w", err)
			}
			configPath = absPath
		}

		// Verify file exists first
		if _, err := os.Stat(configPath); err != nil {
			return nil, fmt.Errorf("config file not found at path: %s", configPath)
		}

		// Directly set the config file path
		v.SetConfigFile(configPath)

		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file at %s: %w", configPath, err)
		}
	}

	cfg := &Configuration{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Load credentials from files if specified
	if cfg.S3 != nil {
		if cfg.S3.AccessKeyIDFile != "" {
			data, err := os.ReadFile(cfg.S3.AccessKeyIDFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read access key file: %w", err)
			}
			cfg.S3.AccessKeyID = string(data)
		}
		if cfg.S3.SecretAccessKeyFile != "" {
			data, err := os.ReadFile(cfg.S3.SecretAccessKeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read secret key file: %w", err)
			}
			cfg.S3.SecretAccessKey = string(data)
		}
	}

	if cfg.Telegram != nil && cfg.Telegram.TokenFile != "" {
		data, err := os.ReadFile(cfg.Telegram.TokenFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read telegram token file: %w", err)
		}
		cfg.Telegram.Token = string(data)
	}

	// Generate DSN for database
	cfg.DB.DSN = fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.DB.Host, cfg.DB.User, cfg.DB.Password,
		cfg.DB.DBName, cfg.DB.Port, cfg.DB.SSLMode,
	)

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.name", "AlfredoBot")
	v.SetDefault("app.version", "0.0.1")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.log_file", "")
	v.SetDefault("app.enable_json_logs", "true")
	v.SetDefault("app.log_level", "0")

	// DB defaults
	v.SetDefault("db.host", "localhost")
	v.SetDefault("db.port", 5432)
	v.SetDefault("db.name", "postgres")
	v.SetDefault("db.sslmode", "disable")

	// Telegram defaults
	v.SetDefault("telegram.timeout", 60)
	v.SetDefault("telegram.use_webhook", false)

	// S3 defaults
	v.SetDefault("s3.endpoint", "")
	v.SetDefault("s3.region", "")
	v.SetDefault("s3.bucket", "")
	v.SetDefault("s3.use_ssl", false)
}

// bindEnv explicitly binds environment variables to config fields
func bindEnv(v *viper.Viper) error {
	var errs []error

	// Helper to collect errors
	bind := func(key, env string) {
		if err := v.BindEnv(key, env); err != nil {
			errs = append(errs, fmt.Errorf("failed to bind %s to %s: %w", env, key, err))
		}
	}

	// App config bindings
	bind("app.name", "APP_NAME")
	bind("app.version", "APP_VERSION")
	bind("app.environment", "APP_ENV")
	bind("app.log_file", "APP_LOG_FILE")
	bind("app.enable_json_logs", "APP_JSON_LOGS")
	bind("app.log_level", "APP_LOG_LEVEL")

	// DB config bindings
	bind("db.host", "DB_HOST")
	bind("db.port", "DB_PORT")
	bind("db.user", "DB_USER")
	bind("db.password", "DB_PASSWORD")
	bind("db.name", "DB_NAME")
	bind("db.sslmode", "DB_SSLMODE")

	// Telegram config bindings
	bind("telegram.token", "TELEGRAM_TOKEN")
	bind("telegram.timeout", "TELEGRAM_TIMEOUT")
	bind("telegram.webhook_url", "TELEGRAM_WEBHOOK_URL")
	bind("telegram.use_webhook", "TELEGRAM_USE_WEBHOOK")

	// S3 config bindings
	bind("s3.endpoint", "S3_ENDPOINT")
	bind("s3.region", "S3_REGION")
	bind("s3.access_key_id", "S3_ACCESS_KEY_ID")
	bind("s3.secret_access_key", "S3_SECRET_ACCESS_KEY")
	bind("s3.bucket", "S3_BUCKET")
	bind("s3.use_ssl", "S3_USE_SSL")

	if len(errs) > 0 {
		return fmt.Errorf("environment binding errors: %v", errs)
	}
	return nil
}

package configs

import (
	"fmt"
	"strconv"

	"github.com/gobuffalo/envy"
	"github.com/rs/zerolog/log"
)

func GetConfig() *Configuration {
	return getFromEnv()
}

func getFromEnv() *Configuration {
	cfg := &Configuration{}

	cfg.App = getAppConfig()
	cfg.DB = getDBConfig()
	cfg.Telegram = getTelegramConfig()
	return cfg
}

func getDBConfig() *DatabaseConfig {
	dbCfg := &DatabaseConfig{}

	dbCfg.Host = envy.Get("DB_HOST", "localhost")
	dbCfg.User = envy.Get("DB_USER", "postgres")
	dbCfg.Password = envy.Get("DB_PASSWORD", "postgres")
	dbCfg.DBName = envy.Get("DB_NAME", "cars")
	dbCfg.SSLMode = envy.Get("DB_SSLMODE", "disable")
	port, err := strconv.Atoi(envy.Get("DB_PORT", "5432"))
	if err != nil {
		log.Panic().Err(err).Msg("cannot convert DB_PORT")
	}
	dbCfg.Port = port

	dbCfg.DSN = getDbDSN(dbCfg)

	return dbCfg
}

func getDbDSN(dbConfig *DatabaseConfig) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		dbConfig.Host, dbConfig.User, dbConfig.Password,
		dbConfig.DBName, dbConfig.Port, dbConfig.SSLMode,
	)
}

func getAppConfig() *App {
	appCfg := &App{}
	
	appCfg.Name = envy.Get("APP_NAME", "TelegramBot")
	appCfg.Version = envy.Get("APP_VERSION", "1.0.0")
	appCfg.Environment = envy.Get("APP_ENV", "development")

	return appCfg
}

// getTelegramConfig loads Telegram bot configuration from environment variables
func getTelegramConfig() *TelegramConfig {
	telegramCfg := &TelegramConfig{}
	
	telegramCfg.Token = envy.Get("TELEGRAM_TOKEN", "")
	telegramCfg.Debug = envy.Get("TELEGRAM_DEBUG", "false") == "true"
	
	timeout, err := strconv.Atoi(envy.Get("TELEGRAM_TIMEOUT", "60"))
	if err != nil {
		log.Warn().Err(err).Msg("cannot convert TELEGRAM_TIMEOUT, using default value")
		timeout = 60
	}
	telegramCfg.Timeout = timeout
	
	telegramCfg.WebhookURL = envy.Get("TELEGRAM_WEBHOOK_URL", "")
	telegramCfg.UseWebhook = envy.Get("TELEGRAM_USE_WEBHOOK", "false") == "true"
	
	return telegramCfg
}


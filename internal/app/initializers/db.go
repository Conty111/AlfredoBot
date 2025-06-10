package initializers

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/rs/zerolog/log"

	"github.com/Conty111/AlfredoBot/internal/configs"
	"github.com/Conty111/AlfredoBot/internal/models"
	"github.com/Conty111/AlfredoBot/pkg/logger"
)

func InitializeDatabase(cfg *configs.Configuration) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.DB.DSN), &gorm.Config{
		Logger: logger.NewZerologGormWrapper(),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("error while connecting to database")
	}
	return db
}

func InitializeMigrations(db *gorm.DB) error {
	var telegramUsers models.TelegramUser
	
	err := db.AutoMigrate(&telegramUsers)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
		return err
	}
	return nil
}


package app

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/rs/zerolog/log"

	"github.com/Conty111/AlfredoBot/internal/app/dependencies"
	"github.com/Conty111/AlfredoBot/internal/app/initializers"
	"github.com/Conty111/AlfredoBot/internal/configs"
	"github.com/Conty111/AlfredoBot/internal/interfaces"
	"github.com/Conty111/AlfredoBot/internal/repositories"
	"github.com/Conty111/AlfredoBot/internal/services/s3"
	"github.com/Conty111/AlfredoBot/internal/services/telegram"
)

// Application is a main struct for the application that contains general information
type Application struct {
	db          *gorm.DB
	Container   *dependencies.Container
	telegramBot *telegram.TelegramBotService
}

// InitializeApplication initializes new application
func InitializeApplication(cfg *configs.Configuration) (*Application, error) {
	initializers.InitializeEnvs()

	// Initialize logging first to capture any errors
	if err := initializers.InitializeLogs(*cfg.App); err != nil {
		return nil, fmt.Errorf("failed to initialize logs: %w", err)
	}

	app, err := BuildApplication(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build application: %w", err)
	}

	// Validate required configuration
	if cfg.Telegram == nil {
		return nil, fmt.Errorf("telegram configuration is missing - check your config")
	}
	if cfg.Telegram.Token == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	err = initializers.InitializeMigrations(app.db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrations: %w", err)
	}

	// Create repositories if they don't exist
	var photoRepository interfaces.PhotoManager
	// Create S3 client if it doesn't exist
	var s3Client interfaces.S3Client
	if app.Container.S3Client == nil {
		s3Client, err = createS3Client(cfg.S3)
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 client: %w", err)
		}
	} else {
		s3Client = app.Container.S3Client
	}

	if app.Container.PhotoRepository == nil {
		photoRepository = repositories.NewPhotoRepository(app.db, s3Client)
	} else {
		photoRepository = app.Container.PhotoRepository
	}

	var articleRepository interfaces.ArticleNumberManager
	if app.Container.ArticleNumberRepository == nil {
		articleRepository = repositories.NewArticleNumberRepository(app.db)
	} else {
		articleRepository = app.Container.ArticleNumberRepository
	}

	// Initialize Telegram bot service
	telegramBot, err := telegram.NewTelegramBotService(
		cfg.Telegram,
		cfg.S3,
		app.Container.TelegramUserRepository,
		photoRepository,
		articleRepository,
		s3Client,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telegram bot: %w", err)
	}
	app.telegramBot = telegramBot

	log.Info().Msg("Application initialized successfully")
	return app, nil
}

// Start starts application services
func (a *Application) Start(ctx context.Context, cli bool) {
	if cli {
		return
	}

	log.Info().Msg("Starting application")

	// Start Telegram bot if configured
	if a.telegramBot != nil {
		if err := a.telegramBot.Start(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to start Telegram bot")
		}
	}
}

// Stop stops application services
func (a *Application) Stop() (err error) {
	log.Info().Msg("Gracefully stopping application")

	// Stop Telegram bot if it was started
	if a.telegramBot != nil {
		if err := a.telegramBot.Stop(); err != nil {
			log.Error().Err(err).Msg("Failed to stop Telegram bot")
		}
	}

	return nil
}

// createS3Client creates a new S3 client
func createS3Client(cfg *configs.S3Config) (interfaces.S3Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("S3 config is nil")
	}

	// Import the s3 package
	s3Client, err := s3.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return s3.NewS3Client(s3Client), nil
}

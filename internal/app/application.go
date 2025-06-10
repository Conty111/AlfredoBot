package app

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/rs/zerolog/log"

	"github.com/Conty111/AlfredoBot/internal/app/dependencies"
	"github.com/Conty111/AlfredoBot/internal/app/initializers"
	"github.com/Conty111/AlfredoBot/internal/services"
)

// Application is a main struct for the application that contains general information
type Application struct {
	db          *gorm.DB
	Container   *dependencies.Container
	telegramBot *services.TelegramBotService
}

func (a *Application) InitializeApplication() (*Application, any) {
	panic("unimplemented")
}

// InitializeApplication initializes new application
func InitializeApplication() (*Application, error) {
	initializers.InitializeEnvs()

	// Initialize logging first to capture any errors
	if err := initializers.InitializeLogs(); err != nil {
		return nil, fmt.Errorf("failed to initialize logs: %w", err)
	}

	app, err := BuildApplication()
	if err != nil {
		return nil, fmt.Errorf("failed to build application: %w", err)
	}

	// Validate required configuration
	if app.Container.Config == nil {
		return nil, fmt.Errorf("application configuration is missing")
	}
	if app.Container.Config.Telegram == nil {
		return nil, fmt.Errorf("telegram configuration is missing - check your .env file")
	}
	if app.Container.Config.Telegram.Token == "" {
		return nil, fmt.Errorf("telegram bot token is required - set TELEGRAM_TOKEN in .env file")
	}

	err = initializers.InitializeMigrations(app.db)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize migrations: %w", err)
	}

	// Initialize Telegram bot service
	telegramBot, err := services.NewTelegramBotService(
		app.Container.Config.Telegram,
		app.Container.TelegramUserRepository,
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

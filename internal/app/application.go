package app

import (
	"context"

	"gorm.io/gorm"

	"github.com/rs/zerolog/log"

	"github.com/Conty111/TelegramBotTemplate/internal/app/dependencies"
	"github.com/Conty111/TelegramBotTemplate/internal/app/initializers"
	"github.com/Conty111/TelegramBotTemplate/internal/services"
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

	if err := initializers.InitializeLogs(); err != nil {
		return nil, err
	}

	app, err := BuildApplication()
	if err != nil {
		return nil, err
	}

	err = initializers.InitializeMigrations(app.db)
	if err != nil {
		return nil, err
	}

	// Initialize Telegram bot service
	if app.Container.Config.Telegram != nil && app.Container.Config.Telegram.Token != "" {
		telegramBot, err := services.NewTelegramBotService(
			app.Container.Config.Telegram,
			app.Container.TelegramUserRepository,
		)
		if err != nil {
			return nil, err
		}
		app.telegramBot = telegramBot
	}

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

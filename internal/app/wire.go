//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"

	"github.com/Conty111/AlfredoBot/internal/app/dependencies"
	"github.com/Conty111/AlfredoBot/internal/app/initializers"
	"github.com/Conty111/AlfredoBot/internal/configs"
	"github.com/Conty111/AlfredoBot/internal/interfaces"
	"github.com/Conty111/AlfredoBot/internal/repositories"
)

func BuildApplication() (*Application, error) {
	wire.Build(
		initializers.InitializeBuildInfo,
		configs.GetConfig,
		initializers.InitializeDatabase,
		
		// Telegram user repository
		repositories.NewTelegramUserRepository,
		wire.Bind(new(interfaces.TelegramUserManager), new(*repositories.TelegramUserRepository)),
		
		// Container and application
		wire.Struct(new(dependencies.Container), "*"),
		wire.Struct(new(Application), "db", "Container"),
	)

	return &Application{}, nil
}


package dependencies

import (
	"github.com/Conty111/AlfredoBot/internal/app/build"
	"github.com/Conty111/AlfredoBot/internal/configs"
	"github.com/Conty111/AlfredoBot/internal/interfaces"
)

// Container is a DI container for application
type Container struct {
	BuildInfo              *build.Info
	Config                 *configs.Configuration
	TelegramUserRepository interfaces.TelegramUserManager
}

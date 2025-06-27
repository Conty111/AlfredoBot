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
	"github.com/Conty111/AlfredoBot/internal/services/s3"
)

func BuildApplication(cfg *configs.Configuration) (*Application, error) {
	wire.Build(
		initializers.InitializeBuildInfo,
		initializers.InitializeDatabase,

		// S3 Client
		provideS3Client,

		// Repositories
		repositories.NewTelegramUserRepository,
		wire.Bind(new(interfaces.TelegramUserManager), new(*repositories.TelegramUserRepository)),

		wire.Struct(new(repositories.PhotoRepository), "db", "s3Client"),
		wire.Bind(new(interfaces.PhotoManager), new(*repositories.PhotoRepository)),

		repositories.NewArticleNumberRepository,
		wire.Bind(new(interfaces.ArticleNumberManager), new(*repositories.ArticleNumberRepository)),

		// Container and application
		wire.Struct(new(dependencies.Container), "*"),
		wire.Struct(new(Application), "db", "Container"),
	)
	return &Application{}, nil
}

// provideS3Client creates a new S3 client
func provideS3Client(cfg *configs.Configuration) (interfaces.S3Client, error) {
	if cfg.S3 == nil {
		return nil, nil
	}

	// Create the S3 client
	s3Client, err := s3.NewClient(cfg.S3)
	if err != nil {
		return nil, err
	}

	// Wrap it in the S3Client interface implementation
	return s3.NewS3Client(s3Client), nil
}

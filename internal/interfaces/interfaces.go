package interfaces

import (
	"github.com/google/uuid"

	"github.com/Conty111/AlfredoBot/internal/models"
)

type TelegramUserProvider interface {
	GetByID(id uuid.UUID) (*models.TelegramUser, error)
	GetByTelegramID(telegramID int64) (*models.TelegramUser, error)
	GetByUsername(username string) (*models.TelegramUser, error)
	CreateUser(user *models.TelegramUser) error
	GetUsersByState(state string) ([]*models.TelegramUser, error)
}

type TelegramUserManager interface {
	GetByID(id uuid.UUID) (*models.TelegramUser, error)
	GetByTelegramID(telegramID int64) (*models.TelegramUser, error)
	GetByUsername(username string) (*models.TelegramUser, error)
	CreateUser(user *models.TelegramUser) error
	UpdateByID(id uuid.UUID, updates interface{}) error
	UpdateByTelegramID(telegramID int64, updates interface{}) error
	DeleteByID(id uuid.UUID) error
	DeleteByTelegramID(telegramID int64) error
	GetUsersByState(state string) ([]*models.TelegramUser, error)
}


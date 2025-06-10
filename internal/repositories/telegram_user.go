package repositories

import (
	"gorm.io/gorm"

	"github.com/google/uuid"

	"github.com/dimagic/Repositories/Alfredo/internal/errs"
	"github.com/dimagic/Repositories/Alfredo/internal/models"
)

// TelegramUserRepository handles database operations for Telegram users
type TelegramUserRepository struct {
	db *gorm.DB
}

// NewTelegramUserRepository creates a new TelegramUserRepository
func NewTelegramUserRepository(db *gorm.DB) *TelegramUserRepository {
	return &TelegramUserRepository{db: db}
}

// GetByID retrieves a Telegram user by UUID
func (r *TelegramUserRepository) GetByID(id uuid.UUID) (*models.TelegramUser, error) {
	user := &models.TelegramUser{}
	tx := r.db.Where("id = ?", id).First(user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, errs.UserNotFound
	}
	return user, nil
}

// GetByTelegramID retrieves a Telegram user by their Telegram ID
func (r *TelegramUserRepository) GetByTelegramID(telegramID int64) (*models.TelegramUser, error) {
	user := &models.TelegramUser{}
	tx := r.db.Where("telegram_id = ?", telegramID).First(user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, errs.UserNotFound
	}
	return user, nil
}

// GetByUsername retrieves a Telegram user by their username
func (r *TelegramUserRepository) GetByUsername(username string) (*models.TelegramUser, error) {
	user := &models.TelegramUser{}
	tx := r.db.Where("username = ?", username).First(user)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, errs.UserNotFound
	}
	return user, nil
}

// CreateUser creates a new Telegram user
func (r *TelegramUserRepository) CreateUser(user *models.TelegramUser) error {
	return r.db.Create(user).Error
}

// UpdateByID updates a Telegram user by ID
func (r *TelegramUserRepository) UpdateByID(id uuid.UUID, updates interface{}) error {
	return r.db.
		Model(&models.TelegramUser{}).
		Where("id = ?", id).
		Updates(updates).
		Error
}

// UpdateByTelegramID updates a Telegram user by Telegram ID
func (r *TelegramUserRepository) UpdateByTelegramID(telegramID int64, updates interface{}) error {
	return r.db.
		Model(&models.TelegramUser{}).
		Where("telegram_id = ?", telegramID).
		Updates(updates).
		Error
}

// DeleteByID deletes a Telegram user by ID
func (r *TelegramUserRepository) DeleteByID(id uuid.UUID) error {
	user := models.TelegramUser{}
	user.ID = id

	return r.db.
		Model(&models.TelegramUser{}).
		Delete(&user).
		Error
}

// DeleteByTelegramID deletes a Telegram user by Telegram ID
func (r *TelegramUserRepository) DeleteByTelegramID(telegramID int64) error {
	return r.db.
		Where("telegram_id = ?", telegramID).
		Delete(&models.TelegramUser{}).
		Error
}

// GetUsersByState retrieves all users with a specific state
func (r *TelegramUserRepository) GetUsersByState(state string) ([]*models.TelegramUser, error) {
	var users []*models.TelegramUser
	tx := r.db.Where("state = ?", state).Find(&users)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return users, nil
}

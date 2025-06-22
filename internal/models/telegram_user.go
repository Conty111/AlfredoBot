package models

import (
	"gorm.io/gorm"

	"github.com/google/uuid"
)

// TelegramUser represents a Telegram user in the database
type TelegramUser struct {
	BaseModel
	TelegramID   int64  `gorm:"column:telegram_id;uniqueIndex"`
	Username     string `gorm:"column:username"`
	FirstName    string `gorm:"column:first_name"`
	LastName     string `gorm:"column:last_name"`
	LanguageCode string `gorm:"column:language_code"`
	IsBot        bool   `gorm:"column:is_bot"`
	Photos	   []Photo `gorm:"foreignKey:UserID"`
	State string `gorm:"column:state"`
}

func (u *TelegramUser) BeforeCreate(_ *gorm.DB) (err error) {
	u.ID = uuid.New()
	return nil
}

// TelegramUserFilter for filtering telegram users
type TelegramUserFilter struct {
	TelegramID   int64
	Username     string
	State string
}

const (
	TelegramUserStateUploading = "uploading"
	TelegramUserStateSearching = "searching"
	TelegramUserStateDefault	= "default"
)
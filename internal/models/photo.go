package models

import (
	"gorm.io/gorm"

	"github.com/google/uuid"
)

// Photo represents an Photo in the database
type Photo struct {
	BaseModel
	S3Key	   uuid.UUID	   `gorm:"column:s3_key"`
	ArticleNumbers []ArticleNumber `gorm:"many2many:article_number_photos;"`
	TelegramUser   TelegramUser   `gorm:"foreignKey:UserID"`
	UserID     uuid.UUID     `gorm:"column:user_id"`
	State string `gorm:"column:state"`
}

func (i *Photo) BeforeCreate(tx *gorm.DB) (err error) {
	i.ID = uuid.New()
	return nil
}

const (
	PhotoNotApplied = "not_applied"
	PhotoApplied = "applied"
)
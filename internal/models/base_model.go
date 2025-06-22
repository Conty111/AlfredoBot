package models

import (
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

type BaseModel struct {
	ID        uuid.UUID `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

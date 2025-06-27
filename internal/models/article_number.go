package models

import (
	"gorm.io/gorm"

	"github.com/google/uuid"
)

// ArticleNumber represents an article number in the database
type ArticleNumber struct {
	BaseModel
	Number string  `gorm:"column:number;uniqueIndex"`
	Photos []Photo `gorm:"many2many:article_number_photos;"`
}

func (a *ArticleNumber) BeforeCreate(tx *gorm.DB) (err error) {
	a.ID = uuid.New()
	return nil
}

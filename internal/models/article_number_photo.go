package models

import "github.com/google/uuid"

// ArticleNumberPhoto is the join table for ArticleNumber and Photo
type ArticleNumberPhoto struct {
	ArticleNumberID uuid.UUID `gorm:"primaryKey"`
	PhotoID         uuid.UUID `gorm:"primaryKey"`
	
	ArticleNumber ArticleNumber `gorm:"foreignKey:ArticleNumberID"`
	Photo         Photo         `gorm:"foreignKey:PhotoID"`
}
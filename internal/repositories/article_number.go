package repositories

import (
	"gorm.io/gorm"

	"github.com/google/uuid"

	"github.com/Conty111/AlfredoBot/internal/models"
)

// ArticleNumberRepository handles database operations for ArticleNumbers
type ArticleNumberRepository struct {
	db *gorm.DB
}

// NewArticleNumberRepository creates a new ArticleNumberRepository
func NewArticleNumberRepository(db *gorm.DB) *ArticleNumberRepository {
	return &ArticleNumberRepository{db: db}
}

// GetByID retrieves an ArticleNumber by UUID
func (r *ArticleNumberRepository) GetByID(id uuid.UUID) (*models.ArticleNumber, error) {
	articleNumber := &models.ArticleNumber{}
	tx := r.db.Where("id = ?", id).First(articleNumber)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return articleNumber, nil
}

// GetByNumber retrieves an ArticleNumber by its number string
func (r *ArticleNumberRepository) GetByNumber(number string) (*models.ArticleNumber, error) {
	articleNumber := &models.ArticleNumber{}
	tx := r.db.Where("number = ?", number).First(articleNumber)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return articleNumber, nil
}

// GetArticleNumbersByPhoto retrieves all ArticleNumbers associated with a photo
func (r *ArticleNumberRepository) GetArticleNumbersByPhoto(photoID uuid.UUID) ([]*models.ArticleNumber, error) {
	var articleNumbers []*models.ArticleNumber
	tx := r.db.Joins("JOIN photo_article_number ON photo_article_number.article_number_id = article_numbers.id").
		Where("photo_article_number.photo_id = ?", photoID).
		Find(&articleNumbers)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return articleNumbers, nil
}

// CreateArticleNumber creates a new ArticleNumber
func (r *ArticleNumberRepository) CreateArticleNumber(articleNumber *models.ArticleNumber) error {
	return r.db.Create(articleNumber).Error
}

// UpdateArticleNumber updates an ArticleNumber
func (r *ArticleNumberRepository) UpdateArticleNumber(articleNumber *models.ArticleNumber) error {
	return r.db.Save(articleNumber).Error
}

// DeleteArticleNumber deletes an ArticleNumber
func (r *ArticleNumberRepository) DeleteArticleNumber(id uuid.UUID) error {
	return r.db.Delete(&models.ArticleNumber{}, id).Error
}

// GetOrCreateArticleNumber gets an existing article number by number string or creates a new one
func (r *ArticleNumberRepository) GetOrCreateArticleNumber(number string) (*models.ArticleNumber, error) {
	articleNumber := &models.ArticleNumber{}
	
	// Try to find existing article number
	tx := r.db.Where("number = ?", number).First(articleNumber)
	if tx.Error == nil {
		return articleNumber, nil
	}
	
	// If not found, create a new one
	if tx.Error == gorm.ErrRecordNotFound {
		articleNumber = &models.ArticleNumber{
			Number: number,
		}
		if err := r.db.Create(articleNumber).Error; err != nil {
			return nil, err
		}
		return articleNumber, nil
	}
	
	// Return any other error
	return nil, tx.Error
}

// GetArticleNumberWithPhotos retrieves an article number with its associated photos
func (r *ArticleNumberRepository) GetArticleNumberWithPhotos(articleNumberID uuid.UUID) (*models.ArticleNumber, error) {
	articleNumber := &models.ArticleNumber{}
	tx := r.db.
		Preload("Photos.ArticleNumbers").
		Where("id = ?", articleNumberID).
		First(articleNumber)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return articleNumber, nil
}

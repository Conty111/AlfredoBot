package repositories

import (
	"context"
	"fmt"
	"io"

	"gorm.io/gorm"

	"github.com/google/uuid"

	"github.com/Conty111/AlfredoBot/internal/interfaces"
	"github.com/Conty111/AlfredoBot/internal/models"
)

// PhotoRepository handles database operations for Photos
type PhotoRepository struct {
	DB       *gorm.DB
	S3Client interfaces.S3Client
}

// NewPhotoRepository creates a new PhotoRepository
func NewPhotoRepository(db *gorm.DB, s3Client interfaces.S3Client) *PhotoRepository {
	return &PhotoRepository{
		DB:       db,
		S3Client: s3Client,
	}
}

// GetByID retrieves a Photo by UUID
func (r *PhotoRepository) GetByID(id uuid.UUID) (*models.Photo, error) {
	photo := &models.Photo{}
	tx := r.DB.Preload("ArticleNumbers").Where("id = ?", id).First(photo)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return photo, nil
}

// GetUsersPhotosByState retrieves photos for a user filtered by state
func (r *PhotoRepository) GetUsersPhotosByState(
	id uuid.UUID,
	state string,
) ([]*models.Photo, error) {
	var photos []*models.Photo

	err := r.DB.
		Preload("ArticleNumbers").
		Joins("JOIN telegram_users ON telegram_users.id = photos.user_id").
		Where("telegram_users.id = ? AND photos.state = ?", id, state).
		Find(&photos).
		Error

	if err != nil {
		return nil, err
	}
	return photos, nil
}

// GetPhotosByArticleNumber retrieves all Photos associated with an article number
func (r *PhotoRepository) GetPhotosByArticleNumber(articleNumberID uuid.UUID) ([]*models.Photo, error) {
	var photos []*models.Photo
	tx := r.DB.Preload("ArticleNumbers").
		Joins("JOIN photo_article_number ON photo_article_number.photo_id = photo.id").
		Where("photo_article_number.article_number_id = ?", articleNumberID).
		Find(&photos)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return photos, nil
}

// CreatePhoto creates a new Photo
func (r *PhotoRepository) CreatePhoto(photo *models.Photo) error {
	return r.DB.Create(photo).Error
}

// UpdatePhoto updates a Photo
func (r *PhotoRepository) UpdatePhoto(photo *models.Photo) error {
	return r.DB.Save(photo).Error
}

func (r *PhotoRepository) DeletePhoto(
	id uuid.UUID,
	bucket string) error {

	photo, err := r.GetByID(id)
	if err != nil {
		return err
	}

	if bucket != "" {
		s3ObjectKey := photo.UserID.String() + "/" + photo.S3Key.String() + ".jpg"
		if err := r.S3Client.DeleteFile(context.Background(), bucket, s3ObjectKey); err != nil {
			return fmt.Errorf("failed to delete from S3: %w", err)
		}
	}

	return r.DB.Delete(&models.Photo{}, id).Error
}

// AddArticleNumberToPhoto associates an article number with a photo
func (r *PhotoRepository) AddArticleNumberToPhoto(photoID, articleNumberID uuid.UUID) error {
	// Check if article number exists
	var count int64
	if err := r.DB.Model(&models.ArticleNumber{}).
		Where("id = ?", articleNumberID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("article number not found")
	}

	relation := models.ArticleNumberPhoto{
		PhotoID:         photoID,
		ArticleNumberID: articleNumberID,
	}
	return r.DB.Create(&relation).Error
}

// RemoveArticleNumberFromPhoto removes an association between an article number and a photo
func (r *PhotoRepository) RemoveArticleNumberFromPhoto(photoID, articleNumberID uuid.UUID) error {
	return r.DB.Where("photo_id = ? AND article_number_id = ?", photoID, articleNumberID).
		Delete(&models.ArticleNumberPhoto{}).Error
}

// UploadPhotoToS3 uploads a photo to S3 storage and associates it with article numbers
func (r *PhotoRepository) UploadPhotoToS3(
	ctx context.Context,
	userID uuid.UUID,
	s3Key uuid.UUID,
	bucket string,
	photoData io.Reader) error {
	// Generate S3 object key
	s3ObjectKey := userID.String() + "/" + s3Key.String() + ".jpg"

	// Upload the file to S3
	if err := r.S3Client.UploadFile(ctx, bucket, s3ObjectKey, photoData); err != nil {
		return err
	}

	return nil
}

// GetPhotoFromS3 downloads a photo from S3 storage
func (r *PhotoRepository) GetPhotoFromS3(ctx context.Context, userID uuid.UUID, s3Key uuid.UUID, bucket string) (io.ReadCloser, error) {
	s3ObjectKey := userID.String() + "/" + s3Key.String() + ".jpg"
	return r.S3Client.DownloadFile(ctx, bucket, s3ObjectKey)
}

// GetPhotoURL generates a URL for a photo in S3
func (r *PhotoRepository) GetPhotoURL(ctx context.Context, userID uuid.UUID, s3Key uuid.UUID, bucket string, endpoint string) string {
	s3ObjectKey := userID.String() + "/" + s3Key.String() + ".jpg"
	return endpoint + "/" + bucket + "/" + s3ObjectKey
}

// GetPhotoWithArticleNumbers retrieves a photo with its associated article numbers
func (r *PhotoRepository) GetPhotoWithArticleNumbers(photoID uuid.UUID) (*models.Photo, error) {
	photo := &models.Photo{}
	tx := r.DB.Preload("ArticleNumbers").Where("id = ?", photoID).First(photo)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return photo, nil
}

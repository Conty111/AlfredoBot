package interfaces

import (
	"context"
	"io"

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

type PhotoProvider interface {
	GetByID(id uuid.UUID) (*models.Photo, error)
	GetPhotosByArticleNumber(articleNumberID uuid.UUID) ([]*models.Photo, error)
	GetPhotoWithArticleNumbers(photoID uuid.UUID) (*models.Photo, error)
}

type PhotoManager interface {
	PhotoProvider
	CreatePhoto(photo *models.Photo) error
	GetUsersPhotosByState(userID uuid.UUID, state string) ([]*models.Photo, error)
	UpdatePhoto(photo *models.Photo) error
	DeletePhoto(id uuid.UUID, bucket string) error
	AddArticleNumberToPhoto(photoID, articleNumberID uuid.UUID) error
	RemoveArticleNumberFromPhoto(photoID, articleNumberID uuid.UUID) error
	UploadPhotoToS3(ctx context.Context, userID uuid.UUID, s3Key uuid.UUID, bucket string, photoData io.Reader) error
	GetPhotoFromS3(ctx context.Context, userID uuid.UUID, s3Key uuid.UUID, bucket string) (io.ReadCloser, error)
	GetPhotoURL(ctx context.Context, userID uuid.UUID, s3Key uuid.UUID, bucket string, endpoint string) string
}

type ArticleNumberProvider interface {
	GetByID(id uuid.UUID) (*models.ArticleNumber, error)
	GetByNumber(number string) (*models.ArticleNumber, error)
	GetArticleNumbersByPhoto(photoID uuid.UUID) ([]*models.ArticleNumber, error)
	GetArticleNumberWithPhotos(articleNumberID uuid.UUID) (*models.ArticleNumber, error)
}

type ArticleNumberManager interface {
	ArticleNumberProvider
	CreateArticleNumber(articleNumber *models.ArticleNumber) error
	UpdateArticleNumber(articleNumber *models.ArticleNumber) error
	DeleteArticleNumber(id uuid.UUID) error
	GetOrCreateArticleNumber(number string) (*models.ArticleNumber, error)
}

type S3Client interface {
	UploadFile(ctx context.Context, bucket, key string, file io.Reader) error
	DownloadFile(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, bucket, key string) error
	GeneratePresignedURL(ctx context.Context, bucket, key string, expiresIn int64) (string, error)
}

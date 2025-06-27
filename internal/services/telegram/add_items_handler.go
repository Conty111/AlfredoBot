package telegram

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"gorm.io/gorm"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/Conty111/AlfredoBot/internal/models"
)

func (s *TelegramBotService) addItemHandler(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Пожалуйста, отправьте все фото товара, а затем отдельным сообщением или с подписью артикулы в формате: articul1, articul2, ...\n\nПример: 1.2345, 6.7890",
		ReplyMarkup: cancelMenu,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send message")
	}
	err = s.userRepository.UpdateByTelegramID(
		update.Message.From.ID,
		map[string]interface{}{"state": models.TelegramUserStateUploading},
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update user state")
	}
}

func (s *TelegramBotService) photoMessageHandler(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
	user, err := s.userRepository.GetByTelegramID(update.Message.From.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Произошла ошибка. Пожалуйста, попробуйте снова.",
			ReplyMarkup: mainMenu,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to send message")
		}
		return
	}

	var file *tgmodels.File
	if len(update.Message.Photo) > 0 {
		photo := update.Message.Photo[len(update.Message.Photo)-1]
		file, err = b.GetFile(ctx, &bot.GetFileParams{
			FileID: photo.FileID,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get file from Telegram")
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        "Не удалось получить файл из Telegram. Пожалуйста, попробуйте снова.",
				ReplyMarkup: mainMenu,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
			return
		}
	} else if update.Message.Document != nil && update.Message.Document.FileID != "" {
		// Handle document attachments (like image files)
		file, err = b.GetFile(ctx, &bot.GetFileParams{
			FileID: update.Message.Document.FileID,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to get document file from Telegram")
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        "Не удалось получить файл из Telegram. Пожалуйста, попробуйте снова.",
				ReplyMarkup: mainMenu,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
			return
		}
	}
	if file != nil {
		fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", s.config.Token, file.FilePath)
		resp, err := http.Get(fileURL)
		if err != nil {
			log.Error().Err(err).Msg("Failed to download file from Telegram")
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        "Не удалось загрузить файл из Telegram. Пожалуйста, попробуйте снова.",
				ReplyMarkup: mainMenu,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
			return
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Error().Err(err).Msg("Failed to close response body")
			}
		}()

		// Read photo data to generate hash
		photoData, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read photo data for hashing")
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        "Не удалось обработать фото. Пожалуйста, попробуйте снова.",
				ReplyMarkup: mainMenu,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
			return
		}

		resp.Body = io.NopCloser(bytes.NewReader(photoData))

		photoModel := &models.Photo{
			UserID: user.ID,
			State:  models.PhotoNotApplied,
		}

		s3Key := uuid.New()
		photoModel.S3Key = s3Key

		err = s.photoRepository.CreatePhoto(photoModel)
		if err != nil {
			log.Error().Err(err).Msg("Failed to save photo to database")
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        "Не удалось сохранить фото в базе данных. Пожалуйста, попробуйте снова.",
				ReplyMarkup: mainMenu,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
			return
		}

		if err := s.photoRepository.UploadPhotoToS3(
			ctx,
			user.ID,
			s3Key,
			s.s3Config.Bucket,
			resp.Body,
		); err != nil {
			log.Error().Err(err).Msg("Failed to upload file to S3")
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        "Не удалось загрузить фото в хранилище. Пожалуйста, попробуйте снова.",
				ReplyMarkup: mainMenu,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
			return
		}
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Фото успешно сохранено!",
			ReplyMarkup: cancelMenu,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to send message")
		}
	}
	if update.Message.Text != "" || update.Message.Caption != "" {
		var articleNumbers []string
		if update.Message.Text == cancelText || update.Message.Caption == cancelText {
			s.cancelAddPhotos(ctx, user.ID, update, b)
			return
		}
		if update.Message.Text != "" {
			articleNumbers = parseArticleNumbers(update.Message.Text)
		} else {
			articleNumbers = parseArticleNumbers(update.Message.Caption)
		}
		if len(articleNumbers) == 0 {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        "Не удалось найти артикулы в сообщении. Пожалуйста, попробуйте снова.",
				ReplyMarkup: mainMenu,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
			return
		}
		s.applyPhotos(ctx, articleNumbers, user.ID, update, b)
		return
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Отправьте еще фото или текст с артикулами. Или фото с подписью",
		ReplyMarkup: cancelMenu,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send message")
	}
}

func (s *TelegramBotService) applyPhotos(
	ctx context.Context,
	articleNumbers []string,
	userID uuid.UUID,
	update *tgmodels.Update,
	b *bot.Bot,
) {
	successfulApplies := 0

	var articleNumberModels []models.ArticleNumber
	for _, articleNumberStr := range articleNumbers {
		articleNumberModel, err := s.articleRepository.GetOrCreateArticleNumber(articleNumberStr)
		if err != nil {
			log.Error().
				Err(err).
				Str("article_number", articleNumberStr).
				Msg("Failed to get or create article number")

			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        fmt.Sprintf("Ошибка обработки артикула '%s'. Пожалуйста, попробуйте снова.", articleNumberStr),
				ReplyMarkup: mainMenu,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
			return
		}
		articleNumberModels = append(articleNumberModels, *articleNumberModel)
	}

	// Process all photos in state NotApplied
	photos, err := s.photoRepository.GetUsersPhotosByState(userID, models.PhotoNotApplied)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get photos from database")
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Не удалось загрузить фото. Пожалуйста, попробуйте снова.",
			ReplyMarkup: mainMenu,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to send message")
		}
		return
	}

	for _, photo := range photos {
		photo.State = models.PhotoApplied
		err := s.photoRepository.UpdatePhoto(photo)
		if err != nil {
			log.Error().Err(err).Msg("Failed to update photo in database")
			continue
		}
		for _, articleNumber := range articleNumberModels {
			err := s.photoRepository.AddArticleNumberToPhoto(photo.ID, articleNumber.ID)
			if err != nil {
				log.Error().Err(err).Msg("Failed to add article number to photo")
				continue
			}
		}
		successfulApplies++
	}

	if successfulApplies != len(photos) {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        fmt.Sprintf("Не удалось загрузить %d фото. Пожалуйста, попробуйте снова.", len(photos)-successfulApplies),
			ReplyMarkup: mainMenu,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to send message")
		}
	}
	err = s.userRepository.UpdateByTelegramID(update.Message.From.ID, map[string]interface{}{
		"state": models.TelegramUserStateDefault,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to update user")
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        fmt.Sprintf("Успешно загружено %d фото!", successfulApplies),
		ReplyMarkup: mainMenu,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send message")
	}
}

func (s *TelegramBotService) cancelAddPhotos(
	ctx context.Context,
	userID uuid.UUID,
	update *tgmodels.Update,
	b *bot.Bot,
) {

	photos, err := s.photoRepository.GetUsersPhotosByState(userID, models.PhotoNotApplied)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error().Err(err).Msg("Failed to get photos for cleanup")
		return
	} else {
		for _, photo := range photos {
			if err := s.photoRepository.DeletePhoto(photo.ID, s.s3Config.Bucket); err != nil {
				log.Error().Err(err).Str("photo_id", photo.ID.String()).Msg("Failed to delete photo")
			}
		}
	}

	if err := s.userRepository.UpdateByTelegramID(update.Message.From.ID, map[string]interface{}{
		"state": models.TelegramUserStateDefault,
	}); err != nil {
		log.Error().Err(err).Msg("Failed to reset user state")
		return
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Добавление фото отменено",
		ReplyMarkup: mainMenu,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send cancellation message")
	}
}

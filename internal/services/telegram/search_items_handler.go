package telegram

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	appmodels "github.com/Conty111/AlfredoBot/internal/models"
)

func (s *TelegramBotService) searchByArticleNumberHandler(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
	err := s.userRepository.UpdateByTelegramID(update.Message.From.ID, map[string]interface{}{
		"state": appmodels.TelegramUserStateSearching,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to update user state")
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Пожалуйста, введите артикул товара для поиска:",
		ReplyMarkup: cancelMenu,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send message")
		return
	}
}

func (s *TelegramBotService) handleArticleNumberSearch(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {

	user, err := s.userRepository.GetByTelegramID(update.Message.From.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		return
	}

	if update.Message.Text == "" {
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Пожалуйста, введите артикулы через запятую",
			ReplyMarkup: cancelMenu,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to send message")
		}
		return
	}
	if update.Message.Text == cancelText {
		s.cancelSearchPhotos(ctx, update, b)
		return
	}

	articleNumbers := parseArticleNumbers(update.Message.Text)
	totalPhotos := map[uuid.UUID]string{}

	for _, article := range articleNumbers {
		articleNumber, err := s.articleRepository.GetByNumber(article)
		if err != nil {
			log.Debug().
				Err(err).
				Str("article_number", article).
				Msg("Article number not found")
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        fmt.Sprintf("Артикул '%s' не найден в базе данных.", article),
				ReplyMarkup: mainMenu,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
			continue
		}

		// Get photos associated with this article number
		articleNumberWithPhotos, err := s.articleRepository.GetArticleNumberWithPhotos(articleNumber.ID)
		if err != nil {
			log.Error().
				Err(err).
				Str("article_number_id", articleNumber.ID.String()).
				Msg("Failed to get photos for article number")

			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        "Произошла ошибка при поиске фотографий для артикула " + articleNumber.Number,
				ReplyMarkup: mainMenu,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
			return
		}

		if len(articleNumberWithPhotos.Photos) == 0 {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:      update.Message.Chat.ID,
				Text:        fmt.Sprintf("Для артикула '%s' не найдено фотографий.", articleNumberWithPhotos.Number),
				ReplyMarkup: mainMenu,
			})
			log.Debug().
				Str("article_number", articleNumberWithPhotos.Number).
				Str("id", articleNumber.ID.String()).
				Msg("No photos found for article number")
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
		} else {
			// Send each photo
			for _, photo := range articleNumberWithPhotos.Photos {
				if photo.State != appmodels.PhotoApplied {
					continue
				}
				photoFull, err := s.photoRepository.GetByID(photo.ID)
				if err != nil {
					log.Error().
						Err(err).
						Str("photo_id", photo.ID.String()).
						Msg("Failed to get photo")
					continue
				}
				photoFull.State = appmodels.PhotoApplied
				if len(photo.ArticleNumbers) == 0 {
					log.Warn().
						Str("photo_id", photo.ID.String()).
						Msg("Photo has no associated article numbers - skipping")
					continue
				}

				articlesStr := ""
				for _, article := range photo.ArticleNumbers[:len(photo.ArticleNumbers)-1] {
					articlesStr += article.Number + ", "
				}
				articlesStr += photo.ArticleNumbers[len(photo.ArticleNumbers)-1].Number
				totalPhotos[photo.S3Key] = articlesStr
			}
		}
	}

	for s3Key, articlesStr := range totalPhotos {
		fileReader, err := s.photoRepository.GetPhotoFromS3(ctx, user.ID, s3Key, s.s3Config.Bucket)
		if err != nil {
			log.Error().
				Err(err).
				Str("s3_key", s3Key.String()).
				Msg("Failed to download file from S3")
			continue
		}
		defer func() {
			if err := fileReader.Close(); err != nil {
				log.Error().Err(err).Msg("Failed to close file reader")
			}
		}()

		// Send photo with articles as caption
		_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID: update.Message.Chat.ID,
			Photo: &tgmodels.InputFileUpload{
				Data:     fileReader,
				Filename: s3Key.String() + ".jpg",
			},
			Caption: articlesStr,
		})
		if err != nil {
			log.Error().
				Err(err).
				Str("s3_key", s3Key.String()).
				Msg("Failed to send photo")
		}
	}

	// Reset user state
	if err := s.userRepository.UpdateByTelegramID(update.Message.From.ID, map[string]interface{}{
		"state": appmodels.TelegramUserStateDefault,
	}); err != nil {
		log.Error().Err(err).Msg("Failed to reset user state")
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Поиск завершен!",
		ReplyMarkup: mainMenu,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send message with photos")
	}
}

func (s *TelegramBotService) cancelSearchPhotos(
	ctx context.Context,
	update *tgmodels.Update,
	b *bot.Bot,
) {
	// Reset user state
	if err := s.userRepository.UpdateByTelegramID(update.Message.From.ID, map[string]interface{}{
		"state": appmodels.TelegramUserStateDefault,
	}); err != nil {
		log.Error().Err(err).Msg("Failed to reset user state")
	}

	// Send confirmation message
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Поиск отменен",
		ReplyMarkup: mainMenu,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send cancellation message")
	}
}

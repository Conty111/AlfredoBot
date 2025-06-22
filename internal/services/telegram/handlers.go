package telegram

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/Conty111/AlfredoBot/internal/models"
	appmodels "github.com/Conty111/AlfredoBot/internal/models"
)

var searchByArticleNumberText = "Поиск по артикулу 🔎"
var addItemText = "Добавить товар ®️"
var helpText = "Help ❓"
var supportText = "Support 🆘"

var mainMenu = &tgmodels.ReplyKeyboardMarkup{
	Keyboard: [][]tgmodels.KeyboardButton{
		{
			{Text: searchByArticleNumberText},
			{Text: addItemText},
		},
		{
			{Text: helpText},
			{Text: supportText},
		},
	},
}


func (s *TelegramBotService) saveUserMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
		if err := s.SaveUser(ctx, update.Message.From); err != nil {
			log.Error().Err(err).Msg("failed to save user")
		}
		next(ctx, b, update)
	}
}

func (s *TelegramBotService) routerMiddleware(next bot.HandlerFunc) bot.HandlerFunc {
	
	return func(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
		if update.Message == nil {
			next(ctx, b, update)
			return
		}
		
		user, err := s.userRepository.GetByTelegramID(update.Message.From.ID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get user")
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Произошла ошибка. Пожалуйста, попробуйте снова.",
				ReplyMarkup: mainMenu,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
			next(ctx, b, update)
			return
		}
		if user.State == appmodels.TelegramUserStateUploading {
			s.photoMessageHandler(ctx, b, update)
			return
		}
		if user.State == appmodels.TelegramUserStateSearching {
			s.handleArticleNumberSearch(ctx, b, update)
			return
		}
		next(ctx, b, update)
	}
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Привет, " + update.Message.From.FirstName + "! 👋",
		ReplyMarkup: mainMenu,
	})
}

func helpHandler(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: `Этот бот помогает искать товары по артикулу.

Доступные команды:
- Поиск по артикулу 🔎 - найти товар по его артикулу
- Help ❓ - показать это сообщение
- Support 🆘 - связаться с поддержкой`,
		ReplyMarkup: mainMenu,
	})
}

func supportHandler(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: `Свяжитесь с нашей поддержкой:

Email: support@example.com
Телефон: +7 (123) 456-78-90
Telegram: @support_bot`,
		ReplyMarkup: mainMenu,
	})
}

func (s *TelegramBotService) searchByArticleNumberHandler(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
	// Update user state to indicate they're in search mode
	err := s.userRepository.UpdateByTelegramID(update.Message.From.ID, map[string]interface{}{
		"state": appmodels.TelegramUserStateSearching,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to update user state")
	}

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Пожалуйста, введите артикул товара для поиска:",
		ReplyMarkup: &tgmodels.ReplyKeyboardMarkup{
			Keyboard: [][]tgmodels.KeyboardButton{
				{
					{Text: "Отмена"},
				},
			},
			ResizeKeyboard: true,
		},
	})
}

// addItemHandler handles the "Add item" button
func (s *TelegramBotService) addItemHandler(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
	// Send message asking for photo with article numbers
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Пожалуйста, отправьте все фото товара, а затем отдельным сообщением или с подписью артикулы в формате: articul1, articul2, ...\n\nПример: 1.2345, 6.7890",
		ReplyMarkup: &tgmodels.ReplyKeyboardMarkup{
			Keyboard: [][]tgmodels.KeyboardButton{
				{
					{Text: "Отмена"},
				},
			},
			ResizeKeyboard: true,
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send message")
	}
	err = s.userRepository.UpdateByTelegramID(
		update.Message.From.ID, 
		map[string]interface{}{"state": appmodels.TelegramUserStateUploading},
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
			ChatID: update.Message.Chat.ID,
			Text:   "Произошла ошибка. Пожалуйста, попробуйте снова.",
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
				ChatID: update.Message.Chat.ID,
				Text:   "Не удалось получить файл из Telegram. Пожалуйста, попробуйте снова.",
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
				ChatID: update.Message.Chat.ID,
				Text:   "Не удалось получить файл из Telegram. Пожалуйста, попробуйте снова.",
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
				ChatID: update.Message.Chat.ID,
				Text:   "Не удалось загрузить файл из Telegram. Пожалуйста, попробуйте снова.",
				ReplyMarkup: mainMenu,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
			return
		}
		defer resp.Body.Close()

		// Read photo data to generate hash
		photoData, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read photo data for hashing")
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Не удалось обработать фото. Пожалуйста, попробуйте снова.",
				ReplyMarkup: mainMenu,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
			return
		}

		resp.Body = io.NopCloser(bytes.NewReader(photoData))

		photoModel := &appmodels.Photo{
			UserID: user.ID,
			State: appmodels.PhotoNotApplied,
		}

		s3Key := uuid.New()
		photoModel.S3Key = s3Key

		err = s.photoRepository.CreatePhoto(photoModel)
		if err != nil {
			log.Error().Err(err).Msg("Failed to save photo to database")
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Не удалось сохранить фото в базе данных. Пожалуйста, попробуйте снова.",
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
				ChatID: update.Message.Chat.ID,
				Text:   "Не удалось загрузить фото в хранилище. Пожалуйста, попробуйте снова.",
				ReplyMarkup: mainMenu,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
			return
		}
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Фото успешно сохранено. Пожалуйста, отправьте еще фото или текст с артикулами.",
			ReplyMarkup: &tgmodels.ReplyKeyboardMarkup{
				Keyboard: [][]tgmodels.KeyboardButton{
					{
						{Text: "Отмена"},
					},
				},
				ResizeKeyboard: true,
			},
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to send message")
		}
		return
	}
	if update.Message.Text != "" || update.Message.Caption != "" {
		var articleNumbers []string
		if update.Message.Text != "" {
			articleNumbers = parseArticleNumbers(update.Message.Text)
		} else {
			articleNumbers = parseArticleNumbers(update.Message.Caption)
		}
		s.applyPhotos(ctx, articleNumbers, user.ID, update, b)
		return
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Отправьте еще фото или текст с артикулами. Или фото с подписью",
		ReplyMarkup: mainMenu,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send message")
	}
	return
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
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("Ошибка обработки артикула '%s'. Пожалуйста, попробуйте снова.", articleNumberStr),
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
	photos, err := s.photoRepository.GetUsersPhotosByState(userID, appmodels.PhotoNotApplied)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get photos from database")
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Не удалось загрузить фото. Пожалуйста, попробуйте снова.",
			ReplyMarkup: mainMenu,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to send message")
		}
		return
	}

	for _, photo := range photos {
		photo.State = appmodels.PhotoApplied
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
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Не удалось загрузить %d фото. Пожалуйста, попробуйте снова.", len(photos)-successfulApplies),
			ReplyMarkup: mainMenu,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to send message")
		}
	}
	err = s.userRepository.UpdateByTelegramID(update.Message.From.ID, map[string]interface{}{
		"state": appmodels.TelegramUserStateDefault,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to update user")
	}
	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Успешно загружено %d фото!", successfulApplies),
		ReplyMarkup: mainMenu,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send message")
	}
}

// parseArticleNumbers extracts article numbers from a caption string
func parseArticleNumbers(caption string) []string {
	// Split by comma and trim whitespace
	parts := strings.Split(caption, ",")
	var articleNumbers []string
	
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			articleNumbers = append(articleNumbers, trimmed)
		}
	}
	
	return articleNumbers
}

// handleArticleNumberSearch processes article number search requests
func (s *TelegramBotService) handleArticleNumberSearch(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {

	user, err := s.userRepository.GetByTelegramID(update.Message.From.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
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
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("Артикул '%s' не найден в базе данных.", article),
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
				ChatID: update.Message.Chat.ID,
				Text:   "Произошла ошибка при поиске фотографий для артикула " + articleNumber.Number,
				ReplyMarkup: mainMenu,
			})
			if err != nil {
				log.Error().Err(err).Msg("Failed to send message")
			}
			return
		}

		if len(articleNumberWithPhotos.Photos) == 0 {
			_, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   fmt.Sprintf("Для артикула '%s' не найдено фотографий.", articleNumberWithPhotos.Number),
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
		defer fileReader.Close()

		// Send photo with articles as caption
		_, err = b.SendPhoto(ctx, &bot.SendPhotoParams{
			ChatID: update.Message.Chat.ID,
			Photo:  &tgmodels.InputFileUpload{
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
		ChatID: update.Message.Chat.ID,
		Text:   "Поиск завершен!",
		ReplyMarkup: mainMenu,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send message with photos")
	}
}


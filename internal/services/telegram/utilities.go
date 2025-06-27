package telegram

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"

	"github.com/Conty111/AlfredoBot/internal/models"
)

func (s *TelegramBotService) SaveUser(ctx context.Context, tgUser *tgmodels.User) error {
	if tgUser == nil {
		return fmt.Errorf("user cannot be nil")
	}

	user := &models.TelegramUser{
		TelegramID:   tgUser.ID,
		Username:     tgUser.Username,
		FirstName:    tgUser.FirstName,
		LastName:     tgUser.LastName,
		LanguageCode: tgUser.LanguageCode,
		IsBot:        tgUser.IsBot,
	}

	existingUser, err := s.userRepository.GetByTelegramID(tgUser.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error().
			Err(err).
			Int64("user_id", tgUser.ID).
			Msg("Failed to check if Telegram user exists")
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	if existingUser != nil {
		// Update existing user
		err = s.userRepository.UpdateByTelegramID(tgUser.ID, map[string]interface{}{
			"username":      user.Username,
			"first_name":    user.FirstName,
			"last_name":     user.LastName,
			"language_code": user.LanguageCode,
			"is_bot":        user.IsBot,
		})
		if err != nil {
			log.Debug().
				Err(err).
				Int64("user_id", tgUser.ID).
				Msg("Failed to update Telegram user")
			return fmt.Errorf("failed to update user: %w", err)
		}
		log.Debug().
			Int64("user_id", tgUser.ID).
			Str("username", tgUser.Username).
			Msg("Telegram user updated successfully")
	} else {
		err = s.userRepository.CreateUser(user)
		if err != nil {
			log.Debug().
				Err(err).
				Int64("user_id", tgUser.ID).
				Str("username", tgUser.Username).
				Msg("Failed to create Telegram user")
			return fmt.Errorf("failed to create user: %w", err)
		}
		log.Debug().
			Int64("user_id", tgUser.ID).
			Str("username", tgUser.Username).
			Msg("Telegram user created successfully")
	}

	return nil
}

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

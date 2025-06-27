package telegram

import (
	"context"

	appmodels "github.com/Conty111/AlfredoBot/internal/models"
	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

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
				ChatID:      update.Message.Chat.ID,
				Text:        "Произошла ошибка. Пожалуйста, попробуйте снова.",
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

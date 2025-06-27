package telegram

import (
	"context"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"
)

const searchByArticleNumberText = "Поиск по артикулу 🔎"
const addItemText = "Добавить товар ®️"
const helpText = "Help ❓"
const supportText = "Support 🆘"
const cancelText = "Отмена"

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

var cancelMenu = &tgmodels.ReplyKeyboardMarkup{
	Keyboard: [][]tgmodels.KeyboardButton{
		{
			{Text: cancelText},
		},
	},
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Привет, " + update.Message.From.FirstName + "! 👋",
		ReplyMarkup: mainMenu,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send default welcome message")
	}
}

func helpHandler(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: `Привет, ` + update.Message.From.FirstName + `.

Доступные команды:
-  ` + addItemText + ` - добавить фото товара с артикулом(-ами)
- ` + searchByArticleNumberText + ` - найти товар по его артикулу
- ` + helpText + ` - показать справку
- ` + supportText + ` - связаться с поддержкой`,
		ReplyMarkup: mainMenu,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send help message")
	}
}

func supportHandler(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: `Свяжитесь с нашей поддержкой:

Email: support@example.com
Телефон: +7 (123) 456-78-90
Telegram: @support_bot`,
		ReplyMarkup: mainMenu,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send support message")
	}
}

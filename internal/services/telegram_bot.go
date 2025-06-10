package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"

	"github.com/Conty111/TelegramBotTemplate/internal/configs"
	"github.com/Conty111/TelegramBotTemplate/internal/interfaces"
	"github.com/Conty111/TelegramBotTemplate/internal/models"
)

// TelegramBotService handles Telegram bot operations
type TelegramBotService struct {
	bot            *tgbotapi.BotAPI
	config         *configs.TelegramConfig
	userRepository interfaces.TelegramUserManager
	updates        tgbotapi.UpdatesChannel
	handlers       map[string]CommandHandler
	middlewares    []Middleware
	wg             sync.WaitGroup
	stopCh         chan struct{}
}

// CommandHandler is a function that handles a specific command
type CommandHandler func(update tgbotapi.Update) error

// Middleware is a function that processes updates before they reach handlers
type Middleware func(update tgbotapi.Update, next CommandHandler) error

// NewTelegramBotService creates a new TelegramBotService
func NewTelegramBotService(
	config *configs.TelegramConfig,
	userRepository interfaces.TelegramUserManager,
) (*TelegramBotService, error) {
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	bot.Debug = config.Debug

	service := &TelegramBotService{
		bot:            bot,
		config:         config,
		userRepository: userRepository,
		handlers:       make(map[string]CommandHandler),
		middlewares:    make([]Middleware, 0),
		stopCh:         make(chan struct{}),
	}

	// Register default handlers
	service.RegisterHandler("start", service.handleStart)
	service.RegisterHandler("help", service.handleHelp)

	return service, nil
}

// Start starts the bot service
func (s *TelegramBotService) Start(ctx context.Context) error {
	log.Info().Str("username", s.bot.Self.UserName).Msg("Telegram bot started")

	var updates tgbotapi.UpdatesChannel
	var err error

	if s.config.UseWebhook && s.config.WebhookURL != "" {
		// Webhook configuration
		wh, _ := tgbotapi.NewWebhook(s.config.WebhookURL)
		_, err = s.bot.Request(wh)
		if err != nil {
			return fmt.Errorf("failed to set webhook: %w", err)
		}

		info, err := s.bot.GetWebhookInfo()
		if err != nil {
			return fmt.Errorf("failed to get webhook info: %w", err)
		}

		if info.LastErrorDate != 0 {
			log.Error().
				Time("last_error_date", time.Unix(int64(info.LastErrorDate), 0)).
				Str("last_error_message", info.LastErrorMessage).
				Msg("Webhook error")
		}

		updates = s.bot.ListenForWebhook("/")
	} else {
		// Long polling configuration
		u := tgbotapi.NewUpdate(0)
		u.Timeout = s.config.Timeout
		updates = s.bot.GetUpdatesChan(u)
	}

	s.updates = updates

	// Start processing updates
	s.wg.Add(1)
	go s.processUpdates(ctx)

	return nil
}

// Stop stops the bot service
func (s *TelegramBotService) Stop() error {
	log.Info().Msg("Stopping Telegram bot")
	close(s.stopCh)
	s.wg.Wait()
	return nil
}

// RegisterHandler registers a command handler
func (s *TelegramBotService) RegisterHandler(command string, handler CommandHandler) {
	s.handlers[command] = handler
}

// RegisterMiddleware registers a middleware
func (s *TelegramBotService) RegisterMiddleware(middleware Middleware) {
	s.middlewares = append(s.middlewares, middleware)
}

// processUpdates processes incoming updates
func (s *TelegramBotService) processUpdates(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case update := <-s.updates:
			go s.handleUpdate(update)
		}
	}
}

// handleUpdate processes a single update
func (s *TelegramBotService) handleUpdate(update tgbotapi.Update) {
	// Save or update user if message exists
	if update.Message != nil && update.Message.From != nil {
		s.saveUser(update.Message.From)
	} else if update.CallbackQuery != nil && update.CallbackQuery.From != nil {
		s.saveUser(update.CallbackQuery.From)
	}

	// Handle commands
	if update.Message != nil && update.Message.IsCommand() {
		command := update.Message.Command()
		handler, exists := s.handlers[command]
		if !exists {
			handler = s.handleUnknownCommand
		}

		// Apply middlewares
		for i := len(s.middlewares) - 1; i >= 0; i-- {
			middleware := s.middlewares[i]
			currentHandler := handler
			handler = func(u tgbotapi.Update) error {
				return middleware(u, currentHandler)
			}
		}

		if err := handler(update); err != nil {
			log.Error().Err(err).Str("command", command).Msg("Failed to handle command")
		}
	}
}

// saveUser saves or updates a user in the database
func (s *TelegramBotService) saveUser(user *tgbotapi.User) {
	existingUser, err := s.userRepository.GetByTelegramID(user.ID)
	if err != nil {
		// User doesn't exist, create a new one
		newUser := &models.TelegramUser{
			TelegramID:   user.ID,
			Username:     user.UserName,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			LanguageCode: user.LanguageCode,
			IsBot:        user.IsBot,
			State:        "new",
		}

		if err := s.userRepository.CreateUser(newUser); err != nil {
			log.Error().Err(err).Int64("telegram_id", user.ID).Msg("Failed to create user")
		}
	} else {
		// User exists, update if needed
		updates := map[string]interface{}{
			"username":      user.UserName,
			"first_name":    user.FirstName,
			"last_name":     user.LastName,
			"language_code": user.LanguageCode,
		}

		if err := s.userRepository.UpdateByID(existingUser.ID, updates); err != nil {
			log.Error().Err(err).Int64("telegram_id", user.ID).Msg("Failed to update user")
		}
	}
}

// handleStart handles the /start command
func (s *TelegramBotService) handleStart(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to the bot! Type /help to see available commands.")
	_, err := s.bot.Send(msg)
	return err
}

// handleHelp handles the /help command
func (s *TelegramBotService) handleHelp(update tgbotapi.Update) error {
	helpText := "Available commands:\n" +
		"/start - Start the bot\n" +
		"/help - Show this help message"

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
	_, err := s.bot.Send(msg)
	return err
}

// handleUnknownCommand handles unknown commands
func (s *TelegramBotService) handleUnknownCommand(update tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command. Type /help to see available commands.")
	_, err := s.bot.Send(msg)
	return err
}

// SendMessage sends a message to a chat
func (s *TelegramBotService) SendMessage(chatID int64, text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	return s.bot.Send(msg)
}

// SendMessageWithMarkup sends a message with reply markup
func (s *TelegramBotService) SendMessageWithMarkup(chatID int64, text string, markup interface{}) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = markup
	return s.bot.Send(msg)
}
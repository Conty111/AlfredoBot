package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"

	"github.com/Conty111/AlfredoBot/internal/configs"
	"github.com/Conty111/AlfredoBot/internal/interfaces"
	"github.com/Conty111/AlfredoBot/internal/models"
)

// TelegramBotService handles Telegram bot operations
type TelegramBotService struct {
	bot            *bot.Bot
	config         *configs.TelegramConfig
	userRepository interfaces.TelegramUserManager
	handlers       map[string]CommandHandler
	middlewares    []Middleware
	wg             sync.WaitGroup
	stopCh         chan struct{}
	cancel         context.CancelFunc
	botUser        *tgmodels.User
}

// CommandHandler is a function that handles a specific command
type CommandHandler func(ctx context.Context, b *bot.Bot, update *tgmodels.Update) error

// Middleware is a function that processes updates before they reach handlers
type Middleware func(ctx context.Context, b *bot.Bot, update *tgmodels.Update, next CommandHandler) error

// NewTelegramBotService creates a new TelegramBotService
func NewTelegramBotService(
	config *configs.TelegramConfig,
	userRepository interfaces.TelegramUserManager,
) (*TelegramBotService, error) {
	if config == nil {
		return nil, fmt.Errorf("telegram config is nil")
	}
	if config.Token == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	log.Debug().
		Str("token_prefix", config.Token[:4]+"...").
		Bool("webhook", config.UseWebhook).
		Msg("Initializing Telegram bot")

	// Create bot with options
	opts := []bot.Option{
		bot.WithDebug(),
	}

	b, err := bot.New(config.Token, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Telegram bot (token: %s...): %w",
			config.Token[:4], err)
	}

	service := &TelegramBotService{
		bot:            b,
		config:         config,
		userRepository: userRepository,
		handlers:       make(map[string]CommandHandler),
		middlewares:    make([]Middleware, 0),
		stopCh:         make(chan struct{}),
	}

	// Get and cache bot info
	var getMeErr error
	service.botUser, getMeErr = b.GetMe(context.Background())
	if getMeErr != nil {
		return nil, fmt.Errorf("failed to get bot info: %w", getMeErr)
	}

	// Register default handlers
	service.RegisterHandler("start", service.handleStart)
	service.RegisterHandler("help", service.handleHelp)

	// Get bot info for logging
	botUser, err := b.GetMe(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get bot info: %w", err)
	}

	log.Info().
		Str("bot_username", botUser.Username).
		Msg("Telegram bot initialized successfully")

	return service, nil
}

// Start starts the bot service
func (s *TelegramBotService) Start(parentCtx context.Context) error {
	botUser, err := s.bot.GetMe(parentCtx)
	if err != nil {
		return fmt.Errorf("failed to get bot info: %w", err)
	}

	log.Info().Str("username", botUser.Username).Msg("Telegram bot started")

	// Create cancellable context for the bot
	ctx, cancel := context.WithCancel(parentCtx)
	s.cancel = cancel

	// Register update handler
	s.bot.RegisterHandler(bot.HandlerTypeMessageText, "/", bot.MatchTypePrefix, s.handleMessage)

	// Start processing updates in background
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer log.Debug().Msg("Bot update processing stopped")
		
		// Start bot - it doesn't return error but we still check context
		s.bot.Start(ctx)
		if ctx.Err() != nil {
			log.Debug().Msg("Bot stopped due to context cancellation")
		}
	}()

	return nil
}

// Stop stops the bot service
func (s *TelegramBotService) Stop() error {
	log.Info().Msg("Initiating Telegram bot shutdown")
	
	// Cancel the context first to stop updates
	if s.cancel != nil {
		s.cancel()
	}
	
	// Close the stop channel
	close(s.stopCh)
	
	// Shutdown the bot
	if _, err := s.bot.Close(context.Background()); err != nil {
		log.Error().Err(err).Msg("Error during bot shutdown")
	}
	
	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info().Msg("Telegram bot shutdown completed")
		return nil
	case <-time.After(5 * time.Second):
		log.Warn().Msg("Timeout waiting for bot shutdown")
		return fmt.Errorf("timeout waiting for bot shutdown")
	}
}

// RegisterHandler registers a command handler
func (s *TelegramBotService) RegisterHandler(command string, handler CommandHandler) {
	s.handlers[command] = handler
}

// RegisterMiddleware registers a middleware
func (s *TelegramBotService) RegisterMiddleware(middleware Middleware) {
	s.middlewares = append(s.middlewares, middleware)
}

// handleMessage handles incoming messages
func (s *TelegramBotService) handleMessage(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
	// Save or update user if message exists
	if update.Message != nil && update.Message.From != nil {
		s.saveUser(update.Message.From)
	} else if update.CallbackQuery != nil {
		s.saveUser(update.CallbackQuery.From)
	}

	// Handle commands
	if update.Message != nil && update.Message.Text != "" && update.Message.Text[0] == '/' {
		command := update.Message.Text[1:] // Remove leading '/'
		handler, exists := s.handlers[command]
		if !exists {
			handler = s.handleUnknownCommand
		}

		// Apply middlewares
		for i := len(s.middlewares) - 1; i >= 0; i-- {
			middleware := s.middlewares[i]
			currentHandler := handler
			handler = func(ctx context.Context, b *bot.Bot, u *tgmodels.Update) error {
				return middleware(ctx, b, u, currentHandler)
			}
		}

		if err := handler(ctx, b, update); err != nil {
			log.Error().Err(err).Str("command", command).Msg("Failed to handle command")
		}
	}
}

// saveUser saves or updates a user in the database
func (s *TelegramBotService) saveUser(user interface{}) {
	var userID int64
	var username, firstName, lastName, languageCode string
	var isBot bool

	switch u := user.(type) {
	case *tgmodels.User:
		userID = u.ID
		username = u.Username
		firstName = u.FirstName
		lastName = u.LastName
		languageCode = u.LanguageCode
		isBot = u.IsBot
	case tgmodels.User:
		userID = u.ID
		username = u.Username
		firstName = u.FirstName
		lastName = u.LastName
		languageCode = u.LanguageCode
		isBot = u.IsBot
	default:
		log.Error().Msgf("Invalid user type: %T", user)
		return
	}

	existingUser, err := s.userRepository.GetByTelegramID(userID)
	if err != nil {
		// User doesn't exist, create a new one
		newUser := &models.TelegramUser{
			TelegramID:   userID,
			Username:     username,
			FirstName:    firstName,
			LastName:     lastName,
			LanguageCode: languageCode,
			IsBot:        isBot,
			State:        "new",
		}

		if err := s.userRepository.CreateUser(newUser); err != nil {
			log.Error().Err(err).Int64("telegram_id", userID).Msg("Failed to create user")
		}
	} else {
		// User exists, update if needed
		updates := map[string]interface{}{
			"username":      username,
			"first_name":    firstName,
			"last_name":     lastName,
			"language_code": languageCode,
		}

		if err := s.userRepository.UpdateByID(existingUser.ID, updates); err != nil {
			log.Error().Err(err).Int64("telegram_id", userID).Msg("Failed to update user")
		}
	}
}

// handleStart handles the /start command
func (s *TelegramBotService) handleStart(ctx context.Context, b *bot.Bot, update *tgmodels.Update) error {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Welcome to the bot! Type /help to see available commands.",
	})
	return err
}

// handleHelp handles the /help command
func (s *TelegramBotService) handleHelp(ctx context.Context, b *bot.Bot, update *tgmodels.Update) error {
	helpText := "Available commands:\n" +
		"/start - Start the bot\n" +
		"/help - Show this help message"

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   helpText,
	})
	return err
}

// handleUnknownCommand handles unknown commands
func (s *TelegramBotService) handleUnknownCommand(ctx context.Context, b *bot.Bot, update *tgmodels.Update) error {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Unknown command. Type /help to see available commands.",
	})
	return err
}

// SendMessage sends a message to a chat
func (s *TelegramBotService) SendMessage(ctx context.Context, chatID int64, text string) (*tgmodels.Message, error) {
	return s.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
}

// SendMessageWithMarkup sends a message with reply markup
func (s *TelegramBotService) SendMessageWithMarkup(ctx context.Context, chatID int64, text string, markup interface{}) (*tgmodels.Message, error) {
	return s.bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ReplyMarkup: markup,
	})
}
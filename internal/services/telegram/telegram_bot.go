package telegram

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/rs/zerolog/log"

	"github.com/Conty111/AlfredoBot/internal/configs"
	"github.com/Conty111/AlfredoBot/internal/errs"
	"github.com/Conty111/AlfredoBot/internal/interfaces"
	"github.com/Conty111/AlfredoBot/internal/models"
)

// TelegramBotService is the main service that manages the Telegram bot operations.
// It handles command registration, message processing, and bot lifecycle management.
type TelegramBotService struct {
	bot                  *bot.Bot
	config               *configs.TelegramConfig
	s3Config             *configs.S3Config
	userRepository       interfaces.TelegramUserManager
	photoRepository      interfaces.PhotoManager
	articleRepository    interfaces.ArticleNumberManager
	wg                   sync.WaitGroup
	stopCh               chan struct{}
	cancel               context.CancelFunc
	botUser              *tgmodels.User
}

func NewTelegramBotService(
	config *configs.TelegramConfig,
	s3Config *configs.S3Config,
	userRepository interfaces.TelegramUserManager,
	photoRepository interfaces.PhotoManager,
	articleRepository interfaces.ArticleNumberManager,
	s3Client interfaces.S3Client,
) (*TelegramBotService, error) {
	if config == nil {
		return nil, fmt.Errorf("telegram config is nil")
	}
	if config.Token == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	service := &TelegramBotService{
		config:            config,
		s3Config:          s3Config,
		userRepository:    userRepository,
		photoRepository:   photoRepository,
		articleRepository: articleRepository,
		stopCh:            make(chan struct{}),
	}

	log.Debug().
		Str("token_prefix", config.Token[:4]+"...").
		Bool("webhook", config.UseWebhook).
		Msg("Initializing Telegram bot")

	opts := []bot.Option{
		// bot.WithDebug(),
	}
	opts = service.RegisterHandlers(opts)

	b, err := bot.New(config.Token, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Telegram bot (token: %s...): %w",
			config.Token[:4], err)
	}

	service.bot = b

	var getMeErr error
	service.botUser, getMeErr = b.GetMe(context.Background())
	if getMeErr != nil {
		return nil, fmt.Errorf("failed to get bot info: %w", getMeErr)
	}

	log.Info().
		Str("bot_username", service.botUser.Username).
		Msg("Telegram bot initialized successfully")

	return service, nil
}

func (s *TelegramBotService) Start(parentCtx context.Context) error {
	ctx, cancel := context.WithCancel(parentCtx)
	s.cancel = cancel

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer log.Debug().Msg("Bot update processing stopped")
		
		s.bot.Start(ctx)
		if ctx.Err() != nil {
			log.Debug().Msg("Bot stopped due to context cancellation")
		}
	}()

	return nil
}

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

// SaveUser saves or updates a Telegram user in the database
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
	if err != nil && !errors.Is(err, errs.UserNotFound) {
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
		// Create new user
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

func (s *TelegramBotService) RegisterHandlers(opts []bot.Option) []bot.Option {
	return append(opts,
		[]bot.Option{
			bot.WithMiddlewares(s.saveUserMiddleware, s.routerMiddleware),
			bot.WithDefaultHandler(defaultHandler),
			bot.WithMessageTextHandler(helpText, bot.MatchTypeExact, helpHandler),
			bot.WithMessageTextHandler(supportText, bot.MatchTypeExact, supportHandler),
			bot.WithMessageTextHandler(searchByArticleNumberText, bot.MatchTypeExact, s.searchByArticleNumberHandler),
			bot.WithMessageTextHandler(addItemText, bot.MatchTypeExact, s.addItemHandler),
		}...
	)
}

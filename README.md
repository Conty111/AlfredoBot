# Telegram Bot Template

A reusable template for creating Telegram bots in Go.

TODO:
1. Create bucket when MinIO bootstraps
2. Fix code clean
3. Implement Cancel button while sending photo or article

## Features

* Clean architecture with dependency injection
* Database integration with GORM
* Telegram Bot API integration
* Command handling system with middleware support
* User state management
* Support for both long polling and webhook modes
* Docker support for easy deployment
* S3-compatible storage with MinIO

## How to run

### Requirements

* **_Local running_**: Go v1.22 and PostgreSQL server
* **_Running in Docker_**: Docker with Docker Compose

### Docker run

```
cp .docker.env.example .docker.env
# Edit .docker.env to add your Telegram bot token
# Generate self-signed certificates for MinIO
make gen-certs

# Start the services
docker-compose up -d
```

To stop:
```
docker-compose down
```

### Local run

1. Create a **.env** file from the example
    ```
    cp .env.example .env
    ```
2. Make sure PostgreSQL is running. Update connection details in the **.env** file
3. Add your Telegram bot token to the **.env** file
4. Install dependencies
    ```
    go mod tidy
    ```
5. Run the application
    ```
    make run
    ```
    or build and run the binary
    ```
    make build
    chmod +x ./build/app
    ./build/app serve
    ```

## Creating a Telegram Bot

1. Talk to [@BotFather](https://t.me/BotFather) on Telegram to create a new bot
2. Get the token and add it to your .env file as `TELEGRAM_TOKEN`
3. Customize the bot by implementing your own command handlers

## Customizing the Bot

### Adding New Commands

To add a new command handler, modify the `NewTelegramBotService` function in `internal/services/telegram_bot.go`:

```go
// Register your custom handlers
service.RegisterHandler("mycommand", service.handleMyCommand)
```

Then implement your handler function:

```go
func (s *TelegramBotService) handleMyCommand(update tgbotapi.Update) error {
    // Your command logic here
    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "This is my custom command!")
    _, err := s.bot.Send(msg)
    return err
}
```

### Adding Middleware

You can add middleware to process updates before they reach handlers:

```go
// Add middleware
service.RegisterMiddleware(func(update tgbotapi.Update, next CommandHandler) error {
    // Do something before handling the command
    log.Info().Str("command", update.Message.Command()).Msg("Processing command")
    
    // Call the next handler
    return next(update)
})
```

## Make commands

```
# runs application
make run

# install dev tools (wire, ginkgo)
make install-tools

# build application
make build

# run all unit tests
make test-unit

# run go generate
make gen

# generate dependencies with wire
make deps
```

## Project structure

```
├── cmd
│   └── app - application entry point
├── internal
│   ├── app - application assembly
│   │   ├── build - build information
│   │   ├── cli - command line interface
│   │   ├── dependencies - dependency container
│   │   └── initializers - component initializers
│   ├── configs - configuration structures
│   ├── errs - custom errors
│   ├── interfaces - component interfaces
│   ├── models - entity models
│   ├── repositories - storage layer
│   └── services - business logic layer
├── pkg
│   └── logger - logging utilities
└── test - tests and mocks
```

## Tools and packages

* [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) - Telegram Bot API wrapper
* [gorm](https://gorm.io/) - ORM library
* [cobra](https://github.com/spf13/cobra) - CLI framework
* [envy](https://github.com/gobuffalo/envy) - Environment variable management
* [zerolog](https://github.com/rs/zerolog) - Zero allocation JSON logger
* [wire](https://github.com/google/wire) - Dependency injection
* [ginkgo](https://github.com/onsi/ginkgo) - Testing framework
* [docker](https://www.docker.com/) - Containerization
* [minio](https://min.io/) - S3-compatible object storage

## S3 Storage with MinIO

This project uses MinIO as an S3-compatible object storage service for storing files. MinIO is configured to use TLS for secure communication.

### Setting up MinIO with TLS

1. Generate self-signed certificates:
   ```
   make gen-certs
   ```

2. The certificates will be placed in the `./certs` directory, which is mounted to the MinIO container.

3. MinIO is configured to use these certificates automatically when started with Docker Compose.

4. The S3 client in the application is configured to use TLS when connecting to MinIO.

### Accessing MinIO Console

The MinIO console is available at https://localhost:9001 (username and password are defined in the .env file).

Note: Since we're using self-signed certificates, you'll need to accept the security warning in your browser.
# Alfredo Bot

Mr Alfredo decided to create telegram bot which is very simple. This bot can add photos with articles and search them by article number. It's not even CRUD, but Alfredo tried well to code that.

## How to run

### Requirements

* **_Local running_**: Go v1.23, MinIO and PostgreSQL server
* **_Running in Docker_**: Docker with Docker Compose

### Docker run

```
cp .env.exmaple .env
# don't forget edit .evn file after copying 
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
2. Make sure PostgreSQL and MinIO are running. Update connection details in the **.env** and **config.yaml** file
3. Add your Telegram bot token and other secrets to the **.env** file
4. Install dependencies
    ```
    go mod tidy
    ```
5. Run the application
    ```
    make run
    ```
    or build and run the binary (recommended)
    ```
    make build
    chmod +x ./build/app
    ./build/app serve --config config.yaml
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
│   ├── configs - configuration structures and loading
│   ├── interfaces - component interfaces
│   ├── models - entity models
│   ├── repositories - storage layer
│   └── services - business logic layer
├── pkg
│   └── logger - logging utilities
├── scripts - some scripts for deploy
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
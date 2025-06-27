PROJECT_PKG = github.com/Conty111/AlfredoBot
BUILD_DIR = build

VERSION ?=$(shell git describe --tags --exact-match 2>/dev/null || git symbolic-ref -q --short HEAD)
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null)
BUILD_DATE ?= $(shell date +%FT%T%z)

# remove debug info from the binary & make it smaller
LDFLAGS += -s -w

# inject build info
LDFLAGS += -X ${PROJECT_PKG}/internal/app/build.Version=${VERSION} -X ${PROJECT_PKG}/internal/app/build.CommitHash=${COMMIT_HASH} -X ${PROJECT_PKG}/internal/app/build.BuildDate=${BUILD_DATE}

run-external-API:
	go run ./test/externalAPIserver.go

run:
	go run ./cmd/app/main.go serve

test-unit:
	go test -v -cover ./...

lint:
	golangci-lint run ./...

format:
	goimports -w -l .
	gofmt -w -s .

check-format:
	test -z $$(goimports -l .) && test -z $$(gofmt -l .)

.PHONY: build
build:
	go build ${GOARGS} -tags "${GOTAGS}" -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/app ./cmd/app

gen:
	go generate ./...

deps:
	wire ./...

install-tools:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.6
	go install github.com/google/wire/cmd/wire@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go get -u github.com/onsi/ginkgo/ginkgo

gen-certs:
	mkdir -p certs
	chmod +x scripts/generate-minio-certs.sh
	./scripts/generate-minio-certs.sh
	
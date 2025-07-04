# image for compiling binary
ARG BUILDER_IMAGE="golang:1.23.10"
# here we'll run binary app
ARG RUNNER_IMAGE="alpine:latest"


# build stage
FROM ${BUILDER_IMAGE} AS builder

ENV GO111MODULE=on
#ENV GOPRIVATE ${GOPRIVATE}

# RUN mkdir src
WORKDIR /src
COPY go.mod go.sum ./
# Get dependencies. Also will be cached if we won't change mod/sum
RUN go mod download
# COPY the source code as the last step
COPY . .

# creates build/main files
RUN make build


# running stage
FROM ${RUNNER_IMAGE}


RUN apk update && apk upgrade && apk add --no-cache ca-certificates
RUN apk add musl-dev && apk add libc6-compat

RUN mkdir -p ./db/migrations
RUN mkdir -p ./certs

COPY --from=builder /src/build/app .

RUN chmod +x app
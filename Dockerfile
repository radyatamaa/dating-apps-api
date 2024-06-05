################
# BUILD BINARY #
################
FROM golang:1.17-alpine AS builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
# Install dependencies
RUN apk update && apk add --no-cache ca-certificates gcc git librdkafka-dev musl-dev openssh-client && update-ca-certificates

# Add credentials on build
# Retrieve github.com host key
RUN mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts

WORKDIR $GOPATH/src/app
COPY . .

RUN go env -w GOPRIVATE=github.com/KB-FMF/*
RUN --mount=type=ssh git config --global --add url."git@github.com:".insteadOf "https://github.com/"

# Install swag
RUN --mount=type=ssh go install github.com/swaggo/swag/cmd/swag@v1.8.0

# Fetch dependencies.
RUN --mount=type=ssh go mod download
RUN --mount=type=ssh go mod tidy

# CMD go build -v
RUN GO111MODULE=on CGO_ENABLED=1 GOOS=linux go build -tags dynamic -o dating-apps-api

# Generate swagger
RUN swag init -g main.go --output swagger


#####################
# MAKE SMALL BINARY #
#####################
FROM alpine:latest

RUN apk update --no-cache && apk add --no-cache busybox-extras bash librdkafka-dev musl-dev tzdata
ENV TZ=Asia/Jakarta

# Import from builder.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# Copy the executable & swagger.
COPY --from=builder /go/src/app/dating-apps-api /go/src/app/dating-apps-api
COPY --from=builder /go/src/app/conf /go/src/app/conf
COPY --from=builder /go/src/app/external /go/src/app/external
COPY --from=builder /go/src/app/swagger /go/src/app/swagger

WORKDIR /go/src/app

EXPOSE 8082

ENTRYPOINT ["./dating-apps-api"]
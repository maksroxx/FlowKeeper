# Этап сборки
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache gcc musl-dev icu-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -tags "sqlite_icu" -o server_binary ./cmd

FROM alpine:latest

RUN apk add --no-cache icu-libs ca-certificates

WORKDIR /app

COPY --from=builder /app/server_binary .

COPY --from=builder /app/config ./config
COPY --from=builder /app/assets ./assets

RUN mkdir -p data

EXPOSE 8080

CMD ["./server_binary", "--config", "config/config.yaml"]
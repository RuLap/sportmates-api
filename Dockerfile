FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/sportmates-api ./cmd/api/main.go

RUN CGO_ENABLED=0 GOOS=linux go install github.com/pressly/goose/v3/cmd/goose@latest

FROM alpine:latest
RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /app

COPY --from=build /app/sportmates-api .
COPY --from=build /go/bin/goose /usr/local/bin/goose
COPY internal/pkg/config/config.yaml ./config/config.yaml
COPY migrations/ ./migrations/
COPY internal/app/mail/mailer/templates/ ./internal/app/mail/mailer/templates/

RUN chmod +x /app/sportmates-api

EXPOSE 8080

CMD ["./sportmates-api"]

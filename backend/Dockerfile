# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && mkdir -p /app/storage/attachments

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/seed ./cmd/seed

# Production stage
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && mkdir -p /app/storage/attachments

COPY --from=builder /app/bin/api /app/api
COPY --from=builder /app/bin/seed /app/seed
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

CMD ["/app/api"]

# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git gcc musl-dev opus-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -o meow ./cmd/meow

# Runtime stage
FROM alpine:latest

# Install dependencies including nodejs (for yt-dlp JavaScript runtime)
RUN apk add --no-cache \
    ffmpeg \
    opus \
    python3 \
    py3-pip \
    ca-certificates \
    nodejs \
    npm \
    bash \
    && pip3 install --break-system-packages yt-dlp \
    && rm -rf /var/cache/apk/*

# Verify node works
RUN node --version

WORKDIR /app

# Create data directory for writable cookies
RUN mkdir -p /app/data

COPY --from=builder /app/meow .
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]

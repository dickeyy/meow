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

RUN apk add --no-cache \
    ffmpeg \
    opus \
    python3 \
    py3-pip \
    ca-certificates \
    curl \
    unzip \
    && pip3 install --break-system-packages yt-dlp \
    && rm -rf /var/cache/apk/*

# Install deno (JavaScript runtime required by yt-dlp for YouTube)
RUN curl -fsSL https://deno.land/install.sh | DENO_INSTALL=/usr/local sh

WORKDIR /app

COPY --from=builder /app/meow .

CMD ["./meow"]

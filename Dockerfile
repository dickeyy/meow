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

# Install dependencies
RUN apk add --no-cache \
    ffmpeg \
    opus \
    python3 \
    py3-pip \
    ca-certificates \
    curl \
    unzip \
    bash \
    gcompat \
    && pip3 install --break-system-packages yt-dlp \
    && rm -rf /var/cache/apk/*

# Install deno (download binary directly for Alpine)
RUN curl -fsSL https://github.com/denoland/deno/releases/latest/download/deno-x86_64-unknown-linux-gnu.zip -o /tmp/deno.zip \
    && unzip /tmp/deno.zip -d /usr/local/bin \
    && chmod +x /usr/local/bin/deno \
    && rm /tmp/deno.zip

# Verify deno works
RUN deno --version

WORKDIR /app

# Create data directory for writable cookies
RUN mkdir -p /app/data

COPY --from=builder /app/meow .
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]

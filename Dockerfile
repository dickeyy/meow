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

# Install dependencies including deno
RUN apk add --no-cache \
    ffmpeg \
    opus \
    python3 \
    py3-pip \
    ca-certificates \
    curl \
    unzip \
    bash \
    && pip3 install --break-system-packages yt-dlp \
    && rm -rf /var/cache/apk/*

# Install deno properly
RUN curl -fsSL https://deno.land/install.sh | DENO_INSTALL=/usr/local sh \
    && chmod +x /usr/local/bin/deno

# Verify deno is installed
RUN deno --version

WORKDIR /app

# Create data directory for writable cookies
RUN mkdir -p /app/data

COPY --from=builder /app/meow .
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]

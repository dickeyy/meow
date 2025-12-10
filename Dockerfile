# Build stage
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git gcc musl-dev opus-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -o meow ./cmd/meow

# Runtime stage - use debian for better compatibility with bun
FROM debian:bookworm-slim

# Install dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    libopus0 \
    python3 \
    python3-pip \
    ca-certificates \
    curl \
    unzip \
    && pip3 install --break-system-packages yt-dlp \
    && rm -rf /var/lib/apt/lists/*

# Install bun
RUN curl -fsSL https://bun.sh/install | bash \
    && mv /root/.bun/bin/bun /usr/local/bin/

# Verify bun works
RUN bun --version

WORKDIR /app

# Create data directory for writable cookies
RUN mkdir -p /app/data

COPY --from=builder /app/meow .
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]

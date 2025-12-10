# Build stage - use Debian to match runtime
FROM golang:1.23-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    gcc \
    libopus-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -o meow ./cmd/meow

# Runtime stage
FROM debian:bookworm-slim

# Install dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    libopus0 \
    python3 \
    python3-pip \
    python3-venv \
    ca-certificates \
    curl \
    unzip \
    git \
    && rm -rf /var/lib/apt/lists/*

# Install bun globally
ENV BUN_INSTALL="/usr/local"
RUN curl -fsSL https://bun.sh/install | bash

# Verify bun is in path
RUN echo "Bun version:" && /usr/local/bin/bun --version

# Install latest yt-dlp from pip (nightly has better YouTube support)
RUN pip3 install --break-system-packages -U "yt-dlp[default]"

# Verify yt-dlp version
RUN yt-dlp --version

# Make sure bun is in PATH for all processes
ENV PATH="/usr/local/bin:${PATH}"

WORKDIR /app

# Create data directory for writable cookies
RUN mkdir -p /app/data

COPY --from=builder /app/meow .
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]

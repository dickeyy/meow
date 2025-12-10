#!/bin/bash
set -e

cd "$(dirname "$0")/.."

# Check if .env exists
if [ ! -f .env ]; then
    echo "Error: .env file not found!"
    echo "Copy .env.example to .env and fill in your tokens:"
    echo "  cp .env.example .env"
    exit 1
fi

# Load environment variables
export $(grep -v '^#' .env | xargs)

# Set local database URLs if not set (using non-standard ports to avoid conflicts)
export POSTGRES_URL="${POSTGRES_URL:-postgres://meow:meow@localhost:5433/meow}"
export REDIS_URL="${REDIS_URL:-redis://localhost:6380}"

# Set cookies path if file exists
if [ -f cookies.txt ]; then
    export YOUTUBE_COOKIES_PATH="$(pwd)/cookies.txt"
    echo "Using cookies from: $YOUTUBE_COOKIES_PATH"
fi

# Start databases if not running
if ! docker ps | grep -q meow-postgres; then
    echo "Starting databases..."
    docker compose -f docker-compose.local.yml up -d
    echo "Waiting for databases to be ready..."
    sleep 5
fi

# Build if binary doesn't exist or source is newer
if [ ! -f meow ] || [ "$(find . -name '*.go' -newer meow 2>/dev/null | head -1)" ]; then
    echo "Building..."
    go build -o meow ./cmd/meow
fi

echo "You can now start the bot by running ./meow"


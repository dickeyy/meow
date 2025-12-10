#!/bin/bash
set -e

echo "=== Meow Bot Ubuntu Setup Script ==="
echo ""

# Update package list
echo "[1/7] Updating package list..."
sudo apt-get update

# Install system dependencies
echo "[2/7] Installing system dependencies..."
sudo apt-get install -y \
    ffmpeg \
    libopus0 \
    libopus-dev \
    python3 \
    python3-pip \
    curl \
    unzip \
    git \
    build-essential

# Install Go if not present
if ! command -v go &> /dev/null; then
    echo "[3/7] Installing Go..."
    curl -fsSL https://go.dev/dl/go1.23.4.linux-amd64.tar.gz -o /tmp/go.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    export PATH=$PATH:/usr/local/go/bin
else
    echo "[3/7] Go already installed: $(go version)"
fi

# Install bun if not present
if ! command -v bun &> /dev/null; then
    echo "[4/7] Installing Bun..."
    curl -fsSL https://bun.sh/install | bash
    export BUN_INSTALL="$HOME/.bun"
    export PATH="$BUN_INSTALL/bin:$PATH"
    echo 'export BUN_INSTALL="$HOME/.bun"' >> ~/.bashrc
    echo 'export PATH="$BUN_INSTALL/bin:$PATH"' >> ~/.bashrc
else
    echo "[4/7] Bun already installed: $(bun --version)"
fi

# Install yt-dlp
echo "[5/7] Installing/updating yt-dlp..."
pip3 install --user -U "yt-dlp[default]"
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
export PATH="$HOME/.local/bin:$PATH"

# Install Docker if not present
if ! command -v docker &> /dev/null; then
    echo "[6/7] Installing Docker..."
    curl -fsSL https://get.docker.com | sudo sh
    sudo usermod -aG docker $USER
    echo "NOTE: You may need to log out and back in for Docker permissions"
else
    echo "[6/7] Docker already installed: $(docker --version)"
fi

# Build the bot
echo "[7/7] Building the bot..."
cd "$(dirname "$0")/.."
go mod tidy
go build -o meow ./cmd/meow

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Next steps:"
echo "1. Copy .env.example to .env and fill in your tokens:"
echo "   cp .env.example .env"
echo "   nano .env"
echo ""
echo "2. Export your YouTube cookies to cookies.txt in this directory"
echo ""
echo "3. Start the databases:"
echo "   docker compose -f docker-compose.local.yml up -d"
echo ""
echo "4. Run the bot:"
echo "   ./meow"
echo ""
echo "Or use: ./scripts/run.sh"


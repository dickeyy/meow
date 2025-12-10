#!/bin/bash

cd "$(dirname "$0")/.."

echo "Stopping databases..."
docker compose -f docker-compose.local.yml down

echo "Done."


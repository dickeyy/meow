#!/bin/bash

# Copy cookies to writable location if source exists
if [ -f /app/cookies-source.txt ]; then
    cp /app/cookies-source.txt /app/data/cookies.txt
    echo "Cookies copied to writable location"
fi

# Run the bot
exec /app/meow


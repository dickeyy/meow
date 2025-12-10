package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dickeyy/meow/internal/bot"
	"github.com/dickeyy/meow/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b, err := bot.New(ctx, cfg)
	if err != nil {
		fmt.Printf("Failed to create bot: %v\n", err)
		os.Exit(1)
	}

	if err := b.Start(); err != nil {
		fmt.Printf("Failed to start bot: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Meow is now running. Press Ctrl+C to exit.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	fmt.Println("\nShutting down...")
	b.Stop()
}


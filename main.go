package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pods, err := NewPodcasts()
	if err != nil {
		return fmt.Errorf("could not create podcasts: %w", err)
	}

	log.Println("starting scraping podcasts")
	for _, p := range pods.Podcasts {
		select {
		case <-ctx.Done():
			log.Println("Shutdown signal received, aborting scraping")
		default:
			p.Scrape(ctx)
			log.Println("final episode count: ", len(p.GetEpisodes()))
			p.Encode()
		}
	}

	log.Println("saving podcasts")
	pods.encode(ChunkingOptions{enabled: true, size: 600})

	log.Println("deleting temp files")
	pods.DeleteTempFiles()

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

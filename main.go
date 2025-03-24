package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const podcastFilePath = "content.json"

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pods, err := NewPodcasts()
	if err != nil {
		return fmt.Errorf("could not create podcasts: %w", err)
	}

	for _, p := range pods.Podcasts {
		select {
		case <-ctx.Done():
			log.Println("Shutdown signal received, aborting scraping")
		default:
			log.Println("starting scraping podcasts")
			p.Scrape(ctx)
		}
	}

	// BUG: doesn't save the pods
	log.Println("saving podcasts")
	pods.encode()
	for _, p := range pods.Podcasts {
		log.Println("deleting temp file")
		// BUG: doesn't delete the temp file
		p.DeletePodcastFile()
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

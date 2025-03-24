package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func encode(p Podcast) error {
	if err := p.PodcastFile().Truncate(0); err != nil {
		return err
	}
	if _, err := p.PodcastFile().Seek(0, 0); err != nil {
		return err
	}

	enc := json.NewEncoder(p.PodcastFile())
	return enc.Encode(p)
}

func decode(p Podcast) error {
	if _, err := p.PodcastFile().Seek(0, 0); err != nil {
		return fmt.Errorf("could not seek file: %v", err)
	}

	dec := json.NewDecoder(p.PodcastFile())
	if err := dec.Decode(p); err != nil {
		return fmt.Errorf("cloud not decode JSON: %w", err)
	}
	return nil
}

func deleteFile(fileName string) {
	if err := os.Remove(fileName); err != nil {
		fmt.Println("could not delete temp file: ", err)
	}
}

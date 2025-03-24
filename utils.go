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

func createFile(path string) (*os.File, error) {
	var podcastFile *os.File
	if _, err := os.Stat(path); err == nil {
		podcastFile, err = os.OpenFile(path, os.O_RDWR, 0666)
		if err != nil {
			return &os.File{}, fmt.Errorf("could not open file: %w", err)
		}
	} else if os.IsNotExist(err) {
		podcastFile, err = os.Create(path)
		if err != nil {
			return &os.File{}, fmt.Errorf("could not create podcast file: %w", err)
		}
	} else {
		return &os.File{}, fmt.Errorf("could not stat file: %w", err)
	}

	return podcastFile, nil
}

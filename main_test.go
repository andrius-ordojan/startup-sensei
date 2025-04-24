package main

import (
	"testing"
)

type Pair[T comparable] struct {
	Key   T
	Count int
}

type Counter[T comparable] struct {
	counts map[T]int
}

func NewCounter[T comparable]() *Counter[T] {
	return &Counter[T]{counts: make(map[T]int)}
}

func (c *Counter[T]) Add(k T) {
	c.counts[k]++
}

func (c *Counter[T]) Count(k T) int {
	return c.counts[k]
}

func (c *Counter[T]) Keys() []T {
	keys := make([]T, 0, len(c.counts))
	for k := range c.counts {
		keys = append(keys, k)
	}
	return keys
}

func (c *Counter[T]) Duplicates() []Pair[T] {
	dupes := make([]Pair[T], 0)
	for k, v := range c.counts {
		if v > 1 {
			dupes = append(dupes, Pair[T]{Key: k, Count: v})
		}
	}
	return dupes
}

func TestNoDuplicatePodcasts(t *testing.T) {
	pods, err := NewPodcasts()
	// TODO: clear temp files
	// how to do final post test thing?
	if err != nil {
		t.Fatalf("could not create podcasts: %s", err)
	}
	defer pods.DeleteTempFiles()

	counter := NewCounter[string]()
	for _, p := range pods.Podcasts {
		for _, e := range p.GetEpisodes() {
			counter.Add(e.Title)
		}
	}

	duplicates := counter.Duplicates()
	if len(duplicates) > 0 {
		t.Fatalf("found duplicate podcasts: %v", duplicates)
	}
}

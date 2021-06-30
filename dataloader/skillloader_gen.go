// Code generated by github.com/vektah/dataloaden, DO NOT EDIT.

package dataloader

import (
	"sync"
	"time"

	"kontrakt-server/prisma/db"
)

// SkillLoaderConfig captures the config to create a new SkillLoader
type SkillLoaderConfig struct {
	// Fetch is a method that provides the data for the loader
	Fetch func(keys []int) ([]db.SkillModel, []error)

	// Wait is how long wait before sending a batch
	Wait time.Duration

	// MaxBatch will limit the maximum number of keys to send in one batch, 0 = not limit
	MaxBatch int
}

// NewSkillLoader creates a new SkillLoader given a fetch, wait, and maxBatch
func NewSkillLoader(config SkillLoaderConfig) *SkillLoader {
	return &SkillLoader{
		fetch:    config.Fetch,
		wait:     config.Wait,
		maxBatch: config.MaxBatch,
	}
}

// SkillLoader batches and caches requests
type SkillLoader struct {
	// this method provides the data for the loader
	fetch func(keys []int) ([]db.SkillModel, []error)

	// how long to done before sending a batch
	wait time.Duration

	// this will limit the maximum number of keys to send in one batch, 0 = no limit
	maxBatch int

	// INTERNAL

	// lazily created cache
	cache map[int]db.SkillModel

	// the current batch. keys will continue to be collected until timeout is hit,
	// then everything will be sent to the fetch method and out to the listeners
	batch *skillLoaderBatch

	// mutex to prevent races
	mu sync.Mutex
}

type skillLoaderBatch struct {
	keys    []int
	data    []db.SkillModel
	error   []error
	closing bool
	done    chan struct{}
}

// Load a SkillModel by key, batching and caching will be applied automatically
func (l *SkillLoader) Load(key int) (db.SkillModel, error) {
	return l.LoadThunk(key)()
}

// LoadThunk returns a function that when called will block waiting for a SkillModel.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *SkillLoader) LoadThunk(key int) func() (db.SkillModel, error) {
	l.mu.Lock()
	if it, ok := l.cache[key]; ok {
		l.mu.Unlock()
		return func() (db.SkillModel, error) {
			return it, nil
		}
	}
	if l.batch == nil {
		l.batch = &skillLoaderBatch{done: make(chan struct{})}
	}
	batch := l.batch
	pos := batch.keyIndex(l, key)
	l.mu.Unlock()

	return func() (db.SkillModel, error) {
		<-batch.done

		var data db.SkillModel
		if pos < len(batch.data) {
			data = batch.data[pos]
		}

		var err error
		// its convenient to be able to return a single error for everything
		if len(batch.error) == 1 {
			err = batch.error[0]
		} else if batch.error != nil {
			err = batch.error[pos]
		}

		if err == nil {
			l.mu.Lock()
			l.unsafeSet(key, data)
			l.mu.Unlock()
		}

		return data, err
	}
}

// LoadAll fetches many keys at once. It will be broken into appropriate sized
// sub batches depending on how the loader is configured
func (l *SkillLoader) LoadAll(keys []int) ([]db.SkillModel, []error) {
	results := make([]func() (db.SkillModel, error), len(keys))

	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}

	skillModels := make([]db.SkillModel, len(keys))
	errors := make([]error, len(keys))
	for i, thunk := range results {
		skillModels[i], errors[i] = thunk()
	}
	return skillModels, errors
}

// LoadAllThunk returns a function that when called will block waiting for a SkillModels.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *SkillLoader) LoadAllThunk(keys []int) func() ([]db.SkillModel, []error) {
	results := make([]func() (db.SkillModel, error), len(keys))
	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}
	return func() ([]db.SkillModel, []error) {
		skillModels := make([]db.SkillModel, len(keys))
		errors := make([]error, len(keys))
		for i, thunk := range results {
			skillModels[i], errors[i] = thunk()
		}
		return skillModels, errors
	}
}

// Prime the cache with the provided key and value. If the key already exists, no change is made
// and false is returned.
// (To forcefully prime the cache, clear the key first with loader.clear(key).prime(key, value).)
func (l *SkillLoader) Prime(key int, value db.SkillModel) bool {
	l.mu.Lock()
	var found bool
	if _, found = l.cache[key]; !found {
		l.unsafeSet(key, value)
	}
	l.mu.Unlock()
	return !found
}

// Clear the value at key from the cache, if it exists
func (l *SkillLoader) Clear(key int) {
	l.mu.Lock()
	delete(l.cache, key)
	l.mu.Unlock()
}

func (l *SkillLoader) unsafeSet(key int, value db.SkillModel) {
	if l.cache == nil {
		l.cache = map[int]db.SkillModel{}
	}
	l.cache[key] = value
}

// keyIndex will return the location of the key in the batch, if its not found
// it will add the key to the batch
func (b *skillLoaderBatch) keyIndex(l *SkillLoader, key int) int {
	for i, existingKey := range b.keys {
		if key == existingKey {
			return i
		}
	}

	pos := len(b.keys)
	b.keys = append(b.keys, key)
	if pos == 0 {
		go b.startTimer(l)
	}

	if l.maxBatch != 0 && pos >= l.maxBatch-1 {
		if !b.closing {
			b.closing = true
			l.batch = nil
			go b.end(l)
		}
	}

	return pos
}

func (b *skillLoaderBatch) startTimer(l *SkillLoader) {
	time.Sleep(l.wait)
	l.mu.Lock()

	// we must have hit a batch limit and are already finalizing this batch
	if b.closing {
		l.mu.Unlock()
		return
	}

	l.batch = nil
	l.mu.Unlock()

	b.end(l)
}

func (b *skillLoaderBatch) end(l *SkillLoader) {
	b.data, b.error = l.fetch(b.keys)
	close(b.done)
}

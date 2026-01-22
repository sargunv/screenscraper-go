package scraper

import (
	"sync"
)

// Deduplicator prevents duplicate in-flight requests
// When multiple goroutines request the same key, only one performs the work
// and others wait for the result
type Deduplicator struct {
	mu       sync.Mutex
	inflight map[string]*inflightRequest
}

type inflightRequest struct {
	done   chan struct{}
	result interface{}
	err    error
}

// NewDeduplicator creates a new deduplicator
func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		inflight: make(map[string]*inflightRequest),
	}
}

// Do ensures only one request is made for a given key.
// Concurrent callers with the same key will wait for the first request to complete.
func (d *Deduplicator) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	d.mu.Lock()
	if req, ok := d.inflight[key]; ok {
		d.mu.Unlock()
		<-req.done // Wait for in-flight request
		return req.result, req.err
	}

	req := &inflightRequest{done: make(chan struct{})}
	d.inflight[key] = req
	d.mu.Unlock()

	// Execute the request
	req.result, req.err = fn()
	close(req.done)

	// Clean up
	d.mu.Lock()
	delete(d.inflight, key)
	d.mu.Unlock()

	return req.result, req.err
}

// DoTyped is a typed wrapper around Do for convenience
func DoTyped[T any](d *Deduplicator, key string, fn func() (T, error)) (T, error) {
	result, err := d.Do(key, func() (interface{}, error) {
		return fn()
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return result.(T), nil
}

package scraper

import (
	"context"
	"sync"
	"time"
)

// RateLimiter controls request rate to respect API limits
type RateLimiter struct {
	mu sync.Mutex

	// Rate limiting
	maxPerMinute  int
	requestTimes  []time.Time
	requestsChan  chan struct{}
	totalRequests int       // total API calls started
	startTime     time.Time // when limiter was created

	// Completed request tracking (for progress/ETA)
	completedRequests int         // total API calls completed
	completedTimes    []time.Time // timestamps of completed requests (for rate calc)

	// Thread limiting
	maxThreads int
	threadSem  chan struct{}

	// Backoff state
	backoffUntil time.Time
	backoffLevel int
}

const (
	maxBackoffLevel = 6 // Max backoff: 2^6 = 64 seconds
	baseBackoff     = 1 * time.Second
	maxBackoff      = 60 * time.Second
)

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxThreads, maxPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		maxPerMinute: maxPerMinute,
		maxThreads:   maxThreads,
		threadSem:    make(chan struct{}, maxThreads),
		requestTimes: make([]time.Time, 0, maxPerMinute),
		startTime:    time.Now(),
	}

	// Pre-fill thread semaphore
	for i := 0; i < maxThreads; i++ {
		rl.threadSem <- struct{}{}
	}

	return rl
}

// Acquire waits for permission to make a request
// Returns an error if the context is cancelled
func (rl *RateLimiter) Acquire(ctx context.Context) error {
	// Wait for a thread slot
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-rl.threadSem:
		// Got a thread slot
	}

	// Check and wait for rate limit
	if err := rl.waitForRateLimit(ctx); err != nil {
		// Return thread slot on error
		rl.threadSem <- struct{}{}
		return err
	}

	// Check and wait for backoff
	if err := rl.waitForBackoff(ctx); err != nil {
		// Return thread slot on error
		rl.threadSem <- struct{}{}
		return err
	}

	// Record this request
	rl.recordRequest()

	return nil
}

// Release returns a thread slot after a request completes
func (rl *RateLimiter) Release() {
	rl.recordCompleted()
	rl.threadSem <- struct{}{}
}

// recordCompleted records a completed request for progress tracking
func (rl *RateLimiter) recordCompleted() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.completedRequests++
	rl.completedTimes = append(rl.completedTimes, time.Now())

	// Keep only last 60 seconds of completed times
	cutoff := time.Now().Add(-time.Minute)
	newTimes := make([]time.Time, 0, len(rl.completedTimes))
	for _, t := range rl.completedTimes {
		if t.After(cutoff) {
			newTimes = append(newTimes, t)
		}
	}
	rl.completedTimes = newTimes
}

// TriggerBackoff triggers exponential backoff (call when receiving 429)
func (rl *RateLimiter) TriggerBackoff() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.backoffLevel++
	if rl.backoffLevel > maxBackoffLevel {
		rl.backoffLevel = maxBackoffLevel
	}

	backoffDuration := baseBackoff * time.Duration(1<<(rl.backoffLevel-1))
	if backoffDuration > maxBackoff {
		backoffDuration = maxBackoff
	}

	rl.backoffUntil = time.Now().Add(backoffDuration)
}

// ResetBackoff resets the backoff level (call after successful request)
func (rl *RateLimiter) ResetBackoff() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.backoffLevel = 0
}

// waitForRateLimit waits if we've exceeded the per-minute rate limit
func (rl *RateLimiter) waitForRateLimit(ctx context.Context) error {
	rl.mu.Lock()

	// Clean up old request times (older than 1 minute)
	cutoff := time.Now().Add(-time.Minute)
	newTimes := make([]time.Time, 0, len(rl.requestTimes))
	for _, t := range rl.requestTimes {
		if t.After(cutoff) {
			newTimes = append(newTimes, t)
		}
	}
	rl.requestTimes = newTimes

	// Check if we need to wait
	if len(rl.requestTimes) >= rl.maxPerMinute {
		// Wait until the oldest request is more than 1 minute old
		waitUntil := rl.requestTimes[0].Add(time.Minute)
		waitDuration := time.Until(waitUntil)
		rl.mu.Unlock()

		if waitDuration > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitDuration):
			}
		}
		return nil
	}

	rl.mu.Unlock()
	return nil
}

// waitForBackoff waits if we're in a backoff period
func (rl *RateLimiter) waitForBackoff(ctx context.Context) error {
	rl.mu.Lock()
	waitUntil := rl.backoffUntil
	rl.mu.Unlock()

	if time.Now().Before(waitUntil) {
		waitDuration := time.Until(waitUntil)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitDuration):
		}
	}

	return nil
}

// recordRequest records a request time for rate limiting
func (rl *RateLimiter) recordRequest() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.requestTimes = append(rl.requestTimes, time.Now())
	rl.totalRequests++
}

// RateLimiterStats contains current rate limiter statistics
type RateLimiterStats struct {
	ActiveThreads     int
	MaxThreads        int
	RequestsLastMin   int
	MaxPerMinute      int
	BackoffLevel      int
	BackoffRemaining  time.Duration
	TotalRequests     int     // total API calls completed
	RequestsPerSecond float64 // completed API calls per second (30s sliding window)
	InFlightRequests  int     // requests started but not yet completed
}

// Stats returns current statistics
func (rl *RateLimiter) Stats() RateLimiterStats {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Count recent started requests (last minute) - for rate limiting info
	cutoffMin := now.Add(-time.Minute)
	recentRequests := 0
	for _, t := range rl.requestTimes {
		if t.After(cutoffMin) {
			recentRequests++
		}
	}

	// Calculate COMPLETED requests per second (30 second sliding window)
	cutoff30s := now.Add(-30 * time.Second)
	completed30s := 0
	for _, t := range rl.completedTimes {
		if t.After(cutoff30s) {
			completed30s++
		}
	}
	var reqPerSec float64
	elapsed := now.Sub(rl.startTime)
	if elapsed < 30*time.Second {
		// Not enough time elapsed, use actual elapsed time
		if elapsed > time.Second {
			reqPerSec = float64(rl.completedRequests) / elapsed.Seconds()
		}
	} else {
		reqPerSec = float64(completed30s) / 30.0
	}

	// Calculate active threads (total - available)
	activeThreads := rl.maxThreads - len(rl.threadSem)

	// In-flight = started but not completed
	inFlight := rl.totalRequests - rl.completedRequests

	// Calculate backoff remaining
	var backoffRemaining time.Duration
	if now.Before(rl.backoffUntil) {
		backoffRemaining = time.Until(rl.backoffUntil)
	}

	return RateLimiterStats{
		ActiveThreads:     activeThreads,
		MaxThreads:        rl.maxThreads,
		RequestsLastMin:   recentRequests,
		MaxPerMinute:      rl.maxPerMinute,
		BackoffLevel:      rl.backoffLevel,
		BackoffRemaining:  backoffRemaining,
		TotalRequests:     rl.completedRequests,
		RequestsPerSecond: reqPerSec,
		InFlightRequests:  inFlight,
	}
}

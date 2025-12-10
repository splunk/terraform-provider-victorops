package victorops

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"golang.org/x/time/rate"
)

const (
	// VictorOps API rate limit is approximately 2 requests per second
	defaultRetryMinDelay = 500 * time.Millisecond
	defaultRetryMaxDelay = 30 * time.Second
	defaultRetryTimeout  = 5 * time.Minute
)

// GlobalRateLimiter is used to limit API requests to 2 per second
// This is a global limiter shared across all resources to respect VictorOps API rate limits
var GlobalRateLimiter = rate.NewLimiter(2, 1) // 2 tokens per second, burst of 1

// WaitForRateLimitWithContext waits for the rate limiter before making an API request
func WaitForRateLimitWithContext(ctx context.Context) error {
	return GlobalRateLimiter.Wait(ctx)
}

// WaitForRateLimit waits for the rate limiter (convenience function with background context)
func WaitForRateLimitGlobal() error {
	return GlobalRateLimiter.Wait(context.Background())
}

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MinDelay time.Duration
	MaxDelay time.Duration
	Timeout  time.Duration
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MinDelay: defaultRetryMinDelay,
		MaxDelay: defaultRetryMaxDelay,
		Timeout:  defaultRetryTimeout,
	}
}

// IsRetryableError determines if an error or status code should be retried
func IsRetryableError(statusCode int, err error) bool {
	if err != nil {
		return true
	}

	switch statusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout,      // 504
		http.StatusBadGateway,          // 502
		http.StatusInternalServerError: // 500
		return true
	}

	return false
}

// RetryWithBackoff executes a function with exponential backoff retry logic
func RetryWithBackoff(ctx context.Context, config RetryConfig, operation func() (interface{}, int, error)) (interface{}, error) {
	var result interface{}

	err := retry.RetryContext(ctx, config.Timeout, func() *retry.RetryError {
		res, statusCode, err := operation()

		if err != nil && IsRetryableError(statusCode, err) {
			log.Printf("[DEBUG] Retryable error encountered (status: %d): %v", statusCode, err)
			return retry.RetryableError(fmt.Errorf("retryable error: %w", err))
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		if IsRetryableError(statusCode, nil) {
			log.Printf("[DEBUG] Retryable status code encountered: %d", statusCode)
			return retry.RetryableError(fmt.Errorf("retryable status code: %d", statusCode))
		}

		result = res
		return nil
	})

	return result, err
}

// WaitForRateLimit implements a simple rate limiter to stay under 2 req/sec
type RateLimiter struct {
	lastRequest time.Time
	minInterval time.Duration
}

// NewRateLimiter creates a new rate limiter with the specified minimum interval
func NewRateLimiter(minInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		minInterval: minInterval,
	}
}

// Wait blocks until it's safe to make another request
func (r *RateLimiter) Wait() {
	if r.lastRequest.IsZero() {
		r.lastRequest = time.Now()
		return
	}

	elapsed := time.Since(r.lastRequest)
	if elapsed < r.minInterval {
		time.Sleep(r.minInterval - elapsed)
	}
	r.lastRequest = time.Now()
}

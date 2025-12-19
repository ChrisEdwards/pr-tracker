package github

import (
	"errors"
	"time"
)

// RetryConfig configures retry behavior for transient failures.
type RetryConfig struct {
	MaxAttempts int           // Maximum number of attempts (default: 3)
	InitialWait time.Duration // Initial wait before first retry (default: 1s)
	MaxWait     time.Duration // Maximum wait between retries (default: 10s)
}

// DefaultRetryConfig provides sensible defaults for retry behavior.
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	InitialWait: time.Second,
	MaxWait:     10 * time.Second,
}

// Retryer handles retry logic with exponential backoff.
type Retryer struct {
	config RetryConfig
	// sleep can be overridden for testing
	sleep func(time.Duration)
}

// NewRetryer creates a new Retryer with the given config.
func NewRetryer(config RetryConfig) *Retryer {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = DefaultRetryConfig.MaxAttempts
	}
	if config.InitialWait <= 0 {
		config.InitialWait = DefaultRetryConfig.InitialWait
	}
	if config.MaxWait <= 0 {
		config.MaxWait = DefaultRetryConfig.MaxWait
	}
	return &Retryer{
		config: config,
		sleep:  time.Sleep,
	}
}

// NewDefaultRetryer creates a Retryer with default configuration.
func NewDefaultRetryer() *Retryer {
	return NewRetryer(DefaultRetryConfig)
}

// Do executes the given function with retry logic.
// Retries on transient errors with exponential backoff.
// Returns the result and error from the last attempt.
func (r *Retryer) Do(fn func() error) error {
	var lastErr error

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		// Don't retry non-retriable errors
		if !IsRetriableError(err) {
			return err
		}

		lastErr = err

		// Sleep before next attempt (except on last attempt)
		if attempt < r.config.MaxAttempts {
			wait := r.calculateBackoff(attempt)
			r.sleep(wait)
		}
	}

	return &NetworkError{
		Cause:   lastErr,
		Retries: r.config.MaxAttempts,
	}
}

// DoWithResult executes the given function with retry logic.
// Similar to Do but for functions that return a result.
func (r *Retryer) DoWithResult(fn func() (interface{}, error)) (interface{}, error) {
	var lastErr error
	var result interface{}

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		res, err := fn()
		if err == nil {
			return res, nil
		}

		// Don't retry non-retriable errors
		if !IsRetriableError(err) {
			return nil, err
		}

		lastErr = err
		result = res

		// Sleep before next attempt (except on last attempt)
		if attempt < r.config.MaxAttempts {
			wait := r.calculateBackoff(attempt)
			r.sleep(wait)
		}
	}

	_ = result // result from failed attempts is discarded
	return nil, &NetworkError{
		Cause:   lastErr,
		Retries: r.config.MaxAttempts,
	}
}

// calculateBackoff computes the wait time for a given attempt using exponential backoff.
// Formula: initialWait * 2^(attempt-1), capped at maxWait.
func (r *Retryer) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: initialWait * 2^(attempt-1)
	multiplier := 1 << (attempt - 1) // 2^(attempt-1)
	wait := r.config.InitialWait * time.Duration(multiplier)

	if wait > r.config.MaxWait {
		wait = r.config.MaxWait
	}
	return wait
}

// IsRetriableError determines if an error should trigger a retry.
// Auth errors, rate limit errors, and repo not found errors are not retriable.
// Network errors and unknown errors are retriable.
func IsRetriableError(err error) bool {
	if err == nil {
		return false
	}

	// Auth errors should not be retried
	var authErr *GHAuthError
	if errors.As(err, &authErr) {
		return false
	}

	// Rate limit should not be retried immediately
	var rateLimitErr *RateLimitError
	if errors.As(err, &rateLimitErr) {
		return false
	}

	// Not found errors should not be retried
	var notFoundErr *RepoNotFoundError
	if errors.As(err, &notFoundErr) {
		return false
	}

	// gh not installed should not be retried
	var ghNotFoundErr *GHNotFoundError
	if errors.As(err, &ghNotFoundErr) {
		return false
	}

	// All other errors (including network errors) are retriable
	return true
}

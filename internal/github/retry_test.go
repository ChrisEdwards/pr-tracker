package github

import (
	"errors"
	"testing"
	"time"
)

func TestNewRetryer(t *testing.T) {
	tests := []struct {
		name           string
		config         RetryConfig
		wantMaxAttempt int
		wantInitWait   time.Duration
		wantMaxWait    time.Duration
	}{
		{
			name:           "default config values used for zero config",
			config:         RetryConfig{},
			wantMaxAttempt: DefaultRetryConfig.MaxAttempts,
			wantInitWait:   DefaultRetryConfig.InitialWait,
			wantMaxWait:    DefaultRetryConfig.MaxWait,
		},
		{
			name: "custom config preserved",
			config: RetryConfig{
				MaxAttempts: 5,
				InitialWait: 2 * time.Second,
				MaxWait:     30 * time.Second,
			},
			wantMaxAttempt: 5,
			wantInitWait:   2 * time.Second,
			wantMaxWait:    30 * time.Second,
		},
		{
			name: "negative values replaced with defaults",
			config: RetryConfig{
				MaxAttempts: -1,
				InitialWait: -1,
				MaxWait:     -1,
			},
			wantMaxAttempt: DefaultRetryConfig.MaxAttempts,
			wantInitWait:   DefaultRetryConfig.InitialWait,
			wantMaxWait:    DefaultRetryConfig.MaxWait,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRetryer(tt.config)
			if r.config.MaxAttempts != tt.wantMaxAttempt {
				t.Errorf("MaxAttempts = %d, want %d", r.config.MaxAttempts, tt.wantMaxAttempt)
			}
			if r.config.InitialWait != tt.wantInitWait {
				t.Errorf("InitialWait = %v, want %v", r.config.InitialWait, tt.wantInitWait)
			}
			if r.config.MaxWait != tt.wantMaxWait {
				t.Errorf("MaxWait = %v, want %v", r.config.MaxWait, tt.wantMaxWait)
			}
		})
	}
}

func TestNewDefaultRetryer(t *testing.T) {
	r := NewDefaultRetryer()
	if r.config.MaxAttempts != DefaultRetryConfig.MaxAttempts {
		t.Errorf("MaxAttempts = %d, want %d", r.config.MaxAttempts, DefaultRetryConfig.MaxAttempts)
	}
}

func TestRetryer_Do_Success(t *testing.T) {
	r := NewDefaultRetryer()
	r.sleep = func(d time.Duration) {} // No-op sleep for tests

	calls := 0
	err := r.Do(func() error {
		calls++
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestRetryer_Do_SuccessAfterRetry(t *testing.T) {
	r := NewDefaultRetryer()
	r.sleep = func(d time.Duration) {} // No-op sleep for tests

	calls := 0
	err := r.Do(func() error {
		calls++
		if calls < 3 {
			return errors.New("transient error")
		}
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetryer_Do_MaxRetriesExhausted(t *testing.T) {
	r := NewDefaultRetryer()
	r.sleep = func(d time.Duration) {} // No-op sleep for tests

	calls := 0
	transientErr := errors.New("always fails")
	err := r.Do(func() error {
		calls++
		return transientErr
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var netErr *NetworkError
	if !errors.As(err, &netErr) {
		t.Fatalf("expected NetworkError, got %T: %v", err, err)
	}
	if netErr.Retries != DefaultRetryConfig.MaxAttempts {
		t.Errorf("retries = %d, want %d", netErr.Retries, DefaultRetryConfig.MaxAttempts)
	}
	if calls != DefaultRetryConfig.MaxAttempts {
		t.Errorf("expected %d calls, got %d", DefaultRetryConfig.MaxAttempts, calls)
	}
}

func TestRetryer_Do_NonRetriableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"auth error", &GHAuthError{Message: "not authenticated"}},
		{"rate limit error", &RateLimitError{}},
		{"repo not found error", &RepoNotFoundError{RepoPath: "/some/path"}},
		{"gh not found error", &GHNotFoundError{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewDefaultRetryer()
			r.sleep = func(d time.Duration) {} // No-op sleep for tests

			calls := 0
			err := r.Do(func() error {
				calls++
				return tt.err
			})

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			// Non-retriable errors should return immediately, no retries
			if calls != 1 {
				t.Errorf("expected 1 call (no retries for non-retriable error), got %d", calls)
			}
		})
	}
}

func TestRetryer_DoWithResult_Success(t *testing.T) {
	r := NewDefaultRetryer()
	r.sleep = func(d time.Duration) {} // No-op sleep for tests

	result, err := r.DoWithResult(func() (interface{}, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got %v", result)
	}
}

func TestRetryer_DoWithResult_SuccessAfterRetry(t *testing.T) {
	r := NewDefaultRetryer()
	r.sleep = func(d time.Duration) {} // No-op sleep for tests

	calls := 0
	result, err := r.DoWithResult(func() (interface{}, error) {
		calls++
		if calls < 2 {
			return nil, errors.New("transient error")
		}
		return "success", nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got %v", result)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestRetryer_calculateBackoff(t *testing.T) {
	r := NewRetryer(RetryConfig{
		InitialWait: 1 * time.Second,
		MaxWait:     10 * time.Second,
	})

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 1 * time.Second},  // 1 * 2^0 = 1
		{2, 2 * time.Second},  // 1 * 2^1 = 2
		{3, 4 * time.Second},  // 1 * 2^2 = 4
		{4, 8 * time.Second},  // 1 * 2^3 = 8
		{5, 10 * time.Second}, // 1 * 2^4 = 16, capped at 10
		{6, 10 * time.Second}, // capped at max
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := r.calculateBackoff(tt.attempt)
			if got != tt.want {
				t.Errorf("calculateBackoff(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestRetryer_SleepCalled(t *testing.T) {
	r := NewRetryer(RetryConfig{
		MaxAttempts: 3,
		InitialWait: 100 * time.Millisecond,
		MaxWait:     1 * time.Second,
	})

	var sleepDurations []time.Duration
	r.sleep = func(d time.Duration) {
		sleepDurations = append(sleepDurations, d)
	}

	_ = r.Do(func() error {
		return errors.New("fail")
	})

	// 3 attempts means 2 sleeps (between 1-2 and 2-3)
	if len(sleepDurations) != 2 {
		t.Errorf("expected 2 sleeps, got %d", len(sleepDurations))
	}
	// First sleep: 100ms * 2^0 = 100ms
	if sleepDurations[0] != 100*time.Millisecond {
		t.Errorf("first sleep = %v, want 100ms", sleepDurations[0])
	}
	// Second sleep: 100ms * 2^1 = 200ms
	if sleepDurations[1] != 200*time.Millisecond {
		t.Errorf("second sleep = %v, want 200ms", sleepDurations[1])
	}
}

func TestIsRetriableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"auth error", &GHAuthError{}, false},
		{"rate limit error", &RateLimitError{}, false},
		{"repo not found error", &RepoNotFoundError{}, false},
		{"gh not found error", &GHNotFoundError{}, false},
		{"network error", &NetworkError{Cause: errors.New("timeout")}, true},
		{"generic error", errors.New("something went wrong"), true},
		{"repo scan error", &RepoScanError{Cause: errors.New("fail")}, true},
		{"wrapped auth error", &RepoScanError{Cause: &GHAuthError{}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRetriableError(tt.err)
			if got != tt.want {
				t.Errorf("IsRetriableError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	if DefaultRetryConfig.MaxAttempts != 3 {
		t.Errorf("MaxAttempts = %d, want 3", DefaultRetryConfig.MaxAttempts)
	}
	if DefaultRetryConfig.InitialWait != time.Second {
		t.Errorf("InitialWait = %v, want 1s", DefaultRetryConfig.InitialWait)
	}
	if DefaultRetryConfig.MaxWait != 10*time.Second {
		t.Errorf("MaxWait = %v, want 10s", DefaultRetryConfig.MaxWait)
	}
}

package display

import (
	"bytes"
	"errors"
	"strings"
	"sync"
	"testing"

	"prt/internal/models"
)

func TestNewProgressDisplay(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		p := NewProgressDisplay(10)
		if p.total != 10 {
			t.Errorf("expected total 10, got %d", p.total)
		}
		if p.barWidth != 40 {
			t.Errorf("expected barWidth 40, got %d", p.barWidth)
		}
		if p.writer != nil {
			t.Error("expected nil writer by default")
		}
	})

	t.Run("with writer option", func(t *testing.T) {
		buf := &bytes.Buffer{}
		p := NewProgressDisplay(5, WithWriter(buf))
		if p.writer != buf {
			t.Error("writer not set correctly")
		}
	})

	t.Run("with bar width option", func(t *testing.T) {
		p := NewProgressDisplay(5, WithBarWidth(20))
		if p.barWidth != 20 {
			t.Errorf("expected barWidth 20, got %d", p.barWidth)
		}
	})

	t.Run("ignores invalid bar width", func(t *testing.T) {
		p := NewProgressDisplay(5, WithBarWidth(0))
		if p.barWidth != 40 {
			t.Errorf("expected barWidth 40 for invalid width, got %d", p.barWidth)
		}

		p = NewProgressDisplay(5, WithBarWidth(-5))
		if p.barWidth != 40 {
			t.Errorf("expected barWidth 40 for negative width, got %d", p.barWidth)
		}
	})
}

func TestProgressDisplay_Update(t *testing.T) {
	t.Run("success status", func(t *testing.T) {
		buf := &bytes.Buffer{}
		p := NewProgressDisplay(2, WithWriter(buf))

		repo := &models.Repository{
			Name:       "test-repo",
			ScanStatus: models.ScanStatusSuccess,
			PRs:        []*models.PR{{Number: 1}, {Number: 2}},
		}
		p.Update(repo)

		output := buf.String()
		if !strings.Contains(output, "test-repo") {
			t.Error("output should contain repo name")
		}
		if !strings.Contains(output, "2 PRs") {
			t.Error("output should contain PR count")
		}
		if !strings.Contains(output, IconSuccess) {
			t.Error("output should contain success icon")
		}
		if !strings.Contains(output, "50%") {
			t.Error("output should show 50% progress")
		}
	})

	t.Run("single PR uses singular", func(t *testing.T) {
		buf := &bytes.Buffer{}
		p := NewProgressDisplay(1, WithWriter(buf))

		repo := &models.Repository{
			Name:       "test-repo",
			ScanStatus: models.ScanStatusSuccess,
			PRs:        []*models.PR{{Number: 1}},
		}
		p.Update(repo)

		output := buf.String()
		if !strings.Contains(output, "1 PR)") {
			t.Errorf("output should use singular 'PR', got: %s", output)
		}
	})

	t.Run("no PRs status", func(t *testing.T) {
		buf := &bytes.Buffer{}
		p := NewProgressDisplay(1, WithWriter(buf))

		repo := &models.Repository{
			Name:       "empty-repo",
			ScanStatus: models.ScanStatusNoPRs,
		}
		p.Update(repo)

		output := buf.String()
		if !strings.Contains(output, "0 PRs") {
			t.Error("output should show 0 PRs")
		}
		if !strings.Contains(output, IconSuccess) {
			t.Error("output should contain success icon")
		}
	})

	t.Run("error status", func(t *testing.T) {
		buf := &bytes.Buffer{}
		p := NewProgressDisplay(1, WithWriter(buf))

		repo := &models.Repository{
			Name:       "bad-repo",
			ScanStatus: models.ScanStatusError,
			ScanError:  errors.New("connection failed"),
		}
		p.Update(repo)

		output := buf.String()
		if !strings.Contains(output, "bad-repo") {
			t.Error("output should contain repo name")
		}
		if !strings.Contains(output, "connection failed") {
			t.Error("output should contain error message")
		}
		if !strings.Contains(output, IconError) {
			t.Error("output should contain error icon")
		}
	})

	t.Run("error status with nil error", func(t *testing.T) {
		buf := &bytes.Buffer{}
		p := NewProgressDisplay(1, WithWriter(buf))

		repo := &models.Repository{
			Name:       "bad-repo",
			ScanStatus: models.ScanStatusError,
			ScanError:  nil,
		}
		p.Update(repo)

		output := buf.String()
		if !strings.Contains(output, "error") {
			t.Error("output should contain generic error message")
		}
	})

	t.Run("error status truncates long message", func(t *testing.T) {
		buf := &bytes.Buffer{}
		p := NewProgressDisplay(1, WithWriter(buf))

		longError := strings.Repeat("x", 100)
		repo := &models.Repository{
			Name:       "bad-repo",
			ScanStatus: models.ScanStatusError,
			ScanError:  errors.New(longError),
		}
		p.Update(repo)

		output := buf.String()
		if strings.Contains(output, longError) {
			t.Error("output should truncate long error messages")
		}
		if !strings.Contains(output, "...") {
			t.Error("truncated error should end with ...")
		}
	})

	t.Run("skipped status", func(t *testing.T) {
		buf := &bytes.Buffer{}
		p := NewProgressDisplay(1, WithWriter(buf))

		repo := &models.Repository{
			Name:       "skip-repo",
			ScanStatus: models.ScanStatusSkipped,
		}
		p.Update(repo)

		output := buf.String()
		if !strings.Contains(output, "skipped") {
			t.Error("output should show skipped")
		}
	})

	t.Run("nil writer doesn't panic", func(t *testing.T) {
		p := NewProgressDisplay(1) // No writer

		repo := &models.Repository{
			Name:       "test-repo",
			ScanStatus: models.ScanStatusSuccess,
		}

		// Should not panic
		p.Update(repo)

		if p.done != 1 {
			t.Error("done counter should still be updated")
		}
	})
}

func TestProgressDisplay_Finish(t *testing.T) {
	buf := &bytes.Buffer{}
	p := NewProgressDisplay(3, WithWriter(buf))

	// Simulate various statuses
	p.Update(&models.Repository{Name: "r1", ScanStatus: models.ScanStatusSuccess, PRs: []*models.PR{{}}})
	p.Update(&models.Repository{Name: "r2", ScanStatus: models.ScanStatusError})
	p.Update(&models.Repository{Name: "r3", ScanStatus: models.ScanStatusNoPRs})

	summary := p.Finish()

	if summary.Total != 3 {
		t.Errorf("expected total 3, got %d", summary.Total)
	}
	if summary.Done != 3 {
		t.Errorf("expected done 3, got %d", summary.Done)
	}
	if summary.Success != 2 {
		t.Errorf("expected success 2, got %d", summary.Success)
	}
	if summary.Errors != 1 {
		t.Errorf("expected errors 1, got %d", summary.Errors)
	}
}

func TestProgressDisplay_Clear(t *testing.T) {
	buf := &bytes.Buffer{}
	p := NewProgressDisplay(1, WithWriter(buf))

	p.Update(&models.Repository{Name: "r1", ScanStatus: models.ScanStatusSuccess})
	initialLen := buf.Len()

	p.Clear()

	// Should have written clear sequence
	if buf.Len() <= initialLen {
		t.Error("Clear should write to buffer")
	}

	// Second clear should be no-op
	afterFirstClear := buf.Len()
	p.Clear()
	if buf.Len() != afterFirstClear {
		t.Error("second Clear should be no-op")
	}
}

func TestProgressDisplay_ThreadSafety(t *testing.T) {
	buf := &bytes.Buffer{}
	p := NewProgressDisplay(100, WithWriter(buf))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			repo := &models.Repository{
				Name:       "repo",
				ScanStatus: models.ScanStatusSuccess,
			}
			p.Update(repo)
		}(i)
	}

	wg.Wait()

	if p.done != 100 {
		t.Errorf("expected done 100, got %d", p.done)
	}
	if len(p.results) != 100 {
		t.Errorf("expected 100 results, got %d", len(p.results))
	}
}

func TestProgressDisplay_ProgressCallback(t *testing.T) {
	buf := &bytes.Buffer{}
	p := NewProgressDisplay(2, WithWriter(buf))

	callback := p.ProgressCallback()

	repo := &models.Repository{Name: "test", ScanStatus: models.ScanStatusSuccess}
	callback(1, 2, repo)

	if p.done != 1 {
		t.Errorf("expected done 1, got %d", p.done)
	}
}

func TestSummary_String(t *testing.T) {
	tests := []struct {
		name     string
		summary  Summary
		contains []string
	}{
		{
			name:     "all types",
			summary:  Summary{Total: 10, Done: 10, Success: 5, Errors: 3, Skipped: 2},
			contains: []string{"10 repos scanned", "5 with PRs", "3 errors", "2 skipped"},
		},
		{
			name:     "no errors or skipped",
			summary:  Summary{Total: 5, Done: 5, Success: 5},
			contains: []string{"5 repos scanned", "5 with PRs"},
		},
		{
			name:     "only errors",
			summary:  Summary{Total: 2, Done: 2, Errors: 2},
			contains: []string{"2 repos scanned", "2 errors"},
		},
		{
			name:     "zero everything",
			summary:  Summary{Total: 0, Done: 0},
			contains: []string{"0 repos scanned"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.summary.String()
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected %q in %q", expected, result)
				}
			}
		})
	}
}

func TestProgressDisplay_ProgressBar(t *testing.T) {
	buf := &bytes.Buffer{}
	p := NewProgressDisplay(4, WithWriter(buf), WithBarWidth(10))

	// 25% complete
	p.Update(&models.Repository{Name: "r1", ScanStatus: models.ScanStatusSuccess})
	output := buf.String()

	// Should have approximately 2-3 filled bars (25% of 10)
	filledCount := strings.Count(output, IconBarFilled)
	emptyCount := strings.Count(output, IconBarEmpty)

	if filledCount < 2 || filledCount > 3 {
		t.Errorf("expected 2-3 filled bars at 25%%, got %d", filledCount)
	}
	if filledCount+emptyCount != 10 {
		t.Errorf("expected 10 total bar chars, got %d", filledCount+emptyCount)
	}
}

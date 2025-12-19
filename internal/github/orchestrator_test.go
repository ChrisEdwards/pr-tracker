package github

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"prt/internal/models"
)

// mockClient implements Client for testing
type mockClient struct {
	listPRsFunc func(repoPath string) ([]*models.PR, error)
}

func (m *mockClient) Check() error {
	return nil
}

func (m *mockClient) GetCurrentUser() (string, error) {
	return "testuser", nil
}

func (m *mockClient) ListPRs(repoPath string) ([]*models.PR, error) {
	if m.listPRsFunc != nil {
		return m.listPRsFunc(repoPath)
	}
	return nil, nil
}

func TestNewOrchestrator(t *testing.T) {
	client := &mockClient{}
	o := NewOrchestrator(client)

	if o.client != client {
		t.Error("client not set correctly")
	}
	if o.concurrency != DefaultConcurrency {
		t.Errorf("expected concurrency %d, got %d", DefaultConcurrency, o.concurrency)
	}
}

func TestNewOrchestratorWithConcurrency(t *testing.T) {
	tests := []struct {
		name            string
		concurrency     int
		wantConcurrency int
	}{
		{"normal concurrency", 5, 5},
		{"zero concurrency", 0, 1},
		{"negative concurrency", -1, 1},
		{"high concurrency", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewOrchestratorWithConcurrency(&mockClient{}, tt.concurrency)
			if o.concurrency != tt.wantConcurrency {
				t.Errorf("expected concurrency %d, got %d", tt.wantConcurrency, o.concurrency)
			}
		})
	}
}

func TestFetchAllPRs_EmptyRepos(t *testing.T) {
	client := &mockClient{}
	o := NewOrchestrator(client)

	called := false
	o.FetchAllPRs(nil, func(done, total int, repo *models.Repository) {
		called = true
	})

	if called {
		t.Error("progress should not be called for empty repos")
	}

	o.FetchAllPRs([]*models.Repository{}, func(done, total int, repo *models.Repository) {
		called = true
	})

	if called {
		t.Error("progress should not be called for empty repos slice")
	}
}

func TestFetchAllPRs_Success(t *testing.T) {
	client := &mockClient{
		listPRsFunc: func(repoPath string) ([]*models.PR, error) {
			// Return fresh PRs for each call to avoid shared state
			return []*models.PR{
				{Number: 1, Title: "PR 1"},
				{Number: 2, Title: "PR 2"},
			}, nil
		},
	}

	repos := []*models.Repository{
		{Name: "repo1", Path: "/path/to/repo1"},
		{Name: "repo2", Path: "/path/to/repo2"},
	}

	o := NewOrchestrator(client)

	var progressCalls int
	var mu sync.Mutex

	o.FetchAllPRs(repos, func(done, total int, repo *models.Repository) {
		mu.Lock()
		progressCalls++
		mu.Unlock()

		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
	})

	if progressCalls != 2 {
		t.Errorf("expected 2 progress calls, got %d", progressCalls)
	}

	for _, repo := range repos {
		if repo.ScanStatus != models.ScanStatusSuccess {
			t.Errorf("repo %s: expected status success, got %s", repo.Name, repo.ScanStatus)
		}
		if len(repo.PRs) != 2 {
			t.Errorf("repo %s: expected 2 PRs, got %d", repo.Name, len(repo.PRs))
		}
		// Check that repo context was set on PRs
		for _, pr := range repo.PRs {
			if pr.RepoName != repo.Name {
				t.Errorf("PR repo name not set correctly, expected %s, got %s", repo.Name, pr.RepoName)
			}
			if pr.RepoPath != repo.Path {
				t.Errorf("PR repo path not set correctly, expected %s, got %s", repo.Path, pr.RepoPath)
			}
		}
	}
}

func TestFetchAllPRs_NoPRs(t *testing.T) {
	client := &mockClient{
		listPRsFunc: func(repoPath string) ([]*models.PR, error) {
			return []*models.PR{}, nil
		},
	}

	repos := []*models.Repository{
		{Name: "repo1", Path: "/path/to/repo1"},
	}

	o := NewOrchestrator(client)
	o.FetchAllPRs(repos, nil)

	if repos[0].ScanStatus != models.ScanStatusNoPRs {
		t.Errorf("expected status no_prs, got %s", repos[0].ScanStatus)
	}
}

func TestFetchAllPRs_Error(t *testing.T) {
	expectedErr := errors.New("API error")
	client := &mockClient{
		listPRsFunc: func(repoPath string) ([]*models.PR, error) {
			return nil, expectedErr
		},
	}

	repos := []*models.Repository{
		{Name: "repo1", Path: "/path/to/repo1"},
	}

	o := NewOrchestrator(client)
	o.FetchAllPRs(repos, nil)

	if repos[0].ScanStatus != models.ScanStatusError {
		t.Errorf("expected status error, got %s", repos[0].ScanStatus)
	}
	if repos[0].ScanError != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, repos[0].ScanError)
	}
}

func TestFetchAllPRs_PartialFailure(t *testing.T) {
	client := &mockClient{
		listPRsFunc: func(repoPath string) ([]*models.PR, error) {
			if repoPath == "/path/to/bad" {
				return nil, errors.New("bad repo")
			}
			return []*models.PR{{Number: 1}}, nil
		},
	}

	repos := []*models.Repository{
		{Name: "good", Path: "/path/to/good"},
		{Name: "bad", Path: "/path/to/bad"},
	}

	o := NewOrchestrator(client)
	o.FetchAllPRs(repos, nil)

	// Both repos should be processed
	goodRepo := repos[0]
	badRepo := repos[1]

	if goodRepo.ScanStatus != models.ScanStatusSuccess {
		t.Errorf("good repo: expected success, got %s", goodRepo.ScanStatus)
	}
	if badRepo.ScanStatus != models.ScanStatusError {
		t.Errorf("bad repo: expected error, got %s", badRepo.ScanStatus)
	}
}

func TestFetchAllPRs_Concurrency(t *testing.T) {
	var concurrent int32
	var maxConcurrent int32

	client := &mockClient{
		listPRsFunc: func(repoPath string) ([]*models.PR, error) {
			// Track concurrent executions
			current := atomic.AddInt32(&concurrent, 1)
			for {
				max := atomic.LoadInt32(&maxConcurrent)
				if current > max {
					if atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
						break
					}
				} else {
					break
				}
			}

			// Simulate work
			time.Sleep(10 * time.Millisecond)

			atomic.AddInt32(&concurrent, -1)
			return []*models.PR{{Number: 1}}, nil
		},
	}

	// Create 20 repos to test concurrency limiting
	repos := make([]*models.Repository, 20)
	for i := 0; i < 20; i++ {
		repos[i] = &models.Repository{
			Name: "repo",
			Path: "/path/to/repo",
		}
	}

	o := NewOrchestratorWithConcurrency(client, 5)
	o.FetchAllPRs(repos, nil)

	max := atomic.LoadInt32(&maxConcurrent)
	if max > 5 {
		t.Errorf("concurrency exceeded limit: max %d, expected <= 5", max)
	}
	if max < 2 {
		// Should have some parallelism
		t.Errorf("concurrency too low: max %d, expected >= 2", max)
	}
}

func TestFetchAllPRs_NilProgress(t *testing.T) {
	client := &mockClient{
		listPRsFunc: func(repoPath string) ([]*models.PR, error) {
			return []*models.PR{{Number: 1}}, nil
		},
	}

	repos := []*models.Repository{
		{Name: "repo1", Path: "/path/to/repo1"},
	}

	o := NewOrchestrator(client)

	// Should not panic with nil progress
	o.FetchAllPRs(repos, nil)

	if repos[0].ScanStatus != models.ScanStatusSuccess {
		t.Errorf("expected success, got %s", repos[0].ScanStatus)
	}
}

func TestFetchAllPRs_ConvenienceFunction(t *testing.T) {
	client := &mockClient{
		listPRsFunc: func(repoPath string) ([]*models.PR, error) {
			return []*models.PR{{Number: 1}}, nil
		},
	}

	repos := []*models.Repository{
		{Name: "repo1", Path: "/path/to/repo1"},
	}

	var called bool
	FetchAllPRs(repos, client, func(done, total int, repo *models.Repository) {
		called = true
	})

	if !called {
		t.Error("progress callback should have been called")
	}
	if repos[0].ScanStatus != models.ScanStatusSuccess {
		t.Errorf("expected success, got %s", repos[0].ScanStatus)
	}
}

func TestFetchAllPRs_ProgressCount(t *testing.T) {
	client := &mockClient{
		listPRsFunc: func(repoPath string) ([]*models.PR, error) {
			return []*models.PR{}, nil
		},
	}

	repos := make([]*models.Repository, 5)
	for i := 0; i < 5; i++ {
		repos[i] = &models.Repository{Name: "repo", Path: "/path"}
	}

	o := NewOrchestrator(client)

	var doneValues []int
	var mu sync.Mutex

	o.FetchAllPRs(repos, func(done, total int, repo *models.Repository) {
		mu.Lock()
		doneValues = append(doneValues, done)
		mu.Unlock()

		if total != 5 {
			t.Errorf("total should always be 5, got %d", total)
		}
	})

	if len(doneValues) != 5 {
		t.Errorf("expected 5 progress calls, got %d", len(doneValues))
	}

	// All values from 1 to 5 should appear (order may vary due to concurrency)
	seen := make(map[int]bool)
	for _, v := range doneValues {
		seen[v] = true
	}
	for i := 1; i <= 5; i++ {
		if !seen[i] {
			t.Errorf("done value %d not seen in progress calls", i)
		}
	}
}

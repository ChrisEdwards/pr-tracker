package github

import (
	"sync"

	"prt/internal/models"
)

// DefaultConcurrency is the default number of concurrent requests.
const DefaultConcurrency = 10

// FetchProgress is a callback invoked after each repository is processed.
// done is the number of repos completed, total is the total count,
// and repo is the repository that was just processed.
type FetchProgress func(done, total int, repo *models.Repository)

// Orchestrator coordinates concurrent PR fetching across repositories.
type Orchestrator struct {
	client      Client
	concurrency int
}

// NewOrchestrator creates an orchestrator with the given client and default concurrency.
func NewOrchestrator(client Client) *Orchestrator {
	return &Orchestrator{
		client:      client,
		concurrency: DefaultConcurrency,
	}
}

// NewOrchestratorWithConcurrency creates an orchestrator with custom concurrency.
func NewOrchestratorWithConcurrency(client Client, concurrency int) *Orchestrator {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Orchestrator{
		client:      client,
		concurrency: concurrency,
	}
}

// FetchAllPRs fetches PRs from all repositories concurrently.
// It uses a semaphore to limit concurrency and avoid rate limiting.
// The progress callback is invoked after each repository completes.
// Errors are stored in individual repository's ScanError field;
// this function does not return an error for partial failures.
func (o *Orchestrator) FetchAllPRs(repos []*models.Repository, progress FetchProgress) {
	if len(repos) == 0 {
		return
	}

	var wg sync.WaitGroup
	results := make(chan *models.Repository, len(repos))

	// Semaphore to limit concurrency (avoid rate limiting)
	sem := make(chan struct{}, o.concurrency)

	for _, repo := range repos {
		wg.Add(1)
		go func(r *models.Repository) {
			defer wg.Done()

			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			prs, err := o.client.ListPRs(r.Path)
			if err != nil {
				r.ScanError = err
				r.ScanStatus = models.ScanStatusError
			} else if len(prs) == 0 {
				r.ScanStatus = models.ScanStatusNoPRs
			} else {
				r.PRs = prs
				r.ScanStatus = models.ScanStatusSuccess
				// Set repo context on each PR
				for _, pr := range prs {
					pr.RepoName = r.Name
					pr.RepoPath = r.Path
				}
			}

			results <- r
		}(repo)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results and call progress callback
	done := 0
	total := len(repos)
	for repo := range results {
		done++
		if progress != nil {
			progress(done, total, repo)
		}
	}
}

// FetchAllPRs is a convenience function that creates a default orchestrator
// and fetches PRs from all repositories.
func FetchAllPRs(repos []*models.Repository, client Client, progress FetchProgress) {
	o := NewOrchestrator(client)
	o.FetchAllPRs(repos, progress)
}

// Package categorizer sorts PRs into categories based on the user's relationship to them.
package categorizer

import (
	"prt/internal/config"
	"prt/internal/models"
	"prt/internal/stacks"
)

// Categorizer organizes PRs into meaningful buckets for display.
type Categorizer interface {
	// Categorize takes repositories with fetched PRs and organizes them into
	// a ScanResult with PRs sorted into categories.
	Categorize(repos []*models.Repository, cfg *config.Config, username string) *models.ScanResult
}

// categorizer implements the Categorizer interface.
type categorizer struct{}

// NewCategorizer creates a new Categorizer instance.
func NewCategorizer() Categorizer {
	return &categorizer{}
}

// Categorize processes repositories and categorizes their PRs based on the user's
// relationship to each PR:
//   - My PRs: PRs authored by the current user
//   - Needs My Attention: PRs where review is requested or user is assigned (and not yet approved)
//   - Team PRs: PRs authored by team members
//   - Other PRs: PRs from everyone else (including bots)
func (c *categorizer) Categorize(repos []*models.Repository, cfg *config.Config, username string) *models.ScanResult {
	result := models.NewScanResult()
	result.Username = username

	teamSet := toSet(cfg.TeamMembers)
	botSet := toSet(cfg.Bots)

	for _, repo := range repos {
		// Handle repos with errors
		if repo.ScanError != nil {
			result.ReposWithErrors = append(result.ReposWithErrors, repo)
			continue
		}

		// Handle repos with no PRs
		if !repo.HasPRs() {
			result.ReposWithoutPRs = append(result.ReposWithoutPRs, repo)
			continue
		}

		result.ReposWithPRs = append(result.ReposWithPRs, repo)
		result.TotalPRsFound += len(repo.PRs)

		// Detect stacks for this repo
		result.Stacks[repo.FullName()] = stacks.DetectStacks(repo.PRs)

		// Categorize each PR
		for _, pr := range repo.PRs {
			pr.RepoName = repo.Name
			pr.RepoOwner = repo.Owner
			pr.RepoPath = repo.Path

			// Compute user-specific fields
			pr.IsReviewRequestedFromMe = contains(pr.ReviewRequests, username)
			pr.IsAssignedToMe = contains(pr.Assignees, username)
			pr.MyReviewStatus = findMyReviewStatus(pr.Reviews, username)

			// Categorize
			c.categorizePR(pr, username, teamSet, botSet, result)
		}
	}

	result.TotalReposScanned = len(repos)

	return result
}

// categorizePR determines which category a PR belongs to and adds it to the result.
func (c *categorizer) categorizePR(pr *models.PR, username string, teamSet, botSet map[string]bool, result *models.ScanResult) {
	switch {
	case pr.Author == username:
		// My PR
		result.MyPRs = append(result.MyPRs, pr)

	case pr.IsReviewRequestedFromMe || pr.IsAssignedToMe:
		// Needs my attention (unless already approved by me)
		if pr.MyReviewStatus != models.ReviewStateApproved {
			result.NeedsMyAttention = append(result.NeedsMyAttention, pr)
		} else {
			// I approved it, categorize based on author
			if teamSet[pr.Author] {
				result.TeamPRs = append(result.TeamPRs, pr)
			} else {
				result.OtherPRs = append(result.OtherPRs, pr)
			}
		}

	case teamSet[pr.Author]:
		result.TeamPRs = append(result.TeamPRs, pr)

	case botSet[pr.Author]:
		result.OtherPRs = append(result.OtherPRs, pr)

	default:
		result.OtherPRs = append(result.OtherPRs, pr)
	}
}

// toSet converts a slice of strings into a set (map) for O(1) lookup.
func toSet(slice []string) map[string]bool {
	set := make(map[string]bool, len(slice))
	for _, s := range slice {
		set[s] = true
	}
	return set
}

// contains checks if a slice contains a specific item.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// findMyReviewStatus finds the user's most recent review status on a PR.
// Returns ReviewStateNone if the user hasn't reviewed the PR.
func findMyReviewStatus(reviews []models.Review, username string) models.ReviewState {
	var latest *models.Review
	for i := range reviews {
		r := &reviews[i]
		if r.Author == username {
			if latest == nil || r.Submitted.After(latest.Submitted) {
				latest = r
			}
		}
	}
	if latest == nil {
		return models.ReviewStateNone
	}
	return latest.State
}

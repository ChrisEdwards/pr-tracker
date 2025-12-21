package categorizer

import (
	"errors"
	"testing"
	"time"

	"prt/internal/config"
	"prt/internal/models"
)

func TestCategorize_EmptyRepos(t *testing.T) {
	c := NewCategorizer()
	cfg := &config.Config{}

	result := c.Categorize(nil, cfg, "testuser")

	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got %s", result.Username)
	}
	if result.TotalReposScanned != 0 {
		t.Errorf("Expected 0 repos scanned, got %d", result.TotalReposScanned)
	}
}

func TestCategorize_MyPRs(t *testing.T) {
	c := NewCategorizer()
	cfg := &config.Config{}

	repos := []*models.Repository{
		{
			Name: "test-repo",
			PRs: []*models.PR{
				{Number: 1, Title: "My PR", Author: "testuser"},
				{Number: 2, Title: "Other PR", Author: "someone"},
			},
		},
	}

	result := c.Categorize(repos, cfg, "testuser")

	if len(result.MyPRs) != 1 {
		t.Errorf("Expected 1 PR in MyPRs, got %d", len(result.MyPRs))
	}
	if result.MyPRs[0].Number != 1 {
		t.Errorf("Expected PR #1 in MyPRs, got #%d", result.MyPRs[0].Number)
	}
	if len(result.OtherPRs) != 1 {
		t.Errorf("Expected 1 PR in OtherPRs, got %d", len(result.OtherPRs))
	}
}

func TestCategorize_NeedsMyAttention_ReviewRequested(t *testing.T) {
	c := NewCategorizer()
	cfg := &config.Config{}

	repos := []*models.Repository{
		{
			Name: "test-repo",
			PRs: []*models.PR{
				{
					Number:         1,
					Title:          "Review me",
					Author:         "alice",
					ReviewRequests: []string{"testuser"},
				},
			},
		},
	}

	result := c.Categorize(repos, cfg, "testuser")

	if len(result.NeedsMyAttention) != 1 {
		t.Errorf("Expected 1 PR in NeedsMyAttention, got %d", len(result.NeedsMyAttention))
	}
	if !result.NeedsMyAttention[0].IsReviewRequestedFromMe {
		t.Error("Expected IsReviewRequestedFromMe to be true")
	}
}

func TestCategorize_NeedsMyAttention_Assigned(t *testing.T) {
	c := NewCategorizer()
	cfg := &config.Config{}

	repos := []*models.Repository{
		{
			Name: "test-repo",
			PRs: []*models.PR{
				{
					Number:    1,
					Title:     "Assigned to me",
					Author:    "bob",
					Assignees: []string{"testuser", "other"},
				},
			},
		},
	}

	result := c.Categorize(repos, cfg, "testuser")

	if len(result.NeedsMyAttention) != 1 {
		t.Errorf("Expected 1 PR in NeedsMyAttention, got %d", len(result.NeedsMyAttention))
	}
	if !result.NeedsMyAttention[0].IsAssignedToMe {
		t.Error("Expected IsAssignedToMe to be true")
	}
}

func TestCategorize_ApprovedPR_NotNeedsAttention(t *testing.T) {
	c := NewCategorizer()
	cfg := &config.Config{
		TeamMembers: []string{"alice"},
	}

	now := time.Now()
	repos := []*models.Repository{
		{
			Name: "test-repo",
			PRs: []*models.PR{
				{
					Number:         1,
					Title:          "Already approved",
					Author:         "alice",
					ReviewRequests: []string{"testuser"},
					Reviews: []models.Review{
						{Author: "testuser", State: models.ReviewStateApproved, Submitted: now},
					},
				},
			},
		},
	}

	result := c.Categorize(repos, cfg, "testuser")

	// PR should NOT be in NeedsMyAttention because I already approved it
	if len(result.NeedsMyAttention) != 0 {
		t.Errorf("Expected 0 PRs in NeedsMyAttention, got %d", len(result.NeedsMyAttention))
	}
	// Since author is a team member, it should go to TeamPRs
	if len(result.TeamPRs) != 1 {
		t.Errorf("Expected 1 PR in TeamPRs, got %d", len(result.TeamPRs))
	}
}

func TestCategorize_TeamPRs(t *testing.T) {
	c := NewCategorizer()
	cfg := &config.Config{
		TeamMembers: []string{"alice", "bob"},
	}

	repos := []*models.Repository{
		{
			Name: "test-repo",
			PRs: []*models.PR{
				{Number: 1, Title: "Alice's PR", Author: "alice"},
				{Number: 2, Title: "Bob's PR", Author: "bob"},
				{Number: 3, Title: "External PR", Author: "external"},
			},
		},
	}

	result := c.Categorize(repos, cfg, "testuser")

	if len(result.TeamPRs) != 2 {
		t.Errorf("Expected 2 PRs in TeamPRs, got %d", len(result.TeamPRs))
	}
	if len(result.OtherPRs) != 1 {
		t.Errorf("Expected 1 PR in OtherPRs, got %d", len(result.OtherPRs))
	}
}

func TestCategorize_BotPRs(t *testing.T) {
	c := NewCategorizer()
	cfg := &config.Config{
		Bots: []string{"dependabot[bot]", "renovate[bot]"},
	}

	repos := []*models.Repository{
		{
			Name: "test-repo",
			PRs: []*models.PR{
				{Number: 1, Title: "Bump deps", Author: "dependabot[bot]"},
				{Number: 2, Title: "Update package", Author: "renovate[bot]"},
			},
		},
	}

	result := c.Categorize(repos, cfg, "testuser")

	if len(result.OtherPRs) != 2 {
		t.Errorf("Expected 2 PRs in OtherPRs (bots), got %d", len(result.OtherPRs))
	}
}

func TestCategorize_RepoWithError(t *testing.T) {
	c := NewCategorizer()
	cfg := &config.Config{}

	repos := []*models.Repository{
		{
			Name:      "error-repo",
			ScanError: errors.New("auth failed"),
		},
	}

	result := c.Categorize(repos, cfg, "testuser")

	if len(result.ReposWithErrors) != 1 {
		t.Errorf("Expected 1 repo in ReposWithErrors, got %d", len(result.ReposWithErrors))
	}
	if result.TotalReposScanned != 1 {
		t.Errorf("Expected TotalReposScanned=1, got %d", result.TotalReposScanned)
	}
}

func TestCategorize_RepoWithNoPRs(t *testing.T) {
	c := NewCategorizer()
	cfg := &config.Config{}

	repos := []*models.Repository{
		{
			Name: "empty-repo",
			PRs:  []*models.PR{},
		},
	}

	result := c.Categorize(repos, cfg, "testuser")

	if len(result.ReposWithoutPRs) != 1 {
		t.Errorf("Expected 1 repo in ReposWithoutPRs, got %d", len(result.ReposWithoutPRs))
	}
	if result.TotalPRsFound != 0 {
		t.Errorf("Expected TotalPRsFound=0, got %d", result.TotalPRsFound)
	}
}

func TestCategorize_RepoContextSet(t *testing.T) {
	c := NewCategorizer()
	cfg := &config.Config{}

	repos := []*models.Repository{
		{
			Name: "my-repo",
			Path: "/home/user/my-repo",
			PRs: []*models.PR{
				{Number: 1, Title: "Test PR", Author: "testuser"},
			},
		},
	}

	result := c.Categorize(repos, cfg, "testuser")

	if len(result.MyPRs) != 1 {
		t.Fatal("Expected 1 PR in MyPRs")
	}
	pr := result.MyPRs[0]
	if pr.RepoName != "my-repo" {
		t.Errorf("Expected RepoName='my-repo', got %s", pr.RepoName)
	}
	if pr.RepoPath != "/home/user/my-repo" {
		t.Errorf("Expected RepoPath='/home/user/my-repo', got %s", pr.RepoPath)
	}
}

func TestCategorize_StacksDetected(t *testing.T) {
	c := NewCategorizer()
	cfg := &config.Config{}

	repos := []*models.Repository{
		{
			Name:  "stacked-repo",
			Owner: "org",
			PRs: []*models.PR{
				{Number: 1, Title: "Base feature", Author: "alice", HeadBranch: "feature-a", BaseBranch: "main"},
				{Number: 2, Title: "Tests for feature", Author: "alice", HeadBranch: "feature-a-tests", BaseBranch: "feature-a"},
			},
		},
	}

	result := c.Categorize(repos, cfg, "testuser")

	stack, ok := result.Stacks["org/stacked-repo"]
	if !ok {
		t.Fatal("Expected stack for 'org/stacked-repo'")
	}
	if len(stack.AllNodes) != 2 {
		t.Errorf("Expected 2 nodes in stack, got %d", len(stack.AllNodes))
	}
}

func TestCategorize_MultipleRepos(t *testing.T) {
	c := NewCategorizer()
	cfg := &config.Config{
		TeamMembers: []string{"teammate"},
	}

	repos := []*models.Repository{
		{
			Name: "repo1",
			PRs: []*models.PR{
				{Number: 1, Title: "My PR", Author: "testuser"},
			},
		},
		{
			Name: "repo2",
			PRs: []*models.PR{
				{Number: 1, Title: "Team PR", Author: "teammate"},
			},
		},
		{
			Name:      "repo3",
			ScanError: errors.New("failed"),
		},
	}

	result := c.Categorize(repos, cfg, "testuser")

	if result.TotalReposScanned != 3 {
		t.Errorf("Expected TotalReposScanned=3, got %d", result.TotalReposScanned)
	}
	if result.TotalPRsFound != 2 {
		t.Errorf("Expected TotalPRsFound=2, got %d", result.TotalPRsFound)
	}
	if len(result.ReposWithPRs) != 2 {
		t.Errorf("Expected 2 ReposWithPRs, got %d", len(result.ReposWithPRs))
	}
	if len(result.ReposWithErrors) != 1 {
		t.Errorf("Expected 1 ReposWithErrors, got %d", len(result.ReposWithErrors))
	}
}

func TestFindMyReviewStatus_NoReviews(t *testing.T) {
	status := findMyReviewStatus(nil, "testuser")
	if status != models.ReviewStateNone {
		t.Errorf("Expected ReviewStateNone, got %s", status)
	}
}

func TestFindMyReviewStatus_SingleReview(t *testing.T) {
	reviews := []models.Review{
		{Author: "testuser", State: models.ReviewStateApproved, Submitted: time.Now()},
	}
	status := findMyReviewStatus(reviews, "testuser")
	if status != models.ReviewStateApproved {
		t.Errorf("Expected ReviewStateApproved, got %s", status)
	}
}

func TestFindMyReviewStatus_MultipleReviews_LatestWins(t *testing.T) {
	now := time.Now()
	reviews := []models.Review{
		{Author: "testuser", State: models.ReviewStateApproved, Submitted: now.Add(-time.Hour)},
		{Author: "testuser", State: models.ReviewStateChangesRequested, Submitted: now},
	}
	status := findMyReviewStatus(reviews, "testuser")
	if status != models.ReviewStateChangesRequested {
		t.Errorf("Expected ReviewStateChangesRequested (most recent), got %s", status)
	}
}

func TestFindMyReviewStatus_OtherUserReviews(t *testing.T) {
	reviews := []models.Review{
		{Author: "other", State: models.ReviewStateApproved, Submitted: time.Now()},
	}
	status := findMyReviewStatus(reviews, "testuser")
	if status != models.ReviewStateNone {
		t.Errorf("Expected ReviewStateNone (no review from testuser), got %s", status)
	}
}

func TestToSet(t *testing.T) {
	set := toSet([]string{"a", "b", "c"})
	if !set["a"] || !set["b"] || !set["c"] {
		t.Error("Expected all items to be in set")
	}
	if set["d"] {
		t.Error("Expected 'd' to NOT be in set")
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		slice    []string
		item     string
		expected bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
		{nil, "a", false},
	}

	for _, tt := range tests {
		result := contains(tt.slice, tt.item)
		if result != tt.expected {
			t.Errorf("contains(%v, %q) = %v, want %v", tt.slice, tt.item, result, tt.expected)
		}
	}
}

func TestCategorize_PriorityOrder(t *testing.T) {
	// Test that categorization priority works:
	// 1. My PRs (even if review requested)
	// 2. Needs Attention (review requested or assigned, not approved)
	// 3. Team PRs
	// 4. Other PRs

	c := NewCategorizer()
	cfg := &config.Config{
		TeamMembers: []string{"testuser"}, // User is also a team member
	}

	repos := []*models.Repository{
		{
			Name: "test-repo",
			PRs: []*models.PR{
				// My PR (should go to MyPRs, not TeamPRs even though I'm a team member)
				{Number: 1, Title: "My Own PR", Author: "testuser"},
			},
		},
	}

	result := c.Categorize(repos, cfg, "testuser")

	if len(result.MyPRs) != 1 {
		t.Errorf("Expected 1 PR in MyPRs, got %d", len(result.MyPRs))
	}
	if len(result.TeamPRs) != 0 {
		t.Errorf("Expected 0 PRs in TeamPRs (my own PR shouldn't be counted as team), got %d", len(result.TeamPRs))
	}
}

func TestIsTooOld(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		createdAt  time.Time
		maxAgeDays int
		expected   bool
	}{
		{
			name:       "no limit (0) - recent PR",
			createdAt:  now.Add(-24 * time.Hour),
			maxAgeDays: 0,
			expected:   false,
		},
		{
			name:       "no limit (0) - old PR",
			createdAt:  now.AddDate(0, 0, -100),
			maxAgeDays: 0,
			expected:   false,
		},
		{
			name:       "negative limit - should not filter",
			createdAt:  now.AddDate(0, 0, -100),
			maxAgeDays: -1,
			expected:   false,
		},
		{
			name:       "30 day limit - PR is 10 days old",
			createdAt:  now.AddDate(0, 0, -10),
			maxAgeDays: 30,
			expected:   false,
		},
		{
			name:       "30 day limit - PR is 29 days old",
			createdAt:  now.AddDate(0, 0, -29),
			maxAgeDays: 30,
			expected:   false, // within limit
		},
		{
			name:       "30 day limit - PR is 31 days old",
			createdAt:  now.AddDate(0, 0, -31),
			maxAgeDays: 30,
			expected:   true,
		},
		{
			name:       "90 day limit - PR is 100 days old",
			createdAt:  now.AddDate(0, 0, -100),
			maxAgeDays: 90,
			expected:   true,
		},
		{
			name:       "1 day limit - PR is brand new",
			createdAt:  now,
			maxAgeDays: 1,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &models.PR{CreatedAt: tt.createdAt}
			got := isTooOld(pr, tt.maxAgeDays)
			if got != tt.expected {
				t.Errorf("isTooOld() = %v, want %v (createdAt: %v, maxAgeDays: %d)",
					got, tt.expected, tt.createdAt, tt.maxAgeDays)
			}
		})
	}
}

func TestCategorize_MaxPRAgeDays(t *testing.T) {
	c := NewCategorizer()
	now := time.Now()

	repos := []*models.Repository{
		{
			Name: "test-repo",
			PRs: []*models.PR{
				{Number: 1, Title: "Recent PR", Author: "testuser", CreatedAt: now.AddDate(0, 0, -5)},
				{Number: 2, Title: "Old PR", Author: "testuser", CreatedAt: now.AddDate(0, 0, -100)},
				{Number: 3, Title: "Very old PR", Author: "alice", CreatedAt: now.AddDate(0, 0, -200)},
			},
		},
	}

	// Test with 30-day limit - should filter old PRs
	cfg := &config.Config{MaxPRAgeDays: 30}
	result := c.Categorize(repos, cfg, "testuser")

	if len(result.MyPRs) != 1 {
		t.Errorf("Expected 1 PR in MyPRs (only recent one), got %d", len(result.MyPRs))
	}
	if len(result.MyPRs) > 0 && result.MyPRs[0].Number != 1 {
		t.Errorf("Expected PR #1 (recent) in MyPRs, got #%d", result.MyPRs[0].Number)
	}
	if len(result.OtherPRs) != 0 {
		t.Errorf("Expected 0 PRs in OtherPRs (old ones filtered), got %d", len(result.OtherPRs))
	}
}

func TestCategorize_MaxPRAgeDays_NoLimit(t *testing.T) {
	c := NewCategorizer()
	now := time.Now()

	repos := []*models.Repository{
		{
			Name: "test-repo",
			PRs: []*models.PR{
				{Number: 1, Title: "Recent PR", Author: "testuser", CreatedAt: now.AddDate(0, 0, -5)},
				{Number: 2, Title: "Old PR", Author: "testuser", CreatedAt: now.AddDate(0, 0, -100)},
			},
		},
	}

	// Test with no limit (0) - should include all PRs
	cfg := &config.Config{MaxPRAgeDays: 0}
	result := c.Categorize(repos, cfg, "testuser")

	if len(result.MyPRs) != 2 {
		t.Errorf("Expected 2 PRs in MyPRs (no age limit), got %d", len(result.MyPRs))
	}
}

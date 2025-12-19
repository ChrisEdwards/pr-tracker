package display

import (
	"strings"
	"testing"
	"time"

	"prt/internal/models"
)

func TestRenderPR_BasicOutput(t *testing.T) {
	pr := &models.PR{
		Number:     42,
		Title:      "Add new feature",
		URL:        "https://github.com/org/repo/pull/42",
		Author:     "testuser",
		State:      models.PRStateOpen,
		IsDraft:    false,
		BaseBranch: "main",
		HeadBranch: "feature-branch",
		CreatedAt:  time.Now().Add(-48 * time.Hour),
		CIStatus:   models.CIStatusPassing,
		Reviews:    []models.Review{},
	}

	output := RenderPR(pr, TreeBranch, false, false, false)

	// Check PR number is present
	if !strings.Contains(output, "#42") {
		t.Error("Output should contain PR number #42")
	}

	// Check title is present
	if !strings.Contains(output, "Add new feature") {
		t.Error("Output should contain title")
	}

	// Check URL is present
	if !strings.Contains(output, "https://github.com/org/repo/pull/42") {
		t.Error("Output should contain URL")
	}
}

func TestRenderPR_WithBranches(t *testing.T) {
	pr := &models.PR{
		Number:     42,
		Title:      "Add new feature",
		URL:        "https://github.com/org/repo/pull/42",
		Author:     "testuser",
		State:      models.PRStateOpen,
		BaseBranch: "main",
		HeadBranch: "feature-branch",
		CreatedAt:  time.Now().Add(-48 * time.Hour),
		CIStatus:   models.CIStatusPassing,
	}

	output := RenderPR(pr, TreeBranch, false, true, false)

	// Check branch info is present
	if !strings.Contains(output, "feature-branch") {
		t.Error("Output should contain head branch")
	}
	if !strings.Contains(output, "main") {
		t.Error("Output should contain base branch")
	}
	if !strings.Contains(output, "@testuser") {
		t.Error("Output should contain author")
	}
}

func TestRenderPR_DraftState(t *testing.T) {
	pr := &models.PR{
		Number:    42,
		Title:     "WIP: Draft PR",
		URL:       "https://github.com/org/repo/pull/42",
		State:     models.PRStateOpen,
		IsDraft:   true,
		CreatedAt: time.Now(),
		CIStatus:  models.CIStatusNone,
	}

	output := RenderPR(pr, TreeBranch, false, false, false)

	if !strings.Contains(output, "Draft") {
		t.Error("Output should indicate draft state")
	}
}

func TestRenderPR_WithIcons(t *testing.T) {
	pr := &models.PR{
		Number:    42,
		Title:     "Feature with icons",
		URL:       "https://github.com/org/repo/pull/42",
		State:     models.PRStateOpen,
		IsDraft:   true,
		CreatedAt: time.Now(),
		CIStatus:  models.CIStatusPassing,
	}

	output := RenderPR(pr, TreeBranch, true, false, false)

	// Icons should be present when showIcons is true
	if !strings.Contains(output, IconDraft) {
		t.Error("Output should contain draft icon when showIcons is true")
	}
}

func TestRenderPR_BlockedStyle(t *testing.T) {
	pr := &models.PR{
		Number:    42,
		Title:     "Blocked PR",
		URL:       "https://github.com/org/repo/pull/42",
		State:     models.PRStateOpen,
		CreatedAt: time.Now(),
		CIStatus:  models.CIStatusPassing,
	}

	blocked := RenderPR(pr, TreeBranch, false, false, true)

	// Verify blocked output contains required content
	if !strings.Contains(blocked, "#42") {
		t.Error("Blocked PR should still contain PR number")
	}
	if !strings.Contains(blocked, "Blocked PR") {
		t.Error("Blocked PR should still contain title")
	}
	// Note: In non-TTY environments, blocked styling may not be visually different
	// The important thing is the code path is exercised without error
}

func TestFormatCIStatus(t *testing.T) {
	tests := []struct {
		name      string
		status    models.CIStatus
		showIcons bool
		contains  string
	}{
		{"Passing no icons", models.CIStatusPassing, false, "✓"},
		{"Passing with icons", models.CIStatusPassing, true, IconCIPassing},
		{"Failing no icons", models.CIStatusFailing, false, "✗"},
		{"Failing with icons", models.CIStatusFailing, true, IconCIFailing},
		{"Pending no icons", models.CIStatusPending, false, "..."},
		{"Pending with icons", models.CIStatusPending, true, IconCIPending},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatCIStatus(tc.status, tc.showIcons)
			if !strings.Contains(result, tc.contains) {
				t.Errorf("Expected CI status to contain %q, got %q", tc.contains, result)
			}
		})
	}
}

func TestFormatCIStatus_None(t *testing.T) {
	result := formatCIStatus(models.CIStatusNone, false)
	if result != "" {
		t.Errorf("CIStatusNone should return empty string, got %q", result)
	}
}

func TestCountApprovals(t *testing.T) {
	tests := []struct {
		name     string
		reviews  []models.Review
		expected int
	}{
		{
			name:     "No reviews",
			reviews:  []models.Review{},
			expected: 0,
		},
		{
			name: "One approval",
			reviews: []models.Review{
				{State: models.ReviewStateApproved},
			},
			expected: 1,
		},
		{
			name: "Multiple approvals",
			reviews: []models.Review{
				{State: models.ReviewStateApproved},
				{State: models.ReviewStateApproved},
				{State: models.ReviewStateCommented},
			},
			expected: 2,
		},
		{
			name: "No approvals with other states",
			reviews: []models.Review{
				{State: models.ReviewStateChangesRequested},
				{State: models.ReviewStateCommented},
			},
			expected: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := countApprovals(tc.reviews)
			if result != tc.expected {
				t.Errorf("Expected %d approvals, got %d", tc.expected, result)
			}
		})
	}
}

func TestGetReviewState(t *testing.T) {
	tests := []struct {
		name     string
		reviews  []models.Review
		expected models.ReviewState
	}{
		{
			name:     "No reviews",
			reviews:  []models.Review{},
			expected: models.ReviewStateNone,
		},
		{
			name: "Only approved",
			reviews: []models.Review{
				{State: models.ReviewStateApproved},
			},
			expected: models.ReviewStateApproved,
		},
		{
			name: "Changes requested takes priority",
			reviews: []models.Review{
				{State: models.ReviewStateApproved},
				{State: models.ReviewStateChangesRequested},
			},
			expected: models.ReviewStateChangesRequested,
		},
		{
			name: "Only comments",
			reviews: []models.Review{
				{State: models.ReviewStateCommented},
			},
			expected: models.ReviewStateNone,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pr := &models.PR{Reviews: tc.reviews}
			result := getReviewState(pr)
			if result != tc.expected {
				t.Errorf("Expected review state %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestPluralize(t *testing.T) {
	if pluralize(0) != "s" {
		t.Error("pluralize(0) should return 's'")
	}
	if pluralize(1) != "" {
		t.Error("pluralize(1) should return empty string")
	}
	if pluralize(2) != "s" {
		t.Error("pluralize(2) should return 's'")
	}
}

func TestFormatState_AllStates(t *testing.T) {
	tests := []struct {
		name     string
		pr       *models.PR
		contains string
	}{
		{
			name:     "Draft",
			pr:       &models.PR{State: models.PRStateOpen, IsDraft: true},
			contains: "Draft",
		},
		{
			name:     "Open waiting review",
			pr:       &models.PR{State: models.PRStateOpen, IsDraft: false},
			contains: "Waiting review",
		},
		{
			name: "Approved",
			pr: &models.PR{
				State:   models.PRStateOpen,
				IsDraft: false,
				Reviews: []models.Review{{State: models.ReviewStateApproved}},
			},
			contains: "Approved",
		},
		{
			name: "Changes requested",
			pr: &models.PR{
				State:   models.PRStateOpen,
				IsDraft: false,
				Reviews: []models.Review{{State: models.ReviewStateChangesRequested}},
			},
			contains: "Changes requested",
		},
		{
			name:     "Merged",
			pr:       &models.PR{State: models.PRStateMerged},
			contains: "Merged",
		},
		{
			name:     "Closed",
			pr:       &models.PR{State: models.PRStateClosed},
			contains: "Closed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatState(tc.pr, false)
			if !strings.Contains(result, tc.contains) {
				t.Errorf("Expected state to contain %q, got %q", tc.contains, result)
			}
		})
	}
}

func TestRenderPRSimple(t *testing.T) {
	pr := &models.PR{
		Number:    42,
		Title:     "Simple PR",
		URL:       "https://github.com/org/repo/pull/42",
		State:     models.PRStateOpen,
		CreatedAt: time.Now(),
		CIStatus:  models.CIStatusPassing,
	}

	output := RenderPRSimple(pr, false, false)

	// Should have content
	if output == "" {
		t.Error("RenderPRSimple should produce output")
	}

	// Should contain PR info
	if !strings.Contains(output, "#42") {
		t.Error("Output should contain PR number")
	}
}

// Tests for the main output orchestrator

func TestRender_NilResult(t *testing.T) {
	_, err := Render(nil, RenderOptions{})
	if err == nil {
		t.Error("Render should return error for nil result")
	}
}

func TestRender_EmptyResult(t *testing.T) {
	result := models.NewScanResult()
	result.TotalReposScanned = 3
	result.TotalPRsFound = 0
	result.ScanDuration = 500 * time.Millisecond

	output, err := Render(result, RenderOptions{})
	if err != nil {
		t.Fatalf("Render should not error for empty result: %v", err)
	}

	// Should have header
	if !strings.Contains(output, "PRT") {
		t.Error("Output should contain PRT header")
	}

	// Should have all sections
	if !strings.Contains(output, "MY PRS") {
		t.Error("Output should contain MY PRS section")
	}
	if !strings.Contains(output, "NEEDS MY ATTENTION") {
		t.Error("Output should contain NEEDS MY ATTENTION section")
	}
	if !strings.Contains(output, "TEAM PRS") {
		t.Error("Output should contain TEAM PRS section")
	}
	if !strings.Contains(output, "OTHER PRS") {
		t.Error("Output should contain OTHER PRS section")
	}

	// Should have footer with stats
	if !strings.Contains(output, "Scanned 3 repos") {
		t.Error("Output should contain repo count in footer")
	}
	if !strings.Contains(output, "Found 0 PRs") {
		t.Error("Output should contain PR count in footer")
	}
}

func TestRender_WithPRs(t *testing.T) {
	result := models.NewScanResult()
	result.TotalReposScanned = 2
	result.TotalPRsFound = 3
	result.ScanDuration = 1200 * time.Millisecond

	result.MyPRs = []*models.PR{
		{
			Number:    1,
			Title:     "My PR",
			URL:       "https://github.com/org/repo/pull/1",
			RepoName:  "repo",
			State:     models.PRStateOpen,
			CreatedAt: time.Now(),
			CIStatus:  models.CIStatusPassing,
		},
	}
	result.NeedsMyAttention = []*models.PR{
		{
			Number:    2,
			Title:     "Review needed",
			URL:       "https://github.com/org/repo/pull/2",
			RepoName:  "repo",
			State:     models.PRStateOpen,
			CreatedAt: time.Now(),
			CIStatus:  models.CIStatusPending,
		},
	}
	result.TeamPRs = []*models.PR{
		{
			Number:    3,
			Title:     "Team PR",
			URL:       "https://github.com/org/repo/pull/3",
			RepoName:  "repo",
			State:     models.PRStateOpen,
			CreatedAt: time.Now(),
			CIStatus:  models.CIStatusFailing,
		},
	}

	output, err := Render(result, RenderOptions{})
	if err != nil {
		t.Fatalf("Render should not error: %v", err)
	}

	// Check PRs are included
	if !strings.Contains(output, "#1") {
		t.Error("Output should contain PR #1")
	}
	if !strings.Contains(output, "#2") {
		t.Error("Output should contain PR #2")
	}
	if !strings.Contains(output, "#3") {
		t.Error("Output should contain PR #3")
	}

	// Check footer
	if !strings.Contains(output, "Scanned 2 repos") {
		t.Error("Output should contain repo count")
	}
	if !strings.Contains(output, "Found 3 PRs") {
		t.Error("Output should contain PR count")
	}
}

func TestRender_WithIcons(t *testing.T) {
	result := models.NewScanResult()
	result.MyPRs = []*models.PR{
		{
			Number:    1,
			Title:     "My PR",
			URL:       "https://github.com/org/repo/pull/1",
			RepoName:  "repo",
			State:     models.PRStateOpen,
			IsDraft:   true,
			CreatedAt: time.Now(),
			CIStatus:  models.CIStatusPassing,
		},
	}

	output, err := Render(result, RenderOptions{ShowIcons: true})
	if err != nil {
		t.Fatalf("Render should not error: %v", err)
	}

	// Check for section icons
	if !strings.Contains(output, IconMyPRs) {
		t.Error("Output should contain My PRs icon when ShowIcons is true")
	}
}

func TestRender_WithBranches(t *testing.T) {
	result := models.NewScanResult()
	result.MyPRs = []*models.PR{
		{
			Number:     1,
			Title:      "My PR",
			URL:        "https://github.com/org/repo/pull/1",
			RepoName:   "repo",
			Author:     "testuser",
			State:      models.PRStateOpen,
			HeadBranch: "feature-branch",
			BaseBranch: "main",
			CreatedAt:  time.Now(),
			CIStatus:   models.CIStatusPassing,
		},
	}

	output, err := Render(result, RenderOptions{ShowBranches: true})
	if err != nil {
		t.Fatalf("Render should not error: %v", err)
	}

	// Check for branch info
	if !strings.Contains(output, "feature-branch") {
		t.Error("Output should contain head branch when ShowBranches is true")
	}
	if !strings.Contains(output, "main") {
		t.Error("Output should contain base branch when ShowBranches is true")
	}
}

func TestRender_JSONMode(t *testing.T) {
	result := models.NewScanResult()
	result.TotalReposScanned = 1
	result.TotalPRsFound = 1
	result.MyPRs = []*models.PR{
		{
			Number:    1,
			Title:     "Test PR",
			URL:       "https://github.com/org/repo/pull/1",
			RepoName:  "repo",
			State:     models.PRStateOpen,
			CreatedAt: time.Now(),
		},
	}

	output, err := Render(result, RenderOptions{JSON: true})
	if err != nil {
		t.Fatalf("Render should not error in JSON mode: %v", err)
	}

	// JSON output should not have styled headers
	if strings.Contains(output, "═") {
		t.Error("JSON output should not contain decorative characters")
	}

	// Should contain JSON structure
	if !strings.Contains(output, "\"my_prs\"") {
		t.Error("JSON output should contain my_prs key")
	}
}

func TestRender_WithReposWithoutPRs(t *testing.T) {
	result := models.NewScanResult()
	result.TotalReposScanned = 3
	result.ReposWithoutPRs = []*models.Repository{
		{Name: "empty-repo-1", Path: "/path/to/empty-repo-1"},
		{Name: "empty-repo-2", Path: "/path/to/empty-repo-2"},
	}

	output, err := Render(result, RenderOptions{})
	if err != nil {
		t.Fatalf("Render should not error: %v", err)
	}

	// Should have repos without PRs section
	if !strings.Contains(output, "REPOS WITH NO OPEN PRS") {
		t.Error("Output should contain REPOS WITH NO OPEN PRS section")
	}
	if !strings.Contains(output, "empty-repo-1") {
		t.Error("Output should list empty-repo-1")
	}
	if !strings.Contains(output, "empty-repo-2") {
		t.Error("Output should list empty-repo-2")
	}
}

func TestRender_NoReposWithoutPRsSection(t *testing.T) {
	result := models.NewScanResult()
	result.TotalReposScanned = 1
	// No repos without PRs - section should be omitted

	output, err := Render(result, RenderOptions{})
	if err != nil {
		t.Fatalf("Render should not error: %v", err)
	}

	// Should NOT have repos without PRs section when empty
	if strings.Contains(output, "REPOS WITH NO OPEN PRS") {
		t.Error("Output should not contain REPOS WITH NO OPEN PRS section when there are none")
	}
}

func TestRenderHeader(t *testing.T) {
	header := renderHeader()

	if !strings.Contains(header, "PRT") {
		t.Error("Header should contain PRT title")
	}
	if !strings.Contains(header, "═") {
		t.Error("Header should contain decorative separator")
	}
}

func TestRenderFooter(t *testing.T) {
	result := &models.ScanResult{
		TotalReposScanned: 5,
		TotalPRsFound:     12,
		ScanDuration:      2500 * time.Millisecond,
	}

	footer := renderFooter(result)

	if !strings.Contains(footer, "Scanned 5 repos") {
		t.Error("Footer should contain repo count")
	}
	if !strings.Contains(footer, "Found 12 PRs") {
		t.Error("Footer should contain PR count")
	}
	if !strings.Contains(footer, "═") {
		t.Error("Footer should contain decorative separator")
	}
}

func TestRenderOptions_Defaults(t *testing.T) {
	opts := RenderOptions{}

	if opts.ShowIcons {
		t.Error("ShowIcons should default to false")
	}
	if opts.ShowBranches {
		t.Error("ShowBranches should default to false")
	}
	if opts.NoColor {
		t.Error("NoColor should default to false")
	}
	if opts.JSON {
		t.Error("JSON should default to false")
	}
}

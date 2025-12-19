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

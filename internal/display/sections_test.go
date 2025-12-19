package display

import (
	"strings"
	"testing"
	"time"

	"prt/internal/models"
)

func TestRenderSectionHeader_WithoutIcon(t *testing.T) {
	result := RenderSectionHeader("", "MY PRS", false)
	if !strings.Contains(result, "MY PRS") {
		t.Error("Header should contain title")
	}
}

func TestRenderSectionHeader_WithIcon(t *testing.T) {
	result := RenderSectionHeader(IconMyPRs, "MY PRS", true)
	if !strings.Contains(result, "MY PRS") {
		t.Error("Header should contain title")
	}
	if !strings.Contains(result, IconMyPRs) {
		t.Error("Header should contain icon when showIcons is true")
	}
}

func TestRenderSectionHeader_IconDisabled(t *testing.T) {
	result := RenderSectionHeader(IconMyPRs, "MY PRS", false)
	if strings.Contains(result, IconMyPRs) {
		t.Error("Header should not contain icon when showIcons is false")
	}
}

func TestRenderSection_Empty(t *testing.T) {
	result := RenderSection("MY PRS", IconMyPRs, nil, nil, false, false)

	if !strings.Contains(result, "MY PRS") {
		t.Error("Section should contain title")
	}
	if !strings.Contains(result, "None") {
		t.Error("Empty section should show 'None'")
	}
}

func TestRenderSection_EmptyNeedsAttention(t *testing.T) {
	result := RenderSection("NEEDS MY ATTENTION", IconNeedsAttention, nil, nil, false, false)

	if !strings.Contains(result, "you're all caught up") {
		t.Error("Empty NEEDS MY ATTENTION should show encouraging message")
	}
}

func TestRenderSection_WithPRs(t *testing.T) {
	prs := []*models.PR{
		{
			Number:    42,
			Title:     "Test PR",
			URL:       "https://github.com/org/repo/pull/42",
			RepoName:  "repo",
			State:     models.PRStateOpen,
			CreatedAt: time.Now(),
			CIStatus:  models.CIStatusPassing,
		},
	}

	result := RenderSection("MY PRS", IconMyPRs, prs, nil, false, false)

	if !strings.Contains(result, "MY PRS") {
		t.Error("Section should contain title")
	}
	if !strings.Contains(result, "[repo]") {
		t.Error("Section should contain repo name")
	}
	if !strings.Contains(result, "#42") {
		t.Error("Section should contain PR number")
	}
	if !strings.Contains(result, "Test PR") {
		t.Error("Section should contain PR title")
	}
}

func TestRenderSection_GroupedByRepo(t *testing.T) {
	prs := []*models.PR{
		{Number: 1, Title: "PR 1", RepoName: "repo-b", URL: "http://x/1", State: models.PRStateOpen, CreatedAt: time.Now()},
		{Number: 2, Title: "PR 2", RepoName: "repo-a", URL: "http://x/2", State: models.PRStateOpen, CreatedAt: time.Now()},
		{Number: 3, Title: "PR 3", RepoName: "repo-b", URL: "http://x/3", State: models.PRStateOpen, CreatedAt: time.Now()},
	}

	result := RenderSection("TEST", "", prs, nil, false, false)

	// repo-a should appear before repo-b (alphabetical)
	idxA := strings.Index(result, "[repo-a]")
	idxB := strings.Index(result, "[repo-b]")
	if idxA > idxB {
		t.Error("Repos should be sorted alphabetically")
	}
}

func TestGroupByRepo(t *testing.T) {
	prs := []*models.PR{
		{Number: 1, RepoName: "repo-a"},
		{Number: 2, RepoName: "repo-b"},
		{Number: 3, RepoName: "repo-a"},
	}

	grouped := groupByRepo(prs)

	if len(grouped["repo-a"]) != 2 {
		t.Error("repo-a should have 2 PRs")
	}
	if len(grouped["repo-b"]) != 1 {
		t.Error("repo-b should have 1 PR")
	}
}

func TestSortedRepoNames(t *testing.T) {
	byRepo := map[string][]*models.PR{
		"zebra": {},
		"alpha": {},
		"beta":  {},
	}

	names := sortedRepoNames(byRepo)

	if len(names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names))
	}
	if names[0] != "alpha" || names[1] != "beta" || names[2] != "zebra" {
		t.Errorf("Names not sorted: %v", names)
	}
}

func TestRenderEmptySection(t *testing.T) {
	result := RenderEmptySection("TEST SECTION", "", false)

	if !strings.Contains(result, "TEST SECTION") {
		t.Error("Should contain title")
	}
	if !strings.Contains(result, "None") {
		t.Error("Should contain 'None'")
	}
}

func TestRenderNoOpenPRsSection_Empty(t *testing.T) {
	result := RenderNoOpenPRsSection(nil, false)
	if result != "" {
		t.Error("Empty repos should return empty string")
	}
}

func TestRenderNoOpenPRsSection_WithRepos(t *testing.T) {
	repos := []*models.Repository{
		{Name: "repo-a", Path: "/path/to/repo-a"},
		{Name: "repo-b", Path: "/path/to/repo-b"},
	}

	result := RenderNoOpenPRsSection(repos, false)

	if !strings.Contains(result, "REPOS WITH NO OPEN PRS") {
		t.Error("Should contain section title")
	}
	if !strings.Contains(result, "repo-a") {
		t.Error("Should contain repo-a name")
	}
	if !strings.Contains(result, "/path/to/repo-a") {
		t.Error("Should contain repo-a path")
	}
}

func TestRenderNoOpenPRsSection_WithIcons(t *testing.T) {
	repos := []*models.Repository{
		{Name: "test", Path: "/test"},
	}

	result := RenderNoOpenPRsSection(repos, true)

	if !strings.Contains(result, IconNoOpenPRs) {
		t.Error("Should contain icon when showIcons is true")
	}
}

func TestIsPRBlocked_NoStack(t *testing.T) {
	pr := &models.PR{Number: 42}
	if isPRBlocked(pr, nil) {
		t.Error("PR should not be blocked when no stack")
	}
}

func TestIsPRBlocked_WithStack(t *testing.T) {
	parentPR := &models.PR{Number: 1, State: models.PRStateOpen}
	childPR := &models.PR{Number: 2, State: models.PRStateOpen}

	parentNode := &models.StackNode{PR: parentPR}
	childNode := &models.StackNode{PR: childPR, Parent: parentNode}

	stack := &models.Stack{
		AllNodes: []*models.StackNode{parentNode, childNode},
	}

	// Parent should not be blocked
	if isPRBlocked(parentPR, stack) {
		t.Error("Parent PR should not be blocked")
	}

	// Child should be blocked
	if !isPRBlocked(childPR, stack) {
		t.Error("Child PR should be blocked")
	}
}

func TestIsPRBlocked_MergedParent(t *testing.T) {
	parentPR := &models.PR{Number: 1, State: models.PRStateMerged}
	childPR := &models.PR{Number: 2, State: models.PRStateOpen}

	parentNode := &models.StackNode{PR: parentPR}
	childNode := &models.StackNode{PR: childPR, Parent: parentNode}

	stack := &models.Stack{
		AllNodes: []*models.StackNode{parentNode, childNode},
	}

	// Child should NOT be blocked if parent is merged
	if isPRBlocked(childPR, stack) {
		t.Error("Child PR should not be blocked when parent is merged")
	}
}

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
	result := RenderSection("MY PRS", IconMyPRs, nil, nil, SectionOptions{ShowIcons: false, ShowBranches: false})

	if !strings.Contains(result, "MY PRS") {
		t.Error("Section should contain title")
	}
	if !strings.Contains(result, "None") {
		t.Error("Empty section should show 'None'")
	}
}

func TestRenderSection_EmptyNeedsAttention(t *testing.T) {
	result := RenderSection("NEEDS MY ATTENTION", IconNeedsAttention, nil, nil, SectionOptions{ShowIcons: false, ShowBranches: false})

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

	result := RenderSection("MY PRS", IconMyPRs, prs, nil, SectionOptions{ShowIcons: false, ShowBranches: false})

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

	result := RenderSection("TEST", "", prs, nil, SectionOptions{ShowIcons: false, ShowBranches: false})

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

// TestRenderSection_StackLookupIntegration tests the full flow of stack lookup
// This verifies that stacks keyed by "owner/repo" are found when PRs have
// RepoOwner and RepoName set correctly.
func TestRenderSection_StackLookupIntegration(t *testing.T) {
	// Create PRs with full repo context
	parentPR := &models.PR{
		Number:     1,
		Title:      "Parent PR",
		URL:        "https://github.com/myorg/myrepo/pull/1",
		RepoName:   "myrepo",
		RepoOwner:  "myorg",
		State:      models.PRStateOpen,
		HeadBranch: "feature-a",
		BaseBranch: "main",
		CreatedAt:  time.Now(),
	}
	childPR := &models.PR{
		Number:     2,
		Title:      "Child PR",
		URL:        "https://github.com/myorg/myrepo/pull/2",
		RepoName:   "myrepo",
		RepoOwner:  "myorg",
		State:      models.PRStateOpen,
		HeadBranch: "feature-b",
		BaseBranch: "feature-a", // Depends on parent
		CreatedAt:  time.Now(),
	}

	// Create stack structure (child is blocked by parent)
	parentNode := &models.StackNode{PR: parentPR, Children: []*models.StackNode{}}
	childNode := &models.StackNode{PR: childPR, Parent: parentNode, Children: []*models.StackNode{}}
	parentNode.Children = append(parentNode.Children, childNode)

	stack := &models.Stack{
		Roots:    []*models.StackNode{parentNode},
		AllNodes: []*models.StackNode{parentNode, childNode},
	}

	// Create stacks map keyed by full name (as categorizer does)
	stacks := map[string]*models.Stack{
		"myorg/myrepo": stack,
	}

	prs := []*models.PR{parentPR, childPR}

	// Render section - this should now correctly look up the stack
	result := RenderSection("MY PRS", "", prs, stacks, SectionOptions{ShowIcons: false, ShowBranches: false})

	// Verify the section renders both PRs
	if !strings.Contains(result, "#1") {
		t.Error("Section should contain parent PR #1")
	}
	if !strings.Contains(result, "#2") {
		t.Error("Section should contain child PR #2")
	}

	// Verify repo grouping uses full name
	if !strings.Contains(result, "[myorg/myrepo]") {
		t.Error("Section should show full repo name [myorg/myrepo]")
	}
}

// TestGroupByRepo_WithOwner tests that groupByRepo uses full name when owner is set
func TestGroupByRepo_WithOwner(t *testing.T) {
	prs := []*models.PR{
		{Number: 1, RepoName: "repo", RepoOwner: "org1"},
		{Number: 2, RepoName: "repo", RepoOwner: "org2"},
		{Number: 3, RepoName: "repo", RepoOwner: "org1"},
	}

	grouped := groupByRepo(prs)

	// Should be grouped by full name, not just repo name
	if _, ok := grouped["org1/repo"]; !ok {
		t.Error("Should have org1/repo group")
	}
	if _, ok := grouped["org2/repo"]; !ok {
		t.Error("Should have org2/repo group")
	}
	if len(grouped["org1/repo"]) != 2 {
		t.Error("org1/repo should have 2 PRs")
	}
	if len(grouped["org2/repo"]) != 1 {
		t.Error("org2/repo should have 1 PR")
	}
}

// TestRenderSection_StackTreeRendering tests that stacked PRs are rendered as a nested tree
func TestRenderSection_StackTreeRendering(t *testing.T) {
	// Note: We don't call DisableColors() here as it can cause flaky tests
	// when followed by tests that depend on style rendering. The tree tests
	// use setupTreeTest which handles this but runs after TestStylesAreDefined.

	// Create a 3-level stack: root -> child -> grandchild
	rootPR := &models.PR{
		Number:     101,
		Title:      "Base feature PR",
		URL:        "https://github.com/myorg/myrepo/pull/101",
		RepoName:   "myrepo",
		RepoOwner:  "myorg",
		State:      models.PRStateOpen,
		HeadBranch: "feature-base",
		BaseBranch: "main",
		CreatedAt:  time.Now(),
	}
	childPR := &models.PR{
		Number:     102,
		Title:      "Child PR",
		URL:        "https://github.com/myorg/myrepo/pull/102",
		RepoName:   "myrepo",
		RepoOwner:  "myorg",
		State:      models.PRStateOpen,
		HeadBranch: "feature-child",
		BaseBranch: "feature-base",
		CreatedAt:  time.Now(),
	}
	grandchildPR := &models.PR{
		Number:     103,
		Title:      "Grandchild PR",
		URL:        "https://github.com/myorg/myrepo/pull/103",
		RepoName:   "myrepo",
		RepoOwner:  "myorg",
		State:      models.PRStateOpen,
		HeadBranch: "feature-grandchild",
		BaseBranch: "feature-child",
		CreatedAt:  time.Now(),
	}

	// Build stack structure
	rootNode := &models.StackNode{PR: rootPR, Depth: 0}
	childNode := &models.StackNode{PR: childPR, Parent: rootNode, Depth: 1}
	grandchildNode := &models.StackNode{PR: grandchildPR, Parent: childNode, Depth: 2}

	childNode.Children = []*models.StackNode{grandchildNode}
	rootNode.Children = []*models.StackNode{childNode}

	stack := &models.Stack{
		Roots:    []*models.StackNode{rootNode},
		AllNodes: []*models.StackNode{rootNode, childNode, grandchildNode},
	}

	stacks := map[string]*models.Stack{
		"myorg/myrepo": stack,
	}

	// Only include the root PR in the section (children should be rendered via stack)
	prs := []*models.PR{rootPR}

	result := RenderSection("MY PRS", "", prs, stacks, SectionOptions{ShowIcons: false, ShowBranches: false})

	// All three PRs should be rendered (root plus its children from the stack)
	if !strings.Contains(result, "#101") {
		t.Error("Section should contain root PR #101")
	}
	if !strings.Contains(result, "#102") {
		t.Error("Section should contain child PR #102 (rendered from stack)")
	}
	if !strings.Contains(result, "#103") {
		t.Error("Section should contain grandchild PR #103 (rendered from stack)")
	}

	// Verify tree order: root should appear before child, child before grandchild
	idx101 := strings.Index(result, "#101")
	idx102 := strings.Index(result, "#102")
	idx103 := strings.Index(result, "#103")

	if idx101 > idx102 || idx102 > idx103 {
		t.Error("PRs should appear in tree order: #101 -> #102 -> #103")
	}
}

// TestRenderSection_StackedAndNonStacked tests rendering when both stacked and non-stacked PRs exist
func TestRenderSection_StackedAndNonStacked(t *testing.T) {
	// Note: We don't call DisableColors() here - see TestRenderSection_StackTreeRendering

	// Create a stacked PR pair
	stackedRoot := &models.PR{
		Number:     1,
		Title:      "Stacked Root",
		URL:        "https://github.com/org/repo/pull/1",
		RepoName:   "repo",
		RepoOwner:  "org",
		State:      models.PRStateOpen,
		HeadBranch: "feature",
		BaseBranch: "main",
		CreatedAt:  time.Now(),
	}
	stackedChild := &models.PR{
		Number:     2,
		Title:      "Stacked Child",
		URL:        "https://github.com/org/repo/pull/2",
		RepoName:   "repo",
		RepoOwner:  "org",
		State:      models.PRStateOpen,
		HeadBranch: "feature-2",
		BaseBranch: "feature",
		CreatedAt:  time.Now(),
	}

	// Create a non-stacked PR
	nonStackedPR := &models.PR{
		Number:     99,
		Title:      "Independent PR",
		URL:        "https://github.com/org/repo/pull/99",
		RepoName:   "repo",
		RepoOwner:  "org",
		State:      models.PRStateOpen,
		HeadBranch: "fix",
		BaseBranch: "main",
		CreatedAt:  time.Now(),
	}

	// Build stack
	rootNode := &models.StackNode{PR: stackedRoot, Depth: 0}
	childNode := &models.StackNode{PR: stackedChild, Parent: rootNode, Depth: 1}
	rootNode.Children = []*models.StackNode{childNode}

	stack := &models.Stack{
		Roots:    []*models.StackNode{rootNode},
		AllNodes: []*models.StackNode{rootNode, childNode},
	}

	stacks := map[string]*models.Stack{
		"org/repo": stack,
	}

	// Include root and non-stacked in the PR list
	prs := []*models.PR{stackedRoot, nonStackedPR}

	result := RenderSection("MY PRS", "", prs, stacks, SectionOptions{ShowIcons: false, ShowBranches: false})

	// All three should appear
	if !strings.Contains(result, "#1") {
		t.Error("Should contain stacked root #1")
	}
	if !strings.Contains(result, "#2") {
		t.Error("Should contain stacked child #2 (from stack tree)")
	}
	if !strings.Contains(result, "#99") {
		t.Error("Should contain non-stacked PR #99")
	}
}

// TestGroupByAuthor tests that PRs are correctly grouped by author
func TestGroupByAuthor(t *testing.T) {
	prs := []*models.PR{
		{Number: 1, Author: "alice", RepoName: "repo-a"},
		{Number: 2, Author: "bob", RepoName: "repo-a"},
		{Number: 3, Author: "alice", RepoName: "repo-b"},
		{Number: 4, Author: "", RepoName: "repo-c"}, // No author - should go to "unknown"
	}

	grouped := groupByAuthor(prs)

	if len(grouped["alice"]) != 2 {
		t.Error("alice should have 2 PRs")
	}
	if len(grouped["bob"]) != 1 {
		t.Error("bob should have 1 PR")
	}
	if len(grouped["unknown"]) != 1 {
		t.Error("unknown should have 1 PR for empty author")
	}
}

// TestSortedAuthorNames tests that author names are sorted alphabetically
func TestSortedAuthorNames(t *testing.T) {
	byAuthor := map[string][]*models.PR{
		"zebra":   {},
		"alice":   {},
		"charlie": {},
	}

	names := sortedAuthorNames(byAuthor)

	if len(names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names))
	}
	if names[0] != "alice" || names[1] != "charlie" || names[2] != "zebra" {
		t.Errorf("Names not sorted: %v", names)
	}
}

// TestRenderSection_GroupByAuthor tests that the section renders correctly when grouped by author
func TestRenderSection_GroupByAuthor(t *testing.T) {
	prs := []*models.PR{
		{
			Number:    1,
			Title:     "Alice's first PR",
			URL:       "https://github.com/org/repo-a/pull/1",
			RepoName:  "repo-a",
			RepoOwner: "org",
			Author:    "alice",
			State:     models.PRStateOpen,
			CreatedAt: time.Now(),
		},
		{
			Number:    2,
			Title:     "Bob's PR",
			URL:       "https://github.com/org/repo-b/pull/2",
			RepoName:  "repo-b",
			RepoOwner: "org",
			Author:    "bob",
			State:     models.PRStateOpen,
			CreatedAt: time.Now(),
		},
		{
			Number:    3,
			Title:     "Alice's second PR",
			URL:       "https://github.com/org/repo-c/pull/3",
			RepoName:  "repo-c",
			RepoOwner: "org",
			Author:    "alice",
			State:     models.PRStateOpen,
			CreatedAt: time.Now(),
		},
	}

	result := RenderSection("TEAM PRS", "", prs, nil, SectionOptions{
		ShowIcons:    false,
		ShowBranches: false,
		GroupBy:      "author",
	})

	// Should have author headers
	if !strings.Contains(result, "[@alice]") {
		t.Error("Section should contain [@alice] author header")
	}
	if !strings.Contains(result, "[@bob]") {
		t.Error("Section should contain [@bob] author header")
	}

	// alice should appear before bob (alphabetical)
	idxAlice := strings.Index(result, "[@alice]")
	idxBob := strings.Index(result, "[@bob]")
	if idxAlice > idxBob {
		t.Error("Authors should be sorted alphabetically (alice before bob)")
	}

	// Should contain all PRs
	if !strings.Contains(result, "#1") {
		t.Error("Section should contain PR #1")
	}
	if !strings.Contains(result, "#2") {
		t.Error("Section should contain PR #2")
	}
	if !strings.Contains(result, "#3") {
		t.Error("Section should contain PR #3")
	}
}

// TestRenderSection_GroupByProject_Default tests default grouping is by project
func TestRenderSection_GroupByProject_Default(t *testing.T) {
	prs := []*models.PR{
		{
			Number:    1,
			Title:     "Test PR",
			URL:       "https://github.com/org/repo/pull/1",
			RepoName:  "repo",
			RepoOwner: "org",
			Author:    "alice",
			State:     models.PRStateOpen,
			CreatedAt: time.Now(),
		},
	}

	// Empty GroupBy should default to project grouping
	result := RenderSection("MY PRS", "", prs, nil, SectionOptions{
		ShowIcons:    false,
		ShowBranches: false,
		GroupBy:      "", // Empty = default to project
	})

	// Should have repo header, not author header
	if !strings.Contains(result, "[org/repo]") {
		t.Error("Default grouping should show repo header")
	}
	// Should NOT have author header as the grouping header
	if strings.Contains(result, "[@alice]") {
		t.Error("Default grouping should NOT have author header as group")
	}
}

// TestRenderSection_GroupByAuthor_WithStacks tests that stacked PRs show parent-child relationships in author mode
func TestRenderSection_GroupByAuthor_WithStacks(t *testing.T) {
	// Create stacked PRs for alice in a single repo
	parentPR := &models.PR{
		Number:     1,
		Title:      "Parent PR",
		URL:        "https://github.com/org/repo/pull/1",
		RepoName:   "repo",
		RepoOwner:  "org",
		Author:     "alice",
		State:      models.PRStateOpen,
		HeadBranch: "feature-a",
		BaseBranch: "main",
		CreatedAt:  time.Now(),
	}
	childPR := &models.PR{
		Number:     2,
		Title:      "Child PR",
		URL:        "https://github.com/org/repo/pull/2",
		RepoName:   "repo",
		RepoOwner:  "org",
		Author:     "alice",
		State:      models.PRStateOpen,
		HeadBranch: "feature-b",
		BaseBranch: "feature-a", // Depends on parent
		CreatedAt:  time.Now(),
	}

	// Create stack structure
	parentNode := &models.StackNode{PR: parentPR, Depth: 0}
	childNode := &models.StackNode{PR: childPR, Parent: parentNode, Depth: 1}
	parentNode.Children = []*models.StackNode{childNode}

	stack := &models.Stack{
		Roots:    []*models.StackNode{parentNode},
		AllNodes: []*models.StackNode{parentNode, childNode},
	}

	stacks := map[string]*models.Stack{
		"org/repo": stack,
	}

	prs := []*models.PR{parentPR, childPR}

	result := RenderSection("TEAM PRS", "", prs, stacks, SectionOptions{
		ShowIcons:    false,
		ShowBranches: false,
		GroupBy:      "author",
	})

	// Should have author header
	if !strings.Contains(result, "[@alice]") {
		t.Error("Author grouping should show [@alice] header")
	}

	// Both PRs should appear
	if !strings.Contains(result, "#1") {
		t.Error("Should contain parent PR #1")
	}
	if !strings.Contains(result, "#2") {
		t.Error("Should contain child PR #2")
	}

	// Parent should appear before child (tree order)
	idx1 := strings.Index(result, "#1")
	idx2 := strings.Index(result, "#2")
	if idx1 > idx2 {
		t.Error("Parent PR #1 should appear before child PR #2 (tree order)")
	}
}

// TestRenderSection_GroupByAuthor_StacksAcrossRepos tests that stacks work when author has PRs in multiple repos
func TestRenderSection_GroupByAuthor_StacksAcrossRepos(t *testing.T) {
	// Alice has:
	// - repo-a: stacked PRs (1 -> 2)
	// - repo-b: independent PR (3)
	stackedRoot := &models.PR{
		Number:     1,
		Title:      "Stacked Root",
		URL:        "https://github.com/org/repo-a/pull/1",
		RepoName:   "repo-a",
		RepoOwner:  "org",
		Author:     "alice",
		State:      models.PRStateOpen,
		HeadBranch: "feature",
		BaseBranch: "main",
		CreatedAt:  time.Now(),
	}
	stackedChild := &models.PR{
		Number:     2,
		Title:      "Stacked Child",
		URL:        "https://github.com/org/repo-a/pull/2",
		RepoName:   "repo-a",
		RepoOwner:  "org",
		Author:     "alice",
		State:      models.PRStateOpen,
		HeadBranch: "feature-2",
		BaseBranch: "feature",
		CreatedAt:  time.Now(),
	}
	independentPR := &models.PR{
		Number:    3,
		Title:     "Independent PR",
		URL:       "https://github.com/org/repo-b/pull/3",
		RepoName:  "repo-b",
		RepoOwner: "org",
		Author:    "alice",
		State:     models.PRStateOpen,
		CreatedAt: time.Now(),
	}

	// Build stack for repo-a
	rootNode := &models.StackNode{PR: stackedRoot, Depth: 0}
	childNode := &models.StackNode{PR: stackedChild, Parent: rootNode, Depth: 1}
	rootNode.Children = []*models.StackNode{childNode}

	stacks := map[string]*models.Stack{
		"org/repo-a": {
			Roots:    []*models.StackNode{rootNode},
			AllNodes: []*models.StackNode{rootNode, childNode},
		},
	}

	prs := []*models.PR{stackedRoot, stackedChild, independentPR}

	result := RenderSection("TEAM PRS", "", prs, stacks, SectionOptions{
		ShowIcons:    false,
		ShowBranches: false,
		GroupBy:      "author",
	})

	// All three PRs should appear
	if !strings.Contains(result, "#1") {
		t.Error("Should contain stacked root #1")
	}
	if !strings.Contains(result, "#2") {
		t.Error("Should contain stacked child #2")
	}
	if !strings.Contains(result, "#3") {
		t.Error("Should contain independent #3")
	}

	// Author header should be present
	if !strings.Contains(result, "[@alice]") {
		t.Error("Should contain author header [@alice]")
	}
}

// TestCountTopLevelItems tests the helper function
func TestCountTopLevelItems(t *testing.T) {
	// Test with no stack
	prs := []*models.PR{{Number: 1}, {Number: 2}, {Number: 3}}
	count := countTopLevelItems(prs, nil)
	if count != 3 {
		t.Errorf("Expected 3 top-level items with no stack, got %d", count)
	}

	// Test with stack where one PR is a child
	parentPR := &models.PR{Number: 1}
	childPR := &models.PR{Number: 2}
	parentNode := &models.StackNode{PR: parentPR}
	childNode := &models.StackNode{PR: childPR, Parent: parentNode}
	parentNode.Children = []*models.StackNode{childNode}

	stack := &models.Stack{
		Roots:    []*models.StackNode{parentNode},
		AllNodes: []*models.StackNode{parentNode, childNode},
	}

	prs2 := []*models.PR{parentPR, childPR, {Number: 3}}
	count = countTopLevelItems(prs2, stack)
	// Expected: 1 stack root + 1 non-stacked = 2 top-level items
	if count != 2 {
		t.Errorf("Expected 2 top-level items (1 stack root + 1 non-stacked), got %d", count)
	}
}

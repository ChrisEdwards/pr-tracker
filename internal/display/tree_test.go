package display

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"prt/internal/models"
)

// setupTreeTest disables colors for consistent test output
func setupTreeTest(t *testing.T) {
	t.Helper()
	DisableColors()
	t.Cleanup(EnableColors)
}

// Helper to create a PR for testing
func testPR(number int, title string) *models.PR {
	return &models.PR{
		Number:     number,
		Title:      title,
		URL:        "https://github.com/org/repo/pull/" + strconv.Itoa(number),
		Author:     "testuser",
		State:      models.PRStateOpen,
		IsDraft:    false,
		CreatedAt:  time.Now().Add(-24 * time.Hour),
		BaseBranch: "main",
		HeadBranch: "feature-" + strconv.Itoa(number),
		CIStatus:   models.CIStatusPassing,
	}
}

func TestRenderStackTree_NilNode(t *testing.T) {
	result := RenderStackTree(nil, false, false)
	if result != "" {
		t.Errorf("expected empty string for nil node, got %q", result)
	}
}

func TestRenderStackTree_NilPR(t *testing.T) {
	node := &models.StackNode{PR: nil}
	result := RenderStackTree(node, false, false)
	if result != "" {
		t.Errorf("expected empty string for nil PR, got %q", result)
	}
}

func TestRenderStackTree_SingleNode(t *testing.T) {
	setupTreeTest(t)

	node := &models.StackNode{
		PR:    testPR(123, "Feature: Auth"),
		Depth: 0,
	}

	result := RenderStackTree(node, false, false)

	// Should contain PR number
	if !strings.Contains(result, "#123") {
		t.Error("expected output to contain PR number")
	}

	// Should contain title
	if !strings.Contains(result, "Feature: Auth") {
		t.Error("expected output to contain title")
	}

	// Should contain URL
	if !strings.Contains(result, "pull/123") {
		t.Errorf("expected output to contain URL path, got: %q", result)
	}
}

func TestRenderStackTree_ParentChild(t *testing.T) {
	setupTreeTest(t)

	parent := &models.StackNode{
		PR:    testPR(100, "Parent PR"),
		Depth: 0,
	}
	child := &models.StackNode{
		PR:     testPR(101, "Child PR"),
		Parent: parent,
		Depth:  1,
	}
	parent.Children = []*models.StackNode{child}

	result := RenderStackTree(parent, false, false)

	// Should contain both PRs
	if !strings.Contains(result, "#100") {
		t.Error("expected output to contain parent PR number")
	}
	if !strings.Contains(result, "#101") {
		t.Error("expected output to contain child PR number")
	}
}

func TestRenderStackTree_MultipleChildren(t *testing.T) {
	setupTreeTest(t)

	parent := &models.StackNode{
		PR:    testPR(100, "Parent PR"),
		Depth: 0,
	}
	child1 := &models.StackNode{
		PR:     testPR(101, "First Child"),
		Parent: parent,
		Depth:  1,
	}
	child2 := &models.StackNode{
		PR:     testPR(102, "Second Child"),
		Parent: parent,
		Depth:  1,
	}
	parent.Children = []*models.StackNode{child1, child2}

	result := RenderStackTree(parent, false, false)

	// Should contain all PRs
	if !strings.Contains(result, "#100") {
		t.Error("expected parent PR")
	}
	if !strings.Contains(result, "#101") {
		t.Error("expected first child PR")
	}
	if !strings.Contains(result, "#102") {
		t.Error("expected second child PR")
	}
}

func TestRenderStackTree_DeepNesting(t *testing.T) {
	setupTreeTest(t)

	// Create a 3-level deep stack: root -> child -> grandchild
	root := &models.StackNode{
		PR:    testPR(1, "Root"),
		Depth: 0,
	}
	child := &models.StackNode{
		PR:     testPR(2, "Child"),
		Parent: root,
		Depth:  1,
	}
	grandchild := &models.StackNode{
		PR:     testPR(3, "Grandchild"),
		Parent: child,
		Depth:  2,
	}
	child.Children = []*models.StackNode{grandchild}
	root.Children = []*models.StackNode{child}

	result := RenderStackTree(root, false, false)

	// Should contain all PRs
	if !strings.Contains(result, "Root") {
		t.Error("expected root PR title")
	}
	if !strings.Contains(result, "Child") {
		t.Error("expected child PR title")
	}
	if !strings.Contains(result, "Grandchild") {
		t.Error("expected grandchild PR title")
	}

	// Verify tree structure by checking PR numbers appear in order
	rootIdx := strings.Index(result, "#1 ")
	childIdx := strings.Index(result, "#2 ")
	grandchildIdx := strings.Index(result, "#3 ")

	if rootIdx == -1 || childIdx == -1 || grandchildIdx == -1 {
		t.Fatal("expected all PR numbers to appear")
	}

	if !(rootIdx < childIdx && childIdx < grandchildIdx) {
		t.Error("expected PRs to appear in tree order: root, child, grandchild")
	}
}

func TestRenderStackTree_WithIcons(t *testing.T) {
	setupTreeTest(t)

	node := &models.StackNode{
		PR:    testPR(123, "Test PR"),
		Depth: 0,
	}

	result := RenderStackTree(node, true, false)

	// Should contain the PR info
	if !strings.Contains(result, "#123") {
		t.Error("expected output to contain PR number")
	}
}

func TestRenderStackTree_WithBranches(t *testing.T) {
	setupTreeTest(t)

	node := &models.StackNode{
		PR:    testPR(123, "Test PR"),
		Depth: 0,
	}

	result := RenderStackTree(node, false, true)

	// Should contain branch info
	if !strings.Contains(result, "feature-123") {
		t.Error("expected output to contain head branch")
	}
}

func TestRenderFullStack_Nil(t *testing.T) {
	result := RenderFullStack(nil, false, false)
	if result != "" {
		t.Errorf("expected empty string for nil stack, got %q", result)
	}
}

func TestRenderFullStack_Empty(t *testing.T) {
	stack := &models.Stack{
		Roots:    []*models.StackNode{},
		AllNodes: []*models.StackNode{},
	}
	result := RenderFullStack(stack, false, false)
	if result != "" {
		t.Errorf("expected empty string for empty stack, got %q", result)
	}
}

func TestRenderFullStack_SingleRoot(t *testing.T) {
	setupTreeTest(t)

	node := &models.StackNode{
		PR:    testPR(100, "Single Root"),
		Depth: 0,
	}
	stack := &models.Stack{
		Roots:    []*models.StackNode{node},
		AllNodes: []*models.StackNode{node},
	}

	result := RenderFullStack(stack, false, false)

	if !strings.Contains(result, "#100") {
		t.Error("expected output to contain root PR")
	}
}

func TestRenderFullStack_MultipleRoots(t *testing.T) {
	setupTreeTest(t)

	root1 := &models.StackNode{
		PR:    testPR(100, "First Root"),
		Depth: 0,
	}
	root2 := &models.StackNode{
		PR:    testPR(200, "Second Root"),
		Depth: 0,
	}
	stack := &models.Stack{
		Roots:    []*models.StackNode{root1, root2},
		AllNodes: []*models.StackNode{root1, root2},
	}

	result := RenderFullStack(stack, false, false)

	// Should contain both roots
	if !strings.Contains(result, "#100") {
		t.Error("expected first root PR")
	}
	if !strings.Contains(result, "#200") {
		t.Error("expected second root PR")
	}
}

func TestRenderOrphanIndicator(t *testing.T) {
	t.Run("without icons", func(t *testing.T) {
		result := RenderOrphanIndicator(false)
		if !strings.Contains(result, "orphan") {
			t.Error("expected orphan text")
		}
	})

	t.Run("with icons", func(t *testing.T) {
		result := RenderOrphanIndicator(true)
		if !strings.Contains(result, "orphan") {
			t.Error("expected orphan text")
		}
	})
}

func TestRenderBlockedIndicator(t *testing.T) {
	t.Run("without icons", func(t *testing.T) {
		result := RenderBlockedIndicator(123, false)
		if !strings.Contains(result, "blocked") {
			t.Error("expected blocked text")
		}
		if !strings.Contains(result, "#123") {
			t.Error("expected PR number")
		}
	})

	t.Run("with icons", func(t *testing.T) {
		result := RenderBlockedIndicator(456, true)
		if !strings.Contains(result, "blocked") {
			t.Error("expected blocked text")
		}
		if !strings.Contains(result, "#456") {
			t.Error("expected PR number")
		}
	})
}

func TestRenderStackTree_BlockedPR(t *testing.T) {
	setupTreeTest(t)

	// Parent is open (not merged) so child should be blocked
	parent := &models.StackNode{
		PR:    testPR(100, "Parent"),
		Depth: 0,
	}
	parent.PR.State = models.PRStateOpen

	child := &models.StackNode{
		PR:     testPR(101, "Child"),
		Parent: parent,
		Depth:  1,
	}
	parent.Children = []*models.StackNode{child}

	// The child's IsBlocked() should return true
	if !child.IsBlocked() {
		t.Error("child should be blocked because parent is not merged")
	}

	// Render and verify it doesn't crash
	result := RenderStackTree(parent, false, false)
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestRenderStackTree_OrphanPR(t *testing.T) {
	setupTreeTest(t)

	// Parent was merged, child is orphan
	parent := &models.StackNode{
		PR:    testPR(100, "Parent"),
		Depth: 0,
	}
	parent.PR.State = models.PRStateMerged

	child := &models.StackNode{
		PR:       testPR(101, "Orphan Child"),
		Parent:   parent,
		Depth:    1,
		IsOrphan: true,
	}
	parent.Children = []*models.StackNode{child}

	// Orphan should not be blocked (parent is merged)
	if child.IsBlocked() {
		t.Error("orphan child should not be blocked because parent is merged")
	}

	// Render and verify
	result := RenderStackTree(parent, false, false)
	if !strings.Contains(result, "#101") {
		t.Error("expected orphan child in output")
	}
}

func TestRenderStackTree_ComplexTree(t *testing.T) {
	setupTreeTest(t)

	// Create a complex tree:
	//       root
	//      /    \
	//   child1  child2
	//    /
	// grandchild

	root := &models.StackNode{PR: testPR(1, "Root"), Depth: 0}
	child1 := &models.StackNode{PR: testPR(2, "Child 1"), Parent: root, Depth: 1}
	child2 := &models.StackNode{PR: testPR(3, "Child 2"), Parent: root, Depth: 1}
	grandchild := &models.StackNode{PR: testPR(4, "Grandchild"), Parent: child1, Depth: 2}

	child1.Children = []*models.StackNode{grandchild}
	root.Children = []*models.StackNode{child1, child2}

	result := RenderStackTree(root, false, false)

	// All PRs should be present
	for i := 1; i <= 4; i++ {
		if !strings.Contains(result, "#"+strconv.Itoa(i)) {
			t.Errorf("expected PR #%d in output", i)
		}
	}
}

package stacks

import (
	"testing"
	"time"

	"prt/internal/models"
)

// Helper to create a test PR
func testPR(number int, head, base string) *models.PR {
	return &models.PR{
		Number:     number,
		Title:      "Test PR",
		HeadBranch: head,
		BaseBranch: base,
		State:      models.PRStateOpen,
		CreatedAt:  time.Now(),
	}
}

func TestDetectStacks_Empty(t *testing.T) {
	stack := DetectStacks(nil)
	if !stack.IsEmpty() {
		t.Error("expected empty stack for nil input")
	}

	stack = DetectStacks([]*models.PR{})
	if !stack.IsEmpty() {
		t.Error("expected empty stack for empty input")
	}
}

func TestDetectStacks_NoPRsStacked(t *testing.T) {
	// All PRs target main - no stacks
	prs := []*models.PR{
		testPR(1, "feature-a", "main"),
		testPR(2, "feature-b", "main"),
		testPR(3, "fix-bug", "main"),
	}

	stack := DetectStacks(prs)

	if stack.Size() != 3 {
		t.Errorf("expected 3 nodes, got %d", stack.Size())
	}

	if len(stack.Roots) != 3 {
		t.Errorf("expected 3 roots, got %d", len(stack.Roots))
	}

	// No PR should have children or parents
	for _, node := range stack.AllNodes {
		if node.Parent != nil {
			t.Errorf("PR #%d should not have a parent", node.PR.Number)
		}
		if len(node.Children) > 0 {
			t.Errorf("PR #%d should not have children", node.PR.Number)
		}
		if node.Depth != 0 {
			t.Errorf("PR #%d depth should be 0, got %d", node.PR.Number, node.Depth)
		}
	}
}

func TestDetectStacks_SimpleStack(t *testing.T) {
	// PR 2 is stacked on PR 1
	prs := []*models.PR{
		testPR(1, "feature-auth", "main"),
		testPR(2, "feature-auth-tests", "feature-auth"),
	}

	stack := DetectStacks(prs)

	if stack.Size() != 2 {
		t.Errorf("expected 2 nodes, got %d", stack.Size())
	}

	if len(stack.Roots) != 1 {
		t.Errorf("expected 1 root, got %d", len(stack.Roots))
	}

	// Find PR 1 and PR 2 nodes
	var node1, node2 *models.StackNode
	for _, n := range stack.AllNodes {
		if n.PR.Number == 1 {
			node1 = n
		}
		if n.PR.Number == 2 {
			node2 = n
		}
	}

	if node1 == nil || node2 == nil {
		t.Fatal("could not find expected nodes")
	}

	// PR 1 should be root
	if node1.Parent != nil {
		t.Error("PR 1 should be root (no parent)")
	}
	if node1.Depth != 0 {
		t.Errorf("PR 1 depth should be 0, got %d", node1.Depth)
	}

	// PR 1 should have PR 2 as child
	if len(node1.Children) != 1 || node1.Children[0].PR.Number != 2 {
		t.Error("PR 1 should have PR 2 as its only child")
	}

	// PR 2 should have PR 1 as parent
	if node2.Parent != node1 {
		t.Error("PR 2 should have PR 1 as parent")
	}
	if node2.Depth != 1 {
		t.Errorf("PR 2 depth should be 1, got %d", node2.Depth)
	}
}

func TestDetectStacks_DeepStack(t *testing.T) {
	// PR 1 -> PR 2 -> PR 3 -> PR 4 (4 levels deep)
	prs := []*models.PR{
		testPR(1, "feat-1", "main"),
		testPR(2, "feat-2", "feat-1"),
		testPR(3, "feat-3", "feat-2"),
		testPR(4, "feat-4", "feat-3"),
	}

	stack := DetectStacks(prs)

	if len(stack.Roots) != 1 {
		t.Errorf("expected 1 root, got %d", len(stack.Roots))
	}

	// Verify depths
	depths := make(map[int]int)
	for _, node := range stack.AllNodes {
		depths[node.PR.Number] = node.Depth
	}

	expected := map[int]int{1: 0, 2: 1, 3: 2, 4: 3}
	for pr, expectedDepth := range expected {
		if depths[pr] != expectedDepth {
			t.Errorf("PR #%d depth = %d, want %d", pr, depths[pr], expectedDepth)
		}
	}
}

func TestDetectStacks_DiamondPattern(t *testing.T) {
	// PR 2 and PR 3 both target PR 1's branch
	prs := []*models.PR{
		testPR(1, "feature-base", "main"),
		testPR(2, "feature-a", "feature-base"),
		testPR(3, "feature-b", "feature-base"),
	}

	stack := DetectStacks(prs)

	if len(stack.Roots) != 1 {
		t.Errorf("expected 1 root, got %d", len(stack.Roots))
	}

	// Find PR 1 node
	var node1 *models.StackNode
	for _, n := range stack.AllNodes {
		if n.PR.Number == 1 {
			node1 = n
		}
	}

	if node1 == nil {
		t.Fatal("could not find PR 1 node")
	}

	// PR 1 should have 2 children
	if len(node1.Children) != 2 {
		t.Errorf("PR 1 should have 2 children, got %d", len(node1.Children))
	}

	// Both children should have depth 1
	for _, child := range node1.Children {
		if child.Depth != 1 {
			t.Errorf("Child PR #%d depth should be 1, got %d", child.PR.Number, child.Depth)
		}
	}
}

func TestDetectStacks_MultipleIndependentStacks(t *testing.T) {
	// Two independent stacks:
	// Stack 1: PR 1 -> PR 2
	// Stack 2: PR 3 -> PR 4
	prs := []*models.PR{
		testPR(1, "stack1-base", "main"),
		testPR(2, "stack1-child", "stack1-base"),
		testPR(3, "stack2-base", "main"),
		testPR(4, "stack2-child", "stack2-base"),
	}

	stack := DetectStacks(prs)

	if len(stack.Roots) != 2 {
		t.Errorf("expected 2 roots, got %d", len(stack.Roots))
	}

	// Both roots should have 1 child each
	for _, root := range stack.Roots {
		if len(root.Children) != 1 {
			t.Errorf("Root PR #%d should have 1 child, got %d",
				root.PR.Number, len(root.Children))
		}
	}
}

func TestDetectStacks_MixedStackedAndNonStacked(t *testing.T) {
	// PR 1 -> PR 2 (stacked)
	// PR 3 (standalone)
	prs := []*models.PR{
		testPR(1, "feature", "main"),
		testPR(2, "feature-extra", "feature"),
		testPR(3, "bugfix", "main"),
	}

	stack := DetectStacks(prs)

	if len(stack.Roots) != 2 {
		t.Errorf("expected 2 roots, got %d", len(stack.Roots))
	}

	// Find stacked vs non-stacked
	stacked := FindStackedPRs(stack)
	if len(stacked) != 2 {
		t.Errorf("expected 2 stacked PRs, got %d", len(stacked))
	}
}

func TestDetectStacks_Ordering(t *testing.T) {
	// Insert in random order, verify sorted output
	prs := []*models.PR{
		testPR(5, "pr5", "main"),
		testPR(1, "pr1", "main"),
		testPR(3, "pr3", "main"),
	}

	stack := DetectStacks(prs)

	// Roots should be sorted by PR number
	if stack.Roots[0].PR.Number != 1 {
		t.Errorf("first root should be PR 1, got PR %d", stack.Roots[0].PR.Number)
	}
	if stack.Roots[1].PR.Number != 3 {
		t.Errorf("second root should be PR 3, got PR %d", stack.Roots[1].PR.Number)
	}
	if stack.Roots[2].PR.Number != 5 {
		t.Errorf("third root should be PR 5, got PR %d", stack.Roots[2].PR.Number)
	}

	// AllNodes should also be sorted
	for i := 1; i < len(stack.AllNodes); i++ {
		if stack.AllNodes[i-1].PR.Number >= stack.AllNodes[i].PR.Number {
			t.Error("AllNodes should be sorted by PR number")
		}
	}
}

func TestFindStackedPRs(t *testing.T) {
	prs := []*models.PR{
		testPR(1, "feature", "main"),      // Has child, so stacked
		testPR(2, "feature-ext", "feature"), // Has parent, so stacked
		testPR(3, "standalone", "main"),   // No stack
	}

	stack := DetectStacks(prs)
	stacked := FindStackedPRs(stack)

	if len(stacked) != 2 {
		t.Errorf("expected 2 stacked PRs, got %d", len(stacked))
	}

	// Check that standalone PR is not included
	for _, node := range stacked {
		if node.PR.Number == 3 {
			t.Error("standalone PR should not be in stacked list")
		}
	}
}

func TestGetStackForPR(t *testing.T) {
	prs := []*models.PR{
		testPR(1, "feature", "main"),
		testPR(2, "feature-ext", "feature"),
		testPR(3, "other", "main"),
	}

	stack := DetectStacks(prs)

	// PR 2's root should be PR 1
	root := GetStackForPR(stack, 2)
	if root == nil {
		t.Fatal("expected to find stack for PR 2")
	}
	if root.PR.Number != 1 {
		t.Errorf("PR 2's root should be PR 1, got PR %d", root.PR.Number)
	}

	// PR 1 is its own root
	root = GetStackForPR(stack, 1)
	if root == nil || root.PR.Number != 1 {
		t.Error("PR 1 should be its own root")
	}

	// Non-existent PR
	root = GetStackForPR(stack, 999)
	if root != nil {
		t.Error("non-existent PR should return nil")
	}
}

func TestCountBlockedPRs(t *testing.T) {
	prs := []*models.PR{
		testPR(1, "feature", "main"),
		testPR(2, "feature-ext", "feature"),
		testPR(3, "other", "main"),
	}
	// PR 1 is open, so PR 2 is blocked

	stack := DetectStacks(prs)
	blocked := CountBlockedPRs(stack)

	if blocked != 1 {
		t.Errorf("expected 1 blocked PR, got %d", blocked)
	}
}

func TestCountBlockedPRs_ParentMerged(t *testing.T) {
	pr1 := testPR(1, "feature", "main")
	pr1.State = models.PRStateMerged // Parent is merged
	prs := []*models.PR{
		pr1,
		testPR(2, "feature-ext", "feature"),
	}

	stack := DetectStacks(prs)
	blocked := CountBlockedPRs(stack)

	if blocked != 0 {
		t.Errorf("expected 0 blocked PRs (parent is merged), got %d", blocked)
	}
}

func TestChildrenSorting(t *testing.T) {
	// Parent with multiple children - verify children are sorted
	prs := []*models.PR{
		testPR(1, "base", "main"),
		testPR(5, "child-5", "base"),
		testPR(3, "child-3", "base"),
		testPR(4, "child-4", "base"),
		testPR(2, "child-2", "base"),
	}

	stack := DetectStacks(prs)

	// Find parent node
	var parent *models.StackNode
	for _, node := range stack.Roots {
		if node.PR.Number == 1 {
			parent = node
			break
		}
	}

	if parent == nil {
		t.Fatal("could not find parent node")
	}

	// Children should be sorted: 2, 3, 4, 5
	expected := []int{2, 3, 4, 5}
	for i, child := range parent.Children {
		if child.PR.Number != expected[i] {
			t.Errorf("child %d should be PR %d, got PR %d", i, expected[i], child.PR.Number)
		}
	}
}

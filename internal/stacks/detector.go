// Package stacks implements detection and management of stacked PRs.
// Stacked PRs are PRs that depend on other PRs (targeting another PR's branch
// instead of main/master).
package stacks

import (
	"sort"

	"prt/internal/models"
)

// DetectStacks analyzes a set of PRs and builds a Stack representing their
// parent-child relationships. A PR is considered a "child" of another PR if
// its base branch matches the parent's head branch.
//
// Example:
//
//	PR_A: feature-auth -> main (head=feature-auth, base=main)
//	PR_B: feature-auth-tests -> feature-auth (head=feature-auth-tests, base=feature-auth)
//
//	Since PR_B.base == PR_A.head, PR_B is a child of PR_A.
func DetectStacks(prs []*models.PR) *models.Stack {
	stack := &models.Stack{
		Roots:    []*models.StackNode{},
		AllNodes: []*models.StackNode{},
	}

	if len(prs) == 0 {
		return stack
	}

	// Map: headBranch -> PR (for finding parents)
	headBranchToPR := make(map[string]*models.PR)
	for _, pr := range prs {
		headBranchToPR[pr.HeadBranch] = pr
	}

	// Create nodes for all PRs
	nodes := make(map[int]*models.StackNode)
	for _, pr := range prs {
		nodes[pr.Number] = &models.StackNode{
			PR:       pr,
			Children: []*models.StackNode{},
		}
	}

	// Build parent-child relationships
	for _, pr := range prs {
		// Is there a PR whose head branch is our base branch?
		if parentPR, ok := headBranchToPR[pr.BaseBranch]; ok {
			parentNode := nodes[parentPR.Number]
			childNode := nodes[pr.Number]

			childNode.Parent = parentNode
			parentNode.Children = append(parentNode.Children, childNode)
		}
	}

	// Find roots (nodes with no parent) and collect all nodes
	for _, node := range nodes {
		stack.AllNodes = append(stack.AllNodes, node)
		if node.Parent == nil {
			stack.Roots = append(stack.Roots, node)
		}
	}

	// Calculate depths starting from roots
	for _, root := range stack.Roots {
		setDepths(root, 0)
	}

	// Sort roots by PR number for consistent ordering
	sort.Slice(stack.Roots, func(i, j int) bool {
		return stack.Roots[i].PR.Number < stack.Roots[j].PR.Number
	})

	// Sort children within each node for consistent ordering
	sortChildren(stack.Roots)

	// Sort AllNodes for consistent iteration
	sort.Slice(stack.AllNodes, func(i, j int) bool {
		return stack.AllNodes[i].PR.Number < stack.AllNodes[j].PR.Number
	})

	return stack
}

// setDepths recursively sets the depth of each node in the tree.
func setDepths(node *models.StackNode, depth int) {
	node.Depth = depth
	for _, child := range node.Children {
		setDepths(child, depth+1)
	}
}

// sortChildren recursively sorts children by PR number for consistent output.
func sortChildren(nodes []*models.StackNode) {
	for _, node := range nodes {
		if len(node.Children) > 0 {
			sort.Slice(node.Children, func(i, j int) bool {
				return node.Children[i].PR.Number < node.Children[j].PR.Number
			})
			sortChildren(node.Children)
		}
	}
}

// FindStackedPRs returns only PRs that are part of a stack (have parent or children).
// PRs that target main/master with no children are excluded.
func FindStackedPRs(stack *models.Stack) []*models.StackNode {
	var stacked []*models.StackNode
	for _, node := range stack.AllNodes {
		if node.Parent != nil || len(node.Children) > 0 {
			stacked = append(stacked, node)
		}
	}
	return stacked
}

// GetStackForPR returns the root of the stack containing the given PR.
// Returns nil if the PR is not found in the stack.
func GetStackForPR(stack *models.Stack, prNumber int) *models.StackNode {
	for _, node := range stack.AllNodes {
		if node.PR.Number == prNumber {
			return node.GetRoot()
		}
	}
	return nil
}

// CountBlockedPRs returns the number of PRs that are blocked by unmerged parents.
func CountBlockedPRs(stack *models.Stack) int {
	count := 0
	for _, node := range stack.AllNodes {
		if node.IsBlocked() {
			count++
		}
	}
	return count
}

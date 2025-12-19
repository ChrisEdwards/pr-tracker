// Package display provides terminal rendering for PRT output.
package display

import (
	"strings"

	"prt/internal/models"
)

// RenderStackTree renders a stack node and all its children as a tree.
// The root node is rendered with tree branches, and children are nested.
func RenderStackTree(root *models.StackNode, showIcons, showBranches bool) string {
	if root == nil || root.PR == nil {
		return ""
	}

	var b strings.Builder
	renderNode(&b, root, "", true, showIcons, showBranches)
	return b.String()
}

// RenderFullStack renders all root nodes in a stack.
// Use this when rendering multiple independent trees within a repository.
func RenderFullStack(stack *models.Stack, showIcons, showBranches bool) string {
	if stack == nil || stack.IsEmpty() {
		return ""
	}

	var b strings.Builder
	for i, root := range stack.Roots {
		isLast := i == len(stack.Roots)-1
		renderNode(&b, root, "", isLast, showIcons, showBranches)
	}
	return b.String()
}

// renderNode recursively renders a stack node with proper tree formatting.
func renderNode(b *strings.Builder, node *models.StackNode, prefix string, isLast bool, showIcons, showBranches bool) {
	if node == nil || node.PR == nil {
		return
	}

	// Determine branch character
	branch := TreeBranch
	if isLast {
		branch = TreeLastBranch
	}

	// Style the tree characters
	styledBranch := TreeStyle.Render(branch)

	// Determine if this PR is blocked (has unmerged parent)
	isBlocked := node.IsBlocked()

	// Render the PR with tree prefix
	prOutput := RenderPR(node.PR, prefix+styledBranch+" ", showIcons, showBranches, isBlocked)
	b.WriteString(prOutput)

	// Calculate prefix for children
	childPrefix := prefix
	if isLast {
		childPrefix += TreeIndent
	} else {
		childPrefix += TreeStyle.Render(TreeVertical) + "   "
	}

	// Render children recursively
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		renderNode(b, child, childPrefix, isLastChild, showIcons, showBranches)
	}
}

// RenderOrphanIndicator returns a styled indicator for orphan PRs.
// Orphan PRs are those whose parent branch was merged but the PR still targets it.
func RenderOrphanIndicator(showIcons bool) string {
	if showIcons {
		return MetaStyle.Render("(orphan " + IconBlocked + ")")
	}
	return MetaStyle.Render("(orphan)")
}

// RenderBlockedIndicator returns a styled indicator for blocked PRs.
func RenderBlockedIndicator(parentNumber int, showIcons bool) string {
	if showIcons {
		return BlockedStyle.Render("(blocked by #" + itoa(parentNumber) + " " + IconBlocked + ")")
	}
	return BlockedStyle.Render("(blocked by #" + itoa(parentNumber) + ")")
}

// itoa converts an int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

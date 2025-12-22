// Package display provides terminal rendering for PRT output.
package display

import (
	"fmt"
	"sort"
	"strings"

	"prt/internal/config"
	"prt/internal/models"
)

// RenderSectionHeader renders a section header with optional icon.
func RenderSectionHeader(icon, title string, showIcons bool) string {
	if showIcons && icon != "" {
		return HeaderStyle.Render(fmt.Sprintf("%s %s", icon, title))
	}
	return HeaderStyle.Render(title)
}

// SectionOptions configures how a section is rendered.
type SectionOptions struct {
	ShowIcons    bool
	ShowBranches bool
	GroupBy      string // "project" (default) or "author"
}

// RenderSection renders a complete section with header and PRs grouped by repository or author.
// The stacks map provides stack information for determining blocked status.
func RenderSection(title string, icon string, prs []*models.PR, stacks map[string]*models.Stack, opts SectionOptions) string {
	var b strings.Builder

	// Header
	b.WriteString(RenderSectionHeader(icon, title, opts.ShowIcons))
	b.WriteString("\n\n")

	// Empty state
	if len(prs) == 0 {
		emptyMsg := "  None"
		if title == "NEEDS MY ATTENTION" {
			emptyMsg = "  None - you're all caught up!"
		}
		b.WriteString(EmptyStyle.Render(emptyMsg))
		b.WriteString("\n")
		return b.String()
	}

	// Group by author or project (repo)
	if opts.GroupBy == config.GroupByAuthor {
		renderByAuthor(&b, prs, stacks, opts)
	} else {
		renderByProject(&b, prs, stacks, opts)
	}

	return b.String()
}

// renderByProject renders PRs grouped by repository (the default mode).
func renderByProject(b *strings.Builder, prs []*models.PR, stacks map[string]*models.Stack, opts SectionOptions) {
	byRepo := groupByRepo(prs)
	repoNames := sortedRepoNames(byRepo)

	for _, repoName := range repoNames {
		repoPRs := byRepo[repoName]

		// Repo header (no indent - tree lines start directly below)
		b.WriteString(RepoStyle.Render(fmt.Sprintf("[%s]", repoName)))
		b.WriteString("\n")

		// Render PRs
		stack := stacks[repoName]
		renderPRsInSection(b, repoPRs, stack, opts.ShowIcons, opts.ShowBranches, false)

		b.WriteString("\n")
	}
}

// renderByAuthor renders PRs grouped by author.
func renderByAuthor(b *strings.Builder, prs []*models.PR, stacks map[string]*models.Stack, opts SectionOptions) {
	byAuthor := groupByAuthor(prs)
	authorNames := sortedAuthorNames(byAuthor)

	for _, authorName := range authorNames {
		authorPRs := byAuthor[authorName]

		// Author header
		b.WriteString(AuthorStyle.Render(fmt.Sprintf("[@%s]", authorName)))
		b.WriteString("\n")

		// Render PRs with proper stack relationships across repos
		renderPRsForAuthorGroup(b, authorPRs, stacks, opts)

		b.WriteString("\n")
	}
}

// renderPRsForAuthorGroup renders an author's PRs across multiple repos with proper stack relationships.
// PRs are rendered in their input order (preserving sort), interleaving stacks and non-stacked PRs.
func renderPRsForAuthorGroup(b *strings.Builder, prs []*models.PR, stacks map[string]*models.Stack, opts SectionOptions) {
	// Build maps for stack membership and root lookup across all repos
	stackRootNodes := make(map[string]map[int]*models.StackNode) // repoName -> PR number -> stack root node
	stackChildPRs := make(map[string]map[int]bool)               // repoName -> PR numbers that are children

	for repoName, stack := range stacks {
		if stack == nil {
			continue
		}
		stackRootNodes[repoName] = make(map[int]*models.StackNode)
		stackChildPRs[repoName] = make(map[int]bool)

		// Map each root PR number to its node
		for _, root := range stack.Roots {
			if root.PR != nil {
				stackRootNodes[repoName][root.PR.Number] = root
			}
		}

		// Identify all children (PRs in stacks that are NOT roots)
		for _, node := range stack.AllNodes {
			if node.PR != nil {
				if _, isRoot := stackRootNodes[repoName][node.PR.Number]; !isRoot {
					stackChildPRs[repoName][node.PR.Number] = true
				}
			}
		}
	}

	// Count top-level items (roots + non-stacked PRs, excluding children)
	totalItems := 0
	for _, pr := range prs {
		repoName := pr.RepoFullName()
		if childSet, ok := stackChildPRs[repoName]; ok && childSet[pr.Number] {
			continue
		}
		totalItems++
	}

	// Create base render options
	prOpts := PRRenderOptions{
		ShowIcons:               opts.ShowIcons,
		ShowBranches:            opts.ShowBranches,
		ShowRepoInsteadOfAuthor: true,
	}

	// Render PRs in input order, interleaving stacks and non-stacked PRs
	itemIdx := 0
	for _, pr := range prs {
		repoName := pr.RepoFullName()

		// Skip children - they're rendered by their parent stack
		if childSet, ok := stackChildPRs[repoName]; ok && childSet[pr.Number] {
			continue
		}

		isLast := itemIdx == totalItems-1

		// Check if this PR is a stack root in its repo
		if repoRoots, ok := stackRootNodes[repoName]; ok {
			if rootNode, isStackRoot := repoRoots[pr.Number]; isStackRoot {
				// This PR is a stack root - render the entire stack tree
				renderStackNodeInSection(b, rootNode, "", isLast, prOpts)
				itemIdx++
				continue
			}
		}

		// Non-stacked PR - render as single item
		prefix := TreeStyle.Render(TreeBranch) + " "
		if isLast {
			prefix = TreeStyle.Render(TreeLastBranch) + " "
		}

		var continuationPrefix string
		if !isLast {
			continuationPrefix = TreeStyle.Render(TreeVertical) + "   "
		} else {
			continuationPrefix = TreeIndent
		}

		b.WriteString(RenderPRWithContinuation(pr, prefix, continuationPrefix, prOpts))
		itemIdx++
	}
}

// countTopLevelItems counts the number of top-level renderable items (stack roots + non-stacked PRs).
func countTopLevelItems(prs []*models.PR, stack *models.Stack) int {
	stackedPRs := make(map[int]bool)
	stackRootsCount := 0

	if stack != nil {
		prNumberSet := make(map[int]bool)
		for _, pr := range prs {
			prNumberSet[pr.Number] = true
		}

		for _, node := range stack.AllNodes {
			if node.PR != nil {
				stackedPRs[node.PR.Number] = true
			}
		}

		for _, root := range stack.Roots {
			if root.PR != nil && prNumberSet[root.PR.Number] {
				stackRootsCount++
			}
		}
	}

	nonStackedCount := 0
	for _, pr := range prs {
		if !stackedPRs[pr.Number] {
			nonStackedCount++
		}
	}

	return stackRootsCount + nonStackedCount
}

// renderPRsInSection renders a list of PRs within a section.
// It uses stack tree structure for stacked PRs and flat rendering for non-stacked PRs.
// PRs are rendered in their input order (preserving sort), interleaving stacks and non-stacked PRs.
// When showRepoInsteadOfAuthor is true, the repo name is shown instead of author (for author grouping mode).
func renderPRsInSection(b *strings.Builder, prs []*models.PR, stack *models.Stack, showIcons bool, showBranches bool, showRepoInsteadOfAuthor bool) {
	// Build maps for stack membership and root lookup
	stackRootNodes := make(map[int]*models.StackNode) // PR number -> stack root node
	stackChildPRs := make(map[int]bool)               // PR numbers that are children (not roots)

	if stack != nil {
		// Map each root PR number to its node
		for _, root := range stack.Roots {
			if root.PR != nil {
				stackRootNodes[root.PR.Number] = root
			}
		}

		// Identify all children (PRs in stacks that are NOT roots)
		for _, node := range stack.AllNodes {
			if node.PR != nil {
				if _, isRoot := stackRootNodes[node.PR.Number]; !isRoot {
					stackChildPRs[node.PR.Number] = true
				}
			}
		}
	}

	// Count top-level items (roots + non-stacked PRs, excluding children)
	totalItems := 0
	for _, pr := range prs {
		if !stackChildPRs[pr.Number] {
			totalItems++
		}
	}

	// Create base render options
	prOpts := PRRenderOptions{
		ShowIcons:               showIcons,
		ShowBranches:            showBranches,
		ShowRepoInsteadOfAuthor: showRepoInsteadOfAuthor,
	}

	// Render PRs in input order, interleaving stacks and non-stacked PRs
	itemIdx := 0
	for _, pr := range prs {
		// Skip children - they're rendered by their parent stack
		if stackChildPRs[pr.Number] {
			continue
		}

		isLast := itemIdx == totalItems-1

		if rootNode, isStackRoot := stackRootNodes[pr.Number]; isStackRoot {
			// This PR is a stack root - render the entire stack tree
			renderStackNodeInSection(b, rootNode, "", isLast, prOpts)
		} else {
			// Non-stacked PR - render as single item
			prefix := TreeStyle.Render(TreeBranch) + " "
			if isLast {
				prefix = TreeStyle.Render(TreeLastBranch) + " "
			}

			var continuationPrefix string
			if !isLast {
				continuationPrefix = TreeStyle.Render(TreeVertical) + "   "
			} else {
				continuationPrefix = TreeIndent
			}

			b.WriteString(RenderPRWithContinuation(pr, prefix, continuationPrefix, prOpts))
		}
		itemIdx++
	}
}

// renderStackNodeInSection recursively renders a stack node and its children within a section.
// This provides tree-like indentation for stacked PRs.
func renderStackNodeInSection(b *strings.Builder, node *models.StackNode, prefix string, isLast bool, opts PRRenderOptions) {
	if node == nil || node.PR == nil {
		return
	}

	// Style the branch character consistently
	branch := TreeStyle.Render(TreeBranch)
	if isLast {
		branch = TreeStyle.Render(TreeLastBranch)
	}

	// Determine if this PR is blocked (has unmerged parent)
	isBlocked := node.IsBlocked()

	// Calculate continuation prefix for detail lines (status, branches, URL)
	// This needs TWO components at DIFFERENT indentation levels:
	// 1. Vertical at THIS level if there are more siblings below (connects to next sibling)
	// 2. Vertical at CHILD level if this node has children (connects to first child)
	var continuationPrefix string

	// Part 1: Base continuation for this level (sibling connection)
	if !isLast {
		// More siblings below - show vertical bar at this level
		continuationPrefix = prefix + TreeStyle.Render(TreeVertical) + "   "
	} else {
		// Last sibling - just spaces at this level
		continuationPrefix = prefix + TreeIndent
	}

	// Part 2: Child connection (adds vertical at the NEXT indentation level)
	if len(node.Children) > 0 {
		continuationPrefix += TreeStyle.Render(TreeVertical) + "   "
	}

	// Render the PR with tree prefix and continuation for detail lines
	nodeOpts := PRRenderOptions{
		ShowIcons:               opts.ShowIcons,
		ShowBranches:            opts.ShowBranches,
		IsBlocked:               isBlocked,
		ShowRepoInsteadOfAuthor: opts.ShowRepoInsteadOfAuthor,
	}
	prOutput := RenderPRWithContinuation(node.PR, prefix+branch+" ", continuationPrefix, nodeOpts)
	b.WriteString(prOutput)

	// Calculate prefix for children (used in their title lines)
	// Children get the parent's continuation (vertical if not last) plus their own branch
	childPrefix := prefix
	if isLast {
		childPrefix += TreeIndent
	} else {
		childPrefix += TreeStyle.Render(TreeVertical) + "   "
	}

	// Render children recursively
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		renderStackNodeInSection(b, child, childPrefix, isLastChild, opts)
	}
}

// isPRBlocked checks if a PR is blocked based on stack information.
func isPRBlocked(pr *models.PR, stack *models.Stack) bool {
	if stack == nil {
		return false
	}

	// Find the PR's node in the stack
	for _, node := range stack.AllNodes {
		if node.PR != nil && node.PR.Number == pr.Number {
			return node.IsBlocked()
		}
	}

	return false
}

// groupByRepo groups PRs by their repository full name (owner/repo).
func groupByRepo(prs []*models.PR) map[string][]*models.PR {
	result := make(map[string][]*models.PR)
	for _, pr := range prs {
		result[pr.RepoFullName()] = append(result[pr.RepoFullName()], pr)
	}
	return result
}

// sortedRepoNames returns repository names sorted alphabetically.
func sortedRepoNames(byRepo map[string][]*models.PR) []string {
	names := make([]string, 0, len(byRepo))
	for name := range byRepo {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// groupByAuthor groups PRs by their author username.
func groupByAuthor(prs []*models.PR) map[string][]*models.PR {
	result := make(map[string][]*models.PR)
	for _, pr := range prs {
		author := pr.Author
		if author == "" {
			author = "unknown"
		}
		result[author] = append(result[author], pr)
	}
	return result
}

// sortedAuthorNames returns author names sorted alphabetically.
func sortedAuthorNames(byAuthor map[string][]*models.PR) []string {
	names := make([]string, 0, len(byAuthor))
	for name := range byAuthor {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RenderEmptySection renders a section with no content.
func RenderEmptySection(title string, icon string, showIcons bool) string {
	var b strings.Builder
	b.WriteString(RenderSectionHeader(icon, title, showIcons))
	b.WriteString("\n\n")
	b.WriteString(EmptyStyle.Render("  None"))
	b.WriteString("\n")
	return b.String()
}

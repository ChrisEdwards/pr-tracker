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

		// Render PRs (show repo name instead of author since we're grouped by author)
		// Note: stacks are keyed by repo, so we pass nil for stack in author mode
		// Stack visualization doesn't make as much sense when grouped by author
		renderPRsInSection(b, authorPRs, nil, opts.ShowIcons, opts.ShowBranches, true)

		b.WriteString("\n")
	}
}

// renderPRsInSection renders a list of PRs within a section.
// It uses stack tree structure for stacked PRs and flat rendering for non-stacked PRs.
// When showRepoInsteadOfAuthor is true, the repo name is shown instead of author (for author grouping mode).
func renderPRsInSection(b *strings.Builder, prs []*models.PR, stack *models.Stack, showIcons bool, showBranches bool, showRepoInsteadOfAuthor bool) {
	// Build a set of PR numbers that are part of a stack (for filtering)
	stackedPRs := make(map[int]bool)
	var stackRoots []*models.StackNode

	if stack != nil {
		// Find PRs in our list that are stack roots
		prNumberSet := make(map[int]bool)
		for _, pr := range prs {
			prNumberSet[pr.Number] = true
		}

		// Collect all stacked PRs and identify roots in our PR list
		for _, node := range stack.AllNodes {
			if node.PR != nil {
				stackedPRs[node.PR.Number] = true
			}
		}

		// Find roots that are in our PR list
		for _, root := range stack.Roots {
			if root.PR != nil && prNumberSet[root.PR.Number] {
				stackRoots = append(stackRoots, root)
			}
		}
	}

	// Collect non-stacked PRs (PRs not part of any stack)
	var nonStackedPRs []*models.PR
	for _, pr := range prs {
		if !stackedPRs[pr.Number] {
			nonStackedPRs = append(nonStackedPRs, pr)
		}
	}

	// Calculate total items to render for determining last item
	totalItems := len(stackRoots) + len(nonStackedPRs)
	itemIdx := 0

	// Create base render options
	prOpts := PRRenderOptions{
		ShowIcons:               showIcons,
		ShowBranches:            showBranches,
		ShowRepoInsteadOfAuthor: showRepoInsteadOfAuthor,
	}

	// Render stack trees first (each root with its children)
	for _, root := range stackRoots {
		isLast := itemIdx == totalItems-1
		renderStackNodeInSection(b, root, "", isLast, prOpts)
		itemIdx++
	}

	// Render non-stacked PRs
	for _, pr := range nonStackedPRs {
		isLast := itemIdx == totalItems-1
		prefix := TreeStyle.Render(TreeBranch) + " "
		if isLast {
			prefix = TreeStyle.Render(TreeLastBranch) + " "
		}
		b.WriteString(RenderPR(pr, prefix, prOpts))
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


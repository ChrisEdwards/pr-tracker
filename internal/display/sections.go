// Package display provides terminal rendering for PRT output.
package display

import (
	"fmt"
	"sort"
	"strings"

	"prt/internal/models"
)

// RenderSectionHeader renders a section header with optional icon.
func RenderSectionHeader(icon, title string, showIcons bool) string {
	if showIcons && icon != "" {
		return HeaderStyle.Render(fmt.Sprintf("%s %s", icon, title))
	}
	return HeaderStyle.Render(title)
}

// RenderSection renders a complete section with header and PRs grouped by repository.
// The stacks map provides stack information for determining blocked status.
func RenderSection(title string, icon string, prs []*models.PR, stacks map[string]*models.Stack, showIcons bool, showBranches bool) string {
	var b strings.Builder

	// Header
	b.WriteString(RenderSectionHeader(icon, title, showIcons))
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

	// Group by repo
	byRepo := groupByRepo(prs)
	repoNames := sortedRepoNames(byRepo)

	for _, repoName := range repoNames {
		repoPRs := byRepo[repoName]

		// Repo header
		b.WriteString("  ")
		b.WriteString(RepoStyle.Render(fmt.Sprintf("[%s]", repoName)))
		b.WriteString("\n")

		// Render PRs
		stack := stacks[repoName]
		renderPRsInSection(&b, repoPRs, stack, showIcons, showBranches)

		b.WriteString("\n")
	}

	return b.String()
}

// renderPRsInSection renders a list of PRs within a section.
// It uses stack tree structure for stacked PRs and flat rendering for non-stacked PRs.
func renderPRsInSection(b *strings.Builder, prs []*models.PR, stack *models.Stack, showIcons bool, showBranches bool) {
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

	// Render stack trees first (each root with its children)
	for _, root := range stackRoots {
		isLast := itemIdx == totalItems-1
		renderStackNodeInSection(b, root, "", isLast, showIcons, showBranches)
		itemIdx++
	}

	// Render non-stacked PRs
	for _, pr := range nonStackedPRs {
		isLast := itemIdx == totalItems-1
		prefix := TreeBranch
		if isLast {
			prefix = TreeLastBranch
		}
		b.WriteString(RenderPR(pr, prefix, showIcons, showBranches, false))
		itemIdx++
	}
}

// renderStackNodeInSection recursively renders a stack node and its children within a section.
// This provides tree-like indentation for stacked PRs.
func renderStackNodeInSection(b *strings.Builder, node *models.StackNode, prefix string, isLast bool, showIcons bool, showBranches bool) {
	if node == nil || node.PR == nil {
		return
	}

	// Determine branch character for title line
	branch := TreeBranch
	if isLast {
		branch = TreeLastBranch
	}

	// Determine if this PR is blocked (has unmerged parent)
	isBlocked := node.IsBlocked()

	// Calculate continuation prefix for detail lines (status, branches, URL)
	// This shows the vertical tree line if there are more siblings at this level
	var continuationPrefix string
	if isLast {
		continuationPrefix = prefix + TreeIndent // spaces, no more siblings
	} else {
		continuationPrefix = prefix + TreeStyle.Render(TreeVertical) + "   " // vertical line continues
	}

	// Render the PR with tree prefix and continuation for detail lines
	prOutput := RenderPRWithContinuation(node.PR, prefix+branch+" ", continuationPrefix, showIcons, showBranches, isBlocked)
	b.WriteString(prOutput)

	// Calculate prefix for children (used in their title lines)
	childPrefix := prefix
	if isLast {
		childPrefix += TreeIndent
	} else {
		childPrefix += TreeStyle.Render(TreeVertical) + "   "
	}

	// Render children recursively
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		renderStackNodeInSection(b, child, childPrefix, isLastChild, showIcons, showBranches)
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

// RenderEmptySection renders a section with no content.
func RenderEmptySection(title string, icon string, showIcons bool) string {
	var b strings.Builder
	b.WriteString(RenderSectionHeader(icon, title, showIcons))
	b.WriteString("\n\n")
	b.WriteString(EmptyStyle.Render("  None"))
	b.WriteString("\n")
	return b.String()
}

// RenderNoOpenPRsSection renders a special section for repos with no open PRs.
func RenderNoOpenPRsSection(repos []*models.Repository, showIcons bool) string {
	if len(repos) == 0 {
		return ""
	}

	var b strings.Builder

	// Header
	icon := ""
	if showIcons {
		icon = IconNoOpenPRs
	}
	b.WriteString(RenderSectionHeader(icon, "REPOS WITH NO OPEN PRS", showIcons))
	b.WriteString("\n\n")

	// List repos
	for _, repo := range repos {
		b.WriteString(MetaStyle.Render(fmt.Sprintf("  • %s (%s)", repo.Name, repo.Path)))
		b.WriteString("\n")
	}

	return b.String()
}

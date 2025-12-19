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
// It determines blocked status based on stack information.
func renderPRsInSection(b *strings.Builder, prs []*models.PR, stack *models.Stack, showIcons bool, showBranches bool) {
	for i, pr := range prs {
		isLast := i == len(prs)-1
		prefix := TreeBranch
		if isLast {
			prefix = TreeLastBranch
		}

		isBlocked := isPRBlocked(pr, stack)
		b.WriteString(RenderPR(pr, prefix, showIcons, showBranches, isBlocked))
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

// groupByRepo groups PRs by their repository name.
func groupByRepo(prs []*models.PR) map[string][]*models.PR {
	result := make(map[string][]*models.PR)
	for _, pr := range prs {
		result[pr.RepoName] = append(result[pr.RepoName], pr)
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

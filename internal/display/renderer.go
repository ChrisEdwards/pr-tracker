// Package display provides terminal rendering for PRT output.
package display

import (
	"fmt"
	"strings"

	"prt/internal/models"
)

// RenderPR renders a single PR as a formatted row with tree prefix.
// The prefix should be a tree character like TreeBranch or TreeLastBranch.
// If isBlocked is true, the entire PR is rendered with dimmed styling.
func RenderPR(pr *models.PR, prefix string, showIcons bool, showBranches bool, isBlocked bool) string {
	return RenderPRWithContinuation(pr, prefix, "", showIcons, showBranches, isBlocked)
}

// RenderPRWithContinuation renders a PR with a specific continuation prefix for detail lines.
// The continuationPrefix is used for lines 2-4 (status, branches, URL) to maintain tree structure.
// If continuationPrefix is empty, spaces are used (flat list behavior).
func RenderPRWithContinuation(pr *models.PR, prefix string, continuationPrefix string, showIcons bool, showBranches bool, isBlocked bool) string {
	var b strings.Builder

	// Line 1: Number and title
	// Note: prefix contains tree characters (│, └──, etc.) already styled with TreeStyle
	// We must NOT wrap prefix in any style or it will override the tree styling
	b.WriteString(prefix)
	if isBlocked {
		// Blocked PRs: dim the number and title, but preserve tree char styling
		b.WriteString(BlockedStyle.Render(fmt.Sprintf("#%d %s", pr.Number, pr.Title)))
	} else {
		b.WriteString(NumberStyle.Render(fmt.Sprintf("#%d", pr.Number)))
		b.WriteString(" ")
		b.WriteString(pr.Title)
	}
	b.WriteString("\n")

	// Calculate indent for detail lines
	// If we have a continuation prefix (tree context), use it plus spacing
	// Otherwise fall back to spaces based on prefix length
	var indent string
	if continuationPrefix != "" {
		indent = continuationPrefix + "    " // 4 spaces after tree character
	} else {
		indent = strings.Repeat(" ", len(prefix)+4)
	}

	// Line 2: Status details
	b.WriteString(indent)
	b.WriteString(formatStatusLine(pr, showIcons))
	b.WriteString("\n")

	// Line 3: Branch info (optional)
	if showBranches {
		b.WriteString(indent)
		if pr.Author != "" {
			b.WriteString(AuthorStyle.Render(fmt.Sprintf("@%s", pr.Author)))
			b.WriteString(MetaStyle.Render(" · "))
		}
		b.WriteString(BranchStyle.Render(pr.HeadBranch))
		b.WriteString(MetaStyle.Render(" → "))
		b.WriteString(BranchStyle.Render(pr.BaseBranch))
		b.WriteString("\n")
	}

	// Line 4: URL
	b.WriteString(indent)
	b.WriteString(URLStyle.Render(pr.URL))
	b.WriteString("\n")

	return b.String()
}

// formatStatusLine creates the status line showing state, age, CI, and approvals.
func formatStatusLine(pr *models.PR, showIcons bool) string {
	var parts []string

	// State
	state := formatState(pr, showIcons)
	if state != "" {
		parts = append(parts, state)
	}

	// Age
	parts = append(parts, fmt.Sprintf("Created %s", pr.AgeString()))

	// CI Status
	ci := formatCIStatus(pr.CIStatus, showIcons)
	if ci != "" {
		parts = append(parts, ci)
	}

	// Approvals (if any)
	approvals := countApprovals(pr.Reviews)
	if approvals > 0 {
		parts = append(parts, fmt.Sprintf("%d approval%s", approvals, pluralize(approvals)))
	}

	return MetaStyle.Render(strings.Join(parts, " · "))
}

// formatState returns a styled string representing the PR state.
func formatState(pr *models.PR, showIcons bool) string {
	switch pr.EffectiveState() {
	case models.PRStateDraft:
		if showIcons {
			return DraftStyle.Render(IconDraft + " Draft")
		}
		return DraftStyle.Render("Draft")
	case models.PRStateOpen:
		// Check review state
		reviewState := getReviewState(pr)
		switch reviewState {
		case models.ReviewStateApproved:
			if showIcons {
				return ApprovedStyle.Render(IconApproved + " Approved")
			}
			return ApprovedStyle.Render("Approved")
		case models.ReviewStateChangesRequested:
			if showIcons {
				return ChangesRequestedStyle.Render(IconChanges + " Changes requested")
			}
			return ChangesRequestedStyle.Render("Changes requested")
		default:
			if showIcons {
				return NeedsReviewStyle.Render(IconReview + " Waiting review")
			}
			return NeedsReviewStyle.Render("Waiting review")
		}
	case models.PRStateMerged:
		if showIcons {
			return ApprovedStyle.Render(IconMerged + " Merged")
		}
		return ApprovedStyle.Render("Merged")
	case models.PRStateClosed:
		return MetaStyle.Render("Closed")
	default:
		return string(pr.State)
	}
}

// formatCIStatus returns a styled string representing the CI status.
func formatCIStatus(status models.CIStatus, showIcons bool) string {
	switch status {
	case models.CIStatusPassing:
		if showIcons {
			return CIPassingStyle.Render("CI " + IconCIPassing)
		}
		return CIPassingStyle.Render("CI ✓")
	case models.CIStatusFailing:
		if showIcons {
			return CIFailingStyle.Render("CI " + IconCIFailing)
		}
		return CIFailingStyle.Render("CI ✗")
	case models.CIStatusPending:
		if showIcons {
			return CIPendingStyle.Render("CI " + IconCIPending)
		}
		return CIPendingStyle.Render("CI ...")
	case models.CIStatusNone:
		return ""
	default:
		return ""
	}
}

// countApprovals counts the number of approved reviews.
func countApprovals(reviews []models.Review) int {
	count := 0
	for _, r := range reviews {
		if r.State == models.ReviewStateApproved {
			count++
		}
	}
	return count
}

// getReviewState returns the most significant review state for a PR.
// Priority: ChangesRequested > Approved > Pending > None
func getReviewState(pr *models.PR) models.ReviewState {
	hasApproval := false
	for _, r := range pr.Reviews {
		if r.State == models.ReviewStateChangesRequested {
			return models.ReviewStateChangesRequested
		}
		if r.State == models.ReviewStateApproved {
			hasApproval = true
		}
	}
	if hasApproval {
		return models.ReviewStateApproved
	}
	return models.ReviewStateNone
}

// pluralize returns "s" if count != 1, empty string otherwise.
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// RenderPRSimple renders a PR without tree prefix, for use outside of tree context.
func RenderPRSimple(pr *models.PR, showIcons bool, showBranches bool) string {
	return RenderPR(pr, "  ", showIcons, showBranches, false)
}

// RenderOptions configures the output rendering behavior.
type RenderOptions struct {
	ShowIcons    bool // Show emoji icons for sections and status
	ShowBranches bool // Show branch names (head → base)
	ShowOtherPRs bool // Show "Other PRs" section (external contributors, bots)
	NoColor      bool // Disable all color output
	JSON         bool // Output as JSON instead of styled text
}

// Render orchestrates the complete terminal output from a ScanResult.
// This is the main entry point for rendering PRT output.
func Render(result *models.ScanResult, opts RenderOptions) (string, error) {
	if result == nil {
		return "", fmt.Errorf("cannot render nil result")
	}

	// Handle JSON mode
	if opts.JSON {
		return RenderJSON(result)
	}

	// Handle no-color mode
	if opts.NoColor {
		DisableColors()
	}

	var b strings.Builder

	// Header
	b.WriteString(renderHeader())
	b.WriteString("\n\n")

	// My PRs section
	b.WriteString(RenderSection(
		"MY PRS",
		IconMyPRs,
		result.MyPRs,
		result.Stacks,
		opts.ShowIcons,
		opts.ShowBranches,
	))
	b.WriteString("\n")

	// Needs My Attention section
	b.WriteString(RenderSection(
		"NEEDS MY ATTENTION",
		IconNeedsAttention,
		result.NeedsMyAttention,
		result.Stacks,
		opts.ShowIcons,
		opts.ShowBranches,
	))
	b.WriteString("\n")

	// Team PRs section
	b.WriteString(RenderSection(
		"TEAM PRS",
		IconTeam,
		result.TeamPRs,
		result.Stacks,
		opts.ShowIcons,
		opts.ShowBranches,
	))
	b.WriteString("\n")

	// Other PRs section (only if enabled)
	if opts.ShowOtherPRs {
		b.WriteString(RenderSection(
			"OTHER PRS",
			IconOther,
			result.OtherPRs,
			result.Stacks,
			opts.ShowIcons,
			opts.ShowBranches,
		))
		b.WriteString("\n")
	}

	// Footer with summary
	b.WriteString(renderFooter(result))

	return b.String(), nil
}

// renderHeader renders the PRT header with decorative line.
func renderHeader() string {
	title := TitleStyle.Render("PRT")
	separator := strings.Repeat("═", 60)
	return title + " " + separator
}

// renderFooter renders the scan summary footer.
func renderFooter(result *models.ScanResult) string {
	separator := strings.Repeat("═", 65)

	summary := fmt.Sprintf(
		"Scanned %d repos · Found %d PRs · %s",
		result.TotalReposScanned,
		result.TotalPRsFound,
		result.ScanDurationString(),
	)

	return SummaryStyle.Render(separator+"\n"+summary) + "\n"
}

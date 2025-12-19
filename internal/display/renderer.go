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
	var b strings.Builder

	// Line 1: Number and title
	titleLine := fmt.Sprintf("%s #%d %s", prefix, pr.Number, pr.Title)
	if isBlocked {
		b.WriteString(BlockedStyle.Render(titleLine))
	} else {
		b.WriteString(NumberStyle.Render(fmt.Sprintf("%s #%d", prefix, pr.Number)))
		b.WriteString(" ")
		b.WriteString(pr.Title)
	}
	b.WriteString("\n")

	// Calculate indent based on prefix length
	indent := strings.Repeat(" ", len(prefix)+4)

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

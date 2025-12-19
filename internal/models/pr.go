// Package models defines the core domain types for PRT.
// These are pure data structures with no business logic,
// serving as the lingua franca between all other packages.
package models

import (
	"fmt"
	"time"
)

// PRState represents the state of a pull request.
type PRState string

const (
	PRStateOpen   PRState = "OPEN"
	PRStateDraft  PRState = "DRAFT"
	PRStateMerged PRState = "MERGED"
	PRStateClosed PRState = "CLOSED"
)

// CIStatus represents the CI/CD status of a pull request.
type CIStatus string

const (
	CIStatusPassing CIStatus = "passing"
	CIStatusFailing CIStatus = "failing"
	CIStatusPending CIStatus = "pending"
	CIStatusNone    CIStatus = "none"
)

// ReviewState represents the state of a code review.
type ReviewState string

const (
	ReviewStateNone             ReviewState = "NONE"
	ReviewStateApproved         ReviewState = "APPROVED"
	ReviewStateChangesRequested ReviewState = "CHANGES_REQUESTED"
	ReviewStateCommented        ReviewState = "COMMENTED"
	ReviewStatePending          ReviewState = "PENDING"
	ReviewStateDismissed        ReviewState = "DISMISSED"
)

// Review represents a single code review on a PR.
type Review struct {
	Author    string      `json:"author"`
	State     ReviewState `json:"state"`
	Submitted time.Time   `json:"submitted"`
}

// PR represents a GitHub pull request.
type PR struct {
	// Identity
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`

	// Authorship
	Author string `json:"author"`

	// State
	State   PRState `json:"state"`
	IsDraft bool    `json:"is_draft"`

	// Branches
	BaseBranch string `json:"base_branch"` // Target (e.g., "main")
	HeadBranch string `json:"head_branch"` // Source (e.g., "feature-x")

	// Timestamps
	CreatedAt time.Time `json:"created_at"`

	// CI Status
	CIStatus CIStatus `json:"ci_status"`

	// Review Information
	ReviewRequests []string `json:"review_requests"`
	Assignees      []string `json:"assignees"`
	Reviews        []Review `json:"reviews"`

	// Computed (set during categorization)
	IsReviewRequestedFromMe bool        `json:"is_review_requested_from_me"`
	IsAssignedToMe          bool        `json:"is_assigned_to_me"`
	MyReviewStatus          ReviewState `json:"my_review_status"`

	// Repository context (set during aggregation)
	RepoName string `json:"repo_name"`
	RepoPath string `json:"repo_path"`
}

// Age returns the duration since the PR was created.
func (pr *PR) Age() time.Duration {
	return time.Since(pr.CreatedAt)
}

// AgeString returns a human-readable string representing the PR's age.
// Examples: "2d ago", "5h ago", "30m ago", "just now"
func (pr *PR) AgeString() string {
	age := pr.Age()

	days := int(age.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf("%dd ago", days)
	}

	hours := int(age.Hours())
	if hours > 0 {
		return fmt.Sprintf("%dh ago", hours)
	}

	minutes := int(age.Minutes())
	if minutes > 0 {
		return fmt.Sprintf("%dm ago", minutes)
	}

	return "just now"
}

// EffectiveState returns DRAFT if IsDraft is true, otherwise returns State.
// This provides a unified way to check the PR's effective state.
func (pr *PR) EffectiveState() PRState {
	if pr.IsDraft {
		return PRStateDraft
	}
	return pr.State
}

package github

import (
	"encoding/json"
	"fmt"
	"time"

	"prt/internal/models"
)

// ghPR represents the JSON structure returned by `gh pr list --json ...`
type ghPR struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	State             string          `json:"state"`
	IsDraft           bool            `json:"isDraft"`
	CreatedAt         string          `json:"createdAt"`
	BaseRefName       string          `json:"baseRefName"`
	HeadRefName       string          `json:"headRefName"`
	StatusCheckRollup []ghStatusCheck `json:"statusCheckRollup"`
	ReviewRequests    []ghUser        `json:"reviewRequests"`
	Assignees         []ghUser        `json:"assignees"`
	Reviews           []ghReview      `json:"reviews"`
}

// ghStatusCheck represents a CI status check from gh CLI output.
type ghStatusCheck struct {
	Context string `json:"context"`
	State   string `json:"state"`
}

// ghUser represents a user reference from gh CLI output.
type ghUser struct {
	Login string `json:"login"`
}

// ghReview represents a code review from gh CLI output.
type ghReview struct {
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	State       string `json:"state"`
	SubmittedAt string `json:"submittedAt"`
}

// ParsePRList parses the JSON output from `gh pr list --json ...` into PR models.
func ParsePRList(data []byte) ([]*models.PR, error) {
	var ghPRs []ghPR
	if err := json.Unmarshal(data, &ghPRs); err != nil {
		return nil, fmt.Errorf("failed to parse PR list: %w", err)
	}

	prs := make([]*models.PR, 0, len(ghPRs))
	for _, gpr := range ghPRs {
		pr, err := convertPR(gpr)
		if err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	return prs, nil
}

// convertPR converts a ghPR to a models.PR.
func convertPR(gpr ghPR) (*models.PR, error) {
	createdAt, err := time.Parse(time.RFC3339, gpr.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid createdAt %q: %w", gpr.CreatedAt, err)
	}

	// Convert reviewRequests to []string
	reviewRequests := make([]string, len(gpr.ReviewRequests))
	for i, rr := range gpr.ReviewRequests {
		reviewRequests[i] = rr.Login
	}

	// Convert assignees to []string
	assignees := make([]string, len(gpr.Assignees))
	for i, a := range gpr.Assignees {
		assignees[i] = a.Login
	}

	// Convert reviews to []models.Review
	reviews := make([]models.Review, len(gpr.Reviews))
	for i, r := range gpr.Reviews {
		var submitted time.Time
		if r.SubmittedAt != "" {
			submitted, _ = time.Parse(time.RFC3339, r.SubmittedAt)
		}
		reviews[i] = models.Review{
			Author:    r.Author.Login,
			State:     models.ReviewState(r.State),
			Submitted: submitted,
		}
	}

	return &models.PR{
		Number:         gpr.Number,
		Title:          gpr.Title,
		URL:            gpr.URL,
		Author:         gpr.Author.Login,
		State:          models.PRState(gpr.State),
		IsDraft:        gpr.IsDraft,
		BaseBranch:     gpr.BaseRefName,
		HeadBranch:     gpr.HeadRefName,
		CreatedAt:      createdAt,
		CIStatus:       computeCIStatus(gpr.StatusCheckRollup),
		ReviewRequests: reviewRequests,
		Assignees:      assignees,
		Reviews:        reviews,
	}, nil
}

// computeCIStatus determines overall CI status from individual status checks.
// Priority: failing > pending > passing > none
func computeCIStatus(checks []ghStatusCheck) models.CIStatus {
	if len(checks) == 0 {
		return models.CIStatusNone
	}

	hasFailing := false
	hasPending := false

	for _, check := range checks {
		switch check.State {
		case "FAILURE", "ERROR", "CANCELLED", "TIMED_OUT", "ACTION_REQUIRED":
			hasFailing = true
		case "PENDING", "EXPECTED", "QUEUED", "IN_PROGRESS", "WAITING":
			hasPending = true
			// SUCCESS, SKIPPED, NEUTRAL are considered passing
		}
	}

	if hasFailing {
		return models.CIStatusFailing
	}
	if hasPending {
		return models.CIStatusPending
	}
	return models.CIStatusPassing
}

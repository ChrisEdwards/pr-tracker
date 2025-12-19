// Package categorizer provides PR sorting and categorization.
package categorizer

import (
	"sort"

	"prt/internal/config"
	"prt/internal/models"
)

// SortPRs sorts a slice of PRs by creation date.
// Order can be "oldest" (oldest first) or "newest" (newest first).
// Uses stable sort with PR number as secondary key for deterministic ordering.
func SortPRs(prs []*models.PR, order string) {
	sort.SliceStable(prs, func(i, j int) bool {
		// Secondary sort by PR number for stability when times are equal
		if prs[i].CreatedAt.Equal(prs[j].CreatedAt) {
			return prs[i].Number < prs[j].Number
		}

		// Primary sort by creation time
		switch order {
		case config.SortNewest:
			return prs[i].CreatedAt.After(prs[j].CreatedAt)
		default: // SortOldest is default
			return prs[i].CreatedAt.Before(prs[j].CreatedAt)
		}
	})
}

// SortResult sorts all PR categories in a ScanResult.
func SortResult(result *models.ScanResult, order string) {
	SortPRs(result.MyPRs, order)
	SortPRs(result.NeedsMyAttention, order)
	SortPRs(result.TeamPRs, order)
	SortPRs(result.OtherPRs, order)
}

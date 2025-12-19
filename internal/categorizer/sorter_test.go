package categorizer

import (
	"testing"
	"time"

	"prt/internal/config"
	"prt/internal/models"
)

func TestSortPRs_Oldest(t *testing.T) {
	now := time.Now()
	prs := []*models.PR{
		{Number: 3, CreatedAt: now.Add(-1 * time.Hour)}, // middle
		{Number: 1, CreatedAt: now.Add(-3 * time.Hour)}, // oldest
		{Number: 2, CreatedAt: now.Add(-2 * time.Hour)}, // between
	}

	SortPRs(prs, config.SortOldest)

	// Should be oldest first: 1, 2, 3
	expected := []int{1, 2, 3}
	for i, pr := range prs {
		if pr.Number != expected[i] {
			t.Errorf("position %d: got PR #%d, want #%d", i, pr.Number, expected[i])
		}
	}
}

func TestSortPRs_Newest(t *testing.T) {
	now := time.Now()
	prs := []*models.PR{
		{Number: 1, CreatedAt: now.Add(-3 * time.Hour)}, // oldest
		{Number: 3, CreatedAt: now.Add(-1 * time.Hour)}, // newest
		{Number: 2, CreatedAt: now.Add(-2 * time.Hour)}, // middle
	}

	SortPRs(prs, config.SortNewest)

	// Should be newest first: 3, 2, 1
	expected := []int{3, 2, 1}
	for i, pr := range prs {
		if pr.Number != expected[i] {
			t.Errorf("position %d: got PR #%d, want #%d", i, pr.Number, expected[i])
		}
	}
}

func TestSortPRs_StableForEqualTimes(t *testing.T) {
	sameTime := time.Now()
	prs := []*models.PR{
		{Number: 5, CreatedAt: sameTime},
		{Number: 2, CreatedAt: sameTime},
		{Number: 8, CreatedAt: sameTime},
		{Number: 1, CreatedAt: sameTime},
	}

	SortPRs(prs, config.SortOldest)

	// Same time, should sort by PR number: 1, 2, 5, 8
	expected := []int{1, 2, 5, 8}
	for i, pr := range prs {
		if pr.Number != expected[i] {
			t.Errorf("position %d: got PR #%d, want #%d", i, pr.Number, expected[i])
		}
	}
}

func TestSortPRs_StableForEqualTimes_Newest(t *testing.T) {
	sameTime := time.Now()
	prs := []*models.PR{
		{Number: 5, CreatedAt: sameTime},
		{Number: 2, CreatedAt: sameTime},
		{Number: 8, CreatedAt: sameTime},
	}

	SortPRs(prs, config.SortNewest)

	// Same time with newest sort, still sorted by PR number: 2, 5, 8
	expected := []int{2, 5, 8}
	for i, pr := range prs {
		if pr.Number != expected[i] {
			t.Errorf("position %d: got PR #%d, want #%d", i, pr.Number, expected[i])
		}
	}
}

func TestSortPRs_DefaultsToOldest(t *testing.T) {
	now := time.Now()
	prs := []*models.PR{
		{Number: 2, CreatedAt: now.Add(-1 * time.Hour)}, // newer
		{Number: 1, CreatedAt: now.Add(-2 * time.Hour)}, // older
	}

	SortPRs(prs, "invalid") // invalid should default to oldest

	expected := []int{1, 2}
	for i, pr := range prs {
		if pr.Number != expected[i] {
			t.Errorf("position %d: got PR #%d, want #%d", i, pr.Number, expected[i])
		}
	}
}

func TestSortPRs_EmptySlice(t *testing.T) {
	var prs []*models.PR
	SortPRs(prs, config.SortOldest)
	// Should not panic
	if len(prs) != 0 {
		t.Errorf("expected empty slice, got %d items", len(prs))
	}
}

func TestSortPRs_SingleElement(t *testing.T) {
	prs := []*models.PR{{Number: 1, CreatedAt: time.Now()}}
	SortPRs(prs, config.SortOldest)

	if len(prs) != 1 || prs[0].Number != 1 {
		t.Errorf("single element slice modified incorrectly")
	}
}

func TestSortResult(t *testing.T) {
	now := time.Now()
	result := &models.ScanResult{
		MyPRs: []*models.PR{
			{Number: 2, CreatedAt: now.Add(-1 * time.Hour)},
			{Number: 1, CreatedAt: now.Add(-2 * time.Hour)},
		},
		NeedsMyAttention: []*models.PR{
			{Number: 4, CreatedAt: now.Add(-1 * time.Hour)},
			{Number: 3, CreatedAt: now.Add(-2 * time.Hour)},
		},
		TeamPRs: []*models.PR{
			{Number: 6, CreatedAt: now.Add(-1 * time.Hour)},
			{Number: 5, CreatedAt: now.Add(-2 * time.Hour)},
		},
		OtherPRs: []*models.PR{
			{Number: 8, CreatedAt: now.Add(-1 * time.Hour)},
			{Number: 7, CreatedAt: now.Add(-2 * time.Hour)},
		},
	}

	SortResult(result, config.SortOldest)

	// All categories should be sorted oldest first
	categories := []struct {
		name     string
		prs      []*models.PR
		expected []int
	}{
		{"MyPRs", result.MyPRs, []int{1, 2}},
		{"NeedsMyAttention", result.NeedsMyAttention, []int{3, 4}},
		{"TeamPRs", result.TeamPRs, []int{5, 6}},
		{"OtherPRs", result.OtherPRs, []int{7, 8}},
	}

	for _, cat := range categories {
		for i, pr := range cat.prs {
			if pr.Number != cat.expected[i] {
				t.Errorf("%s position %d: got PR #%d, want #%d", cat.name, i, pr.Number, cat.expected[i])
			}
		}
	}
}

func TestSortResult_Newest(t *testing.T) {
	now := time.Now()
	result := &models.ScanResult{
		MyPRs: []*models.PR{
			{Number: 1, CreatedAt: now.Add(-2 * time.Hour)}, // older
			{Number: 2, CreatedAt: now.Add(-1 * time.Hour)}, // newer
		},
	}

	SortResult(result, config.SortNewest)

	if result.MyPRs[0].Number != 2 || result.MyPRs[1].Number != 1 {
		t.Errorf("expected newest first [2, 1], got [%d, %d]",
			result.MyPRs[0].Number, result.MyPRs[1].Number)
	}
}

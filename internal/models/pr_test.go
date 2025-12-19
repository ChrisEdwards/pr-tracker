package models

import (
	"testing"
	"time"
)

func TestPR_Age(t *testing.T) {
	now := time.Now()
	pr := &PR{
		CreatedAt: now.Add(-2 * time.Hour),
	}

	age := pr.Age()
	// Allow some tolerance for test execution time
	if age < 2*time.Hour || age > 2*time.Hour+time.Second {
		t.Errorf("Age() = %v, want approximately 2h", age)
	}
}

func TestPR_AgeString(t *testing.T) {
	tests := []struct {
		name      string
		createdAt time.Time
		want      string
	}{
		{
			name:      "days ago",
			createdAt: time.Now().Add(-3 * 24 * time.Hour),
			want:      "3d ago",
		},
		{
			name:      "hours ago",
			createdAt: time.Now().Add(-5 * time.Hour),
			want:      "5h ago",
		},
		{
			name:      "minutes ago",
			createdAt: time.Now().Add(-30 * time.Minute),
			want:      "30m ago",
		},
		{
			name:      "just now",
			createdAt: time.Now().Add(-10 * time.Second),
			want:      "just now",
		},
		{
			name:      "one day",
			createdAt: time.Now().Add(-25 * time.Hour),
			want:      "1d ago",
		},
		{
			name:      "one hour",
			createdAt: time.Now().Add(-61 * time.Minute),
			want:      "1h ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PR{CreatedAt: tt.createdAt}
			if got := pr.AgeString(); got != tt.want {
				t.Errorf("AgeString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPR_EffectiveState(t *testing.T) {
	tests := []struct {
		name    string
		state   PRState
		isDraft bool
		want    PRState
	}{
		{
			name:    "open PR",
			state:   PRStateOpen,
			isDraft: false,
			want:    PRStateOpen,
		},
		{
			name:    "draft PR",
			state:   PRStateOpen,
			isDraft: true,
			want:    PRStateDraft,
		},
		{
			name:    "merged PR",
			state:   PRStateMerged,
			isDraft: false,
			want:    PRStateMerged,
		},
		{
			name:    "closed PR",
			state:   PRStateClosed,
			isDraft: false,
			want:    PRStateClosed,
		},
		{
			name:    "draft overrides state",
			state:   PRStateMerged,
			isDraft: true,
			want:    PRStateDraft,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PR{
				State:   tt.state,
				IsDraft: tt.isDraft,
			}
			if got := pr.EffectiveState(); got != tt.want {
				t.Errorf("EffectiveState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPRState_Values(t *testing.T) {
	// Verify enum values are as expected
	if PRStateOpen != "OPEN" {
		t.Errorf("PRStateOpen = %v, want OPEN", PRStateOpen)
	}
	if PRStateDraft != "DRAFT" {
		t.Errorf("PRStateDraft = %v, want DRAFT", PRStateDraft)
	}
	if PRStateMerged != "MERGED" {
		t.Errorf("PRStateMerged = %v, want MERGED", PRStateMerged)
	}
	if PRStateClosed != "CLOSED" {
		t.Errorf("PRStateClosed = %v, want CLOSED", PRStateClosed)
	}
}

func TestCIStatus_Values(t *testing.T) {
	if CIStatusPassing != "passing" {
		t.Errorf("CIStatusPassing = %v, want passing", CIStatusPassing)
	}
	if CIStatusFailing != "failing" {
		t.Errorf("CIStatusFailing = %v, want failing", CIStatusFailing)
	}
	if CIStatusPending != "pending" {
		t.Errorf("CIStatusPending = %v, want pending", CIStatusPending)
	}
	if CIStatusNone != "none" {
		t.Errorf("CIStatusNone = %v, want none", CIStatusNone)
	}
}

func TestReviewState_Values(t *testing.T) {
	if ReviewStateNone != "NONE" {
		t.Errorf("ReviewStateNone = %v, want NONE", ReviewStateNone)
	}
	if ReviewStateApproved != "APPROVED" {
		t.Errorf("ReviewStateApproved = %v, want APPROVED", ReviewStateApproved)
	}
	if ReviewStateChangesRequested != "CHANGES_REQUESTED" {
		t.Errorf("ReviewStateChangesRequested = %v, want CHANGES_REQUESTED", ReviewStateChangesRequested)
	}
	if ReviewStateCommented != "COMMENTED" {
		t.Errorf("ReviewStateCommented = %v, want COMMENTED", ReviewStateCommented)
	}
	if ReviewStatePending != "PENDING" {
		t.Errorf("ReviewStatePending = %v, want PENDING", ReviewStatePending)
	}
	if ReviewStateDismissed != "DISMISSED" {
		t.Errorf("ReviewStateDismissed = %v, want DISMISSED", ReviewStateDismissed)
	}
}

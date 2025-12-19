package github

import (
	"testing"
	"time"

	"prt/internal/models"
)

func TestParsePRList(t *testing.T) {
	t.Run("full PR with all fields", func(t *testing.T) {
		data := []byte(`[{
			"number": 402,
			"title": "Feature: Auth",
			"url": "https://github.com/org/repo/pull/402",
			"author": { "login": "jdoe" },
			"state": "OPEN",
			"isDraft": false,
			"createdAt": "2024-12-15T10:30:00Z",
			"baseRefName": "main",
			"headRefName": "feature-auth",
			"statusCheckRollup": [
				{ "context": "ci/build", "state": "SUCCESS" },
				{ "context": "ci/test", "state": "SUCCESS" }
			],
			"reviewRequests": [{ "login": "reviewer1" }],
			"assignees": [{ "login": "assignee1" }],
			"reviews": [{
				"author": { "login": "reviewer1" },
				"state": "APPROVED",
				"submittedAt": "2024-12-16T14:00:00Z"
			}]
		}]`)

		prs, err := ParsePRList(data)
		if err != nil {
			t.Fatalf("ParsePRList() error = %v, want nil", err)
		}

		if len(prs) != 1 {
			t.Fatalf("ParsePRList() returned %d PRs, want 1", len(prs))
		}

		pr := prs[0]

		// Check basic fields
		if pr.Number != 402 {
			t.Errorf("Number = %d, want 402", pr.Number)
		}
		if pr.Title != "Feature: Auth" {
			t.Errorf("Title = %q, want %q", pr.Title, "Feature: Auth")
		}
		if pr.URL != "https://github.com/org/repo/pull/402" {
			t.Errorf("URL = %q, want %q", pr.URL, "https://github.com/org/repo/pull/402")
		}
		if pr.Author != "jdoe" {
			t.Errorf("Author = %q, want %q", pr.Author, "jdoe")
		}
		if pr.State != models.PRStateOpen {
			t.Errorf("State = %q, want %q", pr.State, models.PRStateOpen)
		}
		if pr.IsDraft {
			t.Errorf("IsDraft = true, want false")
		}
		if pr.BaseBranch != "main" {
			t.Errorf("BaseBranch = %q, want %q", pr.BaseBranch, "main")
		}
		if pr.HeadBranch != "feature-auth" {
			t.Errorf("HeadBranch = %q, want %q", pr.HeadBranch, "feature-auth")
		}

		// Check timestamp
		expectedTime, _ := time.Parse(time.RFC3339, "2024-12-15T10:30:00Z")
		if !pr.CreatedAt.Equal(expectedTime) {
			t.Errorf("CreatedAt = %v, want %v", pr.CreatedAt, expectedTime)
		}

		// Check CI status
		if pr.CIStatus != models.CIStatusPassing {
			t.Errorf("CIStatus = %q, want %q", pr.CIStatus, models.CIStatusPassing)
		}

		// Check review requests
		if len(pr.ReviewRequests) != 1 || pr.ReviewRequests[0] != "reviewer1" {
			t.Errorf("ReviewRequests = %v, want [reviewer1]", pr.ReviewRequests)
		}

		// Check assignees
		if len(pr.Assignees) != 1 || pr.Assignees[0] != "assignee1" {
			t.Errorf("Assignees = %v, want [assignee1]", pr.Assignees)
		}

		// Check reviews
		if len(pr.Reviews) != 1 {
			t.Fatalf("Reviews count = %d, want 1", len(pr.Reviews))
		}
		if pr.Reviews[0].Author != "reviewer1" {
			t.Errorf("Reviews[0].Author = %q, want %q", pr.Reviews[0].Author, "reviewer1")
		}
		if pr.Reviews[0].State != models.ReviewStateApproved {
			t.Errorf("Reviews[0].State = %q, want %q", pr.Reviews[0].State, models.ReviewStateApproved)
		}
	})

	t.Run("draft PR", func(t *testing.T) {
		data := []byte(`[{
			"number": 100,
			"title": "WIP: Draft",
			"url": "https://github.com/org/repo/pull/100",
			"author": { "login": "dev" },
			"state": "OPEN",
			"isDraft": true,
			"createdAt": "2024-12-10T08:00:00Z",
			"baseRefName": "main",
			"headRefName": "wip-draft",
			"statusCheckRollup": [],
			"reviewRequests": [],
			"assignees": [],
			"reviews": []
		}]`)

		prs, err := ParsePRList(data)
		if err != nil {
			t.Fatalf("ParsePRList() error = %v", err)
		}

		if len(prs) != 1 {
			t.Fatalf("len(prs) = %d, want 1", len(prs))
		}

		if !prs[0].IsDraft {
			t.Errorf("IsDraft = false, want true")
		}
	})

	t.Run("empty PR list", func(t *testing.T) {
		data := []byte(`[]`)

		prs, err := ParsePRList(data)
		if err != nil {
			t.Fatalf("ParsePRList() error = %v", err)
		}

		if len(prs) != 0 {
			t.Errorf("len(prs) = %d, want 0", len(prs))
		}
	})

	t.Run("multiple PRs", func(t *testing.T) {
		data := []byte(`[
			{
				"number": 1,
				"title": "PR 1",
				"url": "https://github.com/org/repo/pull/1",
				"author": { "login": "user1" },
				"state": "OPEN",
				"isDraft": false,
				"createdAt": "2024-12-01T00:00:00Z",
				"baseRefName": "main",
				"headRefName": "branch-1",
				"statusCheckRollup": [],
				"reviewRequests": [],
				"assignees": [],
				"reviews": []
			},
			{
				"number": 2,
				"title": "PR 2",
				"url": "https://github.com/org/repo/pull/2",
				"author": { "login": "user2" },
				"state": "MERGED",
				"isDraft": false,
				"createdAt": "2024-12-02T00:00:00Z",
				"baseRefName": "main",
				"headRefName": "branch-2",
				"statusCheckRollup": [],
				"reviewRequests": [],
				"assignees": [],
				"reviews": []
			}
		]`)

		prs, err := ParsePRList(data)
		if err != nil {
			t.Fatalf("ParsePRList() error = %v", err)
		}

		if len(prs) != 2 {
			t.Errorf("len(prs) = %d, want 2", len(prs))
		}

		if prs[0].Number != 1 || prs[1].Number != 2 {
			t.Errorf("PR numbers = [%d, %d], want [1, 2]", prs[0].Number, prs[1].Number)
		}

		if prs[1].State != models.PRStateMerged {
			t.Errorf("prs[1].State = %q, want MERGED", prs[1].State)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		data := []byte(`not json`)

		_, err := ParsePRList(data)
		if err == nil {
			t.Error("ParsePRList() expected error for invalid JSON")
		}
	})

	t.Run("invalid createdAt timestamp", func(t *testing.T) {
		data := []byte(`[{
			"number": 1,
			"title": "Test",
			"url": "https://github.com/org/repo/pull/1",
			"author": { "login": "user" },
			"state": "OPEN",
			"isDraft": false,
			"createdAt": "not-a-timestamp",
			"baseRefName": "main",
			"headRefName": "branch",
			"statusCheckRollup": [],
			"reviewRequests": [],
			"assignees": [],
			"reviews": []
		}]`)

		_, err := ParsePRList(data)
		if err == nil {
			t.Error("ParsePRList() expected error for invalid timestamp")
		}
	})

	t.Run("empty arrays handled correctly", func(t *testing.T) {
		data := []byte(`[{
			"number": 1,
			"title": "No Reviews",
			"url": "https://github.com/org/repo/pull/1",
			"author": { "login": "user" },
			"state": "OPEN",
			"isDraft": false,
			"createdAt": "2024-12-01T00:00:00Z",
			"baseRefName": "main",
			"headRefName": "branch",
			"statusCheckRollup": [],
			"reviewRequests": [],
			"assignees": [],
			"reviews": []
		}]`)

		prs, err := ParsePRList(data)
		if err != nil {
			t.Fatalf("ParsePRList() error = %v", err)
		}

		pr := prs[0]
		// CIStatus should be none when no checks
		if pr.CIStatus != models.CIStatusNone {
			t.Errorf("CIStatus = %q, want none when no checks", pr.CIStatus)
		}
		if len(pr.ReviewRequests) != 0 {
			t.Errorf("ReviewRequests should be empty but got %v", pr.ReviewRequests)
		}
		if len(pr.Assignees) != 0 {
			t.Errorf("Assignees should be empty but got %v", pr.Assignees)
		}
		if len(pr.Reviews) != 0 {
			t.Errorf("Reviews should be empty but got %v", pr.Reviews)
		}
	})

	t.Run("review with empty submittedAt", func(t *testing.T) {
		data := []byte(`[{
			"number": 1,
			"title": "Test",
			"url": "https://github.com/org/repo/pull/1",
			"author": { "login": "user" },
			"state": "OPEN",
			"isDraft": false,
			"createdAt": "2024-12-01T00:00:00Z",
			"baseRefName": "main",
			"headRefName": "branch",
			"statusCheckRollup": [],
			"reviewRequests": [],
			"assignees": [],
			"reviews": [{
				"author": { "login": "reviewer" },
				"state": "COMMENTED",
				"submittedAt": ""
			}]
		}]`)

		prs, err := ParsePRList(data)
		if err != nil {
			t.Fatalf("ParsePRList() error = %v", err)
		}

		if len(prs[0].Reviews) != 1 {
			t.Fatal("Expected one review")
		}
		if prs[0].Reviews[0].Author != "reviewer" {
			t.Errorf("Review author = %q, want reviewer", prs[0].Reviews[0].Author)
		}
	})
}

func TestComputeCIStatus(t *testing.T) {
	tests := []struct {
		name   string
		checks []ghStatusCheck
		want   models.CIStatus
	}{
		{
			name:   "no checks",
			checks: []ghStatusCheck{},
			want:   models.CIStatusNone,
		},
		{
			name:   "nil checks",
			checks: nil,
			want:   models.CIStatusNone,
		},
		{
			name: "all passing",
			checks: []ghStatusCheck{
				{Context: "ci/build", State: "SUCCESS"},
				{Context: "ci/test", State: "SUCCESS"},
			},
			want: models.CIStatusPassing,
		},
		{
			name: "one failing",
			checks: []ghStatusCheck{
				{Context: "ci/build", State: "SUCCESS"},
				{Context: "ci/test", State: "FAILURE"},
			},
			want: models.CIStatusFailing,
		},
		{
			name: "one pending with success",
			checks: []ghStatusCheck{
				{Context: "ci/build", State: "SUCCESS"},
				{Context: "ci/test", State: "PENDING"},
			},
			want: models.CIStatusPending,
		},
		{
			name: "failing takes priority over pending",
			checks: []ghStatusCheck{
				{Context: "ci/build", State: "PENDING"},
				{Context: "ci/test", State: "FAILURE"},
			},
			want: models.CIStatusFailing,
		},
		{
			name: "error state is failing",
			checks: []ghStatusCheck{
				{Context: "ci/build", State: "ERROR"},
			},
			want: models.CIStatusFailing,
		},
		{
			name: "cancelled is failing",
			checks: []ghStatusCheck{
				{Context: "ci/build", State: "CANCELLED"},
			},
			want: models.CIStatusFailing,
		},
		{
			name: "timed out is failing",
			checks: []ghStatusCheck{
				{Context: "ci/build", State: "TIMED_OUT"},
			},
			want: models.CIStatusFailing,
		},
		{
			name: "action required is failing",
			checks: []ghStatusCheck{
				{Context: "ci/build", State: "ACTION_REQUIRED"},
			},
			want: models.CIStatusFailing,
		},
		{
			name: "expected is pending",
			checks: []ghStatusCheck{
				{Context: "ci/build", State: "EXPECTED"},
			},
			want: models.CIStatusPending,
		},
		{
			name: "queued is pending",
			checks: []ghStatusCheck{
				{Context: "ci/build", State: "QUEUED"},
			},
			want: models.CIStatusPending,
		},
		{
			name: "in progress is pending",
			checks: []ghStatusCheck{
				{Context: "ci/build", State: "IN_PROGRESS"},
			},
			want: models.CIStatusPending,
		},
		{
			name: "waiting is pending",
			checks: []ghStatusCheck{
				{Context: "ci/build", State: "WAITING"},
			},
			want: models.CIStatusPending,
		},
		{
			name: "skipped is passing",
			checks: []ghStatusCheck{
				{Context: "ci/build", State: "SKIPPED"},
			},
			want: models.CIStatusPassing,
		},
		{
			name: "neutral is passing",
			checks: []ghStatusCheck{
				{Context: "ci/build", State: "NEUTRAL"},
			},
			want: models.CIStatusPassing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeCIStatus(tt.checks)
			if got != tt.want {
				t.Errorf("computeCIStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParsePRList_RealWorldSample(t *testing.T) {
	// This is a more realistic sample that includes various edge cases
	data := []byte(`[
		{
			"number": 123,
			"title": "fix: handle nil pointer in auth flow",
			"url": "https://github.com/example/app/pull/123",
			"author": {"login": "alice"},
			"state": "OPEN",
			"isDraft": false,
			"createdAt": "2024-12-19T09:15:30Z",
			"baseRefName": "main",
			"headRefName": "fix/nil-pointer",
			"statusCheckRollup": [
				{"context": "ci/lint", "state": "SUCCESS"},
				{"context": "ci/test", "state": "SUCCESS"},
				{"context": "ci/build", "state": "SUCCESS"},
				{"context": "security/scan", "state": "SKIPPED"}
			],
			"reviewRequests": [
				{"login": "bob"},
				{"login": "carol"}
			],
			"assignees": [
				{"login": "alice"}
			],
			"reviews": [
				{
					"author": {"login": "bob"},
					"state": "CHANGES_REQUESTED",
					"submittedAt": "2024-12-19T10:00:00Z"
				},
				{
					"author": {"login": "bob"},
					"state": "APPROVED",
					"submittedAt": "2024-12-19T11:30:00Z"
				}
			]
		}
	]`)

	prs, err := ParsePRList(data)
	if err != nil {
		t.Fatalf("ParsePRList() error = %v", err)
	}

	if len(prs) != 1 {
		t.Fatalf("Expected 1 PR, got %d", len(prs))
	}

	pr := prs[0]

	// Verify multiple review requests
	if len(pr.ReviewRequests) != 2 {
		t.Errorf("Expected 2 review requests, got %d", len(pr.ReviewRequests))
	}

	// Verify multiple reviews (history)
	if len(pr.Reviews) != 2 {
		t.Errorf("Expected 2 reviews, got %d", len(pr.Reviews))
	}

	// SKIPPED should still result in passing overall
	if pr.CIStatus != models.CIStatusPassing {
		t.Errorf("CIStatus = %q, want passing", pr.CIStatus)
	}
}

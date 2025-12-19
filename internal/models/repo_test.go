package models

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestRepository_FullName(t *testing.T) {
	tests := []struct {
		name  string
		repo  Repository
		want  string
	}{
		{
			name:  "with owner",
			repo:  Repository{Owner: "myorg", Name: "myrepo"},
			want:  "myorg/myrepo",
		},
		{
			name:  "without owner",
			repo:  Repository{Name: "myrepo"},
			want:  "myrepo",
		},
		{
			name:  "empty owner",
			repo:  Repository{Owner: "", Name: "myrepo"},
			want:  "myrepo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.repo.FullName(); got != tt.want {
				t.Errorf("FullName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_HasPRs(t *testing.T) {
	tests := []struct {
		name string
		repo Repository
		want bool
	}{
		{
			name: "with PRs",
			repo: Repository{PRs: []*PR{{Number: 1}, {Number: 2}}},
			want: true,
		},
		{
			name: "without PRs",
			repo: Repository{PRs: []*PR{}},
			want: false,
		},
		{
			name: "nil PRs",
			repo: Repository{PRs: nil},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.repo.HasPRs(); got != tt.want {
				t.Errorf("HasPRs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanStatus_Values(t *testing.T) {
	if ScanStatusSuccess != "success" {
		t.Errorf("ScanStatusSuccess = %v, want success", ScanStatusSuccess)
	}
	if ScanStatusNoPRs != "no_prs" {
		t.Errorf("ScanStatusNoPRs = %v, want no_prs", ScanStatusNoPRs)
	}
	if ScanStatusError != "error" {
		t.Errorf("ScanStatusError = %v, want error", ScanStatusError)
	}
	if ScanStatusSkipped != "skipped" {
		t.Errorf("ScanStatusSkipped = %v, want skipped", ScanStatusSkipped)
	}
}

func TestRepository_JSONSerialization(t *testing.T) {
	repo := Repository{
		Name:       "prt",
		Path:       "/Users/jdoe/code/prt",
		RemoteURL:  "git@github.com:org/prt.git",
		Owner:      "org",
		PRs:        []*PR{{Number: 1, Title: "Test PR"}},
		ScanError:  errors.New("test error"),
		ScanStatus: ScanStatusSuccess,
	}

	data, err := json.Marshal(repo)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Verify ScanError is NOT in JSON output
	jsonStr := string(data)
	if contains(jsonStr, "test error") {
		t.Error("ScanError should not be serialized to JSON")
	}
	if contains(jsonStr, "scan_error") {
		t.Error("scan_error field should not appear in JSON")
	}

	// Verify other fields are present
	if !contains(jsonStr, "prt") {
		t.Error("Name should be serialized")
	}
	if !contains(jsonStr, "success") {
		t.Error("ScanStatus should be serialized")
	}

	// Verify we can unmarshal
	var decoded Repository
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Name != repo.Name {
		t.Errorf("Name = %v, want %v", decoded.Name, repo.Name)
	}
	if decoded.ScanStatus != repo.ScanStatus {
		t.Errorf("ScanStatus = %v, want %v", decoded.ScanStatus, repo.ScanStatus)
	}
	// ScanError should be nil after unmarshaling
	if decoded.ScanError != nil {
		t.Error("ScanError should be nil after unmarshaling")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

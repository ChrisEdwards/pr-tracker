package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewScanResult(t *testing.T) {
	result := NewScanResult()

	// Verify all slices are initialized (not nil)
	if result.MyPRs == nil {
		t.Error("MyPRs should be initialized")
	}
	if result.NeedsMyAttention == nil {
		t.Error("NeedsMyAttention should be initialized")
	}
	if result.TeamPRs == nil {
		t.Error("TeamPRs should be initialized")
	}
	if result.OtherPRs == nil {
		t.Error("OtherPRs should be initialized")
	}
	if result.ReposWithPRs == nil {
		t.Error("ReposWithPRs should be initialized")
	}
	if result.ReposWithoutPRs == nil {
		t.Error("ReposWithoutPRs should be initialized")
	}
	if result.ReposWithErrors == nil {
		t.Error("ReposWithErrors should be initialized")
	}
	if result.Stacks == nil {
		t.Error("Stacks should be initialized")
	}

	// Verify slices are empty
	if len(result.MyPRs) != 0 {
		t.Error("MyPRs should be empty")
	}
	if len(result.Stacks) != 0 {
		t.Error("Stacks should be empty")
	}
}

func TestScanResult_TotalPRs(t *testing.T) {
	result := NewScanResult()

	if result.TotalPRs() != 0 {
		t.Errorf("TotalPRs() = %d, want 0", result.TotalPRs())
	}

	result.MyPRs = []*PR{{Number: 1}, {Number: 2}}
	result.NeedsMyAttention = []*PR{{Number: 3}}
	result.TeamPRs = []*PR{{Number: 4}, {Number: 5}, {Number: 6}}
	result.OtherPRs = []*PR{{Number: 7}}

	if result.TotalPRs() != 7 {
		t.Errorf("TotalPRs() = %d, want 7", result.TotalPRs())
	}
}

func TestScanResult_HasPRs(t *testing.T) {
	result := NewScanResult()

	if result.HasPRs() {
		t.Error("HasPRs() should return false for empty result")
	}

	result.MyPRs = []*PR{{Number: 1}}
	if !result.HasPRs() {
		t.Error("HasPRs() should return true when MyPRs is not empty")
	}

	result.MyPRs = []*PR{}
	result.OtherPRs = []*PR{{Number: 1}}
	if !result.HasPRs() {
		t.Error("HasPRs() should return true when OtherPRs is not empty")
	}
}

func TestScanResult_HasErrors(t *testing.T) {
	result := NewScanResult()

	if result.HasErrors() {
		t.Error("HasErrors() should return false for empty result")
	}

	result.ReposWithErrors = []*Repository{{Name: "failed-repo"}}
	if !result.HasErrors() {
		t.Error("HasErrors() should return true when ReposWithErrors is not empty")
	}
}

func TestScanResult_TotalRepos(t *testing.T) {
	result := NewScanResult()

	if result.TotalRepos() != 0 {
		t.Errorf("TotalRepos() = %d, want 0", result.TotalRepos())
	}

	result.ReposWithPRs = []*Repository{{Name: "repo1"}, {Name: "repo2"}}
	result.ReposWithoutPRs = []*Repository{{Name: "repo3"}}
	result.ReposWithErrors = []*Repository{{Name: "repo4"}}

	if result.TotalRepos() != 4 {
		t.Errorf("TotalRepos() = %d, want 4", result.TotalRepos())
	}
}

func TestScanResult_ScanDurationString(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "milliseconds",
			duration: 123 * time.Millisecond,
			want:     "123ms",
		},
		{
			name:     "seconds",
			duration: 5 * time.Second,
			want:     "5s",
		},
		{
			name:     "seconds with ms",
			duration: 5*time.Second + 500*time.Millisecond,
			want:     "6s", // Rounds to nearest second
		},
		{
			name:     "minutes",
			duration: 90 * time.Second,
			want:     "1m30s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ScanResult{ScanDuration: tt.duration}
			if got := result.ScanDurationString(); got != tt.want {
				t.Errorf("ScanDurationString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanResult_JSONSerialization(t *testing.T) {
	result := NewScanResult()
	result.MyPRs = []*PR{{Number: 1, Title: "My PR"}}
	result.Username = "testuser"
	result.TotalReposScanned = 5
	result.TotalPRsFound = 10
	result.ScanDuration = 2 * time.Second
	result.Stacks["org/repo"] = &Stack{
		AllNodes: []*StackNode{{PR: &PR{Number: 1}}},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Verify JSON contains expected fields
	jsonStr := string(data)
	if !containsJSON(jsonStr, "my_prs") {
		t.Error("JSON should contain my_prs")
	}
	if !containsJSON(jsonStr, "username") {
		t.Error("JSON should contain username")
	}
	if !containsJSON(jsonStr, "scan_duration_ns") {
		t.Error("JSON should contain scan_duration_ns")
	}
	if !containsJSON(jsonStr, "stacks") {
		t.Error("JSON should contain stacks")
	}

	// Verify we can unmarshal
	var decoded ScanResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if decoded.Username != result.Username {
		t.Errorf("Username = %v, want %v", decoded.Username, result.Username)
	}
	if decoded.TotalReposScanned != result.TotalReposScanned {
		t.Errorf("TotalReposScanned = %v, want %v", decoded.TotalReposScanned, result.TotalReposScanned)
	}
	if len(decoded.MyPRs) != len(result.MyPRs) {
		t.Errorf("MyPRs length = %v, want %v", len(decoded.MyPRs), len(result.MyPRs))
	}
}

func containsJSON(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package display

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"prt/internal/models"
)

func TestRenderJSON_EmptyResult(t *testing.T) {
	result := models.NewScanResult()
	result.Username = "testuser"

	output, err := RenderJSON(result)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Check key fields exist
	if parsed["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got %v", parsed["username"])
	}

	// Check arrays are empty but present
	if myPRs, ok := parsed["my_prs"].([]interface{}); !ok || len(myPRs) != 0 {
		t.Error("Expected empty my_prs array")
	}
}

func TestRenderJSON_WithPRs(t *testing.T) {
	result := models.NewScanResult()
	result.Username = "jdoe"
	result.TotalReposScanned = 2
	result.TotalPRsFound = 3
	result.ScanDuration = 2300 * time.Millisecond

	pr1 := &models.PR{
		Number:     101,
		Title:      "Feature: Add login",
		URL:        "https://github.com/org/repo/pull/101",
		Author:     "jdoe",
		State:      models.PRStateOpen,
		BaseBranch: "main",
		HeadBranch: "feature-login",
		CreatedAt:  time.Now().Add(-24 * time.Hour),
		CIStatus:   models.CIStatusPassing,
	}

	pr2 := &models.PR{
		Number:           102,
		Title:            "Review: Add tests",
		URL:              "https://github.com/org/repo/pull/102",
		Author:           "alice",
		State:            models.PRStateOpen,
		ReviewRequests:   []string{"jdoe"},
		BaseBranch:       "main",
		HeadBranch:       "add-tests",
		CreatedAt:        time.Now().Add(-48 * time.Hour),
		CIStatus:         models.CIStatusFailing,
		MyReviewStatus:   models.ReviewStateNone,
		IsReviewRequestedFromMe: true,
	}

	result.MyPRs = append(result.MyPRs, pr1)
	result.NeedsMyAttention = append(result.NeedsMyAttention, pr2)

	output, err := RenderJSON(result)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var parsed models.ScanResult
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify content
	if len(parsed.MyPRs) != 1 {
		t.Errorf("Expected 1 PR in MyPRs, got %d", len(parsed.MyPRs))
	}
	if parsed.MyPRs[0].Number != 101 {
		t.Errorf("Expected PR #101, got #%d", parsed.MyPRs[0].Number)
	}
	if len(parsed.NeedsMyAttention) != 1 {
		t.Errorf("Expected 1 PR in NeedsMyAttention, got %d", len(parsed.NeedsMyAttention))
	}
	if parsed.TotalReposScanned != 2 {
		t.Errorf("Expected TotalReposScanned=2, got %d", parsed.TotalReposScanned)
	}
	if parsed.ScanDuration != 2300*time.Millisecond {
		t.Errorf("Expected ScanDuration=2300ms, got %v", parsed.ScanDuration)
	}
}

func TestRenderJSON_NilResult(t *testing.T) {
	_, err := RenderJSON(nil)
	if err == nil {
		t.Error("Expected error for nil result")
	}
	if !strings.Contains(err.Error(), "nil result") {
		t.Errorf("Expected error message about nil result, got: %v", err)
	}
}

func TestRenderJSON_IsPrettyPrinted(t *testing.T) {
	result := models.NewScanResult()
	result.Username = "test"

	output, err := RenderJSON(result)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	// Pretty-printed JSON should contain newlines and indentation
	if !strings.Contains(output, "\n") {
		t.Error("Expected pretty-printed JSON with newlines")
	}
	if !strings.Contains(output, "  ") {
		t.Error("Expected pretty-printed JSON with indentation")
	}
}

func TestWriteJSON(t *testing.T) {
	result := models.NewScanResult()
	result.Username = "writer-test"
	result.TotalPRsFound = 5

	var buf bytes.Buffer
	err := WriteJSON(&buf, result)
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	// Verify output
	output := buf.String()
	if !strings.Contains(output, "writer-test") {
		t.Error("Expected username in output")
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}
}

func TestWriteJSON_NilResult(t *testing.T) {
	var buf bytes.Buffer
	err := WriteJSON(&buf, nil)
	if err == nil {
		t.Error("Expected error for nil result")
	}
}

func TestRenderJSONCompact(t *testing.T) {
	result := models.NewScanResult()
	result.Username = "compact-test"
	result.TotalReposScanned = 1

	output, err := RenderJSONCompact(result)
	if err != nil {
		t.Fatalf("RenderJSONCompact failed: %v", err)
	}

	// Compact JSON should NOT contain newlines
	if strings.Contains(output, "\n") {
		t.Error("Expected compact JSON without newlines")
	}

	// Verify it's still valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if parsed["username"] != "compact-test" {
		t.Errorf("Expected username 'compact-test', got %v", parsed["username"])
	}
}

func TestRenderJSONCompact_NilResult(t *testing.T) {
	_, err := RenderJSONCompact(nil)
	if err == nil {
		t.Error("Expected error for nil result")
	}
}

func TestRenderJSON_AllFieldsPresent(t *testing.T) {
	result := models.NewScanResult()
	result.Username = "complete-test"
	result.TotalReposScanned = 5
	result.TotalPRsFound = 12
	result.ScanDuration = 1500 * time.Millisecond

	// Add PRs to all categories
	myPR := &models.PR{Number: 1, Title: "My PR", Author: "complete-test"}
	needsAttentionPR := &models.PR{Number: 2, Title: "Review me", Author: "other"}
	teamPR := &models.PR{Number: 3, Title: "Team PR", Author: "teammate"}
	otherPR := &models.PR{Number: 4, Title: "Other PR", Author: "external"}

	result.MyPRs = append(result.MyPRs, myPR)
	result.NeedsMyAttention = append(result.NeedsMyAttention, needsAttentionPR)
	result.TeamPRs = append(result.TeamPRs, teamPR)
	result.OtherPRs = append(result.OtherPRs, otherPR)

	// Add repos
	result.ReposWithPRs = append(result.ReposWithPRs, &models.Repository{Name: "repo1"})
	result.ReposWithoutPRs = append(result.ReposWithoutPRs, &models.Repository{Name: "repo2"})
	result.ReposWithErrors = append(result.ReposWithErrors, &models.Repository{Name: "repo3"})

	output, err := RenderJSON(result)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	// Verify all top-level fields are present
	requiredFields := []string{
		"my_prs",
		"needs_my_attention",
		"team_prs",
		"other_prs",
		"repos_with_prs",
		"repos_without_prs",
		"repos_with_errors",
		"stacks",
		"total_repos_scanned",
		"total_prs_found",
		"scan_duration_ns",
		"username",
	}

	for _, field := range requiredFields {
		if !strings.Contains(output, `"`+field+`"`) {
			t.Errorf("Expected field %q in JSON output", field)
		}
	}
}

func TestRenderJSON_WorksWithJQ(t *testing.T) {
	// This test verifies the output structure is jq-friendly
	result := models.NewScanResult()
	result.Username = "jq-test"
	result.TotalPRsFound = 3

	pr := &models.PR{
		Number: 42,
		Title:  "JQ Test PR",
		URL:    "https://github.com/test/repo/pull/42",
		Author: "jq-test",
	}
	result.MyPRs = append(result.MyPRs, pr)

	output, err := RenderJSON(result)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	// Parse and verify jq-style access would work
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Test .my_prs[0].number access pattern
	myPRs, ok := parsed["my_prs"].([]interface{})
	if !ok || len(myPRs) == 0 {
		t.Fatal("Could not access my_prs array")
	}

	firstPR, ok := myPRs[0].(map[string]interface{})
	if !ok {
		t.Fatal("Could not access first PR as object")
	}

	// JSON numbers unmarshal as float64
	if number, ok := firstPR["number"].(float64); !ok || int(number) != 42 {
		t.Errorf("Expected PR number 42, got %v", firstPR["number"])
	}

	if url, ok := firstPR["url"].(string); !ok || url != "https://github.com/test/repo/pull/42" {
		t.Errorf("Expected URL, got %v", firstPR["url"])
	}
}

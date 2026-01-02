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

	output, err := RenderJSON(result, JSONOptions{ShowOtherPRs: true})
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

	// Check total_prs is 0
	if parsed["total_prs"].(float64) != 0 {
		t.Errorf("Expected total_prs=0, got %v", parsed["total_prs"])
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
		Number:                  102,
		Title:                   "Review: Add tests",
		URL:                     "https://github.com/org/repo/pull/102",
		Author:                  "alice",
		State:                   models.PRStateOpen,
		ReviewRequests:          []string{"jdoe"},
		BaseBranch:              "main",
		HeadBranch:              "add-tests",
		CreatedAt:               time.Now().Add(-48 * time.Hour),
		CIStatus:                models.CIStatusFailing,
		MyReviewStatus:          models.ReviewStateNone,
		IsReviewRequestedFromMe: true,
	}

	result.MyPRs = append(result.MyPRs, pr1)
	result.NeedsMyAttention = append(result.NeedsMyAttention, pr2)

	output, err := RenderJSON(result, JSONOptions{ShowOtherPRs: true})
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	// Verify it's valid JSON with new structure
	var parsed struct {
		MyPRs            []*models.PR `json:"my_prs"`
		NeedsMyAttention []*models.PR `json:"needs_my_attention"`
		TotalPRs         int          `json:"total_prs"`
		Username         string       `json:"username"`
		ScanSeconds      float64      `json:"scan_seconds"`
	}
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
	if parsed.TotalPRs != 2 {
		t.Errorf("Expected TotalPRs=2, got %d", parsed.TotalPRs)
	}
	if parsed.ScanSeconds < 2.0 || parsed.ScanSeconds > 3.0 {
		t.Errorf("Expected ScanSeconds ~2.3, got %v", parsed.ScanSeconds)
	}
}

func TestRenderJSON_NilResult(t *testing.T) {
	_, err := RenderJSON(nil, JSONOptions{})
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

	output, err := RenderJSON(result, JSONOptions{ShowOtherPRs: true})
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

	output, err := RenderJSON(result, JSONOptions{ShowOtherPRs: true})
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	// Verify all top-level fields are present in clean output
	requiredFields := []string{
		"my_prs",
		"needs_my_attention",
		"team_prs",
		"other_prs",
		"total_prs",
		"username",
		"scan_seconds",
	}

	for _, field := range requiredFields {
		if !strings.Contains(output, `"`+field+`"`) {
			t.Errorf("Expected field %q in JSON output", field)
		}
	}

	// Verify total_prs count is correct
	var parsed map[string]interface{}
	json.Unmarshal([]byte(output), &parsed)
	if parsed["total_prs"].(float64) != 4 {
		t.Errorf("Expected total_prs=4, got %v", parsed["total_prs"])
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

	output, err := RenderJSON(result, JSONOptions{ShowOtherPRs: true})
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

func TestRenderJSON_ShowOtherPRs_True(t *testing.T) {
	result := models.NewScanResult()
	result.Username = "test"
	result.OtherPRs = []*models.PR{
		{Number: 100, Title: "Bot PR", Author: "dependabot"},
	}

	output, err := RenderJSON(result, JSONOptions{ShowOtherPRs: true})
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	var parsed models.ScanResult
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if len(parsed.OtherPRs) != 1 {
		t.Errorf("Expected 1 OtherPR with ShowOtherPRs=true, got %d", len(parsed.OtherPRs))
	}
}

func TestRenderJSON_ShowOtherPRs_False(t *testing.T) {
	result := models.NewScanResult()
	result.Username = "test"
	result.OtherPRs = []*models.PR{
		{Number: 100, Title: "Bot PR", Author: "dependabot"},
	}

	output, err := RenderJSON(result, JSONOptions{ShowOtherPRs: false})
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	var parsed models.ScanResult
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if len(parsed.OtherPRs) != 0 {
		t.Errorf("Expected 0 OtherPRs with ShowOtherPRs=false, got %d", len(parsed.OtherPRs))
	}
}

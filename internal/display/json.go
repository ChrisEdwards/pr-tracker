// Package display provides terminal rendering for PRT output.
package display

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"prt/internal/models"
)

// JSONOptions controls what is included in JSON output.
type JSONOptions struct {
	ShowOtherPRs bool // Include "Other PRs" section
}

// jsonOutput is the clean structure for JSON output.
// Only includes fields useful for scripting.
type jsonOutput struct {
	MyPRs            []*models.PR `json:"my_prs,omitempty"`
	NeedsMyAttention []*models.PR `json:"needs_my_attention,omitempty"`
	TeamPRs          []*models.PR `json:"team_prs,omitempty"`
	OtherPRs         []*models.PR `json:"other_prs,omitempty"`

	// Summary counts
	TotalPRs    int    `json:"total_prs"`
	Username    string `json:"username"`
	ScanSeconds float64 `json:"scan_seconds"`
}

// RenderJSON marshals the ScanResult to pretty-printed JSON.
// The output is suitable for scripting with tools like jq.
//
// Usage examples:
//
//	prt --json | jq '.my_prs | length'
//	prt --json | jq '.needs_my_attention[].url'
//	prt --json | jq '.total_prs'
func RenderJSON(result *models.ScanResult, opts JSONOptions) (string, error) {
	if result == nil {
		return "", fmt.Errorf("cannot render nil result")
	}

	// Build clean output structure
	output := buildJSONOutput(result, opts)

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data) + "\n", nil
}

// buildJSONOutput creates a clean JSON structure from ScanResult.
func buildJSONOutput(result *models.ScanResult, opts JSONOptions) *jsonOutput {
	output := &jsonOutput{
		MyPRs:            result.MyPRs,
		NeedsMyAttention: result.NeedsMyAttention,
		TeamPRs:          result.TeamPRs,
		Username:         result.Username,
		ScanSeconds:      float64(result.ScanDuration) / float64(time.Second),
	}

	if opts.ShowOtherPRs {
		output.OtherPRs = result.OtherPRs
	}

	// Count actual PRs returned
	output.TotalPRs = len(output.MyPRs) + len(output.NeedsMyAttention) + len(output.TeamPRs)
	if opts.ShowOtherPRs {
		output.TotalPRs += len(output.OtherPRs)
	}

	return output
}

// WriteJSON writes the ScanResult as pretty-printed JSON to the given writer.
// This is useful for streaming output directly to stdout or a file.
func WriteJSON(w io.Writer, result *models.ScanResult) error {
	if result == nil {
		return fmt.Errorf("cannot render nil result")
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode result: %w", err)
	}

	return nil
}

// RenderJSONCompact marshals the ScanResult to compact (non-indented) JSON.
// Useful when minimizing output size is more important than readability.
func RenderJSONCompact(result *models.ScanResult) (string, error) {
	if result == nil {
		return "", fmt.Errorf("cannot render nil result")
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data), nil
}

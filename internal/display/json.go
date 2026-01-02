// Package display provides terminal rendering for PRT output.
package display

import (
	"encoding/json"
	"fmt"
	"io"

	"prt/internal/models"
)

// JSONOptions controls what is included in JSON output.
type JSONOptions struct {
	ShowOtherPRs bool // Include "Other PRs" section
}

// RenderJSON marshals the ScanResult to pretty-printed JSON.
// The output is suitable for scripting with tools like jq.
//
// Usage examples:
//
//	prt --json | jq '.my_prs | length'
//	prt --json | jq '.needs_my_attention[].url'
//	prt --json > ~/pr-snapshot.json
//
// Note: scan_duration_ns is in nanoseconds. Convert to seconds:
//
//	jq '.scan_duration_ns / 1000000000'
func RenderJSON(result *models.ScanResult, opts JSONOptions) (string, error) {
	if result == nil {
		return "", fmt.Errorf("cannot render nil result")
	}

	// Apply filters to match CLI output
	filtered := applyJSONFilters(result, opts)

	data, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data), nil
}

// applyJSONFilters creates a filtered copy of the result based on options.
// This ensures JSON output matches what the CLI displays.
func applyJSONFilters(result *models.ScanResult, opts JSONOptions) *models.ScanResult {
	// Create a shallow copy
	filtered := *result

	// Clear OtherPRs if not showing them
	if !opts.ShowOtherPRs {
		filtered.OtherPRs = nil
	}

	// Build set of repos that have PRs in displayed categories
	displayedRepos := make(map[string]bool)
	for _, pr := range filtered.MyPRs {
		displayedRepos[pr.RepoPath] = true
	}
	for _, pr := range filtered.NeedsMyAttention {
		displayedRepos[pr.RepoPath] = true
	}
	for _, pr := range filtered.TeamPRs {
		displayedRepos[pr.RepoPath] = true
	}
	if opts.ShowOtherPRs {
		for _, pr := range filtered.OtherPRs {
			displayedRepos[pr.RepoPath] = true
		}
	}

	// Filter repos_with_prs to only include repos with displayed PRs
	var filteredRepos []*models.Repository
	for _, repo := range filtered.ReposWithPRs {
		if displayedRepos[repo.Path] {
			filteredRepos = append(filteredRepos, repo)
		}
	}
	filtered.ReposWithPRs = filteredRepos

	// Clear repos_without_prs and repos_with_errors - CLI doesn't show these
	filtered.ReposWithoutPRs = nil
	filtered.ReposWithErrors = nil

	return &filtered
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

// Package display provides terminal rendering for PRT output.
package display

import (
	"encoding/json"
	"fmt"
	"io"

	"prt/internal/models"
)

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
func RenderJSON(result *models.ScanResult) (string, error) {
	if result == nil {
		return "", fmt.Errorf("cannot render nil result")
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(data), nil
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

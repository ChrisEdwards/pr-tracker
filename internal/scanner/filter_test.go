package scanner

import (
	"testing"
)

func TestNewRepoFilter(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "empty patterns",
			patterns: []string{},
			wantErr:  false,
		},
		{
			name:     "nil patterns",
			patterns: nil,
			wantErr:  false,
		},
		{
			name:     "valid prefix pattern",
			patterns: []string{"myorg-*"},
			wantErr:  false,
		},
		{
			name:     "valid suffix pattern",
			patterns: []string{"*-api"},
			wantErr:  false,
		},
		{
			name:     "valid exact pattern",
			patterns: []string{"frontend"},
			wantErr:  false,
		},
		{
			name:     "valid contains pattern",
			patterns: []string{"*service*"},
			wantErr:  false,
		},
		{
			name:     "multiple valid patterns",
			patterns: []string{"myorg-*", "*-api", "frontend"},
			wantErr:  false,
		},
		{
			name:     "invalid pattern - unclosed bracket",
			patterns: []string{"test["},
			wantErr:  true,
			errMsg:   `invalid glob pattern "test["`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewRepoFilter(tt.patterns)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
					return
				}
				if tt.errMsg != "" && err.Error()[:len(tt.errMsg)] != tt.errMsg {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if f == nil {
				t.Error("expected non-nil filter")
			}
		})
	}
}

func TestRepoFilter_Matches(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		repoName string
		want     bool
	}{
		// Empty patterns - match all
		{
			name:     "empty patterns match any name",
			patterns: []string{},
			repoName: "anything",
			want:     true,
		},
		{
			name:     "nil patterns match any name",
			patterns: nil,
			repoName: "anything",
			want:     true,
		},

		// Prefix patterns
		{
			name:     "prefix pattern matches",
			patterns: []string{"myorg-*"},
			repoName: "myorg-frontend",
			want:     true,
		},
		{
			name:     "prefix pattern does not match",
			patterns: []string{"myorg-*"},
			repoName: "other-repo",
			want:     false,
		},
		{
			name:     "prefix pattern exact prefix match",
			patterns: []string{"myorg-*"},
			repoName: "myorg-",
			want:     true,
		},

		// Suffix patterns
		{
			name:     "suffix pattern matches",
			patterns: []string{"*-api"},
			repoName: "user-api",
			want:     true,
		},
		{
			name:     "suffix pattern does not match",
			patterns: []string{"*-api"},
			repoName: "user-service",
			want:     false,
		},

		// Exact match
		{
			name:     "exact pattern matches",
			patterns: []string{"frontend"},
			repoName: "frontend",
			want:     true,
		},
		{
			name:     "exact pattern does not match longer name",
			patterns: []string{"frontend"},
			repoName: "frontend-v2",
			want:     false,
		},
		{
			name:     "exact pattern does not match shorter name",
			patterns: []string{"frontend"},
			repoName: "front",
			want:     false,
		},

		// Contains patterns
		{
			name:     "contains pattern matches at start",
			patterns: []string{"*service*"},
			repoName: "service-api",
			want:     true,
		},
		{
			name:     "contains pattern matches in middle",
			patterns: []string{"*service*"},
			repoName: "user-service-api",
			want:     true,
		},
		{
			name:     "contains pattern matches at end",
			patterns: []string{"*service*"},
			repoName: "user-service",
			want:     true,
		},
		{
			name:     "contains pattern matches exact",
			patterns: []string{"*service*"},
			repoName: "service",
			want:     true,
		},
		{
			name:     "contains pattern does not match",
			patterns: []string{"*service*"},
			repoName: "user-api",
			want:     false,
		},

		// Multiple patterns (OR logic)
		{
			name:     "multiple patterns - first matches",
			patterns: []string{"myorg-*", "*-api", "frontend"},
			repoName: "myorg-auth",
			want:     true,
		},
		{
			name:     "multiple patterns - second matches",
			patterns: []string{"myorg-*", "*-api", "frontend"},
			repoName: "user-api",
			want:     true,
		},
		{
			name:     "multiple patterns - third matches",
			patterns: []string{"myorg-*", "*-api", "frontend"},
			repoName: "frontend",
			want:     true,
		},
		{
			name:     "multiple patterns - none match",
			patterns: []string{"myorg-*", "*-api", "frontend"},
			repoName: "other-service",
			want:     false,
		},

		// Case sensitivity
		{
			name:     "case sensitive - exact case matches",
			patterns: []string{"MyOrg-*"},
			repoName: "MyOrg-repo",
			want:     true,
		},
		{
			name:     "case sensitive - different case does not match",
			patterns: []string{"MyOrg-*"},
			repoName: "myorg-repo",
			want:     false,
		},

		// Edge cases
		{
			name:     "empty repo name with wildcard",
			patterns: []string{"*"},
			repoName: "",
			want:     true,
		},
		{
			name:     "empty repo name with prefix pattern",
			patterns: []string{"prefix-*"},
			repoName: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewRepoFilter(tt.patterns)
			if err != nil {
				t.Fatalf("unexpected error creating filter: %v", err)
			}
			got := f.Matches(tt.repoName)
			if got != tt.want {
				t.Errorf("Matches(%q) = %v, want %v", tt.repoName, got, tt.want)
			}
		})
	}
}

func TestRepoFilter_HasPatterns(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		want     bool
	}{
		{
			name:     "empty patterns",
			patterns: []string{},
			want:     false,
		},
		{
			name:     "nil patterns",
			patterns: nil,
			want:     false,
		},
		{
			name:     "one pattern",
			patterns: []string{"test-*"},
			want:     true,
		},
		{
			name:     "multiple patterns",
			patterns: []string{"test-*", "*-api"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewRepoFilter(tt.patterns)
			if err != nil {
				t.Fatalf("unexpected error creating filter: %v", err)
			}
			got := f.HasPatterns()
			if got != tt.want {
				t.Errorf("HasPatterns() = %v, want %v", got, tt.want)
			}
		})
	}
}

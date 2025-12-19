package display

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestStylesAreDefined(t *testing.T) {
	// Test that all styles are non-nil and can render text
	styles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"HeaderStyle", HeaderStyle},
		{"SubheaderStyle", SubheaderStyle},
		{"DraftStyle", DraftStyle},
		{"NeedsReviewStyle", NeedsReviewStyle},
		{"ApprovedStyle", ApprovedStyle},
		{"ChangesRequestedStyle", ChangesRequestedStyle},
		{"BlockedStyle", BlockedStyle},
		{"CIPassingStyle", CIPassingStyle},
		{"CIFailingStyle", CIFailingStyle},
		{"CIPendingStyle", CIPendingStyle},
		{"URLStyle", URLStyle},
		{"TreeStyle", TreeStyle},
		{"EmptyStyle", EmptyStyle},
		{"MetaStyle", MetaStyle},
		{"RepoStyle", RepoStyle},
		{"TitleStyle", TitleStyle},
		{"NumberStyle", NumberStyle},
		{"AuthorStyle", AuthorStyle},
		{"BranchStyle", BranchStyle},
		{"SummaryStyle", SummaryStyle},
	}

	for _, s := range styles {
		t.Run(s.name, func(t *testing.T) {
			result := s.style.Render("test")
			if !strings.Contains(result, "test") {
				t.Errorf("%s should contain the input text", s.name)
			}
		})
	}
}

func TestIconConstantsAreDefined(t *testing.T) {
	// Test that icon constants are non-empty
	icons := []struct {
		name  string
		value string
	}{
		{"IconMyPRs", IconMyPRs},
		{"IconNeedsAttention", IconNeedsAttention},
		{"IconTeam", IconTeam},
		{"IconOther", IconOther},
		{"IconNoOpenPRs", IconNoOpenPRs},
		{"IconDraft", IconDraft},
		{"IconMerged", IconMerged},
		{"IconApproved", IconApproved},
		{"IconChanges", IconChanges},
		{"IconReview", IconReview},
		{"IconBlocked", IconBlocked},
		{"IconCIPassing", IconCIPassing},
		{"IconCIFailing", IconCIFailing},
		{"IconCIPending", IconCIPending},
		{"IconRepo", IconRepo},
		{"IconEmpty", IconEmpty},
	}

	for _, ic := range icons {
		t.Run(ic.name, func(t *testing.T) {
			if ic.value == "" {
				t.Errorf("%s should not be empty", ic.name)
			}
		})
	}
}

func TestTreeCharactersAreDefined(t *testing.T) {
	// Test that tree characters are valid Unicode
	treeChars := []struct {
		name  string
		value string
	}{
		{"TreeVertical", TreeVertical},
		{"TreeBranch", TreeBranch},
		{"TreeLastBranch", TreeLastBranch},
		{"TreeIndent", TreeIndent},
	}

	for _, tc := range treeChars {
		t.Run(tc.name, func(t *testing.T) {
			if tc.value == "" {
				t.Errorf("%s should not be empty", tc.name)
			}
		})
	}
}

func TestTreeCharacterValues(t *testing.T) {
	// Verify specific tree character values
	if TreeVertical != "│" {
		t.Errorf("TreeVertical should be │, got %q", TreeVertical)
	}
	if TreeBranch != "├──" {
		t.Errorf("TreeBranch should be ├──, got %q", TreeBranch)
	}
	if TreeLastBranch != "└──" {
		t.Errorf("TreeLastBranch should be └──, got %q", TreeLastBranch)
	}
	if TreeIndent != "    " {
		t.Errorf("TreeIndent should be 4 spaces, got %q", TreeIndent)
	}
}

func TestDisableColorsDoesNotPanic(t *testing.T) {
	// Simply ensure DisableColors doesn't panic
	DisableColors()
}

func TestEnableColorsDoesNotPanic(t *testing.T) {
	// Simply ensure EnableColors doesn't panic
	EnableColors()
}

func TestStylesRenderWithSpecialCharacters(t *testing.T) {
	// Test that styles can handle various input types
	inputs := []string{
		"",
		"simple",
		"with spaces",
		"with\nnewline",
		"with\ttab",
		"unicode: 日本語",
		"emoji: 🚀",
		"#123",
		"@username",
		"feature/branch-name",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			// Just verify no panic occurs
			_ = HeaderStyle.Render(input)
			_ = RepoStyle.Render(input)
			_ = URLStyle.Render(input)
		})
	}
}

func TestStyleColorsLookReasonable(t *testing.T) {
	// These tests verify the color choices make semantic sense
	// by checking that different states use different styles

	// Passing and failing CI should be distinguishable
	passing := CIPassingStyle.Render("CI")
	failing := CIFailingStyle.Render("CI")
	if passing == failing {
		t.Error("CIPassingStyle and CIFailingStyle should produce different output")
	}

	// Draft and NeedsReview should be distinguishable
	draft := DraftStyle.Render("PR")
	needsReview := NeedsReviewStyle.Render("PR")
	if draft == needsReview {
		t.Error("DraftStyle and NeedsReviewStyle should produce different output")
	}
}

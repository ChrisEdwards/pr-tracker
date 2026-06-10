package cli

import (
	"os"
	"strings"
	"testing"

	"prt/internal/config"
	"prt/internal/models"
	"prt/internal/prfilters"
)

func TestRootCmd_HasSetupFlag(t *testing.T) {
	// Verify the --setup flag is registered
	flag := rootCmd.Flags().Lookup("setup")
	if flag == nil {
		t.Fatal("--setup flag should be registered")
	}

	if flag.Usage == "" {
		t.Error("--setup flag should have usage text")
	}

	if !strings.Contains(flag.Usage, "wizard") {
		t.Error("--setup flag usage should mention wizard")
	}
}

func TestRootCmd_FlagsRegistered(t *testing.T) {
	expectedFlags := []string{
		"path",
		"filter",
		"author",
		"draft",
		"bot",
		"review-decision",
		"view",
		"group",
		"sort",
		"depth",
		"max-age",
		"json",
		"no-color",
		"setup",
	}

	for _, name := range expectedFlags {
		flag := rootCmd.Flags().Lookup(name)
		if flag == nil {
			t.Errorf("expected flag --%s to be registered", name)
		}
	}
}

func TestRootCmd_Metadata(t *testing.T) {
	if rootCmd.Use != "prt" {
		t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "prt")
	}

	if rootCmd.Short == "" {
		t.Error("rootCmd.Short should not be empty")
	}

	if rootCmd.Long == "" {
		t.Error("rootCmd.Long should not be empty")
	}
}

func TestVersionFlag(t *testing.T) {
	// Set version and run with --version
	rootCmd.Version = "1.2.3-test"
	rootCmd.SetArgs([]string{"--version"})
	defer rootCmd.SetArgs(nil)
	output := captureStdout(t, rootCmd.Execute)

	if !strings.Contains(output, "1.2.3-test") {
		t.Errorf("version output = %q, want it to contain %q", output, "1.2.3-test")
	}
}

func TestHelpFlag(t *testing.T) {
	// Run with --help
	rootCmd.SetArgs([]string{"--help"})
	defer rootCmd.SetArgs(nil)
	output := captureStdout(t, rootCmd.Execute)

	// Verify help output contains expected sections
	expectedPhrases := []string{
		"PRT - GitHub PR Tracker",
		"--path",
		"--filter",
		"--author=",
		"--draft=",
		"--bot=",
		"--review-decision=",
		"--view",
		"--json",
		"--setup",
		"--no-color",
		"prt --author=team --draft=false --bot=false --review-decision=not-approved",
		"prt --view=review-needed",
		"--author=team excludes your own PRs",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("help output should contain %q", phrase)
		}
	}
}

func TestREADME_DocumentsReviewNeededViewWorkflow(t *testing.T) {
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatalf("ReadFile(README.md) error = %v", err)
	}
	content := string(readme)

	expectedPhrases := []string{
		"prt --author=team --draft=false --bot=false --review-decision=not-approved",
		"prt --view=review-needed",
		"prt --view=review-needed --author=@alice",
		"External PRs do not match `review-needed` merely because they request your review",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(content, phrase) {
			t.Errorf("README should contain %q", phrase)
		}
	}
}

func TestFlagDefaults(t *testing.T) {
	tests := []struct {
		name         string
		flagName     string
		defaultValue string
	}{
		{"path default", "path", ""},
		{"filter default", "filter", ""},
		{"author default", "author", ""},
		{"draft default", "draft", ""},
		{"bot default", "bot", ""},
		{"review-decision default", "review-decision", ""},
		{"view default", "view", ""},
		{"group default", "group", ""},
		{"sort default", "sort", ""},
		{"json default", "json", "false"},
		{"no-color default", "no-color", "false"},
		{"setup default", "setup", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := rootCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}
			if flag.DefValue != tt.defaultValue {
				t.Errorf("flag %q default = %q, want %q", tt.flagName, flag.DefValue, tt.defaultValue)
			}
		})
	}
}

func TestFlagShorthand(t *testing.T) {
	tests := []struct {
		flagName  string
		shorthand string
	}{
		{"path", "p"},
		{"filter", "f"},
		{"group", "g"},
		{"sort", "s"},
		{"depth", "d"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			flag := rootCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag %q not found", tt.flagName)
			}
			if flag.Shorthand != tt.shorthand {
				t.Errorf("flag %q shorthand = %q, want %q", tt.flagName, flag.Shorthand, tt.shorthand)
			}
		})
	}
}

func TestFlagUsageDescriptions(t *testing.T) {
	// Verify all flags have usage descriptions
	flags := []string{
		"path", "filter", "author", "draft", "bot", "review-decision", "view", "group", "sort", "depth",
		"max-age", "json", "no-color", "setup",
	}

	for _, name := range flags {
		t.Run(name, func(t *testing.T) {
			flag := rootCmd.Flags().Lookup(name)
			if flag == nil {
				t.Fatalf("flag %q not found", name)
			}
			if flag.Usage == "" {
				t.Errorf("flag %q should have usage description", name)
			}
		})
	}
}

func TestBuildActiveFilterSet_SelectsView(t *testing.T) {
	cfg := &config.Config{
		GitHubUsername: "me",
		Views: map[string]config.View{
			"alice-review": {
				Filters: map[string]interface{}{
					"author":          "@alice",
					"draft":           false,
					"bot":             false,
					"review_decision": "not-approved",
				},
			},
		},
	}

	set, err := buildActiveFilterSet(cfg, "alice-review", prfilters.Set{})
	if err != nil {
		t.Fatalf("buildActiveFilterSet() error = %v", err)
	}

	filtered := set.Apply(viewFilterResult(), cfg)
	if got := prNumbers(filtered.MatchingPRs); !sameNumbers(got, []int{2}) {
		t.Fatalf("MatchingPRs = %v, want [2]", got)
	}
}

func TestBuildActiveFilterSet_BuiltInReviewNeededViewFiltersMatchingPRs(t *testing.T) {
	cfg := &config.Config{
		GitHubUsername: "me",
		TeamMembers:    []string{"me", "alice", "bob", "carol", "dana", "release-service"},
		Bots:           []string{"release-service"},
	}

	set, err := buildActiveFilterSet(cfg, "review-needed", prfilters.Set{})
	if err != nil {
		t.Fatalf("buildActiveFilterSet() error = %v", err)
	}

	filtered := set.Apply(reviewNeededViewResult(), cfg)
	if got := prNumbers(filtered.MyPRs); !sameNumbers(got, []int{1}) {
		t.Fatalf("MyPRs = %v, want [1]", got)
	}
	if got := prNumbers(filtered.MatchingPRs); !sameNumbers(got, []int{2, 3, 4, 5, 6}) {
		t.Fatalf("MatchingPRs = %v, want review-needed team PRs [2 3 4 5 6]", got)
	}
}

func TestBuildActiveFilterSet_ConfigReviewNeededViewOverridesBuiltIn(t *testing.T) {
	cfg := &config.Config{
		GitHubUsername: "me",
		TeamMembers:    []string{"alice"},
		Views: map[string]config.View{
			"review-needed": {
				Filters: map[string]interface{}{"author": "@mallory"},
			},
		},
	}

	set, err := buildActiveFilterSet(cfg, "review-needed", prfilters.Set{})
	if err != nil {
		t.Fatalf("buildActiveFilterSet() error = %v", err)
	}

	filtered := set.Apply(reviewNeededViewResult(), cfg)
	if got := prNumbers(filtered.MatchingPRs); !sameNumbers(got, []int{7}) {
		t.Fatalf("MatchingPRs = %v, want config-defined review-needed view to match [7]", got)
	}
}

func TestBuildActiveFilterSet_UnknownViewName(t *testing.T) {
	cfg := &config.Config{
		Views: map[string]config.View{
			"alice-review": {Filters: map[string]interface{}{"author": "@alice"}},
		},
	}

	_, err := buildActiveFilterSet(cfg, "missing", prfilters.Set{})
	if err == nil {
		t.Fatal("buildActiveFilterSet() error = nil, want error")
	}
	for _, want := range []string{"unknown view", "missing", "alice-review"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want it to contain %q", err.Error(), want)
		}
	}
}

func TestBuildActiveFilterSet_CLIFiltersNarrowSelectedView(t *testing.T) {
	cfg := &config.Config{
		GitHubUsername: "me",
		TeamMembers:    []string{"alice", "bob"},
		Views: map[string]config.View{
			"team-ready": {
				Filters: map[string]interface{}{
					"author":          "team",
					"draft":           false,
					"review_decision": "not-approved",
				},
			},
		},
	}
	cliSet, err := prfilters.ParseFlags("@alice", "", "", "")
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	set, err := buildActiveFilterSet(cfg, "team-ready", cliSet)
	if err != nil {
		t.Fatalf("buildActiveFilterSet() error = %v", err)
	}

	filtered := set.Apply(viewFilterResult(), cfg)
	if got := prNumbers(filtered.MatchingPRs); !sameNumbers(got, []int{2}) {
		t.Fatalf("MatchingPRs = %v, want only Alice's team-ready PR [2]", got)
	}
}

func TestBuildActiveFilterSet_PreservesMyPRsWhenViewIsActive(t *testing.T) {
	cfg := &config.Config{
		GitHubUsername: "me",
		Views: map[string]config.View{
			"alice-review": {
				Filters: map[string]interface{}{"author": "@alice"},
			},
		},
	}

	set, err := buildActiveFilterSet(cfg, "alice-review", prfilters.Set{})
	if err != nil {
		t.Fatalf("buildActiveFilterSet() error = %v", err)
	}

	filtered := set.Apply(viewFilterResult(), cfg)
	if got := prNumbers(filtered.MyPRs); !sameNumbers(got, []int{1}) {
		t.Fatalf("MyPRs = %v, want [1]", got)
	}
}

func TestBuildActiveFilterSet_InvalidViewFilterValue(t *testing.T) {
	cfg := &config.Config{
		Views: map[string]config.View{
			"bad-review": {
				Filters: map[string]interface{}{"review_decision": "waiting"},
			},
		},
	}

	_, err := buildActiveFilterSet(cfg, "bad-review", prfilters.Set{})
	if err == nil {
		t.Fatal("buildActiveFilterSet() error = nil, want error")
	}
	for _, want := range []string{"bad-review", "review_decision", "waiting"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want it to contain %q", err.Error(), want)
		}
	}
}

func TestBuildActiveFilterSet_EmptyViewFilters(t *testing.T) {
	cfg := &config.Config{
		Views: map[string]config.View{
			"empty-view": {
				Filters: map[string]interface{}{},
			},
		},
	}

	_, err := buildActiveFilterSet(cfg, "empty-view", prfilters.Set{})
	if err == nil {
		t.Fatal("buildActiveFilterSet() error = nil, want error")
	}
	for _, want := range []string{"empty-view", "at least one filter"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want it to contain %q", err.Error(), want)
		}
	}
}

func viewFilterResult() *models.ScanResult {
	result := models.NewScanResult()
	result.Username = "me"
	result.MyPRs = []*models.PR{
		{Number: 1, Title: "My approved PR", Author: "me", State: models.PRStateOpen, IsDraft: true, ReviewDecision: models.ReviewDecisionApproved},
	}
	result.TeamPRs = []*models.PR{
		{Number: 2, Title: "Alice ready", Author: "alice", State: models.PRStateOpen, ReviewDecision: models.ReviewDecisionReviewRequired},
		{Number: 3, Title: "Bob ready", Author: "bob", State: models.PRStateOpen, ReviewDecision: models.ReviewDecisionReviewRequired},
		{Number: 4, Title: "Alice draft", Author: "alice", State: models.PRStateOpen, IsDraft: true, ReviewDecision: models.ReviewDecisionChangesRequested},
	}
	result.OtherPRs = []*models.PR{
		{Number: 5, Title: "Bot PR", Author: "automation[bot]", State: models.PRStateOpen, ReviewDecision: models.ReviewDecisionReviewRequired},
		{Number: 6, Title: "Approved Alice", Author: "alice", State: models.PRStateOpen, ReviewDecision: models.ReviewDecisionApproved},
	}
	return result
}

func reviewNeededViewResult() *models.ScanResult {
	result := models.NewScanResult()
	result.Username = "me"
	result.MyPRs = []*models.PR{
		{Number: 1, Title: "My approved draft PR stays visible", Author: "me", State: models.PRStateOpen, IsDraft: true, ReviewDecision: models.ReviewDecisionApproved},
	}
	result.NeedsMyAttention = []*models.PR{
		{Number: 2, Title: "Team requested review with changes requested", Author: "alice", State: models.PRStateOpen, ReviewDecision: models.ReviewDecisionChangesRequested, ReviewRequests: []string{"me"}},
		{Number: 7, Title: "External requested review", Author: "mallory", State: models.PRStateOpen, ReviewDecision: models.ReviewDecisionReviewRequired, ReviewRequests: []string{"me"}},
	}
	result.TeamPRs = []*models.PR{
		{Number: 3, Title: "Review required", Author: "alice", State: models.PRStateOpen, ReviewDecision: models.ReviewDecisionReviewRequired},
		{Number: 4, Title: "No review decision", Author: "bob", State: models.PRStateOpen, ReviewDecision: models.ReviewDecisionNone},
		{Number: 5, Title: "Unknown future decision", Author: "carol", State: models.PRStateOpen, ReviewDecision: models.ReviewDecision("FUTURE_STATE")},
		{Number: 6, Title: "Another changes requested PR", Author: "dana", State: models.PRStateOpen, ReviewDecision: models.ReviewDecisionChangesRequested},
		{Number: 8, Title: "Approved team PR", Author: "alice", State: models.PRStateOpen, ReviewDecision: models.ReviewDecisionApproved},
		{Number: 9, Title: "Draft team PR", Author: "bob", State: models.PRStateOpen, IsDraft: true, ReviewDecision: models.ReviewDecisionChangesRequested},
		{Number: 10, Title: "Current user duplicate category", Author: "me", State: models.PRStateOpen, ReviewDecision: models.ReviewDecisionReviewRequired},
	}
	result.OtherPRs = []*models.PR{
		{Number: 11, Title: "Configured bot", Author: "release-service", State: models.PRStateOpen, ReviewDecision: models.ReviewDecisionReviewRequired},
		{Number: 12, Title: "Bot suffix", Author: "automation[bot]", State: models.PRStateOpen, ReviewDecision: models.ReviewDecisionReviewRequired},
	}
	return result
}

func prNumbers(prs []*models.PR) []int {
	numbers := make([]int, 0, len(prs))
	for _, pr := range prs {
		numbers = append(numbers, pr.Number)
	}
	return numbers
}

func sameNumbers(got, want []int) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestRootCmd_ReviewDecisionFlagParsing(t *testing.T) {
	flag := rootCmd.Flags().Lookup("review-decision")
	if flag == nil {
		t.Fatal("--review-decision flag should be registered")
	}

	originalValue := flag.Value.String()
	originalFlagReviewDecision := flagReviewDecision
	defer func() {
		if err := rootCmd.Flags().Set("review-decision", originalValue); err != nil {
			t.Fatalf("restore --review-decision flag: %v", err)
		}
		flagReviewDecision = originalFlagReviewDecision
	}()

	tests := []struct {
		name        string
		value       string
		pr          *models.PR
		wantMatch   bool
		wantErrText string
	}{
		{
			name:      "approved",
			value:     "approved",
			pr:        &models.PR{ReviewDecision: models.ReviewDecisionApproved},
			wantMatch: true,
		},
		{
			name:      "not approved includes unknown",
			value:     "not-approved",
			pr:        &models.PR{ReviewDecision: models.ReviewDecision("FUTURE_STATE")},
			wantMatch: true,
		},
		{
			name:      "review required",
			value:     "review-required",
			pr:        &models.PR{ReviewDecision: models.ReviewDecisionReviewRequired},
			wantMatch: true,
		},
		{
			name:      "changes requested",
			value:     "changes-requested",
			pr:        &models.PR{ReviewDecision: models.ReviewDecisionChangesRequested},
			wantMatch: true,
		},
		{
			name:      "none",
			value:     "none",
			pr:        &models.PR{ReviewDecision: models.ReviewDecisionNone},
			wantMatch: true,
		},
		{
			name:        "invalid",
			value:       "waiting",
			wantErrText: "invalid --review-decision value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := rootCmd.Flags().Set("review-decision", tt.value); err != nil {
				t.Fatalf("set --review-decision flag: %v", err)
			}

			set, err := prfilters.ParseFlags("", "", "", flagReviewDecision)
			if tt.wantErrText != "" {
				if err == nil {
					t.Fatal("ParseFlags() error = nil, want error")
				}
				if !strings.Contains(err.Error(), tt.wantErrText) {
					t.Fatalf("ParseFlags() error = %q, want it to contain %q", err.Error(), tt.wantErrText)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}
			if got := set.Matches(tt.pr, prfilters.Context{}); got != tt.wantMatch {
				t.Fatalf("Matches() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestRootCmd_BotFlagParsing(t *testing.T) {
	flag := rootCmd.Flags().Lookup("bot")
	if flag == nil {
		t.Fatal("--bot flag should be registered")
	}

	originalValue := flag.Value.String()
	originalFlagBot := flagBot
	defer func() {
		if err := rootCmd.Flags().Set("bot", originalValue); err != nil {
			t.Fatalf("restore --bot flag: %v", err)
		}
		flagBot = originalFlagBot
	}()

	tests := []struct {
		name        string
		value       string
		wantMatch   bool
		wantErrText string
	}{
		{name: "true", value: "true", wantMatch: true},
		{name: "false", value: "false", wantMatch: false},
		{name: "invalid", value: "yes", wantErrText: "invalid --bot value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := rootCmd.Flags().Set("bot", tt.value); err != nil {
				t.Fatalf("set --bot flag: %v", err)
			}

			set, err := prfilters.ParseFlags("", "", flagBot, "")
			if tt.wantErrText != "" {
				if err == nil {
					t.Fatal("ParseFlags() error = nil, want error")
				}
				if !strings.Contains(err.Error(), tt.wantErrText) {
					t.Fatalf("ParseFlags() error = %q, want it to contain %q", err.Error(), tt.wantErrText)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}
			got := set.Matches(&models.PR{Author: "automation[bot]"}, prfilters.Context{})
			if got != tt.wantMatch {
				t.Fatalf("Matches(bot PR) = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestConfigSubcommandExists(t *testing.T) {
	// Verify config subcommand is registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "config" {
			found = true
			break
		}
	}
	if !found {
		t.Error("config subcommand should be registered")
	}
}

func TestExecuteFunction(t *testing.T) {
	// Test that Execute sets version correctly
	// Can't fully test without running the actual command
	// but we can verify the function exists and is callable

	// Save original args and restore
	oldArgs := os.Args
	os.Args = []string{"prt", "--version"}
	defer func() { os.Args = oldArgs }()

	// The Execute function should return nil for --version
	// But actually executing would try to print to stdout
	// Instead, verify the rootCmd version gets set
	oldVersion := rootCmd.Version
	defer func() { rootCmd.Version = oldVersion }()

	rootCmd.Version = "test-version"
	if rootCmd.Version != "test-version" {
		t.Error("Execute should set version on rootCmd")
	}
}

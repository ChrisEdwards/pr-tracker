package prfilters

import (
	"strings"
	"testing"

	"prt/internal/config"
	"prt/internal/models"
)

func TestAuthorTeamMatchesConfiguredTeamExcludingCurrentUser(t *testing.T) {
	set, err := ParseFlags("team", "", "", "")
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	ctx := Context{
		Username:    "me",
		TeamMembers: []string{"me", "alice", "bob"},
	}

	tests := []struct {
		name string
		pr   *models.PR
		want bool
	}{
		{name: "team member", pr: &models.PR{Author: "alice"}, want: true},
		{name: "current user excluded", pr: &models.PR{Author: "me"}, want: false},
		{name: "other author", pr: &models.PR{Author: "mallory"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := set.Matches(tt.pr, ctx); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthorFiltersRespectBotHook(t *testing.T) {
	botHook := func(author string) bool {
		return author == "renovate[bot]"
	}

	tests := []struct {
		name   string
		author string
		pr     *models.PR
		ctx    Context
		want   bool
	}{
		{
			name:   "team excludes bot even when configured as team member",
			author: "team",
			pr:     &models.PR{Author: "renovate[bot]"},
			ctx: Context{
				Username:    "me",
				TeamMembers: []string{"alice", "renovate[bot]"},
				IsBot:       botHook,
			},
			want: false,
		},
		{
			name:   "other excludes bot",
			author: "other",
			pr:     &models.PR{Author: "renovate[bot]"},
			ctx: Context{
				Username: "me",
				IsBot:    botHook,
			},
			want: false,
		},
		{
			name:   "team still includes non-bot team member",
			author: "team",
			pr:     &models.PR{Author: "alice"},
			ctx: Context{
				Username:    "me",
				TeamMembers: []string{"alice", "renovate[bot]"},
				IsBot:       botHook,
			},
			want: true,
		},
		{
			name:   "other still includes non-bot non-team author",
			author: "other",
			pr:     &models.PR{Author: "mallory"},
			ctx: Context{
				Username:    "me",
				TeamMembers: []string{"alice"},
				IsBot:       botHook,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set, err := ParseFlags(tt.author, "", "", "")
			if err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}
			if got := set.Matches(tt.pr, tt.ctx); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBotFilterMatchesConfiguredBotAuthors(t *testing.T) {
	result := models.NewScanResult()
	result.Username = "me"
	result.OtherPRs = []*models.PR{
		{Number: 1, Title: "Configured bot", Author: "release-service", State: models.PRStateOpen},
		{Number: 2, Title: "Human author", Author: "mallory", State: models.PRStateOpen},
	}

	set, err := ParseFlags("", "", "true", "")
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	filtered := set.Apply(result, &config.Config{
		GitHubUsername: "me",
		Bots:           []string{"release-service"},
	})

	if got := prNumbers(filtered.MatchingPRs); len(got) != 1 || got[0] != 1 {
		t.Fatalf("MatchingPRs = %v, want configured bot PR #1", got)
	}
}

func TestBotFilterMatchesBotSuffixAuthors(t *testing.T) {
	set, err := ParseFlags("", "", "true", "")
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	tests := []struct {
		name string
		pr   *models.PR
		want bool
	}{
		{name: "suffix bot", pr: &models.PR{Author: "some-app[bot]"}, want: true},
		{name: "human", pr: &models.PR{Author: "alice"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := set.Matches(tt.pr, Context{}); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBotAuthorsAreNotTeamMembers(t *testing.T) {
	set, err := ParseFlags("team", "", "", "")
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	ctx := Context{
		Username:    "me",
		TeamMembers: []string{"alice", "automation[bot]"},
	}

	if got := set.Matches(&models.PR{Author: "automation[bot]"}, ctx); got {
		t.Fatal("team author filter should exclude bot authors even when listed in team_members")
	}
	if got := set.Matches(&models.PR{Author: "alice"}, ctx); !got {
		t.Fatal("team author filter should still include non-bot team members")
	}
}

func TestBotFilterExcludesBotAuthorsAndPreservesMyPRs(t *testing.T) {
	result := models.NewScanResult()
	result.Username = "me"
	result.MyPRs = []*models.PR{
		{Number: 1, Title: "My PR stays visible", Author: "me", State: models.PRStateOpen},
	}
	result.NeedsMyAttention = []*models.PR{
		{Number: 2, Title: "Human needs attention", Author: "alice", State: models.PRStateOpen},
	}
	result.TeamPRs = []*models.PR{
		{Number: 3, Title: "Team bot", Author: "automation[bot]", State: models.PRStateOpen},
	}
	result.OtherPRs = []*models.PR{
		{Number: 4, Title: "Configured bot", Author: "release-service", State: models.PRStateOpen},
		{Number: 5, Title: "Other human", Author: "mallory", State: models.PRStateOpen},
	}

	set, err := ParseFlags("", "", "false", "")
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	filtered := set.Apply(result, &config.Config{
		GitHubUsername: "me",
		Bots:           []string{"release-service"},
	})

	if got := prNumbers(filtered.MyPRs); len(got) != 1 || got[0] != 1 {
		t.Fatalf("MyPRs = %v, want [1]", got)
	}
	if got := prNumbers(filtered.MatchingPRs); !sameNumbers(got, []int{2, 5}) {
		t.Fatalf("MatchingPRs = %v, want [2 5]", got)
	}
}

func TestReviewDecisionFilterMatchesSupportedValues(t *testing.T) {
	tests := []struct {
		name               string
		reviewDecisionFlag string
		pr                 *models.PR
		want               bool
	}{
		{
			name:               "approved matches approved",
			reviewDecisionFlag: "approved",
			pr:                 &models.PR{ReviewDecision: models.ReviewDecisionApproved},
			want:               true,
		},
		{
			name:               "approved rejects review required",
			reviewDecisionFlag: "approved",
			pr:                 &models.PR{ReviewDecision: models.ReviewDecisionReviewRequired},
			want:               false,
		},
		{
			name:               "approved does not infer from review records",
			reviewDecisionFlag: "approved",
			pr: &models.PR{
				ReviewDecision: models.ReviewDecisionNone,
				Reviews:        []models.Review{{State: models.ReviewStateApproved}},
			},
			want: false,
		},
		{
			name:               "not approved rejects approved",
			reviewDecisionFlag: "not-approved",
			pr:                 &models.PR{ReviewDecision: models.ReviewDecisionApproved},
			want:               false,
		},
		{
			name:               "not approved includes review required",
			reviewDecisionFlag: "not-approved",
			pr:                 &models.PR{ReviewDecision: models.ReviewDecisionReviewRequired},
			want:               true,
		},
		{
			name:               "not approved includes changes requested",
			reviewDecisionFlag: "not-approved",
			pr:                 &models.PR{ReviewDecision: models.ReviewDecisionChangesRequested},
			want:               true,
		},
		{
			name:               "not approved includes none",
			reviewDecisionFlag: "not-approved",
			pr:                 &models.PR{ReviewDecision: models.ReviewDecisionNone},
			want:               true,
		},
		{
			name:               "not approved includes unknown",
			reviewDecisionFlag: "not-approved",
			pr:                 &models.PR{ReviewDecision: models.ReviewDecision("FUTURE_STATE")},
			want:               true,
		},
		{
			name:               "not approved includes draft PR",
			reviewDecisionFlag: "not-approved",
			pr:                 &models.PR{State: models.PRStateOpen, IsDraft: true, ReviewDecision: models.ReviewDecisionChangesRequested},
			want:               true,
		},
		{
			name:               "review required only matches review required",
			reviewDecisionFlag: "review-required",
			pr:                 &models.PR{ReviewDecision: models.ReviewDecisionReviewRequired},
			want:               true,
		},
		{
			name:               "review required rejects changes requested",
			reviewDecisionFlag: "review-required",
			pr:                 &models.PR{ReviewDecision: models.ReviewDecisionChangesRequested},
			want:               false,
		},
		{
			name:               "changes requested only matches changes requested",
			reviewDecisionFlag: "changes-requested",
			pr:                 &models.PR{ReviewDecision: models.ReviewDecisionChangesRequested},
			want:               true,
		},
		{
			name:               "changes requested rejects review required",
			reviewDecisionFlag: "changes-requested",
			pr:                 &models.PR{ReviewDecision: models.ReviewDecisionReviewRequired},
			want:               false,
		},
		{
			name:               "none matches empty decision",
			reviewDecisionFlag: "none",
			pr:                 &models.PR{ReviewDecision: models.ReviewDecisionNone},
			want:               true,
		},
		{
			name:               "none rejects unknown decision",
			reviewDecisionFlag: "none",
			pr:                 &models.PR{ReviewDecision: models.ReviewDecision("FUTURE_STATE")},
			want:               false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set, err := ParseFlags("", "", "", tt.reviewDecisionFlag)
			if err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}
			if got := set.Matches(tt.pr, Context{}); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReviewDecisionFilterPreservesMyPRsAndFiltersMatchingPRs(t *testing.T) {
	result := models.NewScanResult()
	result.Username = "me"
	result.MyPRs = []*models.PR{
		{Number: 1, Title: "My approved PR stays visible", Author: "me", ReviewDecision: models.ReviewDecisionApproved},
	}
	result.NeedsMyAttention = []*models.PR{
		{Number: 2, Title: "Needs review", Author: "alice", ReviewDecision: models.ReviewDecisionReviewRequired},
	}
	result.TeamPRs = []*models.PR{
		{Number: 3, Title: "Changes requested", Author: "bob", ReviewDecision: models.ReviewDecisionChangesRequested},
		{Number: 4, Title: "Approved", Author: "carol", ReviewDecision: models.ReviewDecisionApproved},
	}
	result.OtherPRs = []*models.PR{
		{Number: 5, Title: "Unknown future state", Author: "mallory", ReviewDecision: models.ReviewDecision("FUTURE_STATE")},
	}

	set, err := ParseFlags("", "", "", "not-approved")
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	filtered := set.Apply(result, &config.Config{GitHubUsername: "me"})

	if got := prNumbers(filtered.MyPRs); len(got) != 1 || got[0] != 1 {
		t.Fatalf("MyPRs = %v, want [1]", got)
	}
	if got := prNumbers(filtered.MatchingPRs); !sameNumbers(got, []int{2, 3, 5}) {
		t.Fatalf("MatchingPRs = %v, want [2 3 5]", got)
	}
}

func TestParseFlagsAcceptsSupportedBotValues(t *testing.T) {
	tests := []struct {
		name    string
		botFlag string
		pr      *models.PR
		want    bool
	}{
		{
			name:    "bot true matches bots",
			botFlag: "true",
			pr:      &models.PR{Author: "some-app[bot]"},
			want:    true,
		},
		{
			name:    "bot true rejects humans",
			botFlag: "true",
			pr:      &models.PR{Author: "alice"},
			want:    false,
		},
		{
			name:    "bot false excludes bots",
			botFlag: "false",
			pr:      &models.PR{Author: "some-app[bot]"},
			want:    false,
		},
		{
			name:    "bot false includes humans",
			botFlag: "false",
			pr:      &models.PR{Author: "alice"},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set, err := ParseFlags("", "", tt.botFlag, "")
			if err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}
			if got := set.Matches(tt.pr, Context{}); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDraftFilterUsesEffectiveState(t *testing.T) {
	tests := []struct {
		name       string
		draftFlag  string
		pr         *models.PR
		wantMatch  bool
		wantActive bool
	}{
		{
			name:       "draft true matches effective draft from IsDraft",
			draftFlag:  "true",
			pr:         &models.PR{State: models.PRStateOpen, IsDraft: true},
			wantMatch:  true,
			wantActive: true,
		},
		{
			name:       "draft false rejects effective draft",
			draftFlag:  "false",
			pr:         &models.PR{State: models.PRStateOpen, IsDraft: true},
			wantMatch:  false,
			wantActive: true,
		},
		{
			name:       "draft false matches non draft",
			draftFlag:  "false",
			pr:         &models.PR{State: models.PRStateOpen, IsDraft: false},
			wantMatch:  true,
			wantActive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set, err := ParseFlags("", tt.draftFlag, "", "")
			if err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}
			if got := set.Active(); got != tt.wantActive {
				t.Errorf("Active() = %v, want %v", got, tt.wantActive)
			}
			if got := set.Matches(tt.pr, Context{}); got != tt.wantMatch {
				t.Errorf("Matches() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestParseFlagsAcceptsSupportedAuthorValues(t *testing.T) {
	tests := []struct {
		name   string
		author string
		pr     *models.PR
		ctx    Context
		want   bool
	}{
		{
			name:   "me",
			author: "me",
			pr:     &models.PR{Author: "me"},
			ctx:    Context{Username: "me"},
			want:   true,
		},
		{
			name:   "other excludes current user and team",
			author: "other",
			pr:     &models.PR{Author: "mallory"},
			ctx:    Context{Username: "me", TeamMembers: []string{"alice"}},
			want:   true,
		},
		{
			name:   "username",
			author: "@alice",
			pr:     &models.PR{Author: "alice"},
			ctx:    Context{Username: "me"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set, err := ParseFlags(tt.author, "", "", "")
			if err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}
			if got := set.Matches(tt.pr, tt.ctx); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFlagsRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name        string
		authorFlag  string
		draftFlag   string
		botFlag     string
		reviewFlag  string
		wantMessage string
	}{
		{
			name:        "author without @",
			authorFlag:  "alice",
			wantMessage: "invalid --author value",
		},
		{
			name:        "empty username",
			authorFlag:  "@",
			wantMessage: "invalid --author value",
		},
		{
			name:        "draft must be true or false",
			draftFlag:   "yes",
			wantMessage: "invalid --draft value",
		},
		{
			name:        "bot must be true or false",
			botFlag:     "yes",
			wantMessage: "invalid --bot value",
		},
		{
			name:        "review decision must be supported value",
			reviewFlag:  "waiting",
			wantMessage: "invalid --review-decision value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseFlags(tt.authorFlag, tt.draftFlag, tt.botFlag, tt.reviewFlag)
			if err == nil {
				t.Fatal("ParseFlags() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Errorf("ParseFlags() error = %q, want it to contain %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestParseViewFiltersRejectsEmptyFilters(t *testing.T) {
	_, err := ParseViewFilters("empty-view", map[string]interface{}{})
	if err == nil {
		t.Fatal("ParseViewFilters() error = nil, want error")
	}
	if !strings.Contains(err.Error(), `view "empty-view" filters must contain at least one filter`) {
		t.Fatalf("ParseViewFilters() error = %q, want empty filters error", err.Error())
	}
}

func TestApplyPreservesMyPRsAndBuildsMatchingPRsWithANDSemantics(t *testing.T) {
	result := models.NewScanResult()
	result.Username = "me"
	result.MyPRs = []*models.PR{
		{Number: 1, Title: "My draft", Author: "me", State: models.PRStateOpen, IsDraft: true},
	}
	result.NeedsMyAttention = []*models.PR{
		{Number: 2, Title: "Team ready needs attention", Author: "alice", State: models.PRStateOpen},
	}
	result.TeamPRs = []*models.PR{
		{Number: 3, Title: "Team draft", Author: "bob", State: models.PRStateOpen, IsDraft: true},
		{Number: 4, Title: "Current user should not match", Author: "me", State: models.PRStateOpen},
	}
	result.OtherPRs = []*models.PR{
		{Number: 5, Title: "Other ready", Author: "mallory", State: models.PRStateOpen},
	}

	set, err := ParseFlags("team", "false", "", "")
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	filtered := set.Apply(result, &config.Config{
		GitHubUsername: "me",
		TeamMembers:    []string{"me", "alice", "bob"},
	})

	if len(filtered.MyPRs) != 1 || filtered.MyPRs[0].Number != 1 {
		t.Fatalf("filtered MyPRs = %#v, want only PR #1", prNumbers(filtered.MyPRs))
	}
	if len(filtered.MatchingPRs) != 1 || filtered.MatchingPRs[0].Number != 2 {
		t.Fatalf("filtered MatchingPRs = %#v, want only PR #2", prNumbers(filtered.MatchingPRs))
	}
	if len(result.MatchingPRs) != 0 {
		t.Fatal("Apply() should not mutate the input result")
	}
}

func TestApplyUsesConfiguredBotsForAuthorFilters(t *testing.T) {
	result := models.NewScanResult()
	result.Username = "me"
	result.TeamPRs = []*models.PR{
		{Number: 1, Title: "Team bot", Author: "renovate[bot]", State: models.PRStateOpen},
		{Number: 2, Title: "Team human", Author: "alice", State: models.PRStateOpen},
	}
	result.OtherPRs = []*models.PR{
		{Number: 3, Title: "Other bot", Author: "dependabot[bot]", State: models.PRStateOpen},
		{Number: 4, Title: "Other human", Author: "mallory", State: models.PRStateOpen},
	}

	teamSet, err := ParseFlags("team", "", "", "")
	if err != nil {
		t.Fatalf("ParseFlags(team) error = %v", err)
	}
	teamFiltered := teamSet.Apply(result, &config.Config{
		GitHubUsername: "me",
		TeamMembers:    []string{"alice", "renovate[bot]"},
		Bots:           []string{"renovate[bot]", "dependabot[bot]"},
	})
	if got := prNumbers(teamFiltered.MatchingPRs); len(got) != 1 || got[0] != 2 {
		t.Fatalf("team MatchingPRs = %v, want [2]", got)
	}

	otherSet, err := ParseFlags("other", "", "", "")
	if err != nil {
		t.Fatalf("ParseFlags(other) error = %v", err)
	}
	otherFiltered := otherSet.Apply(result, &config.Config{
		GitHubUsername: "me",
		TeamMembers:    []string{"alice", "renovate[bot]"},
		Bots:           []string{"renovate[bot]", "dependabot[bot]"},
	})
	if got := prNumbers(otherFiltered.MatchingPRs); len(got) != 1 || got[0] != 4 {
		t.Fatalf("other MatchingPRs = %v, want [4]", got)
	}
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

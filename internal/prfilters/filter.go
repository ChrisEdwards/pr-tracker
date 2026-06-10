// Package prfilters builds and applies pull-request filters.
package prfilters

import (
	"fmt"
	"sort"
	"strings"

	"prt/internal/config"
	"prt/internal/models"
)

type authorMode string
type reviewDecisionMode string

const (
	authorAny      authorMode = ""
	authorMe       authorMode = "me"
	authorTeam     authorMode = "team"
	authorOther    authorMode = "other"
	authorUsername authorMode = "username"

	reviewDecisionAny              reviewDecisionMode = ""
	reviewDecisionApproved         reviewDecisionMode = "approved"
	reviewDecisionNotApproved      reviewDecisionMode = "not-approved"
	reviewDecisionReviewRequired   reviewDecisionMode = "review-required"
	reviewDecisionChangesRequested reviewDecisionMode = "changes-requested"
	reviewDecisionNone             reviewDecisionMode = "none"
)

// Context provides user and team data needed to evaluate PR filters.
type Context struct {
	Username    string
	TeamMembers []string
	BotAuthors  []string
	IsBot       func(author string) bool
}

// Set is the active pull-request filter pipeline. Filters compose with AND
// semantics: a PR must match every configured criterion.
type Set struct {
	all            []Set
	authorMode     authorMode
	authorUsername string
	draft          *bool
	bot            *bool
	reviewDecision reviewDecisionMode
}

// ParseFlags builds a filter set from CLI flag values.
func ParseFlags(authorValue, draftValue, botValue, reviewDecisionValue string) (Set, error) {
	var set Set

	if authorValue != "" {
		mode, username, err := parseAuthor(authorValue)
		if err != nil {
			return Set{}, err
		}
		set.authorMode = mode
		set.authorUsername = username
	}
	if draftValue != "" {
		draft, err := parseBoolFlag("draft", draftValue)
		if err != nil {
			return Set{}, err
		}
		set.draft = &draft
	}
	if botValue != "" {
		bot, err := parseBoolFlag("bot", botValue)
		if err != nil {
			return Set{}, err
		}
		set.bot = &bot
	}
	if reviewDecisionValue != "" {
		mode, err := parseReviewDecision(reviewDecisionValue)
		if err != nil {
			return Set{}, err
		}
		set.reviewDecision = mode
	}

	return set, nil
}

// ParseViewFilters builds a filter set from a config View's filters map.
func ParseViewFilters(viewName string, filters map[string]interface{}) (Set, error) {
	if len(filters) == 0 {
		return Set{}, fmt.Errorf("view %q filters must contain at least one filter", viewName)
	}

	var set Set
	for _, key := range sortedFilterKeys(filters) {
		value := filters[key]
		switch key {
		case "author":
			author, ok := value.(string)
			if !ok {
				return Set{}, fmt.Errorf("view %q filter author must be a string", viewName)
			}
			mode, username, err := parseAuthor(author)
			if err != nil {
				return Set{}, fmt.Errorf("view %q filter author: %w", viewName, err)
			}
			set.authorMode = mode
			set.authorUsername = username
		case "draft":
			draft, ok := value.(bool)
			if !ok {
				return Set{}, fmt.Errorf("view %q filter draft must be a boolean", viewName)
			}
			set.draft = &draft
		case "bot":
			bot, ok := value.(bool)
			if !ok {
				return Set{}, fmt.Errorf("view %q filter bot must be a boolean", viewName)
			}
			set.bot = &bot
		case "review_decision":
			reviewDecision, ok := value.(string)
			if !ok {
				return Set{}, fmt.Errorf("view %q filter review_decision must be a string", viewName)
			}
			mode, err := parseReviewDecision(reviewDecision)
			if err != nil {
				return Set{}, fmt.Errorf("view %q filter review_decision: %w", viewName, err)
			}
			set.reviewDecision = mode
		default:
			return Set{}, fmt.Errorf("view %q filter %q is not supported (supported filters: author, draft, bot, review_decision)", viewName, key)
		}
	}
	return set, nil
}

// And composes filter sets with AND semantics.
func And(sets ...Set) Set {
	active := make([]Set, 0, len(sets))
	for _, set := range sets {
		if set.Active() {
			active = append(active, set)
		}
	}
	switch len(active) {
	case 0:
		return Set{}
	case 1:
		return active[0]
	default:
		return Set{all: active}
	}
}

// Active reports whether at least one PR filter is configured.
func (s Set) Active() bool {
	for _, set := range s.all {
		if set.Active() {
			return true
		}
	}
	return s.authorMode != authorAny || s.draft != nil || s.bot != nil || s.reviewDecision != reviewDecisionAny
}

// Matches reports whether pr satisfies every active filter.
func (s Set) Matches(pr *models.PR, ctx Context) bool {
	if pr == nil {
		return false
	}
	for _, set := range s.all {
		if !set.Matches(pr, ctx) {
			return false
		}
	}

	if !s.matchesAuthor(pr, ctx) {
		return false
	}
	if s.draft != nil {
		isDraft := pr.EffectiveState() == models.PRStateDraft
		if isDraft != *s.draft {
			return false
		}
	}
	if s.bot != nil {
		if ctx.isBot(pr.Author) != *s.bot {
			return false
		}
	}
	if !s.matchesReviewDecision(pr) {
		return false
	}

	return true
}

// Apply returns a shallow copy of result with MatchingPRs populated from the
// non-My categories. MyPRs are always preserved and never duplicated into
// MatchingPRs.
func (s Set) Apply(result *models.ScanResult, cfg *config.Config) *models.ScanResult {
	if result == nil {
		return nil
	}
	if !s.Active() {
		return result
	}

	filtered := *result
	filtered.MatchingPRs = make([]*models.PR, 0)

	ctx := contextFromResult(result, cfg)
	for _, pr := range nonMyPRs(result) {
		if pr == nil || pr.Author == ctx.Username {
			continue
		}
		if s.Matches(pr, ctx) {
			filtered.MatchingPRs = append(filtered.MatchingPRs, pr)
		}
	}

	return &filtered
}

func (s Set) matchesAuthor(pr *models.PR, ctx Context) bool {
	switch s.authorMode {
	case authorAny:
		return true
	case authorMe:
		return pr.Author == ctx.Username
	case authorTeam:
		return pr.Author != ctx.Username && contains(ctx.TeamMembers, pr.Author) && !ctx.isBot(pr.Author)
	case authorOther:
		return pr.Author != ctx.Username && !contains(ctx.TeamMembers, pr.Author) && !ctx.isBot(pr.Author)
	case authorUsername:
		return pr.Author == s.authorUsername
	default:
		return false
	}
}

func (s Set) matchesReviewDecision(pr *models.PR) bool {
	switch s.reviewDecision {
	case reviewDecisionAny:
		return true
	case reviewDecisionApproved:
		return pr.ReviewDecision == models.ReviewDecisionApproved
	case reviewDecisionNotApproved:
		return pr.ReviewDecision != models.ReviewDecisionApproved
	case reviewDecisionReviewRequired:
		return pr.ReviewDecision == models.ReviewDecisionReviewRequired
	case reviewDecisionChangesRequested:
		return pr.ReviewDecision == models.ReviewDecisionChangesRequested
	case reviewDecisionNone:
		return pr.ReviewDecision == models.ReviewDecisionNone
	default:
		return false
	}
}

func parseAuthor(value string) (authorMode, string, error) {
	switch value {
	case string(authorMe):
		return authorMe, "", nil
	case string(authorTeam):
		return authorTeam, "", nil
	case string(authorOther):
		return authorOther, "", nil
	default:
		if strings.HasPrefix(value, "@") && strings.TrimPrefix(value, "@") != "" {
			return authorUsername, strings.TrimPrefix(value, "@"), nil
		}
		return authorAny, "", fmt.Errorf("invalid --author value %q (must be me, team, other, or @username)", value)
	}
}

func parseBoolFlag(name, value string) (bool, error) {
	switch value {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid --%s value %q (must be true or false)", name, value)
	}
}

func parseReviewDecision(value string) (reviewDecisionMode, error) {
	switch value {
	case string(reviewDecisionApproved):
		return reviewDecisionApproved, nil
	case string(reviewDecisionNotApproved):
		return reviewDecisionNotApproved, nil
	case string(reviewDecisionReviewRequired):
		return reviewDecisionReviewRequired, nil
	case string(reviewDecisionChangesRequested):
		return reviewDecisionChangesRequested, nil
	case string(reviewDecisionNone):
		return reviewDecisionNone, nil
	default:
		return reviewDecisionAny, fmt.Errorf("invalid --review-decision value %q (must be approved, not-approved, review-required, changes-requested, or none)", value)
	}
}

func contextFromResult(result *models.ScanResult, cfg *config.Config) Context {
	ctx := Context{Username: result.Username}
	if cfg == nil {
		return ctx
	}
	if cfg.GitHubUsername != "" {
		ctx.Username = cfg.GitHubUsername
	}
	ctx.TeamMembers = cfg.TeamMembers
	ctx.BotAuthors = cfg.Bots
	return ctx
}

func nonMyPRs(result *models.ScanResult) []*models.PR {
	total := len(result.NeedsMyAttention) + len(result.TeamPRs) + len(result.OtherPRs)
	prs := make([]*models.PR, 0, total)
	prs = append(prs, result.NeedsMyAttention...)
	prs = append(prs, result.TeamPRs...)
	prs = append(prs, result.OtherPRs...)
	return prs
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func (ctx Context) isBot(author string) bool {
	if IsBotAuthor(author, ctx.BotAuthors) {
		return true
	}
	return ctx.IsBot != nil && ctx.IsBot(author)
}

// IsBotAuthor reports whether author is a known Bot Author.
func IsBotAuthor(author string, configuredBots []string) bool {
	return contains(configuredBots, author) || strings.HasSuffix(author, "[bot]")
}

func sortedFilterKeys(filters map[string]interface{}) []string {
	keys := make([]string, 0, len(filters))
	for key := range filters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Package display provides terminal rendering for PRT output.
// It uses lipgloss for consistent styling across different terminal types.
package display

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Style definitions for terminal output.
// These styles provide consistent visual theming for the display system.
var (
	// HeaderStyle renders section headers (MY PRS, NEEDS ATTENTION, etc.)
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")). // White
			Background(lipgloss.Color("57")). // Purple
			Padding(0, 1)

	// SubheaderStyle renders repository names within sections
	SubheaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("244")) // Gray

	// DraftStyle renders draft PRs (dimmed, italic)
	DraftStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")). // Gray
			Italic(true)

	// NeedsReviewStyle renders PRs waiting for review
	NeedsReviewStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")) // Green

	// ApprovedStyle renders approved PRs
	ApprovedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")) // Blue

	// ChangesRequestedStyle renders PRs with requested changes
	ChangesRequestedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")) // Orange

	// BlockedStyle renders blocked PRs (stacked PRs waiting on parent)
	BlockedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")). // Gray
			Faint(true)

	// CIPassingStyle renders passing CI status
	CIPassingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")) // Green

	// CIFailingStyle renders failing CI status
	CIFailingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")) // Red

	// CIPendingStyle renders pending CI status
	CIPendingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")) // Yellow

	// URLStyle renders clickable URLs
	URLStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")). // Blue
			Underline(true)

	// TreeStyle renders tree drawing characters
	TreeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // Dark gray

	// EmptyStyle renders empty state messages
	EmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true)

	// MetaStyle renders metadata (age, author, etc.)
	MetaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	// RepoStyle renders repository names
	RepoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("255")) // White - clean, high contrast header

	// TitleStyle renders the main PRT header
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")) // Pink/magenta

	// NumberStyle renders PR numbers (#123)
	NumberStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")) // Blue

	// AuthorStyle renders author names (@username)
	AuthorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")) // Orange

	// BranchStyle renders branch names
	BranchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("141")) // Light purple

	// SummaryStyle renders the footer summary line
	SummaryStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true)
)

// Icon constants for enhanced visual display.
// These are only shown when show_icons is enabled in config.
const (
	// Section icons
	IconMyPRs          = "\U0001F4CB" // Clipboard
	IconNeedsAttention = "\U0001F440" // Eyes
	IconTeam           = "\U0001F465" // Busts in silhouette
	IconOther          = "\U0001F916" // Robot
	IconNoOpenPRs      = "\U0001F4C2" // Open folder

	// PR state icons
	IconDraft    = "\U0001F4DD" // Memo
	IconMerged   = "\U0001F7E3" // Purple circle
	IconApproved = "\u2705"     // Check mark
	IconChanges  = "\U0001F504" // Arrows counterclockwise
	IconReview   = "\U0001F440" // Eyes
	IconBlocked  = "\U0001F512" // Lock

	// CI status icons
	IconCIPassing = "\u2705" // Check mark
	IconCIFailing = "\u274C" // Cross mark
	IconCIPending = "\u23F3" // Hourglass

	// Other icons
	IconRepo  = "\U0001F4E6" // Package
	IconEmpty = "\u2205"     // Empty set
)

// Tree drawing characters for rendering stacked PR hierarchies.
const (
	TreeVertical   = "\u2502"             // │
	TreeBranch     = "\u251C\u2500\u2500" // ├──
	TreeLastBranch = "\u2514\u2500\u2500" // └──
	TreeIndent     = "    "
)

// DisableColors disables all color output.
// Call this when --no-color flag is set or when output is not a TTY.
func DisableColors() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// EnableColors re-enables color output with automatic detection.
// This uses the terminal's color profile detection.
func EnableColors() {
	// Reset to TrueColor for full color support
	lipgloss.SetColorProfile(termenv.TrueColor)
}

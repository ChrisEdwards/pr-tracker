package display

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"prt/internal/models"

	"github.com/charmbracelet/lipgloss"
)

// Progress icons
const (
	IconSuccess   = "✓"
	IconError     = "✗"
	IconSpinner   = "⠋"
	IconBarFilled = "█"
	IconBarEmpty  = "░"
	IconPause     = "⏸"
)

// ASCII fallback icons
const (
	IconSuccessASCII   = "+"
	IconErrorASCII     = "x"
	IconBarFilledASCII = "="
	IconBarEmptyASCII  = "-"
)

// Progress bar styles
var (
	// ProgressBarStyle renders the progress bar
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")) // Blue

	// ProgressTextStyle renders progress percentage
	ProgressTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")) // Gray

	// SuccessStyle renders successful repo results
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")) // Green

	// ErrorStyle renders failed repo results
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")) // Red

	// WarningStyle renders rate-limited or warning results
	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")) // Yellow

	// DimStyle renders dimmed text (0 PRs, secondary info)
	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // Dark gray

	// ProgressHeaderStyle renders the scanning header
	ProgressHeaderStyle = lipgloss.NewStyle().
				Bold(true)
)

// ProgressDisplay shows scanning progress with a progress bar and results.
type ProgressDisplay struct {
	total     int
	done      int
	results   []string
	mu        sync.Mutex
	writer    io.Writer
	barWidth  int
	cleared   bool
	startTime time.Time
	isTTY     bool
	useASCII  bool

	// PR counts for summary
	totalPRs   int
	yourPRs    int
	needReview int
}

// ProgressOption configures a ProgressDisplay.
type ProgressOption func(*ProgressDisplay)

// WithWriter sets the output writer for the progress display.
func WithWriter(w io.Writer) ProgressOption {
	return func(p *ProgressDisplay) {
		p.writer = w
	}
}

// WithBarWidth sets the width of the progress bar.
func WithBarWidth(width int) ProgressOption {
	return func(p *ProgressDisplay) {
		if width > 0 {
			p.barWidth = width
		}
	}
}

// WithTTY indicates whether the output is a TTY.
// When false, uses simple line-by-line output without screen clearing.
func WithTTY(isTTY bool) ProgressOption {
	return func(p *ProgressDisplay) {
		p.isTTY = isTTY
	}
}

// WithASCII enables ASCII-only output (no Unicode characters).
func WithASCII(useASCII bool) ProgressOption {
	return func(p *ProgressDisplay) {
		p.useASCII = useASCII
	}
}

// NewProgressDisplay creates a new progress display for tracking repo scans.
func NewProgressDisplay(total int, opts ...ProgressOption) *ProgressDisplay {
	p := &ProgressDisplay{
		total:     total,
		results:   make([]string, 0, total),
		writer:    nil, // Will be set to os.Stdout if nil
		barWidth:  40,
		startTime: time.Now(),
		isTTY:     true, // Assume TTY by default
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Update records a completed repository scan and re-renders the display.
func (p *ProgressDisplay) Update(repo *models.Repository) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.done++

	// Get the right icons based on ASCII mode
	successIcon := IconSuccess
	errorIcon := IconError
	if p.useASCII {
		successIcon = IconSuccessASCII
		errorIcon = IconErrorASCII
	}

	// Build result line based on scan status
	var line string
	switch repo.ScanStatus {
	case models.ScanStatusSuccess:
		prCount := len(repo.PRs)
		p.totalPRs += prCount
		plural := "PRs"
		if prCount == 1 {
			plural = "PR"
		}
		if prCount == 0 {
			line = DimStyle.Render(fmt.Sprintf("%s %s (0 PRs)",
				successIcon, repo.Name))
		} else {
			line = SuccessStyle.Render(fmt.Sprintf("%s %s (%d %s)",
				successIcon, repo.Name, prCount, plural))
		}
	case models.ScanStatusNoPRs:
		line = DimStyle.Render(fmt.Sprintf("%s %s (0 PRs)",
			successIcon, repo.Name))
	case models.ScanStatusError:
		errMsg := "error"
		if repo.ScanError != nil {
			errMsg = repo.ScanError.Error()
			// Check for rate limiting
			if strings.Contains(errMsg, "rate limit") {
				line = WarningStyle.Render(fmt.Sprintf("%s %s (rate limited)",
					IconPause, repo.Name))
			} else {
				// Truncate long error messages
				if len(errMsg) > 50 {
					errMsg = errMsg[:47] + "..."
				}
				line = ErrorStyle.Render(fmt.Sprintf("%s %s (%s)",
					errorIcon, repo.Name, errMsg))
			}
		} else {
			line = ErrorStyle.Render(fmt.Sprintf("%s %s (%s)",
				errorIcon, repo.Name, errMsg))
		}
	case models.ScanStatusSkipped:
		line = DimStyle.Render(fmt.Sprintf("- %s (skipped)", repo.Name))
	}

	p.results = append(p.results, line)

	p.render()
}

// render outputs the current progress state.
func (p *ProgressDisplay) render() {
	if p.writer == nil {
		return
	}

	// Non-TTY mode: simple line output
	if !p.isTTY {
		// Just print the latest result
		if len(p.results) > 0 {
			fmt.Fprintln(p.writer, p.results[len(p.results)-1])
		}
		return
	}

	// TTY mode: full interactive display
	p.renderTTY()
}

// renderTTY renders the full interactive progress display for TTY output.
func (p *ProgressDisplay) renderTTY() {
	// Calculate progress percentage
	pct := float64(p.done) / float64(p.total)
	filled := int(pct * float64(p.barWidth))
	if filled > p.barWidth {
		filled = p.barWidth
	}

	// Get the right icons based on ASCII mode
	barFilled := IconBarFilled
	barEmpty := IconBarEmpty
	if p.useASCII {
		barFilled = IconBarFilledASCII
		barEmpty = IconBarEmptyASCII
	}

	// Build progress bar
	bar := strings.Repeat(barFilled, filled) +
		strings.Repeat(barEmpty, p.barWidth-filled)

	// Calculate elapsed time
	elapsed := time.Since(p.startTime)
	elapsedStr := fmt.Sprintf("%.1fs", elapsed.Seconds())

	// Clear screen and move cursor to top-left
	fmt.Fprint(p.writer, "\033[2J\033[H")

	// Header
	fmt.Fprintf(p.writer, "%s\n\n",
		ProgressHeaderStyle.Render(fmt.Sprintf("Fetching PRs from %d repositories...", p.total)))

	// Progress bar with count, percentage, and elapsed time
	fmt.Fprintf(p.writer, "  %s  %d/%d  %s  %s\n\n",
		ProgressBarStyle.Render(bar),
		p.done, p.total,
		ProgressTextStyle.Render(fmt.Sprintf("%d%%", int(pct*100))),
		DimStyle.Render(elapsedStr))

	// Results
	for _, r := range p.results {
		fmt.Fprintf(p.writer, "  %s\n", r)
	}
}

// Finish completes the progress display.
// It clears the progress output and returns the final summary.
func (p *ProgressDisplay) Finish() Summary {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Count successes, errors, etc.
	summary := Summary{
		Total:    p.total,
		Done:     p.done,
		Elapsed:  time.Since(p.startTime),
		TotalPRs: p.totalPRs,
	}

	// Count using icon detection - check ASCII icons too
	for _, r := range p.results {
		if strings.Contains(r, IconSuccess) || strings.Contains(r, IconSuccessASCII) {
			summary.Success++
		} else if strings.Contains(r, IconError) || strings.Contains(r, IconErrorASCII) ||
			strings.Contains(r, IconPause) {
			summary.Errors++
		} else {
			summary.Skipped++
		}
	}

	return summary
}

// Clear clears the progress output without printing summary.
func (p *ProgressDisplay) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.writer != nil && !p.cleared {
		fmt.Fprint(p.writer, "\033[2J\033[H")
		p.cleared = true
	}
}

// Summary contains the final counts from a progress display.
type Summary struct {
	Total    int
	Done     int
	Success  int
	Errors   int
	Skipped  int
	Elapsed  time.Duration
	TotalPRs int
}

// String returns a human-readable summary.
func (s Summary) String() string {
	parts := []string{fmt.Sprintf("%d repos scanned", s.Done)}

	if s.Success > 0 {
		parts = append(parts, fmt.Sprintf("%d with PRs", s.Success))
	}
	if s.Errors > 0 {
		parts = append(parts, fmt.Sprintf("%d errors", s.Errors))
	}
	if s.Skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", s.Skipped))
	}

	return strings.Join(parts, ", ")
}

// RichString returns a formatted summary with elapsed time.
func (s Summary) RichString() string {
	return fmt.Sprintf("Done! Scanned %d repos in %.1fs\n\n  Found %d open PRs across %d repos",
		s.Done, s.Elapsed.Seconds(), s.TotalPRs, s.Success)
}

// ProgressCallback returns a FetchProgress function for use with FetchAllPRs.
func (p *ProgressDisplay) ProgressCallback() func(done, total int, repo *models.Repository) {
	return func(done, total int, repo *models.Repository) {
		p.Update(repo)
	}
}

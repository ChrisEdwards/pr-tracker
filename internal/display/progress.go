package display

import (
	"fmt"
	"io"
	"strings"
	"sync"

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

	// ProgressHeaderStyle renders the scanning header
	ProgressHeaderStyle = lipgloss.NewStyle().
				Bold(true)
)

// ProgressDisplay shows scanning progress with a progress bar and results.
type ProgressDisplay struct {
	total    int
	done     int
	results  []string
	mu       sync.Mutex
	writer   io.Writer
	barWidth int
	cleared  bool
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

// NewProgressDisplay creates a new progress display for tracking repo scans.
func NewProgressDisplay(total int, opts ...ProgressOption) *ProgressDisplay {
	p := &ProgressDisplay{
		total:    total,
		results:  make([]string, 0, total),
		writer:   nil, // Will be set to os.Stdout if nil
		barWidth: 40,
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

	// Build result line based on scan status
	var line string
	switch repo.ScanStatus {
	case models.ScanStatusSuccess:
		prCount := len(repo.PRs)
		plural := "PRs"
		if prCount == 1 {
			plural = "PR"
		}
		line = SuccessStyle.Render(fmt.Sprintf("%s %s (%d %s)",
			IconSuccess, repo.Name, prCount, plural))
	case models.ScanStatusNoPRs:
		line = SuccessStyle.Render(fmt.Sprintf("%s %s (0 PRs)",
			IconSuccess, repo.Name))
	case models.ScanStatusError:
		errMsg := "error"
		if repo.ScanError != nil {
			errMsg = repo.ScanError.Error()
			// Truncate long error messages
			if len(errMsg) > 50 {
				errMsg = errMsg[:47] + "..."
			}
		}
		line = ErrorStyle.Render(fmt.Sprintf("%s %s (%s)",
			IconError, repo.Name, errMsg))
	case models.ScanStatusSkipped:
		line = ProgressTextStyle.Render(fmt.Sprintf("- %s (skipped)", repo.Name))
	}

	p.results = append(p.results, line)

	p.render()
}

// render outputs the current progress state.
func (p *ProgressDisplay) render() {
	if p.writer == nil {
		return
	}

	// Calculate progress percentage
	pct := float64(p.done) / float64(p.total)
	filled := int(pct * float64(p.barWidth))
	if filled > p.barWidth {
		filled = p.barWidth
	}

	// Build progress bar
	bar := strings.Repeat(IconBarFilled, filled) +
		strings.Repeat(IconBarEmpty, p.barWidth-filled)

	// Clear screen and move cursor to top-left
	fmt.Fprint(p.writer, "\033[2J\033[H")

	// Header
	fmt.Fprintln(p.writer, ProgressHeaderStyle.Render("Scanning repositories..."))
	fmt.Fprintln(p.writer)

	// Progress bar
	fmt.Fprintf(p.writer, "%s %s (%d/%d)\n\n",
		ProgressBarStyle.Render(bar),
		ProgressTextStyle.Render(fmt.Sprintf("%d%%", int(pct*100))),
		p.done, p.total)

	// Results
	for _, r := range p.results {
		fmt.Fprintln(p.writer, r)
	}
}

// Finish completes the progress display.
// It clears the progress output and returns the final summary.
func (p *ProgressDisplay) Finish() Summary {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Count successes, errors, etc.
	summary := Summary{
		Total: p.total,
		Done:  p.done,
	}

	for _, r := range p.results {
		if strings.Contains(r, IconSuccess) {
			summary.Success++
		} else if strings.Contains(r, IconError) {
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
	Total   int
	Done    int
	Success int
	Errors  int
	Skipped int
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

// ProgressCallback returns a FetchProgress function for use with FetchAllPRs.
func (p *ProgressDisplay) ProgressCallback() func(done, total int, repo *models.Repository) {
	return func(done, total int, repo *models.Repository) {
		p.Update(repo)
	}
}

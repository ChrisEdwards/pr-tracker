package display

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

// Spinner frames for animated progress indicator.
const (
	SpinnerInterval = 80 * time.Millisecond
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// ASCII fallback for non-Unicode terminals
var spinnerFramesASCII = []string{"|", "/", "-", "\\"}

// Spinner provides an animated spinner for progress indication.
type Spinner struct {
	mu       sync.Mutex
	writer   io.Writer
	frame    int
	message  string
	count    int
	running  bool
	stopCh   chan struct{}
	style    lipgloss.Style
	useASCII bool
}

// NewSpinner creates a new spinner that writes to the given writer.
func NewSpinner(w io.Writer) *Spinner {
	return &Spinner{
		writer:   w,
		style:    lipgloss.NewStyle().Foreground(lipgloss.Color("39")), // Blue
		useASCII: false,
	}
}

// SetASCII enables ASCII-only spinner frames.
func (s *Spinner) SetASCII(ascii bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.useASCII = ascii
}

// Start begins the spinner animation with the given message.
func (s *Spinner) Start(message string) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.message = message
	s.count = 0
	s.frame = 0
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	go s.animate()
}

// UpdateCount updates the count shown in the spinner.
func (s *Spinner) UpdateCount(count int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.count = count
}

// UpdateMessage updates the message shown in the spinner.
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = message
}

// Stop halts the spinner animation.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()
}

// animate runs the spinner animation loop.
func (s *Spinner) animate() {
	ticker := time.NewTicker(SpinnerInterval)
	defer ticker.Stop()

	s.render()

	for {
		select {
		case <-s.stopCh:
			s.clearLine()
			return
		case <-ticker.C:
			s.mu.Lock()
			s.frame++
			s.mu.Unlock()
			s.render()
		}
	}
}

// render draws the current spinner state.
func (s *Spinner) render() {
	s.mu.Lock()
	defer s.mu.Unlock()

	frames := spinnerFrames
	if s.useASCII {
		frames = spinnerFramesASCII
	}
	frame := frames[s.frame%len(frames)]

	// Build the line
	line := fmt.Sprintf("\r%s %s", s.style.Render(frame), s.message)
	if s.count > 0 {
		line += fmt.Sprintf("  %d found", s.count)
	}
	// Pad with spaces to clear any previous longer content
	line += "          "

	fmt.Fprint(s.writer, line)
}

// clearLine clears the current line.
func (s *Spinner) clearLine() {
	fmt.Fprint(s.writer, "\r\033[K")
}

// IsTTY returns true if the given writer is a terminal.
func IsTTY(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
	}
	return false
}

// ShouldUseColors returns true if colors should be used.
// It checks NO_COLOR env var and TTY status.
func ShouldUseColors(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return IsTTY(w)
}

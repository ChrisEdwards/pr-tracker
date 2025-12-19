package display

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func TestNewSpinner(t *testing.T) {
	buf := &bytes.Buffer{}
	s := NewSpinner(buf)

	if s == nil {
		t.Fatal("NewSpinner returned nil")
	}
	if s.writer != buf {
		t.Error("writer not set correctly")
	}
	if s.running {
		t.Error("spinner should not be running initially")
	}
}

func TestSpinner_SetASCII(t *testing.T) {
	buf := &bytes.Buffer{}
	s := NewSpinner(buf)

	s.SetASCII(true)
	if !s.useASCII {
		t.Error("SetASCII(true) should set useASCII to true")
	}

	s.SetASCII(false)
	if s.useASCII {
		t.Error("SetASCII(false) should set useASCII to false")
	}
}

func TestSpinner_StartStop(t *testing.T) {
	buf := &bytes.Buffer{}
	s := NewSpinner(buf)

	// Start the spinner
	s.Start("Testing...")

	if !s.running {
		t.Error("spinner should be running after Start")
	}

	// Give it time to render at least once
	time.Sleep(100 * time.Millisecond)

	// Stop the spinner
	s.Stop()

	if s.running {
		t.Error("spinner should not be running after Stop")
	}

	// Verify something was written
	if buf.Len() == 0 {
		t.Error("spinner should have written to buffer")
	}
}

func TestSpinner_UpdateCount(t *testing.T) {
	buf := &bytes.Buffer{}
	s := NewSpinner(buf)

	s.Start("Testing...")
	s.UpdateCount(42)

	// Give it time to render with the new count
	time.Sleep(100 * time.Millisecond)

	s.Stop()

	output := buf.String()
	if len(output) == 0 {
		t.Error("spinner should have written to buffer")
	}
}

func TestSpinner_UpdateMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	s := NewSpinner(buf)

	s.Start("Initial message")
	s.UpdateMessage("Updated message")

	// Give it time to render with the new message
	time.Sleep(100 * time.Millisecond)

	s.Stop()

	// Verify output was written
	if buf.Len() == 0 {
		t.Error("spinner should have written to buffer")
	}
}

func TestSpinner_DoubleStart(t *testing.T) {
	buf := &bytes.Buffer{}
	s := NewSpinner(buf)

	s.Start("First")
	s.Start("Second") // Should be a no-op

	time.Sleep(50 * time.Millisecond)
	s.Stop()

	// Should not panic and should stop cleanly
	if s.running {
		t.Error("spinner should have stopped")
	}
}

func TestSpinner_DoubleStop(t *testing.T) {
	buf := &bytes.Buffer{}
	s := NewSpinner(buf)

	s.Start("Test")
	time.Sleep(50 * time.Millisecond)
	s.Stop()
	s.Stop() // Should be a no-op, not panic

	if s.running {
		t.Error("spinner should still be stopped")
	}
}

func TestSpinner_StopWithoutStart(t *testing.T) {
	buf := &bytes.Buffer{}
	s := NewSpinner(buf)

	// Should not panic
	s.Stop()

	if s.running {
		t.Error("spinner should not be running")
	}
}

func TestIsTTY(t *testing.T) {
	// Test with a buffer (not a TTY)
	buf := &bytes.Buffer{}
	if IsTTY(buf) {
		t.Error("buffer should not be detected as TTY")
	}

	// Test with os.Stdout - result depends on environment
	// We can't make strong assertions here, but it shouldn't panic
	_ = IsTTY(os.Stdout)
}

func TestShouldUseColors(t *testing.T) {
	// Test with a buffer (not a TTY)
	buf := &bytes.Buffer{}
	if ShouldUseColors(buf) {
		t.Error("buffer should not use colors")
	}

	// Test NO_COLOR handling would require manipulating env vars
	// which is not safe in parallel tests
}

func TestSpinnerFrames(t *testing.T) {
	// Verify spinner frames are defined
	if len(spinnerFrames) == 0 {
		t.Error("spinnerFrames should not be empty")
	}
	if len(spinnerFramesASCII) == 0 {
		t.Error("spinnerFramesASCII should not be empty")
	}
}

func TestSpinner_ASCIIFrames(t *testing.T) {
	buf := &bytes.Buffer{}
	s := NewSpinner(buf)
	s.SetASCII(true)

	s.Start("Testing...")
	time.Sleep(100 * time.Millisecond)
	s.Stop()

	output := buf.String()
	if len(output) == 0 {
		t.Error("spinner should have written to buffer")
	}
}

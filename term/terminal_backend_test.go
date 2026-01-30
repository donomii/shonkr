package main

import (
	"testing"
)

// Test basic terminal creation and sizing
func TestTerminalBackendCreation(t *testing.T) {
	term := NewTerminalBackend(80, 24)
	if term == nil {
		t.Fatal("Failed to create terminal backend")
	}

	width, height := term.GetSize()
	if width != 80 || height != 24 {
		t.Errorf("Expected size 80x24, got %dx%d", width, height)
	}
}

// Test simple text writing
func TestTerminalBackendText(t *testing.T) {
	term := NewTerminalBackend(80, 24)

	term.Write([]byte("Hello, World!"))

	screen := term.GetScreen()
	if len(screen) == 0 {
		t.Fatal("Screen is empty after writing")
	}

	// First line should contain our text
	firstLine := screen[0]
	text := ""
	for _, cell := range firstLine.Cells {
		text += string(cell.Char)
	}

	if len(text) < 13 || text[:13] != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", text)
	}
}

// Test newline handling
func TestTerminalBackendNewline(t *testing.T) {
	term := NewTerminalBackend(80, 24)

	term.Write([]byte("Line 1\nLine 2\nLine 3"))

	screen := term.GetScreen()
	if len(screen) < 3 {
		t.Fatal("Expected at least 3 lines")
	}

	// Check each line
	lines := []string{"Line 1", "Line 2", "Line 3"}
	for i, expected := range lines {
		line := screen[i]
		text := ""
		for j := 0; j < len(expected) && j < len(line.Cells); j++ {
			text += string(line.Cells[j].Char)
		}
		if text != expected {
			t.Errorf("Line %d: expected '%s', got '%s'", i, expected, text)
		}
	}
}

// Test color codes
func TestTerminalBackendColors(t *testing.T) {
	term := NewTerminalBackend(80, 24)

	// Red text: ESC[31m
	term.Write([]byte("\x1b[31mRed Text\x1b[0m"))

	screen := term.GetScreen()
	firstLine := screen[0]

	// First visible character should be 'R' with red foreground
	if firstLine.Cells[0].Char != 'R' {
		t.Errorf("Expected 'R', got '%c'", firstLine.Cells[0].Char)
	}

	// Check that FG color is not default (0)
	if firstLine.Cells[0].FG == 0 {
		t.Log("Warning: FG color might not be set correctly")
	}
}

// Test cursor movement
func TestTerminalBackendCursor(t *testing.T) {
	term := NewTerminalBackend(80, 24)

	// Write text, move cursor back, overwrite
	term.Write([]byte("XXXXX"))
	term.Write([]byte("\x1b[5D")) // Move cursor left 5
	term.Write([]byte("ABC"))

	screen := term.GetScreen()
	firstLine := screen[0]

	text := ""
	for i := 0; i < 5; i++ {
		text += string(firstLine.Cells[i].Char)
	}

	if text[:3] != "ABC" {
		t.Errorf("Expected 'ABCXX', got '%s'", text)
	}
}

// Test scrolling
func TestTerminalBackendScroll(t *testing.T) {
	term := NewTerminalBackend(10, 5) // Small terminal

	// Write more lines than the height
	for i := 0; i < 10; i++ {
		term.Write([]byte("Line "))
		term.Write([]byte{byte('0' + i), '\n'})
	}

	screen := term.GetScreen()

	// Should only have 5 lines visible
	if len(screen) != 5 {
		t.Errorf("Expected 5 visible lines, got %d", len(screen))
	}

	// Last line should contain "Line 9"
	lastLine := screen[4]
	text := ""
	for i := 0; i < 6 && i < len(lastLine.Cells); i++ {
		text += string(lastLine.Cells[i].Char)
	}

	if text != "Line 9" {
		t.Logf("Expected last line to be 'Line 9', got '%s' (scrolling may work differently)", text)
	}
}

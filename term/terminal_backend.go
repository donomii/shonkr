package main

import (
	"github.com/hinshun/vt10x"
)

// TerminalCell represents a single character cell with styling
type TerminalCell struct {
	Char rune
	FG   int // Foreground color
	BG   int // Background color
	Mode int // Mode flags (contains bold, underline, etc)
}

// TerminalLine represents a line of cells
type TerminalLine struct {
	Cells []TerminalCell
}

// TerminalBackend wraps vt10x to provide terminal emulation
type TerminalBackend struct {
	term   vt10x.Terminal
	width  int
	height int
}

// NewTerminalBackend creates a new terminal backend
func NewTerminalBackend(width, height int) *TerminalBackend {
	term := vt10x.New(vt10x.WithSize(width, height))

	return &TerminalBackend{
		term:   term,
		width:  width,
		height: height,
	}
}

// Write writes data to the terminal (processes VT100 sequences)
func (tb *TerminalBackend) Write(data []byte) (int, error) {
	return tb.term.Write(data)
}

func (tb *TerminalBackend) Resize(cols, rows int) {
	if tb.term != nil {
		tb.term.Resize(cols, rows)
		tb.width = cols
		tb.height = rows
	}
}

// GetSize returns the terminal dimensions
func (tb *TerminalBackend) GetSize() (int, int) {
	return tb.width, tb.height
}

// SetSize resizes the terminal
func (tb *TerminalBackend) SetSize(width, height int) {
	tb.width = width
	tb.height = height
	tb.term.Resize(width, height)
}

// GetScreen returns the current screen state
func (tb *TerminalBackend) GetScreen() []TerminalLine {
	lines := make([]TerminalLine, tb.height)

	tb.term.Lock()
	defer tb.term.Unlock()

	for row := 0; row < tb.height; row++ {
		cells := make([]TerminalCell, tb.width)
		for col := 0; col < tb.width; col++ {
			glyph := tb.term.Cell(col, row)

			cells[col] = TerminalCell{
				Char: glyph.Char,
				FG:   int(glyph.FG),
				BG:   int(glyph.BG),
				Mode: int(glyph.Mode),
			}
		}
		lines[row] = TerminalLine{Cells: cells}
	}

	return lines
}

// GetCursor returns the cursor position
func (tb *TerminalBackend) GetCursor() (int, int) {
	tb.term.Lock()
	defer tb.term.Unlock()

	cursor := tb.term.Cursor()
	return cursor.X, cursor.Y
}

// String returns the visible text as a string (for debugging)
func (tb *TerminalBackend) String() string {
	var result string
	screen := tb.GetScreen()
	for _, line := range screen {
		for _, cell := range line.Cells {
			result += string(cell.Char)
		}
		result += "\n"
	}
	return result
}

package screenwriter

import (
	"io"
	"strings"
)

// ScreenWriter allows writing to the screen and then overwriting that content.
type ScreenWriter struct {
	writer   io.StringWriter
	previous string
}

// New creates a new ScreenWriter which is ready to be used.
func New(writer io.StringWriter) *ScreenWriter {
	return &ScreenWriter{
		writer: writer,
	}
}

func (w *ScreenWriter) clearPrevious() error {
	if w.previous == "" {
		return nil
	}

	lines := strings.Split(w.previous, "\n")

	// Move the cursor to the start of the line
	finalLine := lines[len(lines)-1]
	finalLineLength := len(finalLine)
	err := moveCursor(w.writer, directionLeft, finalLineLength)
	if err != nil {
		return err
	}

	// We need to clear the last line in case the string does not end with a "\n"
	err = clearLine(w.writer)
	if err != nil {
		return err
	}

	// Go up and clear n lines
	for i := 0; i < len(lines)-1; i++ {
		// Go up one line
		err = moveCursor(w.writer, directionUp, 1)
		if err != nil {
			return err
		}
		// Clear entire line
		err = clearLine(w.writer)
		if err != nil {
			return err
		}
	}

	return nil
}

// Display clears the previously printed content from the screen (if necessary), then writes the provided content to the screen.
func (w *ScreenWriter) Display(s string) error {
	err := w.clearPrevious()
	if err != nil {
		return err
	}

	w.previous = s

	_, err = w.writer.WriteString(s)
	return err
}

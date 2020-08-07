package screenwriter

import (
	"fmt"
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

	// Get number of lines in previous
	lines := strings.Count(w.previous, "\n")

	// Handle the final line
	finalLineLength := 0
	splitLines := strings.Split(w.previous, "\n")
	finalLine := splitLines[len(splitLines)-1]
	finalLineLength = len(finalLine)
	_, err := w.writer.WriteString(fmt.Sprintf("\u001b[%dD", finalLineLength))
	if err != nil {
		return err
	}

	// Clear entire line
	// This is necessary if the last line has content (i.e. the string does not end with \n)
	_, err = w.writer.WriteString("\u001b[2K")
	if err != nil {
		return err
	}

	// Go up and clear n lines
	for i := 0; i < lines; i++ {
		// Go up one line
		_, err = w.writer.WriteString("\u001b[1A")
		if err != nil {
			return err
		}
		// Clear entire line
		_, err = w.writer.WriteString("\u001b[2K")
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

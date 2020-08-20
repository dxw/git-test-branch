package screenwriter

import (
	"io"
	"strings"
)

// ScreenWriter allows writing messages to the screen and then overwriting those messages.
type ScreenWriter struct {
	writer          io.StringWriter
	previousMessage string
}

// New creates a new ScreenWriter which is ready to be used.
func New(writer io.StringWriter) *ScreenWriter {
	return &ScreenWriter{
		writer: writer,
	}
}

func (w *ScreenWriter) clearPrevious() error {
	if w.previousMessage == "" {
		return nil
	}

	lines := strings.Split(w.previousMessage, "\n")

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

// Display clears the previously printed message from the screen (if necessary), then writes the provided message to the screen.
func (w *ScreenWriter) Display(message string) error {
	err := w.clearPrevious()
	if err != nil {
		return err
	}

	w.previousMessage = message

	_, err = w.writer.WriteString(message)
	return err
}

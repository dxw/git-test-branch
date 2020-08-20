package screenwriter

import (
	"fmt"
	"io"
)

type direction string

const (
	directionUp   direction = "A"
	directionLeft           = "D"
)

const ansiEscape = "\u001b["
const lineClear = "2K"

func moveCursor(w io.StringWriter, direction direction, count int) error {
	_, err := w.WriteString(fmt.Sprintf("%s%d%s", ansiEscape, count, string(direction)))
	return err
}

func clearLine(w io.StringWriter) error {
	_, err := w.WriteString(fmt.Sprintf("%s%s", ansiEscape, lineClear))
	return err
}

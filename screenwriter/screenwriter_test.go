package screenwriter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisplayOnce(t *testing.T) {
	w := strings.Builder{}
	sw := New(&w)
	err := sw.Display("xyz")
	assert.Nil(t, err)
	assert.Equal(t, "xyz", w.String())
}

func TestDisplayTwice(t *testing.T) {
	w := strings.Builder{}
	sw := New(&w)
	err := sw.Display("xyz")
	assert.Nil(t, err)
	err = sw.Display("abc")
	assert.Nil(t, err)
	assert.Equal(t, "xyz\u001b[3D\u001b[2Kabc", w.String())
}

func TestDisplayTwiceNewlines(t *testing.T) {
	w := strings.Builder{}
	sw := New(&w)
	err := sw.Display("xyz\nXYZ\n")
	assert.Nil(t, err)
	err = sw.Display("abc\nABC\n")
	assert.Nil(t, err)
	assert.Equal(t, "xyz\nXYZ\n\x1b[0D\x1b[2K\x1b[1A\x1b[2K\x1b[1A\x1b[2Kabc\nABC\n", w.String())
}

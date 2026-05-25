package output

import (
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

// ANSI color sequences (empty when colors disabled).
const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiCyan   = "\033[36m"
	ansiYellow = "\033[33m"
)

// colorPalette bar fill colors by rank (0 = largest share).
var colorPalette = []string{
	"\033[38;5;42m",  // green
	"\033[38;5;78m",
	"\033[38;5;114m",
	"\033[38;5;150m",
	"\033[38;5;186m",
}

const ansiBarEmpty = "\033[38;5;238m"

type styler struct {
	enabled bool
}

func newStyler(w io.Writer) styler {
	return styler{enabled: colorsEnabled(w)}
}

func colorsEnabled(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("FORCE_COLOR") != "" && os.Getenv("FORCE_COLOR") != "0" {
		return true
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd())
}

func (s styler) paint(code, text string) string {
	if !s.enabled || text == "" {
		return text
	}
	return code + text + ansiReset
}

func (s styler) bold(text string) string   { return s.paint(ansiBold, text) }
func (s styler) dim(text string) string    { return s.paint(ansiDim, text) }
func (s styler) cyan(text string) string   { return s.paint(ansiCyan, text) }
func (s styler) yellow(text string) string { return s.paint(ansiYellow, text) }

func (s styler) barColor(rank int) string {
	if !s.enabled {
		return ""
	}
	if rank < len(colorPalette) {
		return colorPalette[rank]
	}
	return ansiDim
}

// stripANSI removes escape sequences (for tests).
func stripANSI(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			if i < len(s) {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

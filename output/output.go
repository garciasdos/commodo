package output

import (
	"fmt"
	"io"
)

// ANSI color codes
const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorReset  = "\033[0m"
)

type Printer struct {
	w io.Writer
}

func New(w io.Writer) *Printer {
	return &Printer{w: w}
}

func (p *Printer) Success(msg string) {
	fmt.Fprintf(p.w, "%s✓ %s%s\n", colorGreen, msg, colorReset)
}

func (p *Printer) Error(msg string) {
	fmt.Fprintf(p.w, "%s✗ %s%s\n", colorRed, msg, colorReset)
}

func (p *Printer) Info(msg string) {
	fmt.Fprintf(p.w, "%sℹ %s%s\n", colorCyan, msg, colorReset)
}

func (p *Printer) Warn(msg string) {
	fmt.Fprintf(p.w, "%s⟳ %s%s\n", colorYellow, msg, colorReset)
}

func (p *Printer) Secondary(msg string) {
	fmt.Fprintf(p.w, "%s  %s%s\n", colorGray, msg, colorReset)
}

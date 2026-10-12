// Package ui provides terminal colors and user prompts.
package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/cli/go-gh/v2/pkg/prompter"
	"github.com/cli/go-gh/v2/pkg/term"
)

// Prompter asks the user for input.
type Prompter interface {
	// Select returns the index of the chosen option.
	Select(prompt, defaultValue string, options []string) (int, error)
	// Input returns the entered text, or defaultValue when the user just presses Enter.
	Input(prompt, defaultValue string) (string, error)
}

// IO bundles the streams and capabilities of the current terminal.
type IO struct {
	In          io.Reader
	Out         io.Writer
	ErrOut      io.Writer
	Interactive bool
	Theme       Theme
}

// System returns IO for the process's standard streams. Arrow-key prompts are
// only used when both stdin and stdout are terminals; colors follow gh's rules
// (NO_COLOR, CLICOLOR, CLICOLOR_FORCE, GH_FORCE_TTY).
func System() *IO {
	t := term.FromEnv()
	return &IO{
		In:          os.Stdin,
		Out:         t.Out(),
		ErrOut:      t.ErrOut(),
		Interactive: t.IsTerminalOutput() && term.IsTerminal(os.Stdin),
		Theme:       Theme{Enabled: t.IsColorEnabled()},
	}
}

// Prompter returns an arrow-key prompter for interactive terminals and a
// line-based one otherwise, so piped input keeps working.
func (s *IO) Prompter() Prompter {
	if s.Interactive {
		if in, ok := s.In.(*os.File); ok {
			if out, ok := s.Out.(*os.File); ok {
				errOut, _ := s.ErrOut.(*os.File)
				if errOut == nil {
					errOut = os.Stderr
				}
				return prompter.New(in, out, errOut)
			}
		}
	}
	return NewLinePrompter(s.In, s.Out)
}

// IsInterrupt reports whether err was caused by the user pressing Ctrl+C in a prompt.
func IsInterrupt(err error) bool {
	return errors.Is(err, terminal.InterruptErr)
}

// LinePrompter reads answers line by line. It is used when stdin is not a terminal.
type LinePrompter struct {
	in  *bufio.Reader
	out io.Writer
}

// NewLinePrompter returns a LinePrompter reading from in and writing prompts to out.
func NewLinePrompter(in io.Reader, out io.Writer) *LinePrompter {
	return &LinePrompter{in: bufio.NewReader(in), out: out}
}

// Select returns the index of defaultValue without asking, since a list cannot
// be navigated without a terminal.
func (p *LinePrompter) Select(_ string, defaultValue string, options []string) (int, error) {
	for i, o := range options {
		if o == defaultValue {
			return i, nil
		}
	}
	return 0, nil
}

// Input prints prompt and reads one line. An empty answer returns defaultValue.
// Reaching end of input without an answer is an error unless there is a default.
func (p *LinePrompter) Input(prompt, defaultValue string) (string, error) {
	fmt.Fprintln(p.out, prompt)
	if defaultValue != "" {
		fmt.Fprintf(p.out, "Suggested: [%s]\npress Enter to use suggested value, or write new one", defaultValue)
	}
	fmt.Fprint(p.out, ": ")

	line, err := p.in.ReadString('\n')
	if err != nil && !(errors.Is(err, io.EOF) && (line != "" || defaultValue != "")) {
		return "", fmt.Errorf("reading input: %w", err)
	}
	if answer := strings.TrimSpace(line); answer != "" {
		return answer, nil
	}
	return defaultValue, nil
}

// Theme applies ANSI colors when enabled.
type Theme struct {
	Enabled bool
}

func (t Theme) paint(code, s string) string {
	if !t.Enabled {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func (t Theme) Bold(s string) string   { return t.paint("1", s) }
func (t Theme) Green(s string) string  { return t.paint("32", s) }
func (t Theme) Yellow(s string) string { return t.paint("33", s) }
func (t Theme) Red(s string) string    { return t.paint("31", s) }
func (t Theme) Cyan(s string) string   { return t.paint("36", s) }
func (t Theme) Gray(s string) string   { return t.paint("90", s) }

// SuccessIcon, WarningIcon and FailureIcon match the symbols gh uses.
func (t Theme) SuccessIcon() string { return t.Green("✓") }
func (t Theme) WarningIcon() string { return t.Yellow("!") }
func (t Theme) FailureIcon() string { return t.Red("X") }

package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestLinePrompterInput(t *testing.T) {
	out := &bytes.Buffer{}
	p := NewLinePrompter(strings.NewReader("\nmy title\n  \n"), out)

	tests := []struct {
		defaultValue string
		want         string
	}{
		{"suggested", "suggested"}, // Enter keeps the suggestion
		{"suggested", "my title"},  // typed value wins
		{"", ""},                   // blank line without default
	}
	for _, tt := range tests {
		got, err := p.Input("Prompt", tt.defaultValue)
		if err != nil {
			t.Fatal(err)
		}
		if got != tt.want {
			t.Errorf("Input(%q) = %q, want %q", tt.defaultValue, got, tt.want)
		}
	}
	if !strings.Contains(out.String(), "Suggested: [suggested]") {
		t.Errorf("suggestion not shown: %q", out.String())
	}
}

func TestLinePrompterEOF(t *testing.T) {
	p := NewLinePrompter(strings.NewReader("last line without newline"), &bytes.Buffer{})
	if got, err := p.Input("Prompt", ""); err != nil || got != "last line without newline" {
		t.Errorf("got %q, %v", got, err)
	}
	if got, err := p.Input("Prompt", "default"); err != nil || got != "default" {
		t.Errorf("EOF with default: got %q, %v", got, err)
	}
	if _, err := p.Input("Prompt", ""); err == nil {
		t.Error("EOF without default should fail")
	}
}

func TestLinePrompterSelectReturnsDefault(t *testing.T) {
	p := NewLinePrompter(strings.NewReader(""), &bytes.Buffer{})
	if i, _ := p.Select("Pick", "b", []string{"a", "b", "c"}); i != 1 {
		t.Errorf("Select = %d, want 1", i)
	}
	if i, _ := p.Select("Pick", "missing", []string{"a", "b"}); i != 0 {
		t.Errorf("Select = %d, want 0", i)
	}
}

func TestThemeDisabledLeavesTextUnchanged(t *testing.T) {
	if got := (Theme{}).Green("ok"); got != "ok" {
		t.Errorf("got %q", got)
	}
	if got := (Theme{Enabled: true}).Green("ok"); got != "\x1b[32mok\x1b[0m" {
		t.Errorf("got %q", got)
	}
}

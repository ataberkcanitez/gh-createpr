package config

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newStore(t *testing.T) (*Store, *bytes.Buffer) {
	t.Helper()
	out := &bytes.Buffer{}
	return &Store{Path: filepath.Join(t.TempDir(), FileName), Out: out}, out
}

func TestLoadCreatesDefaultConfig(t *testing.T) {
	s, out := newStore(t)

	cfg, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Assignee != "@me" || len(cfg.Reviewers) != 0 || cfg.TargetBranch != "" {
		t.Errorf("unexpected default config: %+v", cfg)
	}
	if !strings.Contains(out.String(), "Default config created.") {
		t.Errorf("expected creation message, got %q", out.String())
	}

	data, err := os.ReadFile(s.Path)
	if err != nil {
		t.Fatal(err)
	}
	if want := "reviewers: []\nassignee: '@me'\n"; string(data) != want {
		t.Errorf("default file = %q, want %q", data, want)
	}
}

func TestLoadExistingConfig(t *testing.T) {
	s, out := newStore(t)
	content := "reviewers:\n    - alice\n    - bob\nassignee: carol\ntarget_branch: develop\n"
	if err := os.WriteFile(s.Path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(cfg.Reviewers, ","); got != "alice,bob" {
		t.Errorf("reviewers = %q", got)
	}
	if cfg.Assignee != "carol" || cfg.TargetBranch != "develop" {
		t.Errorf("unexpected config: %+v", cfg)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output, got %q", out.String())
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	s, _ := newStore(t)
	if err := os.WriteFile(s.Path, []byte("reviewers: [unclosed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(); err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestAddAndRemoveReviewer(t *testing.T) {
	s, _ := newStore(t)

	if err := s.AddReviewer("alice"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddReviewer("bob"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddReviewer("alice"); !errors.Is(err, ErrReviewerExists) {
		t.Errorf("duplicate add: got %v, want ErrReviewerExists", err)
	}

	if err := s.RemoveReviewer("alice"); err != nil {
		t.Fatal(err)
	}
	if err := s.RemoveReviewer("alice"); !errors.Is(err, ErrReviewerNotFound) {
		t.Errorf("missing remove: got %v, want ErrReviewerNotFound", err)
	}

	cfg, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(cfg.Reviewers, ","); got != "bob" {
		t.Errorf("reviewers = %q, want bob", got)
	}
}

func TestSettersPreserveOtherFields(t *testing.T) {
	s, _ := newStore(t)
	if err := s.AddReviewer("alice"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetAssignee("bob"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetTargetBranch("develop"); err != nil {
		t.Fatal(err)
	}

	cfg, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Reviewers) != 1 || cfg.Assignee != "bob" || cfg.TargetBranch != "develop" {
		t.Errorf("unexpected config: %+v", cfg)
	}
}

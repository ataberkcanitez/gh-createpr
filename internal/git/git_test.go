package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseLog(t *testing.T) {
	out := "aaaaaaaaaa\x1fAT: Execute go mod tidy\x1f\x1e\n" +
		"bbbbbbbbbb\x1fAT: Real job done\x1fLonger description\n\nwith paragraphs\n\x1e\n"

	commits := parseLog(out)
	if len(commits) != 2 {
		t.Fatalf("got %d commits, want 2", len(commits))
	}
	if commits[0].Subject != "AT: Execute go mod tidy" || commits[0].Body != "" {
		t.Errorf("first commit = %+v", commits[0])
	}
	if commits[1].Subject != "AT: Real job done" || commits[1].Body != "Longer description\n\nwith paragraphs" {
		t.Errorf("second commit = %+v", commits[1])
	}
	if commits[1].ShortHash() != "bbbbbbb" {
		t.Errorf("short hash = %q", commits[1].ShortHash())
	}
}

// fakeRunner answers git commands by their first non-format argument.
type fakeRunner struct {
	refs  map[string]bool
	logs  map[string]string // keyed by the revision range ("" for none)
	calls [][]string
}

func (f *fakeRunner) run(args ...string) (string, error) {
	f.calls = append(f.calls, args)
	switch args[0] {
	case "rev-parse":
		if f.refs[strings.TrimSuffix(args[3], "^{commit}")] {
			return "sha\n", nil
		}
		return "", errors.New("unknown ref")
	case "symbolic-ref":
		return "", errors.New("not set")
	case "log":
		rangeArg := ""
		if len(args) > 3 {
			rangeArg = args[3]
		}
		return f.logs[rangeArg], nil
	}
	return "", errors.New("unexpected command")
}

func record(hash, subject string) string {
	return hash + fieldSep + subject + fieldSep + recordSep
}

func TestBranchCommitsUsesBaseRange(t *testing.T) {
	f := &fakeRunner{
		refs: map[string]bool{"origin/develop": true},
		logs: map[string]string{
			"origin/develop..HEAD": record("h2", "second") + record("h1", "first"),
			"":                     record("h2", "second") + record("h1", "first") + record("h0", "base"),
		},
	}
	c := &Client{Run: f.run}

	commits, err := c.BranchCommits("develop", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 2 || commits[0].Subject != "second" {
		t.Errorf("commits = %+v", commits)
	}
}

// A configured target branch that does not exist in this repository (e.g. a
// global "dev" setting used in a repo that only has "main") must still show
// only the branch's own commits.
func TestBranchCommitsMissingBaseUsesDefaultBranch(t *testing.T) {
	f := &fakeRunner{
		refs: map[string]bool{"origin/main": true},
		logs: map[string]string{
			"origin/main..HEAD": record("h2", "mine"),
			"":                  record("h2", "mine") + record("h1", "unrelated"),
		},
	}
	c := &Client{Run: f.run}

	commits, err := c.BranchCommits("dev", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 || commits[0].Subject != "mine" {
		t.Errorf("commits = %+v", commits)
	}
}

func TestBranchCommitsUnresolvableBaseOffersOnlyLatest(t *testing.T) {
	f := &fakeRunner{
		logs: map[string]string{"": record("h1", "latest") + record("h0", "older")},
	}
	c := &Client{Run: f.run}

	commits, err := c.BranchCommits("missing-branch", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 || commits[0].Subject != "latest" {
		t.Errorf("commits = %+v", commits)
	}
}

func TestBranchCommitsFallsBackWhenNothingAhead(t *testing.T) {
	f := &fakeRunner{
		refs: map[string]bool{"origin/main": true},
		logs: map[string]string{"": record("h1", "latest")},
	}
	c := &Client{Run: f.run}

	commits, err := c.BranchCommits("main", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 || commits[0].Subject != "latest" {
		t.Errorf("commits = %+v", commits)
	}
}

func TestBranchCommitsNoCommits(t *testing.T) {
	c := &Client{Run: (&fakeRunner{}).run}
	if _, err := c.BranchCommits("", 10); err == nil {
		t.Fatal("expected error when repository has no commits")
	}
}

// TestBranchCommitsRealRepository runs against a throwaway git repository.
func TestBranchCommitsRealRepository(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	gitIn := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
			"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	commit := func(msg string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, "f"), []byte(msg), 0o644); err != nil {
			t.Fatal(err)
		}
		gitIn("add", "f")
		gitIn("commit", "-q", "-m", msg)
	}

	gitIn("init", "-q", "-b", "main")
	commit("initial")
	gitIn("checkout", "-q", "-b", "feature")
	commit("AT: Real job done\n\nDetails here.")
	commit("AT: Execute go mod tidy")

	c := &Client{Run: func(args ...string) (string, error) {
		return ExecRunner(append([]string{"-C", dir}, args...)...)
	}}

	commits, err := c.BranchCommits("main", 10)
	if err != nil {
		t.Fatal(err)
	}
	var subjects []string
	for _, c := range commits {
		subjects = append(subjects, c.Subject)
	}
	if got, want := strings.Join(subjects, " | "), "AT: Execute go mod tidy | AT: Real job done"; got != want {
		t.Errorf("subjects = %q, want %q", got, want)
	}
	if commits[1].Body != "Details here." {
		t.Errorf("body = %q", commits[1].Body)
	}
}

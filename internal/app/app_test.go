package app

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gh-createpr/internal/config"
	"gh-createpr/internal/git"
	"gh-createpr/internal/ui"
)

type fakePrompter struct {
	selects []int    // answers for Select, in order
	inputs  []string // answers for Input; "" means "press Enter"
	asked   []string
}

func (p *fakePrompter) Select(prompt, defaultValue string, options []string) (int, error) {
	p.asked = append(p.asked, fmt.Sprintf("select: %s [%s]", strings.Join(options, " / "), defaultValue))
	if len(p.selects) == 0 {
		return 0, errors.New("unexpected select")
	}
	i := p.selects[0]
	p.selects = p.selects[1:]
	return i, nil
}

func (p *fakePrompter) Input(prompt, defaultValue string) (string, error) {
	p.asked = append(p.asked, fmt.Sprintf("input: %s [%s]", prompt, defaultValue))
	if len(p.inputs) == 0 {
		return "", errors.New("unexpected input")
	}
	answer := p.inputs[0]
	p.inputs = p.inputs[1:]
	if answer == "" {
		return defaultValue, nil
	}
	return answer, nil
}

type fakeGit struct {
	commits []git.Commit
	err     error
	base    string
}

func (g *fakeGit) BranchCommits(base string, _ int) ([]git.Commit, error) {
	g.base = base
	return g.commits, g.err
}

type fakeGitHub struct {
	title, body, base string
	reviewers         []string
	assignee          string
	createErr         error
	reviewerErr       error
	created           bool
}

func (g *fakeGitHub) CreatePullRequest(title, body, base string) (string, error) {
	g.title, g.body, g.base = title, body, base
	if g.createErr != nil {
		return "", g.createErr
	}
	g.created = true
	return "https://github.com/o/r/pull/1", nil
}

func (g *fakeGitHub) AddReviewers(_ string, reviewers []string) error {
	g.reviewers = reviewers
	return g.reviewerErr
}

func (g *fakeGitHub) AddAssignee(_ string, assignee string) error {
	g.assignee = assignee
	return nil
}

type harness struct {
	app    *App
	out    *bytes.Buffer
	errOut *bytes.Buffer
	prompt *fakePrompter
	git    *fakeGit
	gh     *fakeGitHub
	store  *config.Store
}

func newHarness(t *testing.T, interactive bool, configYAML string) *harness {
	t.Helper()
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	path := filepath.Join(t.TempDir(), config.FileName)
	if configYAML != "" {
		if err := os.WriteFile(path, []byte(configYAML), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{
		out:    out,
		errOut: errOut,
		prompt: &fakePrompter{},
		git: &fakeGit{commits: []git.Commit{
			{Hash: "2222222222", Subject: "AT: Execute go mod tidy"},
			{Hash: "1111111111", Subject: "AT: Real job done"},
		}},
		gh:    &fakeGitHub{},
		store: &config.Store{Path: path, Out: out},
	}
	h.app = &App{
		IO:       &ui.IO{Out: out, ErrOut: errOut, Interactive: interactive},
		Prompter: h.prompt,
		Config:   h.store,
		Git:      h.git,
		GitHub:   h.gh,
	}
	return h
}

func (h *harness) run(args ...string) int {
	return h.app.Run("gh-createpr", args)
}

func TestNonInteractiveSuggestsLatestCommit(t *testing.T) {
	h := newHarness(t, false, "reviewers: [alice]\nassignee: '@me'\n")
	h.prompt.inputs = []string{"", ""}

	if code := h.run(); code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, h.errOut)
	}
	if h.gh.title != "AT: Execute go mod tidy" || h.gh.body != "" || h.gh.base != "" {
		t.Errorf("created PR with %+v", h.gh)
	}
	if !slices.Equal(h.gh.reviewers, []string{"alice"}) || h.gh.assignee != "@me" {
		t.Errorf("reviewers=%v assignee=%q", h.gh.reviewers, h.gh.assignee)
	}
	if !strings.Contains(h.out.String(), "Pull Request created: https://github.com/o/r/pull/1") {
		t.Errorf("stdout = %q", h.out)
	}
	for _, q := range h.prompt.asked {
		if strings.HasPrefix(q, "select") {
			t.Errorf("non-interactive run must not show a picker: %v", h.prompt.asked)
		}
	}
}

func TestInteractivePickOlderCommit(t *testing.T) {
	h := newHarness(t, true, "reviewers: []\nassignee: '@me'\ntarget_branch: develop\n")
	h.prompt.selects = []int{2}
	h.prompt.inputs = []string{"", "Some body"}

	if code := h.run(); code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, h.errOut)
	}
	if h.gh.title != "AT: Real job done" || h.gh.body != "Some body" || h.gh.base != "develop" {
		t.Errorf("created PR with %+v", h.gh)
	}
	if h.git.base != "develop" {
		t.Errorf("commits compared against %q, want develop", h.git.base)
	}
	if want := "select: " + customTitleOption + " / 2222222  AT: Execute go mod tidy / 1111111  AT: Real job done [2222222  AT: Execute go mod tidy]"; h.prompt.asked[0] != want {
		t.Errorf("picker = %q, want %q", h.prompt.asked[0], want)
	}
	if h.gh.reviewers != nil {
		t.Errorf("no reviewers configured, but AddReviewers was called with %v", h.gh.reviewers)
	}
}

func TestInteractiveCustomTitle(t *testing.T) {
	h := newHarness(t, true, "reviewers: []\nassignee: ''\n")
	h.prompt.selects = []int{0}
	h.prompt.inputs = []string{"Hand written", ""}

	if code := h.run(); code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, h.errOut)
	}
	if h.gh.title != "Hand written" {
		t.Errorf("title = %q", h.gh.title)
	}
	if h.prompt.asked[1] != "input: Enter Pull Request Title []" {
		t.Errorf("custom title should have no default: %q", h.prompt.asked[1])
	}
	if h.gh.assignee != "" {
		t.Errorf("empty assignee should be skipped, got %q", h.gh.assignee)
	}
}

func TestInteractiveSingleCommitSkipsPicker(t *testing.T) {
	h := newHarness(t, true, "reviewers: []\nassignee: '@me'\n")
	h.git.commits = h.git.commits[:1]
	h.prompt.inputs = []string{"", ""}

	if code := h.run(); code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, h.errOut)
	}
	if len(h.prompt.asked) != 2 || h.gh.title != "AT: Execute go mod tidy" {
		t.Errorf("asked=%v title=%q", h.prompt.asked, h.gh.title)
	}
}

func TestTitleAndBodyFlagsSkipPrompts(t *testing.T) {
	h := newHarness(t, true, "reviewers: []\nassignee: '@me'\n")

	if code := h.run("--title", "From flag", "--body", ""); code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, h.errOut)
	}
	if h.gh.title != "From flag" || h.gh.body != "" || len(h.prompt.asked) != 0 {
		t.Errorf("title=%q body=%q asked=%v", h.gh.title, h.gh.body, h.prompt.asked)
	}
}

func TestAssigneeFlagUpdatesConfigAndCreatesPR(t *testing.T) {
	h := newHarness(t, false, "reviewers: []\nassignee: '@me'\n")
	h.prompt.inputs = []string{"", ""}

	if code := h.run("--assignee", "bob"); code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, h.errOut)
	}
	cfg, _ := h.store.Load()
	if cfg.Assignee != "bob" || h.gh.assignee != "bob" || !h.gh.created {
		t.Errorf("config assignee=%q pr assignee=%q created=%v", cfg.Assignee, h.gh.assignee, h.gh.created)
	}
}

func TestCreateFailureExitsNonZero(t *testing.T) {
	h := newHarness(t, false, "reviewers: [alice]\nassignee: '@me'\n")
	h.prompt.inputs = []string{"", ""}
	h.gh.createErr = errors.New("a pull request already exists")

	if code := h.run(); code != 1 {
		t.Errorf("exit code %d, want 1", code)
	}
	if !strings.Contains(h.errOut.String(), "a pull request already exists") {
		t.Errorf("stderr = %q", h.errOut)
	}
	if h.gh.reviewers != nil {
		t.Error("reviewers must not be added when creation fails")
	}
}

func TestReviewerFailureIsAWarning(t *testing.T) {
	h := newHarness(t, false, "reviewers: [ghost]\nassignee: '@me'\n")
	h.prompt.inputs = []string{"", ""}
	h.gh.reviewerErr = errors.New("could not resolve user")

	if code := h.run(); code != 0 {
		t.Errorf("exit code %d, want 0", code)
	}
	if !strings.Contains(h.errOut.String(), "Could not add reviewers") || h.gh.assignee != "@me" {
		t.Errorf("stderr=%q assignee=%q", h.errOut, h.gh.assignee)
	}
}

func TestGitFailureExitsNonZero(t *testing.T) {
	h := newHarness(t, false, "reviewers: []\nassignee: '@me'\n")
	h.git.err = errors.New("not a git repository")

	if code := h.run(); code != 1 || h.gh.created {
		t.Errorf("exit code %d, created=%v", code, h.gh.created)
	}
}

func TestConfigCommands(t *testing.T) {
	h := newHarness(t, false, "")

	steps := []struct {
		args     []string
		wantCode int
		wantOut  string
	}{
		{[]string{"--add-reviewer", "alice"}, 0, "Reviewer alice added."},
		{[]string{"--add-reviewer", "alice"}, 0, "Reviewer alice already exists."},
		{[]string{"--list-reviewers"}, 0, "- alice"},
		{[]string{"--set-target-branch", "develop"}, 0, "Target branch set to: develop"},
		{[]string{"--list"}, 0, "target_branch: develop"},
		{[]string{"--remove-reviewer", "alice"}, 0, "Reviewer alice removed."},
		{[]string{"--remove-reviewer", "alice"}, 1, "Reviewer alice not found."},
		{[]string{"--list-reviewers"}, 0, "[]"},
	}
	for _, s := range steps {
		h.out.Reset()
		h.errOut.Reset()
		code := h.run(s.args...)
		output := h.out.String() + h.errOut.String()
		if code != s.wantCode || !strings.Contains(output, s.wantOut) {
			t.Errorf("%v: code=%d output=%q, want code=%d containing %q", s.args, code, output, s.wantCode, s.wantOut)
		}
	}
	if h.gh.created {
		t.Error("config commands must not create a pull request")
	}
}

func TestFlagErrors(t *testing.T) {
	h := newHarness(t, false, "")
	if code := h.run("--help"); code != 0 {
		t.Errorf("--help exit code %d, want 0", code)
	}
	if code := h.run("--unknown"); code != 2 {
		t.Errorf("unknown flag exit code %d, want 2", code)
	}
	if h.gh.created {
		t.Error("flag errors must not create a pull request")
	}
}

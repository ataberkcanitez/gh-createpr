package github

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"
)

type recorder struct {
	calls  [][]string
	stdout string
	stderr string
	err    error
}

func (r *recorder) exec(args ...string) (stdout, stderr bytes.Buffer, err error) {
	r.calls = append(r.calls, args)
	stdout.WriteString(r.stdout)
	stderr.WriteString(r.stderr)
	return stdout, stderr, r.err
}

func TestCreatePullRequest(t *testing.T) {
	tests := []struct {
		name string
		base string
		want []string
	}{
		{"default base", "", []string{"pr", "create", "--title", "T", "--body", ""}},
		{"custom base", "develop", []string{"pr", "create", "--title", "T", "--body", "", "--base", "develop"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &recorder{stdout: "https://github.com/o/r/pull/1\n"}
			url, err := (&Client{Exec: r.exec}).CreatePullRequest("T", "", tt.base)
			if err != nil {
				t.Fatal(err)
			}
			if url != "https://github.com/o/r/pull/1" {
				t.Errorf("url = %q", url)
			}
			if !slices.Equal(r.calls[0], tt.want) {
				t.Errorf("args = %q, want %q", r.calls[0], tt.want)
			}
		})
	}
}

func TestCreatePullRequestIncludesStderr(t *testing.T) {
	r := &recorder{stderr: "a pull request already exists\n", err: errors.New("exit status 1")}
	_, err := (&Client{Exec: r.exec}).CreatePullRequest("T", "", "")
	if err == nil || !strings.Contains(err.Error(), "a pull request already exists") {
		t.Errorf("err = %v", err)
	}
}

func TestAddReviewersAndAssignee(t *testing.T) {
	r := &recorder{}
	c := &Client{Exec: r.exec}

	if err := c.AddReviewers("URL", []string{"alice", "org/team"}); err != nil {
		t.Fatal(err)
	}
	if err := c.AddAssignee("URL", "@me"); err != nil {
		t.Fatal(err)
	}

	want := [][]string{
		{"pr", "edit", "URL", "--add-reviewer", "alice", "--add-reviewer", "org/team"},
		{"pr", "edit", "URL", "--add-assignee", "@me"},
	}
	if !slices.EqualFunc(r.calls, want, slices.Equal) {
		t.Errorf("calls = %q, want %q", r.calls, want)
	}
}

// Package git reads commit information from the local repository.
package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Runner executes git with args and returns its stdout.
type Runner func(args ...string) (string, error)

// ExecRunner runs the git binary found on PATH.
func ExecRunner(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}

// Commit is a single commit on the current branch.
type Commit struct {
	Hash    string
	Subject string
	Body    string
}

// ShortHash returns the abbreviated commit hash.
func (c Commit) ShortHash() string {
	if len(c.Hash) > 7 {
		return c.Hash[:7]
	}
	return c.Hash
}

// Client reads commits through a Runner.
type Client struct {
	Run Runner
}

// NewClient returns a Client backed by the git binary.
func NewClient() *Client {
	return &Client{Run: ExecRunner}
}

// Field and record separators that cannot appear in commit messages.
const (
	fieldSep  = "\x1f"
	recordSep = "\x1e"
	logFormat = "--pretty=format:%H" + fieldSep + "%s" + fieldSep + "%b" + recordSep
)

// LatestCommit returns the commit at HEAD.
func (c *Client) LatestCommit() (Commit, error) {
	commits, err := c.log("-1")
	if err != nil {
		return Commit{}, err
	}
	if len(commits) == 0 {
		return Commit{}, fmt.Errorf("no commits found")
	}
	return commits[0], nil
}

// BranchCommits returns up to limit commits, newest first, that are on HEAD but
// not on base. The first element is always HEAD. When base cannot be resolved
// locally, or HEAD has no commits ahead of it, the latest limit commits are
// returned instead so the caller always has a suggestion.
func (c *Client) BranchCommits(base string, limit int) ([]Commit, error) {
	n := fmt.Sprintf("-%d", limit)
	if ref := c.resolveBase(base); ref != "" {
		if commits, err := c.log(n, ref+"..HEAD"); err == nil && len(commits) > 0 {
			return commits, nil
		}
	}
	commits, err := c.log(n)
	if err != nil {
		return nil, err
	}
	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits found")
	}
	return commits, nil
}

// resolveBase finds a local ref for the base branch, preferring the remote
// tracking branch. It returns "" when nothing matches.
func (c *Client) resolveBase(base string) string {
	var candidates []string
	if base != "" {
		candidates = []string{"origin/" + base, base}
	} else {
		if head, err := c.Run("symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"); err == nil {
			candidates = append(candidates, strings.TrimSpace(head))
		}
		candidates = append(candidates, "origin/main", "origin/master", "main", "master")
	}
	for _, ref := range candidates {
		if ref == "" {
			continue
		}
		if _, err := c.Run("rev-parse", "--verify", "--quiet", ref+"^{commit}"); err == nil {
			return ref
		}
	}
	return ""
}

func (c *Client) log(args ...string) ([]Commit, error) {
	out, err := c.Run(append([]string{"log", logFormat}, args...)...)
	if err != nil {
		return nil, err
	}
	return parseLog(out), nil
}

func parseLog(out string) []Commit {
	var commits []Commit
	for record := range strings.SplitSeq(out, recordSep) {
		record = strings.TrimLeft(record, "\n")
		if record == "" {
			continue
		}
		fields := strings.SplitN(record, fieldSep, 3)
		if len(fields) < 2 {
			continue
		}
		commit := Commit{Hash: fields[0], Subject: strings.TrimSpace(fields[1])}
		if len(fields) == 3 {
			commit.Body = strings.TrimSpace(fields[2])
		}
		commits = append(commits, commit)
	}
	return commits
}

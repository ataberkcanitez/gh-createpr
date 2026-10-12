// Package github creates and edits pull requests through the gh CLI.
package github

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2"
)

// Executor runs a gh command. Its signature matches gh.Exec.
type Executor func(args ...string) (stdout, stderr bytes.Buffer, err error)

// Client wraps gh pull request commands.
type Client struct {
	Exec Executor
}

// NewClient returns a Client backed by the gh CLI.
func NewClient() *Client {
	return &Client{Exec: gh.Exec}
}

// CreatePullRequest opens a pull request for the current branch and returns its URL.
// An empty base uses the repository's default branch.
func (c *Client) CreatePullRequest(title, body, base string) (string, error) {
	// --body is always passed (even when empty) so gh never opens an editor.
	args := []string{"pr", "create", "--title", title, "--body", body}
	if base != "" {
		args = append(args, "--base", base)
	}
	stdout, err := c.run(args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout), nil
}

// AddReviewers requests reviews from reviewers on the pull request at url.
func (c *Client) AddReviewers(url string, reviewers []string) error {
	args := []string{"pr", "edit", url}
	for _, r := range reviewers {
		args = append(args, "--add-reviewer", r)
	}
	_, err := c.run(args...)
	return err
}

// AddAssignee assigns assignee to the pull request at url.
func (c *Client) AddAssignee(url, assignee string) error {
	_, err := c.run("pr", "edit", url, "--add-assignee", assignee)
	return err
}

func (c *Client) run(args ...string) (string, error) {
	stdout, stderr, err := c.Exec(args...)
	if err != nil {
		if detail := strings.TrimSpace(stderr.String()); detail != "" {
			return "", fmt.Errorf("%w\n%s", err, detail)
		}
		return "", err
	}
	return stdout.String(), nil
}

// Package app implements the gh-createpr command.
package app

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"gh-createpr/internal/config"
	"gh-createpr/internal/git"
	"gh-createpr/internal/ui"
)

// maxCommitChoices limits how many commits are offered as title suggestions.
const maxCommitChoices = 15

// customTitleOption is the last entry of the commit picker.
const customTitleOption = "Write my own title"

// CommitSource lists commits that can be used as title suggestions.
type CommitSource interface {
	BranchCommits(base string, limit int) ([]git.Commit, error)
}

// PullRequests creates and edits pull requests.
type PullRequests interface {
	CreatePullRequest(title, body, base string) (string, error)
	AddReviewers(url string, reviewers []string) error
	AddAssignee(url, assignee string) error
}

// App holds the command's dependencies.
type App struct {
	IO       *ui.IO
	Prompter ui.Prompter
	Config   *config.Store
	Git      CommitSource
	GitHub   PullRequests
}

type options struct {
	list           bool
	listReviewers  bool
	addReviewer    string
	removeReviewer string
	assignee       string
	targetBranch   string
	title          string
	body           string
	titleSet       bool
	bodySet        bool
}

// Run executes the command and returns the process exit code.
func (a *App) Run(name string, args []string) int {
	opts, code, ok := a.parseFlags(name, args)
	if !ok {
		return code
	}

	switch {
	case opts.list:
		return a.listConfig()
	case opts.listReviewers:
		return a.listReviewers()
	case opts.addReviewer != "":
		return a.addReviewer(opts.addReviewer)
	case opts.removeReviewer != "":
		return a.removeReviewer(opts.removeReviewer)
	case opts.targetBranch != "":
		return a.setTargetBranch(opts.targetBranch)
	}

	if opts.assignee != "" {
		if err := a.Config.SetAssignee(opts.assignee); err != nil {
			a.errorf("Error updating config file: %v", err)
			return 1
		}
	}
	return a.createPullRequest(opts)
}

func (a *App) parseFlags(name string, args []string) (opts options, code int, ok bool) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(a.IO.ErrOut)
	fs.BoolVar(&opts.list, "list", false, "List configs")
	fs.BoolVar(&opts.listReviewers, "list-reviewers", false, "List reviewers")
	fs.StringVar(&opts.addReviewer, "add-reviewer", "", "Add reviewer")
	fs.StringVar(&opts.removeReviewer, "remove-reviewer", "", "Remove reviewer")
	fs.StringVar(&opts.assignee, "assignee", "", "Assignee")
	fs.StringVar(&opts.targetBranch, "set-target-branch", "", "Set target branch for PRs")
	fs.StringVar(&opts.title, "title", "", "Title of the pull request (skips the title prompt)")
	fs.StringVar(&opts.body, "body", "", "Body of the pull request (skips the body prompt)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return opts, 0, false
		}
		return opts, 2, false
	}
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "title":
			opts.titleSet = true
		case "body":
			opts.bodySet = true
		}
	})
	return opts, 0, true
}

func (a *App) listConfig() int {
	data, err := a.Config.Raw()
	if err != nil {
		a.errorf("Error reading config file: %v", err)
		return 1
	}
	fmt.Fprintln(a.IO.Out, string(data))
	return 0
}

func (a *App) listReviewers() int {
	cfg, err := a.Config.Load()
	if err != nil {
		a.errorf("Error reading config file: %v", err)
		return 1
	}
	data, err := yaml.Marshal(cfg.Reviewers)
	if err != nil {
		a.errorf("Error marshalling reviewers: %v", err)
		return 1
	}
	fmt.Fprintln(a.IO.Out, string(data))
	return 0
}

func (a *App) addReviewer(reviewer string) int {
	switch err := a.Config.AddReviewer(reviewer); {
	case errors.Is(err, config.ErrReviewerExists):
		fmt.Fprintf(a.IO.Out, "Reviewer %s already exists.\n", reviewer)
		return 0
	case err != nil:
		a.errorf("Error updating config file: %v", err)
		return 1
	}
	a.successf("Reviewer %s added.", reviewer)
	return 0
}

func (a *App) removeReviewer(reviewer string) int {
	switch err := a.Config.RemoveReviewer(reviewer); {
	case errors.Is(err, config.ErrReviewerNotFound):
		a.errorf("Reviewer %s not found.", reviewer)
		return 1
	case err != nil:
		a.errorf("Error updating config file: %v", err)
		return 1
	}
	a.successf("Reviewer %s removed.", reviewer)
	return 0
}

func (a *App) setTargetBranch(branch string) int {
	if err := a.Config.SetTargetBranch(branch); err != nil {
		a.errorf("Error updating config file: %v", err)
		return 1
	}
	a.successf("Target branch set to: %s", branch)
	return 0
}

func (a *App) createPullRequest(opts options) int {
	cfg, err := a.Config.Load()
	if err != nil {
		a.errorf("Error reading config file: %v", err)
		return 1
	}

	title := opts.title
	if !opts.titleSet {
		if title, err = a.askTitle(cfg.TargetBranch); err != nil {
			return a.promptFailed(err)
		}
	}

	body := opts.body
	if !opts.bodySet {
		if body, err = a.Prompter.Input("Enter Pull Request Body Or Press Enter Leave it empty", ""); err != nil {
			return a.promptFailed(err)
		}
	}

	t := a.IO.Theme
	fmt.Fprintf(a.IO.Out, "\nCreating PR with title: %s", t.Bold(title))
	if cfg.TargetBranch != "" {
		fmt.Fprintf(a.IO.Out, " into %s", t.Cyan(cfg.TargetBranch))
	}
	fmt.Fprintln(a.IO.Out)

	url, err := a.GitHub.CreatePullRequest(title, body, cfg.TargetBranch)
	if err != nil {
		a.errorf("Error creating pull request: %v", err)
		return 1
	}
	fmt.Fprintf(a.IO.Out, "%s Pull Request created: %s\n", t.SuccessIcon(), t.Cyan(url))

	// Reviewers and assignee are best effort: the PR already exists, so
	// failures are reported but do not change the exit code.
	if len(cfg.Reviewers) > 0 {
		if err := a.GitHub.AddReviewers(url, cfg.Reviewers); err != nil {
			a.warnf("Could not add reviewers: %v", err)
		} else {
			a.successf("Reviewers added: %s", strings.Join(cfg.Reviewers, ", "))
		}
	}
	if cfg.Assignee != "" {
		if err := a.GitHub.AddAssignee(url, cfg.Assignee); err != nil {
			a.warnf("Could not add assignee %s: %v", cfg.Assignee, err)
		} else {
			a.successf("Assignee added: %s", cfg.Assignee)
		}
	}
	return 0
}

// askTitle suggests a title from the branch's commits. With several commits on
// an interactive terminal the user picks one from a list; the chosen subject is
// then offered as an editable default.
func (a *App) askTitle(base string) (string, error) {
	commits, err := a.Git.BranchCommits(base, maxCommitChoices)
	if err != nil {
		return "", fmt.Errorf("getting last commit message: %w", err)
	}

	suggestion := commits[0].Subject
	if a.IO.Interactive && len(commits) > 1 {
		options := make([]string, 0, len(commits)+1)
		for _, c := range commits {
			options = append(options, c.ShortHash()+"  "+c.Subject)
		}
		options = append(options, customTitleOption)

		i, err := a.Prompter.Select("Which commit should the title come from?", options[0], options)
		if err != nil {
			return "", err
		}
		if i < len(commits) {
			suggestion = commits[i].Subject
		} else {
			suggestion = ""
		}
	}

	for {
		title, err := a.Prompter.Input("Enter Pull Request Title", suggestion)
		if err != nil {
			return "", err
		}
		if title = strings.TrimSpace(title); title != "" || !a.IO.Interactive {
			return title, nil
		}
		a.warnf("Title cannot be empty.")
	}
}

func (a *App) promptFailed(err error) int {
	if ui.IsInterrupt(err) {
		fmt.Fprintln(a.IO.ErrOut, "Cancelled.")
		return 1
	}
	a.errorf("Error: %v", err)
	return 1
}

func (a *App) successf(format string, args ...any) {
	fmt.Fprintf(a.IO.Out, "%s %s\n", a.IO.Theme.SuccessIcon(), fmt.Sprintf(format, args...))
}

func (a *App) warnf(format string, args ...any) {
	fmt.Fprintf(a.IO.ErrOut, "%s %s\n", a.IO.Theme.WarningIcon(), fmt.Sprintf(format, args...))
}

func (a *App) errorf(format string, args ...any) {
	fmt.Fprintf(a.IO.ErrOut, "%s %s\n", a.IO.Theme.FailureIcon(), fmt.Sprintf(format, args...))
}

package main

import (
	"fmt"
	"github.com/cli/go-gh/v2"
	"os"
	"strings"
)

func createPullRequest(title, body string) string {
	url, stdErr, err := gh.Exec("pr", "create", "--title", title, "--body", body)
	if err != nil {
		fmt.Println("Error:", err)
		if stdErr.Len() > 0 {
			fmt.Println("Detail:", stdErr.String())
		}
		os.Exit(1)
	}
	return strings.TrimSpace(url.String())
}

func addReviewerToPullRequest(prUrl string) {
	reviewers, err := readReviewersFromConfig()
	if err != nil {
		fmt.Println("can't add reviewers")
		fmt.Println("Error reading config file: ", err)
		os.Exit(1)
	}

	args := []string{"pr", "edit", prUrl}
	for _, reviewer := range reviewers {
		args = append(args, "--add-reviewer", reviewer)
	}

	_, stdErr, err := gh.Exec(args...)
	if err != nil {
		fmt.Println("Error:", err)
		if stdErr.Len() > 0 {
			fmt.Println(stdErr.String())
		}
		return
	}
}
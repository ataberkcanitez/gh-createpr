package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	title, body, assignee := handleOptions()

	prUrl := createPullRequest(title, body)
	fmt.Println("Pull Request created:", prUrl)
	fmt.Println("Adding reviewers...")
	addReviewerToPullRequest(prUrl)
	fmt.Println("Reviewers added.")
	addAssignee(prUrl, assignee)

}

func getUserInput(prompt string) string {
	fmt.Println(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input: ", err)
		os.Exit(1)
	}
	return strings.TrimSpace(input)
}

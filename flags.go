package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

var (
	titleFlag          string
	bodyFlag           string
	listConfigsFlag    bool
	listReviewersFlag  bool
	addReviewerFlag    string
	removeReviewerFlag string
	updateAssignee     string
	setTargetBranch    string
)

func handleOptions() (string, string, string, string) {
	flag.BoolVar(&listConfigsFlag, "list", false, "List configs")
	flag.BoolVar(&listReviewersFlag, "list-reviewers", false, "List reviewers")
	flag.StringVar(&addReviewerFlag, "add-reviewer", "", "Add reviewer")
	flag.StringVar(&removeReviewerFlag, "remove-reviewer", "", "Remove reviewer")
	flag.StringVar(&updateAssignee, "assignee", "", "Assignee")
	flag.StringVar(&setTargetBranch, "set-target-branch", "", "Set target branch for PRs")
	flag.Parse()

	if listConfigsFlag {
		listConfigs()
		os.Exit(0)
	}

	if listReviewersFlag {
		listReviewers()
		os.Exit(0)
	}

	if addReviewerFlag != "" {
		addReviewer(addReviewerFlag)
		os.Exit(0)
	}

	if removeReviewerFlag != "" {
		removeReviewer(removeReviewerFlag)
		os.Exit(0)
	}

	if setTargetBranch != "" {
		setTargetBranchConfig(setTargetBranch)
		os.Exit(0)
	}

	if updateAssignee != "" {
		updateAssigneeConfig(updateAssignee)
	}
	updateAssignee = getAssigneeFromConfig()

	targetBranch := getTargetBranchFromConfig()

	message := getLastCommitMessage()
	titleFlag = getUserInputWithSuggestion("Enter Pull Request Title: ", message)
	bodyFlag = getUserInput("Enter Pull Request Body Or Press Enter Leave it empty: ")

	fmt.Printf("creating PR with title: [%s] ", titleFlag)
	if bodyFlag != "" {
		fmt.Println(" and body: ", bodyFlag)
	}
	fmt.Println()
	return titleFlag, bodyFlag, updateAssignee, targetBranch
}

func updateAssigneeConfig(assignee string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		os.Exit(1)
	}
	config.Assignee = assignee
	err = updateConfig(config)
	if err != nil {
		fmt.Println("Error updating config file: ", err)
		os.Exit(1)
	}
}

func getAssigneeFromConfig() string {
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		os.Exit(1)
	}
	return config.Assignee
}

func removeReviewer(reviewerToRemove string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		os.Exit(1)
	}

	reviewers := config.Reviewers

	if !slices.Contains(reviewers, reviewerToRemove) {
		fmt.Printf("Reviewer %s not found.\n", reviewerToRemove)
		os.Exit(1)
	}

	var updatedReviewers []string
	for _, reviewer := range reviewers {
		if reviewer != reviewerToRemove {
			updatedReviewers = append(updatedReviewers, reviewer)
		}
	}

	config.Reviewers = updatedReviewers
	err = updateConfig(config)
	if err != nil {
		fmt.Println("Error updating config file: ", err)
		os.Exit(1)
	}
	fmt.Printf("Reviewer %s removed.\n", reviewerToRemove)
}

func listReviewers() {
	reviewers, err := readReviewersFromConfig()
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		os.Exit(1)
	}

	ymlData, err := yaml.Marshal(reviewers)
	if err != nil {
		fmt.Println("Error marshalling reviewers: ", err)
		os.Exit(1)
	}
	fmt.Println(string(ymlData))
}

func listConfigs() {
	file, err := loadFile()
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		os.Exit(1)
	}
	fmt.Println(string(file))
}

func addReviewer(newReviewer string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		os.Exit(1)
	}
	reviewers := config.Reviewers
	if slices.Contains(reviewers, newReviewer) {
		fmt.Printf("Reviewer %s already exists.\n", newReviewer)
		os.Exit(0)
	}

	reviewers = append(reviewers, newReviewer)

	config.Reviewers = reviewers
	err = updateConfig(config)
	if err != nil {
		fmt.Println("Error updating config file: ", err)
		os.Exit(1)
	}

	fmt.Printf("Reviewer %s added.\n", newReviewer)
}

func setTargetBranchConfig(targetBranch string) {
	err := updateTargetBranchConfig(targetBranch)
	if err != nil {
		fmt.Println("Error updating config file: ", err)
		os.Exit(1)
	}
	fmt.Printf("Target branch set to: %s\n", targetBranch)
}

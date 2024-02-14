package main

import (
	"flag"
	"fmt"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
	"os"
)

var (
	titleFlag          string
	bodyFlag           string
	listConfigsFlag    bool
	listReviewersFlag  bool
	addReviewerFlag    string
	removeReviewerFlag string
)

func handleOptions() (string, string) {
	titleFlag := flag.String("title", "", "Pull Request title")
	bodyFlag := flag.String("body", "", "Pull Request body")
	listConfigsFlag := flag.Bool("list", false, "List configs")
	listReviewersFlag := flag.Bool("list-reviewers", false, "List configs")
	addReviewerFlag := flag.String("add-reviewer", "", "Add reviewer")
	removeReviewerFlag := flag.String("remove-reviewer", "", "Remove reviewer")
	flag.Parse()

	if *listConfigsFlag {
		listConfigs()
		os.Exit(0)
	}

	if *listReviewersFlag {
		listReviewers()
		os.Exit(0)
	}

	if *addReviewerFlag != "" {
		addReviewer(*addReviewerFlag)
		os.Exit(0)
	}

	if *removeReviewerFlag != "" {
		removeReviewer(*removeReviewerFlag)
		os.Exit(0)
	}

	if *titleFlag == "" {
		*titleFlag = getUserInput("Enter Pull Request Title: ")
	}

	if *bodyFlag == "" {
		*bodyFlag = getUserInput("Enter Pull Request Body: ")
	}

	return *titleFlag, *bodyFlag
}

func removeReviewer(reviewerToRemove string) {
	reviewers, err := readReviewersFromConfig()
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		os.Exit(1)
	}

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

	err = updateConfig(updatedReviewers)
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
	reviewers, err := readReviewersFromConfig()
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		os.Exit(1)
	}

	if slices.Contains(reviewers, newReviewer) {
		fmt.Printf("Reviewer %s already exists.\n", newReviewer)
		os.Exit(0)
	}

	reviewers = append(reviewers, newReviewer)

	err = updateConfig(reviewers)
	if err != nil {
		fmt.Println("Error updating config file: ", err)
		os.Exit(1)
	}

	fmt.Printf("Reviewer %s added.\n", newReviewer)
}
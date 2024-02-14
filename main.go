package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/cli/go-gh/v2"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type Config struct {
	Reviewers []string `yaml:"reviewers"`
}

const configFileName = ".config/gh-createpr/gh-createpr-configuration.yml"

func main() {
	titleFlag := flag.String("title", "", "Pull Request title")
	bodyFlag := flag.String("body", "", "Pull Request body")
	listReviewersFlag := flag.Bool("list", false, "List configs")
	addReviewerFlag := flag.String("add-reviewer", "", "Add reviewer")
	removeReviewerFlag := flag.String("remove-reviewer", "", "Remove reviewer")
	flag.Parse()

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

	prUrl := createPullRequest(*titleFlag, *bodyFlag)
	fmt.Println("Pull Request created:", prUrl)
	fmt.Println("Adding reviewers...")
	addReviewerToPullRequest(prUrl)
	fmt.Println("Reviewers added.")
}

func addReviewerToPullRequest(prUrl string) {
	reviewers, err := readReviewersFromConfig()
	if err != nil {
		fmt.Println("can't add reviewers")
		fmt.Println("Error reading config file: ", err)
		os.Exit(1)
	}
	for _, reviewer := range reviewers {
		gh.Exec("pr", "edit", prUrl, "--add-reviewer", reviewer)
	}
}

func createPullRequest(title, body string) string {
	url, _, err := gh.Exec("pr", "create", "--title", title, "--body", body)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	return strings.TrimSpace(url.String())
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

func removeReviewer(reviewerToRemove string) {
	reviewers, err := readReviewersFromConfig()
	if err != nil {
		fmt.Println("Error reading config file: ", err)
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

func addReviewer(newReviewer string) {
	reviewers, err := readReviewersFromConfig()
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		os.Exit(1)
	}
	reviewers = append(reviewers, newReviewer)

	err = updateConfig(reviewers)
	if err != nil {
		fmt.Println("Error updating config file: ", err)
		os.Exit(1)
	}

	fmt.Printf("Reviewer %s added.\n", newReviewer)
}

func updateConfig(reviewers []string) error {
	config := Config{
		Reviewers: reviewers,
	}
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	err = os.WriteFile(configFileName, yamlData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func listReviewers() {
	reviewers, err := readReviewersFromConfig()
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		os.Exit(1)
	}

	fmt.Println(reviewers)
}

func readReviewersFromConfig() ([]string, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	return config.Reviewers, nil
}

func loadConfig() (*Config, error) {
	content, err := os.ReadFile(configFileName)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Config file not found. Creating a new one.")
			err := createDefaultConfig()
			if err != nil {
				return nil, err
			}
			return loadConfig()
		}
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func createDefaultConfig() error {
	config := Config{
		Reviewers: []string{},
	}

	yamlData, err := yaml.Marshal(config)
	if err != nil {
		fmt.Println("Error creating default config: ", err)
		return err
	}

	err = os.WriteFile(configFileName, yamlData, 0644)
	if err != nil {
		return err
	}
	fmt.Println("Default config created.")
	return nil
}

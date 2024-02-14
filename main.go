package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/cli/go-gh/v2"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type Config struct {
	Reviewers []string `yaml:"reviewers"`
}

const configFileName = "gh-createpr-configuration.yml"

func main() {
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

	prUrl := createPullRequest(*titleFlag, *bodyFlag)
	fmt.Println("Pull Request created:", prUrl)
	fmt.Println("Adding reviewers...")
	addReviewerToPullRequest(prUrl)
	fmt.Println("Reviewers added.")
}

func listConfigs() {
	file, err := loadFile()
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		os.Exit(1)
	}
	fmt.Println(string(file))
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

	ymlData, err := yaml.Marshal(reviewers)
	if err != nil {
		fmt.Println("Error marshalling reviewers: ", err)
		os.Exit(1)
	}
	fmt.Println(string(ymlData))
}

func readReviewersFromConfig() ([]string, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	return config.Reviewers, nil
}

func loadFile() ([]byte, error) {
	file, err := os.ReadFile(configFileName)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Config file not found. Creating a new one.")
			err := createDefaultConfig()
			if err != nil {
				return nil, err
			}
			return loadFile()
		}
		return nil, err
	}
	return file, err

}

func loadConfig() (*Config, error) {
	file, err := loadFile()
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configFileName = "gh-createpr-configuration.yml"

type Config struct {
	Reviewers    []string `yaml:"reviewers"`
	Assignee     string   `yaml:"assignee"`
	TargetBranch string   `yaml:"target_branch,omitempty"`
}

func getConfigFileName() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(homeDir, configFileName)
}

func updateConfig(config *Config) error {
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	err = os.WriteFile(getConfigFileName(), yamlData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func loadFile() ([]byte, error) {
	file, err := os.ReadFile(getConfigFileName())
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
		Reviewers:    []string{},
		Assignee:     "@me",
		TargetBranch: "",
	}

	yamlData, err := yaml.Marshal(config)
	if err != nil {
		fmt.Println("Error creating default config: ", err)
		return err
	}

	err = os.WriteFile(getConfigFileName(), yamlData, 0644)
	if err != nil {
		return err
	}
	fmt.Println("Default config created.")
	return nil
}

func readReviewersFromConfig() ([]string, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	return config.Reviewers, nil
}

func getTargetBranchFromConfig() string {
	config, err := loadConfig()
	if err != nil {
		return ""
	}
	return config.TargetBranch
}

func updateTargetBranchConfig(targetBranch string) error {
	config, err := loadConfig()
	if err != nil {
		return err
	}
	config.TargetBranch = targetBranch
	return updateConfig(config)
}

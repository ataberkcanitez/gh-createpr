package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

const configFileName = "gh-createpr-configuration.yml"

type Config struct {
	Reviewers []string `yaml:"reviewers"`
	Assignee  string   `yaml:"assignee"`
}

func updateConfig(config *Config) error {
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
		Assignee:  "@me",
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

func readReviewersFromConfig() ([]string, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	return config.Reviewers, nil
}

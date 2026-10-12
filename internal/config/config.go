// Package config reads and writes the gh-createpr YAML configuration file.
package config

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"
)

// FileName is the name of the configuration file stored in the user's home directory.
const FileName = "gh-createpr-configuration.yml"

// DefaultAssignee is used when a new configuration file is created.
const DefaultAssignee = "@me"

var (
	// ErrReviewerExists is returned when adding a reviewer that is already configured.
	ErrReviewerExists = errors.New("reviewer already exists")
	// ErrReviewerNotFound is returned when removing a reviewer that is not configured.
	ErrReviewerNotFound = errors.New("reviewer not found")
)

// Config is the persisted configuration.
type Config struct {
	Reviewers    []string `yaml:"reviewers"`
	Assignee     string   `yaml:"assignee"`
	TargetBranch string   `yaml:"target_branch,omitempty"`
}

// DefaultPath returns the path of the configuration file in the user's home directory.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locating home directory: %w", err)
	}
	return filepath.Join(home, FileName), nil
}

// Store loads and saves the configuration file at Path.
// Informational messages (e.g. "Default config created.") are written to Out.
type Store struct {
	Path string
	Out  io.Writer
}

// Raw returns the file contents, creating a default configuration if the file does not exist.
func (s *Store) Raw() ([]byte, error) {
	data, err := os.ReadFile(s.Path)
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintln(s.Out, "Config file not found. Creating a new one.")
		if err := s.createDefault(); err != nil {
			return nil, err
		}
		return os.ReadFile(s.Path)
	}
	return data, err
}

// Load parses the configuration, creating a default one if the file does not exist.
func (s *Store) Load() (*Config, error) {
	data, err := s.Raw()
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", s.Path, err)
	}
	return &cfg, nil
}

// Save writes the configuration to disk.
func (s *Store) Save(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0o644)
}

func (s *Store) createDefault() error {
	cfg := &Config{Reviewers: []string{}, Assignee: DefaultAssignee}
	if err := s.Save(cfg); err != nil {
		return err
	}
	fmt.Fprintln(s.Out, "Default config created.")
	return nil
}

// Update loads the configuration, applies fn and saves the result.
// Nothing is saved when fn returns an error.
func (s *Store) Update(fn func(*Config) error) error {
	cfg, err := s.Load()
	if err != nil {
		return err
	}
	if err := fn(cfg); err != nil {
		return err
	}
	return s.Save(cfg)
}

// AddReviewer appends reviewer to the configuration.
func (s *Store) AddReviewer(reviewer string) error {
	return s.Update(func(c *Config) error {
		if slices.Contains(c.Reviewers, reviewer) {
			return ErrReviewerExists
		}
		c.Reviewers = append(c.Reviewers, reviewer)
		return nil
	})
}

// RemoveReviewer removes every occurrence of reviewer from the configuration.
func (s *Store) RemoveReviewer(reviewer string) error {
	return s.Update(func(c *Config) error {
		if !slices.Contains(c.Reviewers, reviewer) {
			return ErrReviewerNotFound
		}
		c.Reviewers = slices.DeleteFunc(c.Reviewers, func(r string) bool { return r == reviewer })
		return nil
	})
}

// SetAssignee stores the default assignee.
func (s *Store) SetAssignee(assignee string) error {
	return s.Update(func(c *Config) error {
		c.Assignee = assignee
		return nil
	})
}

// SetTargetBranch stores the default base branch for new pull requests.
func (s *Store) SetTargetBranch(branch string) error {
	return s.Update(func(c *Config) error {
		c.TargetBranch = branch
		return nil
	})
}

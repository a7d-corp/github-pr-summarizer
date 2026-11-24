// Command github-pr-summarizer is a terminal dashboard for open GitHub pull
// requests in a configured list of repos.
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/a7d-corp/github-pr-summarizer/internal/config"
	"github.com/a7d-corp/github-pr-summarizer/internal/github"
	"github.com/a7d-corp/github-pr-summarizer/internal/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	configPath := flag.String("config", "", "path to the repo allowlist YAML file (default: ~/.config/github-pr-summarizer/repos.yaml)")
	repo := flag.String("repo", "", "check only this repo (owner/name); the config file is ignored")
	flag.Parse()

	cfg, err := loadConfig(*configPath, *repo)
	if err != nil {
		return err
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}

	client := github.NewClient(token, cfg.Repos)
	model := ui.NewModel(client, client, cfg.Allowlist())

	_, err = tea.NewProgram(model, tea.WithAltScreen()).Run()
	return err
}

// loadConfig uses repo when it is set. Otherwise it reads the config file.
func loadConfig(configPath, repo string) (config.Config, error) {
	if repo != "" {
		return config.FromRepo(repo)
	}

	if configPath == "" {
		defaultPath, err := config.DefaultPath()
		if err != nil {
			return config.Config{}, fmt.Errorf("resolve default config path: %w", err)
		}
		configPath = defaultPath
	}

	return config.Load(configPath)
}

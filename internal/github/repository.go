package github

import (
	"fmt"
	"net/url"
	"strings"
)

// Repository represents a GitHub repository with owner and name.
type Repository struct {
	Owner string
	Repo  string
}

// String returns "owner/repo" format.
func (r Repository) String() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Repo)
}

// TryParse parses a repository from "owner/repo" format or a GitHub URL.
func TryParse(src string) (Repository, error) {
	if src == "" {
		return Repository{}, fmt.Errorf("invalid repository: cannot be empty")
	}

	if strings.HasPrefix(src, "http://github.com") || strings.HasPrefix(src, "https://github.com") {
		return parseURL(src)
	}
	return parse(src)
}

func parse(input string) (Repository, error) {
	if !strings.Contains(input, "/") {
		return Repository{}, fmt.Errorf("invalid repository: use {owner}/{repo} format")
	}

	parts := strings.Split(input, "/")
	var filtered []string
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) < 2 {
		return Repository{}, fmt.Errorf("invalid repository: missing owner or repo")
	}

	return Repository{Owner: filtered[0], Repo: filtered[1]}, nil
}

func parseURL(input string) (Repository, error) {
	u, err := url.Parse(input)
	if err != nil {
		return Repository{}, fmt.Errorf("invalid repository URL: %w", err)
	}

	parts := strings.Split(u.Path, "/")
	var filtered []string
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}

	if len(filtered) < 2 {
		return Repository{}, fmt.Errorf("invalid repository URL: missing owner or repo")
	}

	return Repository{Owner: filtered[0], Repo: filtered[1]}, nil
}

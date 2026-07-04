package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// GithubClient handles GitHub API requests with optional authentication.
type GithubClient struct {
	Token  string
	Client *http.Client
}

// NewGithubClient creates a client with an optional token.
func NewGithubClient(token string) *GithubClient {
	return &GithubClient{
		Token:  token,
		Client: &http.Client{},
	}
}

// GithubClientFromEnvironment creates a client by reading auth from environment.
func GithubClientFromEnvironment() *GithubClient {
	if isAuthDisabled() {
		return NewGithubClient("")
	}

	token := os.Getenv("DRA_GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	if token == "" {
		token = githubCLIToken()
	}

	return NewGithubClient(token)
}

// GetRelease fetches a release (latest or by tag).
func (c *GithubClient) GetRelease(repo Repository, tag *Tag) (Release, error) {
	releaseURL := getReleaseURL(repo, tag)

	req, err := http.NewRequest("GET", releaseURL, nil)
	if err != nil {
		return Release{}, fmt.Errorf("failed to create request: %w", err)
	}
	c.setHeaders(req)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Release{}, fmt.Errorf("repository or release not found: %s", repo)
	}
	if resp.StatusCode == http.StatusForbidden {
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			return Release{}, fmt.Errorf("GitHub API rate limit exceeded. Set GITHUB_TOKEN to increase limits")
		}
		return Release{}, fmt.Errorf("unauthorized: check your GitHub token")
	}
	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("GitHub API error: HTTP %d", resp.StatusCode)
	}

	var releaseResp ReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&releaseResp); err != nil {
		return Release{}, fmt.Errorf("failed to parse release response: %w", err)
	}

	return ReleaseFromResponse(releaseResp, repo), nil
}

// DownloadAsset streams an asset download, returning the body and content length.
func (c *GithubClient) DownloadAsset(asset Asset) (io.ReadCloser, int64, error) {
	req, err := http.NewRequest("GET", asset.DownloadURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.raw")
	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to download asset: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, 0, fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	return resp.Body, resp.ContentLength, nil
}

func (c *GithubClient) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github+json")
	if c.Token != "" {
		req.Header.Set("Authorization", "token "+c.Token)
	}
}

func getReleaseURL(repo Repository, tag *Tag) string {
	release := "latest"
	if tag != nil {
		release = "tags/" + tag.Value
	}
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/%s",
		repo.Owner, repo.Repo, release)
}

func isAuthDisabled() bool {
	v := strings.ToLower(os.Getenv("DRA_DISABLE_GITHUB_AUTHENTICATION"))
	return v == "true" || v == "1" || v == "yes"
}

func githubCLIToken() string {
	cmd := exec.Command("gh", "auth", "token")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

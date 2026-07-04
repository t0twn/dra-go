package github

import (
	"fmt"
	"strings"
)

// Tag represents a release tag.
type Tag struct {
	Value string
}

// Version returns the tag with leading 'v' removed.
func (t Tag) Version() string {
	return strings.TrimPrefix(t.Value, "v")
}

// Asset represents a downloadable release asset.
type Asset struct {
	Name        string
	DisplayName string
	DownloadURL string
}

// ShowName returns display_name if set, otherwise name.
func (a Asset) ShowName() string {
	if a.DisplayName != "" {
		return a.DisplayName
	}
	return a.Name
}

// IsSameName checks if asset matches the given name by display_name or name.
func (a Asset) IsSameName(name string) bool {
	if a.DisplayName != "" && a.DisplayName == name {
		return true
	}
	return a.Name == name
}

// Release represents a GitHub release with its assets.
type Release struct {
	Tag    Tag
	Assets []Asset
}

// ReleaseResponse is the JSON response from GitHub API.
type ReleaseResponse struct {
	TagName    string         `json:"tag_name"`
	TarballURL string         `json:"tarball_url"`
	ZipballURL string         `json:"zipball_url"`
	Assets     []AssetResponse `json:"assets"`
}

// AssetResponse is a single asset in the GitHub API response.
type AssetResponse struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// ReleaseFromResponse builds a Release from the API response.
func ReleaseFromResponse(resp ReleaseResponse, repo Repository) Release {
	tag := Tag{Value: resp.TagName}

	sourceBase := sourceCode(repo, tag)
	tarball := tarballAsset(resp.TarballURL, sourceBase)
	zipball := zipballAsset(resp.ZipballURL, sourceBase)

	assets := make([]Asset, 0, len(resp.Assets)+2)
	for _, a := range resp.Assets {
		assets = append(assets, Asset{
			Name:        a.Name,
			DownloadURL: a.BrowserDownloadURL,
		})
	}
	assets = append(assets, tarball, zipball)

	return Release{Tag: tag, Assets: assets}
}

func tarballAsset(url, baseName string) Asset {
	return Asset{
		Name:        fmt.Sprintf("%s.tar.gz", baseName),
		DownloadURL: url,
		DisplayName: "Source code (tar.gz)",
	}
}

func zipballAsset(url, baseName string) Asset {
	return Asset{
		Name:        fmt.Sprintf("%s.zip", baseName),
		DownloadURL: url,
		DisplayName: "Source code (zip)",
	}
}

func sourceCode(repo Repository, tag Tag) string {
	return fmt.Sprintf("%s-%s-source-code", repo.Repo, tag.Version())
}

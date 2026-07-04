package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"dra/internal/github"
	"dra/internal/installer"
	"dra/internal/progress"
	"dra/internal/system"
	"dra/internal/wildcard"

	"github.com/spf13/cobra"
)

func newDownloadCmd() *cobra.Command {
	var (
		selectPattern string
		automatic     bool
		tag           string
		output        string
		installFlag   bool
		installFiles  []string
	)

	cmd := &cobra.Command{
		Use:   "download <owner/repo>",
		Short: "Select and download an asset",
		Long: `Select and download a release asset from a GitHub repository.

SUPPORTED PATTERNS for --select:
  - Literal: exact asset name (e.g., helloworld.tar.gz)
  - Untagged: version-free pattern with {tag} (e.g., helloworld_{tag}.tar.gz)
  - Wildcard: uses * and ? (e.g., helloworld*_amd64.deb)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := github.TryParse(args[0])
			if err != nil {
				return err
			}

			return runDownload(repo, selectPattern, automatic, tag, output, installFlag, installFiles)
		},
	}

	cmd.Flags().StringVarP(&selectPattern, "select", "s", "", "Select asset by pattern (literal, {tag}, or wildcard)")
	cmd.Flags().BoolVarP(&automatic, "automatic", "a", false, "Auto-select based on OS and architecture")
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "Fetch a specific release tag (default: latest)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Custom output path (file or directory)")
	cmd.Flags().BoolVarP(&installFlag, "install", "i", false, "Install the downloaded asset")
	cmd.Flags().StringArrayVarP(&installFiles, "install-file", "I", nil, "Select specific executable to install from archive")

	// select and automatic are mutually exclusive
	cmd.MarkFlagsMutuallyExclusive("select", "automatic")

	return cmd
}

func runDownload(repo github.Repository, selectPattern string, automatic bool, tagStr, output string, installFlag bool, installFiles []string) error {
	githubClient := github.GithubClientFromEnvironment()

	// Determine tag
	var releaseTag *github.Tag
	if tagStr != "" {
		t := github.Tag{Value: tagStr}
		releaseTag = &t
	}

	// Fetch release
	fmt.Printf("Fetching release for %s...\n", repo)
	release, err := githubClient.GetRelease(repo, releaseTag)
	if err != nil {
		return fmt.Errorf("failed to fetch release: %w", err)
	}
	fmt.Printf("Release tag is %s\n", release.Tag.Value)

	// Select asset
	var selectedAsset *github.Asset

	switch {
	case selectPattern != "":
		// Pattern-based selection
		tagged := github.TaggedAsset{}
		assetName := tagged.Tag(release.Tag, selectPattern)
		for i := range release.Assets {
			if wildcard.Match(assetName, release.Assets[i].Name) {
				selectedAsset = &release.Assets[i]
				break
			}
		}
		if selectedAsset == nil {
			return fmt.Errorf("no asset found matching pattern: %s", selectPattern)
		}

	case automatic:
		// Auto-select based on OS/arch
		sys, err := system.Detect()
		if err != nil {
			return fmt.Errorf("error detecting system: %w", err)
		}
		selectedAsset = system.FindAsset(sys, release.Assets)
		if selectedAsset == nil {
			return fmt.Errorf("cannot find asset matching your system %s %s", sys.GetOS(), sys.GetArch())
		}
		fmt.Printf("Auto-selected: %s\n", selectedAsset.Name)

	default:
		// Interactive selection
		asset, err := askSelectAsset(release.Assets)
		if err != nil {
			return err
		}
		selectedAsset = asset
	}

	if selectedAsset == nil {
		return fmt.Errorf("no asset selected")
	}

	// Determine output path
	installMode := installFlag || len(installFiles) > 0
	outputPath := chooseOutputPath(output, installMode, selectedAsset.Name)

	// Download
	fmt.Printf("Downloading %s...\n", selectedAsset.Name)
	if err := downloadAsset(githubClient, selectedAsset, outputPath); err != nil {
		return err
	}

	// Install if requested
	if installMode {
		return installAsset(selectedAsset.Name, outputPath, output, installFiles, repo)
	}

	fmt.Printf("\033[32mDownloaded %s\033[0m\n", selectedAsset.Name)
	return nil
}

func downloadAsset(client *github.GithubClient, asset *github.Asset, outputPath string) error {
	body, contentLength, err := client.DownloadAsset(*asset)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer body.Close()

	// Create output directory if needed
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", outputPath, err)
	}
	defer outFile.Close()

	// Set up progress tracking
	pw := progress.NewWriter(asset.Name, contentLength)
	reader := pw.NewReader(body)

	if _, err := io.Copy(outFile, reader); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	pw.Finish()

	return nil
}

func chooseOutputPath(output string, installMode bool, assetName string) string {
	if installMode {
		return installer.TempFilePath()
	}

	if output == "" {
		return assetName
	}

	// Check if output is a directory
	info, err := os.Stat(output)
	if err == nil && info.IsDir() {
		return filepath.Join(output, assetName)
	}

	return output
}

func installAsset(assetName, downloadPath, output string, installFiles []string, repo github.Repository) error {
	fmt.Println("Installing...")

	// Determine destination
	var dest installer.Destination
	if output != "" {
		info, err := os.Stat(output)
		if err == nil && info.IsDir() {
			dest = installer.Destination{Directory: output}
		} else {
			dest = installer.Destination{File: output}
		}
	} else {
		cwd, _ := os.Getwd()
		dest = installer.Destination{Directory: cwd}
	}

	// Validate: multiple executables need directory destination
	if len(installFiles) > 1 && !dest.IsDirectory() {
		return fmt.Errorf("%s is not a directory. When installing multiple executables, output must be a directory", output)
	}

	// Build executable list
	var executables []installer.Executable
	if len(installFiles) > 0 {
		for _, name := range installFiles {
			executables = append(executables, installer.Executable{Selected: name})
		}
	} else {
		executables = append(executables, installer.Executable{Automatic: repo.Repo})
	}

	result, err := installer.Install(assetName, downloadPath, dest, executables)
	if err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	// Clean up temp download file
	os.Remove(downloadPath)

	fmt.Printf("\033[32m%s\033[0m\n", result.Message)
	fmt.Printf("\033[32mInstallation completed!\033[0m\n")
	return nil
}

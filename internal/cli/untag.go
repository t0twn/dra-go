package cli

import (
	"fmt"

	"dra/internal/github"

	"github.com/spf13/cobra"
)

func newUntagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "untag <owner/repo>",
		Short: "Select an asset and generate an untagged version of it",
		Long: `Select a release asset interactively and print its version-free pattern.
The version string is replaced with {tag} placeholder.

EXAMPLE:
  dra untag devmatteini/dra-tests
  # Output: helloworld_{tag}.tar.gz`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := github.TryParse(args[0])
			if err != nil {
				return err
			}

			return runUntag(repo)
		},
	}

	return cmd
}

func runUntag(repo github.Repository) error {
	githubClient := github.GithubClientFromEnvironment()

	fmt.Printf("Fetching release for %s...\n", repo)
	release, err := githubClient.GetRelease(repo, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch release: %w", err)
	}

	if len(release.Assets) == 0 {
		return fmt.Errorf("release %s has no assets", release.Tag.Value)
	}

	asset, err := askSelectAsset(release.Assets)
	if err != nil {
		return err
	}

	tagged := github.TaggedAsset{}
	untagged := tagged.Untag(release.Tag, asset.Name)
	fmt.Println(untagged)

	return nil
}

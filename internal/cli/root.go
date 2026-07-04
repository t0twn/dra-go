package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.10.2-go"

// NewRootCmd creates the root cobra command.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "dra",
		Short: "A command line tool to download release assets from GitHub",
		Long: `dra - Download Release Assets

A command line tool to download release assets from GitHub repositories.
Supports interactive selection, pattern matching, automatic OS/arch detection,
and installation of downloaded assets.

EXAMPLES:
  # Interactive download
  dra download devmatteini/dra-tests

  # Automatic download based on OS/arch
  dra download -a BurntSushi/ripgrep

  # Pattern-based download
  dra download -s "*linux*amd64*.tar.gz" cli/cli

  # Download and install
  dra download --install spf13/cobra`,
		Version: version,
	}

	rootCmd.AddCommand(newDownloadCmd())
	rootCmd.AddCommand(newUntagCmd())

	// Silence default error printing, we handle it ourselves
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true

	return rootCmd
}

// Execute runs the root command.
func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\033[31mError: %s\033[0m\n", err)
		os.Exit(1)
	}
}

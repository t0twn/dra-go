package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"dra/internal/github"
)

// askSelectAsset prompts the user to select an asset interactively.
func askSelectAsset(assets []github.Asset) (*github.Asset, error) {
	// Check if stdin is a terminal
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return nil, fmt.Errorf("interactive mode requires a terminal; use --select or --automatic for non-interactive mode")
	}

	fmt.Println("Pick the asset to download:")
	for i, a := range assets {
		fmt.Printf("  %d) %s\n", i+1, a.ShowName())
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter number (or 'q' to quit): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "q" {
			return nil, fmt.Errorf("no asset selected")
		}

		num, err := strconv.Atoi(input)
		if err != nil || num < 1 || num > len(assets) {
			fmt.Printf("Invalid selection. Please enter a number between 1 and %d\n", len(assets))
			continue
		}

		return &assets[num-1], nil
	}
}

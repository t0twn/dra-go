package system

import (
	"sort"
	"strings"

	"dra/internal/github"
)

// FindAsset filters and selects the best matching asset for the given system.
func FindAsset(sys System, assets []github.Asset) *github.Asset {
	var ignored = []string{"sha256", "sha512", "checksums"}

	var matches []github.Asset
	for _, asset := range assets {
		skip := false
		for _, ig := range ignored {
			if strings.Contains(strings.ToLower(asset.Name), ig) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		if sys.Matches(strings.ToLower(asset.Name)) {
			matches = append(matches, asset)
		}
	}

	if len(matches) == 0 {
		return nil
	}

	sort.SliceStable(matches, func(i, j int) bool {
		return sys.AssetPriority(matches[i].Name) < sys.AssetPriority(matches[j].Name)
	})

	return &matches[0]
}

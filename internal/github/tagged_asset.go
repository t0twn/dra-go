package github

import "strings"

const placeholder = "{tag}"

// TaggedAsset provides functions for tagging and untaging asset names.
type TaggedAsset struct{}

// Tag replaces {tag} placeholder with the version string.
func (TaggedAsset) Tag(tag Tag, untagged string) string {
	return strings.ReplaceAll(untagged, placeholder, tag.Version())
}

// Untag replaces the version string with {tag} placeholder.
func (TaggedAsset) Untag(tag Tag, assetName string) string {
	return strings.ReplaceAll(assetName, tag.Version(), placeholder)
}

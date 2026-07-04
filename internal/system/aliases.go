package system

import "strings"

// OSAliases returns lowercase aliases for an OS.
func OSAliases(os OS) []string {
	switch os {
	case OSLinux:
		return []string{"linux", "alpine", "ubuntu", "debian", "unknown-linux"}
	case OSMac:
		return []string{"macos", "darwin", "apple", "osx", "mac"}
	case OSWindows:
		return []string{"windows", "win64", "win-64bit", "win"}
	default:
		return nil
	}
}

// ArchAliases returns lowercase aliases for an architecture.
func ArchAliases(arch Arch) []string {
	switch arch {
	case ArchX86_64:
		return []string{"x86_64", "amd64", "x64"}
	case ArchArm64:
		return []string{"aarch64", "arm64"}
	case ArchArmV6:
		return []string{"arm", "armv6", "armv7", "armhf"}
	default:
		return nil
	}
}

// containsAny checks if the asset name (lowercased) contains any of the given aliases.
func containsAny(assetNameLower string, aliases []string) bool {
	for _, alias := range aliases {
		if strings.Contains(assetNameLower, alias) {
			return true
		}
	}
	return false
}

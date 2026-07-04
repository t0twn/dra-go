package system

import "strings"

// LinuxX86_64 represents Linux on x86_64.
type LinuxX86_64 struct{}

func (s *LinuxX86_64) GetOS() OS   { return OSLinux }
func (s *LinuxX86_64) GetArch() Arch { return ArchX86_64 }

func (s *LinuxX86_64) Matches(assetName string) bool {
	return linuxMatches(OSLinux, ArchX86_64, assetName)
}

func (s *LinuxX86_64) AssetPriority(assetName string) int {
	return linuxAssetPriority(assetName)
}

// LinuxArmV6 represents Linux on ARM v6.
type LinuxArmV6 struct{}

func (s *LinuxArmV6) GetOS() OS   { return OSLinux }
func (s *LinuxArmV6) GetArch() Arch { return ArchArmV6 }

func (s *LinuxArmV6) Matches(assetName string) bool {
	return linuxMatches(OSLinux, ArchArmV6, assetName)
}

func (s *LinuxArmV6) AssetPriority(assetName string) int {
	return linuxAssetPriority(assetName)
}

// LinuxArm64 represents Linux on ARM64.
type LinuxArm64 struct{}

func (s *LinuxArm64) GetOS() OS   { return OSLinux }
func (s *LinuxArm64) GetArch() Arch { return ArchArm64 }

func (s *LinuxArm64) Matches(assetName string) bool {
	return linuxMatches(OSLinux, ArchArm64, assetName)
}

func (s *LinuxArm64) AssetPriority(assetName string) int {
	return linuxAssetPriority(assetName)
}

func linuxMatches(os OS, arch Arch, assetName string) bool {
	name := strings.ToLower(assetName)
	sameArch := containsAny(name, ArchAliases(arch))
	sameOS := containsAny(name, OSAliases(os)) && sameArch
	sameArchAndExt := sameArch && hasLinuxExtension(name)
	return sameOS || sameArchAndExt
}

func hasLinuxExtension(name string) bool {
	return strings.HasSuffix(name, ".appimage")
}

func linuxAssetPriority(name string) int {
	isArchive := hasArchiveExtension(name)
	isMusl := strings.Contains(name, "musl")

	if isMusl && isArchive {
		return 1
	}
	if isMusl {
		return 2
	}
	if isArchive {
		return 3
	}
	return 4
}

// MacOSX86_64 represents macOS on x86_64.
type MacOSX86_64 struct{}

func (s *MacOSX86_64) GetOS() OS   { return OSMac }
func (s *MacOSX86_64) GetArch() Arch { return ArchX86_64 }

func (s *MacOSX86_64) Matches(assetName string) bool {
	return macMatches(OSMac, ArchX86_64, assetName)
}

func (s *MacOSX86_64) AssetPriority(assetName string) int {
	return macAssetPriority(assetName)
}

// MacOSArm64 represents macOS on ARM64.
type MacOSArm64 struct{}

func (s *MacOSArm64) GetOS() OS   { return OSMac }
func (s *MacOSArm64) GetArch() Arch { return ArchArm64 }

func (s *MacOSArm64) Matches(assetName string) bool {
	return macMatches(OSMac, ArchArm64, assetName)
}

func (s *MacOSArm64) AssetPriority(assetName string) int {
	return macAssetPriority(assetName)
}

func macMatches(os OS, arch Arch, assetName string) bool {
	name := strings.ToLower(assetName)
	sameArch := containsAny(name, ArchAliases(arch))
	sameOS := containsAny(name, OSAliases(os)) && sameArch
	sameArchAndExt := sameArch && hasMacExtension(name)
	return sameOS || sameArchAndExt
}

func hasMacExtension(name string) bool {
	return strings.HasSuffix(name, ".dmg")
}

func macAssetPriority(name string) int {
	isArchive := hasArchiveExtension(name)
	isDmg := strings.HasSuffix(name, ".dmg")

	if isArchive {
		return 1
	}
	if isDmg {
		return 2
	}
	return 3
}

// WindowsX86_64 represents Windows on x86_64.
type WindowsX86_64 struct{}

func (s *WindowsX86_64) GetOS() OS   { return OSWindows }
func (s *WindowsX86_64) GetArch() Arch { return ArchX86_64 }

func (s *WindowsX86_64) Matches(assetName string) bool {
	name := strings.ToLower(assetName)
	sameArch := containsAny(name, ArchAliases(ArchX86_64))
	sameOS := containsAny(name, OSAliases(OSWindows)) && sameArch
	sameArchAndExt := sameArch && hasWindowsExtension(name)
	return sameOS || sameArchAndExt
}

func (s *WindowsX86_64) AssetPriority(assetName string) int {
	name := strings.ToLower(assetName)
	isArchive := hasArchiveExtension(name)
	isExe := strings.HasSuffix(name, ".exe") || strings.HasSuffix(name, ".msi")

	if isArchive {
		return 1
	}
	if isExe {
		return 2
	}
	return 3
}

func hasWindowsExtension(name string) bool {
	return strings.HasSuffix(name, ".exe") || strings.HasSuffix(name, ".msi")
}

// Common helpers

var archiveExtensions = []string{".gz", ".tgz", ".bz2", ".tbz", ".xz", ".txz", ".zip"}

func hasArchiveExtension(name string) bool {
	for _, ext := range archiveExtensions {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

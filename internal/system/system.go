package system

// OS represents an operating system.
type OS int

const (
	OSLinux OS = iota
	OSMac
	OSWindows
)

func (o OS) String() string {
	switch o {
	case OSLinux:
		return "linux"
	case OSMac:
		return "macos"
	case OSWindows:
		return "windows"
	default:
		return "unknown"
	}
}

// Arch represents a CPU architecture.
type Arch int

const (
	ArchX86_64 Arch = iota
	ArchArmV6
	ArchArm64
)

func (a Arch) String() string {
	switch a {
	case ArchX86_64:
		return "x86_64"
	case ArchArmV6:
		return "arm"
	case ArchArm64:
		return "aarch64"
	default:
		return "unknown"
	}
}

// System defines the interface for OS/arch-specific asset matching.
type System interface {
	GetOS() OS
	GetArch() Arch
	Matches(assetName string) bool
	AssetPriority(assetName string) int
}

package system

import (
	"fmt"
	"runtime"
)

// Detect returns the current system's System implementation.
func Detect() (System, error) {
	os := runtime.GOOS
	arch := runtime.GOARCH

	switch os {
	case "linux":
		switch arch {
		case "amd64":
			return &LinuxX86_64{}, nil
		case "arm", "armv6", "armv7":
			return &LinuxArmV6{}, nil
		case "arm64":
			return &LinuxArm64{}, nil
		}
	case "darwin":
		switch arch {
		case "amd64":
			return &MacOSX86_64{}, nil
		case "arm64":
			return &MacOSArm64{}, nil
		}
	case "windows":
		switch arch {
		case "amd64":
			return &WindowsX86_64{}, nil
		}
	}

	return nil, fmt.Errorf("unknown operating system or architecture: %s %s", os, arch)
}

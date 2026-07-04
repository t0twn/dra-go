package installer

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Install dispatches installation to the appropriate handler based on file type.
func Install(assetName, sourcePath string, dest Destination, executables []Executable) (*InstallOutput, error) {
	fileType, compression, err := DetectFileType(assetName, sourcePath)
	if err != nil {
		return nil, err
	}

	switch fileType {
	case FileTypeDebian:
		return installDebian(sourcePath)
	case FileTypeRpm:
		return installRpm(sourcePath)
	case FileTypeTarArchive:
		return installTar(sourcePath, dest, compression, executables, assetName)
	case FileTypeZipArchive:
		return installZip(sourcePath, dest, executables, assetName)
	case FileTypeCompressedFile:
		return installCompressed(sourcePath, dest, compression)
	case FileTypeExecutableFile:
		return installExecutable(sourcePath, dest, assetName)
	default:
		return nil, fmt.Errorf("unsupported file type for installation: %s", assetName)
	}
}

// installDebian installs a .deb package via dpkg.
func installDebian(path string) (*InstallOutput, error) {
	if _, err := exec.LookPath("dpkg"); err != nil {
		return nil, fmt.Errorf("dpkg not found in PATH: cannot install .deb packages")
	}
	cmd := exec.Command("dpkg", "--install", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("dpkg install failed: %w", err)
	}
	return &InstallOutput{Message: fmt.Sprintf("Installed %s via dpkg", filepath.Base(path))}, nil
}

// installRpm installs a .rpm package via rpm.
func installRpm(path string) (*InstallOutput, error) {
	if _, err := exec.LookPath("rpm"); err != nil {
		return nil, fmt.Errorf("rpm not found in PATH: cannot install .rpm packages")
	}
	cmd := exec.Command("rpm", "--install", "--replacepkgs", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("rpm install failed: %w", err)
	}
	return &InstallOutput{Message: fmt.Sprintf("Installed %s via rpm", filepath.Base(path))}, nil
}

// installCompressed decompresses a single compressed file and installs it.
func installCompressed(sourcePath string, dest Destination, compression Compression) (*InstallOutput, error) {
	reader, err := openDecompressor(sourcePath, compression)
	if err != nil {
		return nil, fmt.Errorf("failed to open decompressor: %w", err)
	}
	defer reader.Close()

	outPath := resolveOutputPath(dest, strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(
		filepath.Base(sourcePath), ".gz"), ".xz"), ".bz2"))

	outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, reader); err != nil {
		return nil, fmt.Errorf("failed to decompress: %w", err)
	}

	return &InstallOutput{Message: fmt.Sprintf("Installed %s", outPath)}, nil
}

// installExecutable copies an executable file to the destination and sets +x.
func installExecutable(sourcePath string, dest Destination, assetName string) (*InstallOutput, error) {
	outPath := resolveOutputPath(dest, assetName)

	if err := copyFile(sourcePath, outPath); err != nil {
		return nil, fmt.Errorf("failed to copy executable: %w", err)
	}

	if err := os.Chmod(outPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to set executable permission: %w", err)
	}

	return &InstallOutput{Message: fmt.Sprintf("Installed %s", outPath)}, nil
}

// findExecutables walks a directory and finds executable files.
func findExecutables(dir string, maxDepth int) ([]string, error) {
	var executables []string

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate depth
		rel, _ := filepath.Rel(dir, path)
		depth := strings.Count(rel, string(filepath.Separator))
		if d.IsDir() && depth >= maxDepth {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Check if file is executable
		if info.Mode()&0111 != 0 {
			executables = append(executables, path)
			return nil
		}

		// Check ELF/Mach-O magic
		if isExecutableByMagic(path) {
			executables = append(executables, path)
		}

		return nil
	})

	return executables, err
}

// resolveExecutable determines the output path for an executable from an archive.
func resolveExecutable(exePath, assetName string, dest Destination, executables []Executable) (string, error) {
	exeName := filepath.Base(exePath)

	if len(executables) > 0 {
		for _, e := range executables {
			if e.IsSelected() {
				if exeName == e.Selected || filepath.Base(exePath) == e.Selected {
					return resolveOutputPath(dest, exeName), nil
				}
				return "", nil // not the selected one
			}
			// Automatic: match by repo name hint
			if e.Automatic != "" {
				lower := strings.ToLower(exeName)
				hint := strings.ToLower(e.Automatic)
				if lower == hint || strings.Contains(lower, hint) {
					return resolveOutputPath(dest, exeName), nil
				}
			}
		}
		// If we have specific selections and this doesn't match, skip it
		hasSelected := false
		for _, e := range executables {
			if e.IsSelected() {
				hasSelected = true
				break
			}
		}
		if hasSelected {
			return "", nil
		}
	}

	// Automatic mode: if only one executable, use it
	if dest.IsDirectory() {
		return filepath.Join(dest.Directory, exeName), nil
	}
	return dest.File, nil
}

// resolveOutputPath computes the final output path.
func resolveOutputPath(dest Destination, name string) string {
	if dest.IsDirectory() {
		return filepath.Join(dest.Directory, name)
	}
	return dest.File
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

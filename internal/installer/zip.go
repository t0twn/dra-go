package installer

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// installZip extracts a zip archive and installs executables.
func installZip(sourcePath string, dest Destination, executables []Executable, assetName string) (*InstallOutput, error) {
	r, err := zip.OpenReader(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	// Create temp dir for extraction
	tmpDir, err := os.MkdirTemp("", "dra-zip-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract zip
	for _, f := range r.File {
		target := filepath.Join(tmpDir, f.Name)

		// Security: reject path traversal
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(tmpDir)+string(os.PathSeparator)) &&
			filepath.Clean(target) != filepath.Clean(tmpDir) {
			return nil, fmt.Errorf("zip entry attempts path traversal: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		dir := filepath.Dir(target)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create parent dir: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open zip entry: %w", err)
		}

		mode := f.Mode()
		if mode == 0 {
			mode = 0644
		}

		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
		if err != nil {
			rc.Close()
			return nil, fmt.Errorf("failed to create file: %w", err)
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			outFile.Close()
			rc.Close()
			return nil, fmt.Errorf("failed to write file: %w", err)
		}
		outFile.Close()
		rc.Close()
	}

	// Find executables
	foundExecs, err := findExecutables(tmpDir, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to find executables: %w", err)
	}

	if len(foundExecs) == 0 {
		return nil, fmt.Errorf("no executables found in zip archive %s", assetName)
	}

	// If multiple executables and single destination file, error
	if !dest.IsDirectory() && len(foundExecs) > 1 {
		hasSelections := false
		for _, e := range executables {
			if e.IsSelected() {
				hasSelections = true
				break
			}
		}
		if !hasSelections {
			return nil, fmt.Errorf("archive contains %d executables; use --install-file to select which to install, or use --output with a directory", len(foundExecs))
		}
	}

	// Install executables
	var installed []string
	for _, exePath := range foundExecs {
		outPath, err := resolveExecutable(exePath, assetName, dest, executables)
		if err != nil {
			return nil, err
		}
		if outPath == "" {
			continue
		}

		if err := copyFile(exePath, outPath); err != nil {
			return nil, fmt.Errorf("failed to copy executable: %w", err)
		}
		if err := os.Chmod(outPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to chmod: %w", err)
		}
		installed = append(installed, outPath)
	}

	if len(installed) == 0 {
		return nil, fmt.Errorf("no matching executables found in zip archive %s", assetName)
	}

	return &InstallOutput{Message: fmt.Sprintf("Installed %s", strings.Join(installed, ", "))}, nil
}

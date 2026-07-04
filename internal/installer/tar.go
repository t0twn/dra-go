package installer

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// installTar extracts a tar archive and installs executables.
func installTar(sourcePath string, dest Destination, compression Compression, executables []Executable, assetName string) (*InstallOutput, error) {
	reader, err := openDecompressor(sourcePath, compression)
	if err != nil {
		return nil, fmt.Errorf("failed to open tar decompressor: %w", err)
	}
	defer reader.Close()

	// Create temp dir for extraction
	tmpDir, err := os.MkdirTemp("", "dra-tar-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract tar
	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tar read error: %w", err)
		}

		// Security: reject path traversal
		target := filepath.Join(tmpDir, header.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(tmpDir)+string(os.PathSeparator)) &&
			filepath.Clean(target) != filepath.Clean(tmpDir) {
			return nil, fmt.Errorf("tar entry attempts path traversal: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create parent dir: %w", err)
			}
			mode := os.FileMode(header.Mode)
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if err != nil {
				return nil, fmt.Errorf("failed to create file: %w", err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return nil, fmt.Errorf("failed to write file: %w", err)
			}
			f.Close()
		case tar.TypeSymlink:
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create parent dir for symlink: %w", err)
			}
			if err := os.Symlink(header.Linkname, target); err != nil {
				// Non-fatal: some symlinks may fail on certain systems
				continue
			}
		}
	}

	// Find executables in extracted directory
	foundExecs, err := findExecutables(tmpDir, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to find executables: %w", err)
	}

	if len(foundExecs) == 0 {
		return nil, fmt.Errorf("no executables found in archive %s", assetName)
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
			continue // skip non-selected executables
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
		return nil, fmt.Errorf("no matching executables found in archive %s", assetName)
	}

	return &InstallOutput{Message: fmt.Sprintf("Installed %s", strings.Join(installed, ", "))}, nil
}

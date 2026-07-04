package installer

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Compression type for archive/compressed files.
type Compression int

const (
	CompressionGz Compression = iota
	CompressionXz
	CompressionBz2
)

func (c Compression) String() string {
	switch c {
	case CompressionGz:
		return "gz"
	case CompressionXz:
		return "xz"
	case CompressionBz2:
		return "bz2"
	default:
		return "unknown"
	}
}

// FileType represents the type of a downloadable file.
type FileType int

const (
	FileTypeDebian FileType = iota
	FileTypeRpm
	FileTypeTarArchive
	FileTypeZipArchive
	FileTypeCompressedFile
	FileTypeExecutableFile
)

// FileInfo holds path and name of a downloaded file.
type FileInfo struct {
	Path     string
	Name     string
	FileType FileType
	Compression Compression
}

// Destination for installation: directory or specific file path.
type Destination struct {
	Directory string // non-empty if destination is a directory
	File      string // non-empty if destination is a specific file
}

// IsDirectory returns true if the destination is a directory.
func (d Destination) IsDirectory() bool {
	return d.Directory != ""
}

// Executable represents which executable(s) to install from an archive.
type Executable struct {
	Automatic string // repo name hint for auto-detection
	Selected  string // exact name selection
}

// IsSelected returns true if a specific executable was selected.
func (e Executable) IsSelected() bool {
	return e.Selected != ""
}

// InstallOutput holds the result of a successful installation.
type InstallOutput struct {
	Message string
}

// ELF magic number
var elfMagic = []byte{0x7F, 'E', 'L', 'F'}

// Mach-O magic numbers
var machoMagics = [][]byte{
	{0xFE, 0xED, 0xFA, 0xCE}, // 32-bit big endian
	{0xFE, 0xED, 0xFA, 0xCF}, // 64-bit big endian
	{0xCE, 0xFA, 0xED, 0xFE}, // 32-bit little endian
	{0xCF, 0xFA, 0xED, 0xFE}, // 64-bit little endian
	{0xCA, 0xFE, 0xBA, 0xBE}, // fat binary
}

// DetectFileType detects the file type from the file name and content.
func DetectFileType(name, path string) (FileType, Compression, error) {
	lower := strings.ToLower(name)

	// Check by extension first
	if strings.HasSuffix(lower, ".deb") {
		return FileTypeDebian, 0, nil
	}
	if strings.HasSuffix(lower, ".rpm") {
		return FileTypeRpm, 0, nil
	}
	if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
		return FileTypeTarArchive, CompressionGz, nil
	}
	if strings.HasSuffix(lower, ".tar.bz2") || strings.HasSuffix(lower, ".tbz") {
		return FileTypeTarArchive, CompressionBz2, nil
	}
	if strings.HasSuffix(lower, ".tar.xz") || strings.HasSuffix(lower, ".txz") {
		return FileTypeTarArchive, CompressionXz, nil
	}
	if strings.HasSuffix(lower, ".zip") {
		return FileTypeZipArchive, 0, nil
	}
	// Single compressed files (must check AFTER tar.* variants)
	if strings.HasSuffix(lower, ".gz") {
		return FileTypeCompressedFile, CompressionGz, nil
	}
	if strings.HasSuffix(lower, ".bz2") {
		return FileTypeCompressedFile, CompressionBz2, nil
	}
	if strings.HasSuffix(lower, ".xz") {
		return FileTypeCompressedFile, CompressionXz, nil
	}

	// Check if executable by magic bytes or extension
	if isExecutableByMagic(path) ||
		strings.HasSuffix(lower, ".appimage") ||
		strings.HasSuffix(lower, ".exe") ||
		filepath.Ext(lower) == "" {
		return FileTypeExecutableFile, 0, nil
	}

	return 0, 0, fmt.Errorf("unsupported file type: %s", name)
}

func isExecutableByMagic(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	header := make([]byte, 4)
	n, err := f.Read(header)
	if err != nil || n < 4 {
		return false
	}

	// Check ELF
	if header[0] == elfMagic[0] && header[1] == elfMagic[1] &&
		header[2] == elfMagic[2] && header[3] == elfMagic[3] {
		return true
	}

	// Check Mach-O
	for _, magic := range machoMagics {
		if header[0] == magic[0] && header[1] == magic[1] &&
			header[2] == magic[2] && header[3] == magic[3] {
			return true
		}
	}

	return false
}

// TempFilePath generates a temp file path for install mode downloads.
func TempFilePath() string {
	dir := os.TempDir()
	return filepath.Join(dir, fmt.Sprintf("dra-%d", binary.BigEndian.Uint64(generateID())))
}

func generateID() []byte {
	b := make([]byte, 8)
	f, _ := os.Open("/dev/urandom")
	if f != nil {
		f.Read(b)
		f.Close()
	}
	return b
}

package installer

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/ulikunitz/xz"
)

// openDecompressor opens a file with the appropriate decompressor.
func openDecompressor(path string, compression Compression) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	switch compression {
	case CompressionGz:
		gz, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		return &wrappedReadCloser{gz, f}, nil

	case CompressionXz:
		xzReader, err := xz.NewReader(f)
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("failed to create xz reader: %w", err)
		}
		return &wrappedReadCloser{xzReader, f}, nil

	case CompressionBz2:
		// Use command-line bzip2 as fallback since pure Go bzip2 libraries are limited
		return openBz2Decompressor(path)

	default:
		f.Close()
		return nil, fmt.Errorf("unsupported compression: %s", compression)
	}
}

// wrappedReadCloser wraps a reader and closes both the reader and underlying file.
type wrappedReadCloser struct {
	reader io.Reader
	closer io.Closer
}

func (w *wrappedReadCloser) Read(p []byte) (int, error) {
	return w.reader.Read(p)
}

func (w *wrappedReadCloser) Close() error {
	// Try to close the reader if it implements io.Closer
	if rc, ok := w.reader.(io.Closer); ok {
		rc.Close()
	}
	return w.closer.Close()
}

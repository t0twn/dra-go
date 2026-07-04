package installer

import (
	"fmt"
	"io"
	"os/exec"
)

// openBz2Decompressor uses the bzip2 command-line tool.
func openBz2Decompressor(path string) (io.ReadCloser, error) {
	if _, err := exec.LookPath("bzip2"); err != nil {
		if _, err := exec.LookPath("bzcat"); err != nil {
			return nil, fmt.Errorf("bzip2/bzcat not found in PATH: cannot decompress .bz2 files. Install bzip2 first")
		}
		return newCmdReader("bzcat", path)
	}
	return newCmdReader("bzip2", "-dc", path)
}

// cmdReadCloser wraps an exec.Cmd pipe as an io.ReadCloser.
type cmdReadCloser struct {
	io.Reader
	cmd *exec.Cmd
}

func (c *cmdReadCloser) Close() error {
	return nil // process will exit when pipe is closed
}

// newCmdReader creates a reader from a command that decompresses to stdout.
func newCmdReader(name string, args ...string) (io.ReadCloser, error) {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipe for %s: %w", name, err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start %s: %w", name, err)
	}

	return &cmdReadCloser{Reader: stdout, cmd: cmd}, nil
}

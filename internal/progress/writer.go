package progress

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

// Writer wraps an io.Reader and displays download progress.
type Writer struct {
	Total       int64
	Current     int64
	Description string
	mu          sync.Mutex
	done        bool
}

// NewWriter creates a progress writer with a description.
func NewWriter(description string, total int64) *Writer {
	return &Writer{
		Total:       total,
		Description: description,
	}
}

// NewReader wraps an io.Reader with progress tracking.
func (w *Writer) NewReader(r io.Reader) io.Reader {
	return &progressReader{reader: r, writer: w}
}

// Finish marks the progress as complete.
func (w *Writer) Finish() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.done = true
	w.print()
	fmt.Println()
}

func (w *Writer) print() {
	if w.Total > 0 {
		pct := float64(w.Current) / float64(w.Total) * 100
		bar := progressBar(pct, 30)
		fmt.Printf("\r%s %s %.1f%% %s/%s",
			w.Description,
			bar,
			pct,
			humanBytes(w.Current),
			humanBytes(w.Total),
		)
	} else {
		fmt.Printf("\r%s %s", w.Description, humanBytes(w.Current))
	}
}

type progressReader struct {
	reader io.Reader
	writer *Writer
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.writer.mu.Lock()
		pr.writer.Current += int64(n)
		pr.writer.print()
		pr.writer.mu.Unlock()
	}
	return n, err
}

func progressBar(pct float64, width int) string {
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}

func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

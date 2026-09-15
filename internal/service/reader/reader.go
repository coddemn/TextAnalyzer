package reader

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"
)

type FileReadResult struct {
	Text      string
	LineCount int
	Err       error
}

type FileReader struct {
	maxRetries int
	baseDelay  time.Duration
}

func NewFileReader(maxRetries int, baseDelay time.Duration) *FileReader {
	return &FileReader{
		maxRetries: maxRetries,
		baseDelay:  baseDelay,
	}
}

func (r *FileReader) ReadWithRetry(path string) FileReadResult {
	var lastErr error

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		readRes := r.ReadWithStats(path)
		if readRes.Err == nil {

			return readRes
		}

		lastErr = readRes.Err

		if attempt < r.maxRetries {
			delay := r.calcDelay(attempt)
			time.Sleep(delay)
		}
	}

	return FileReadResult{Err: fmt.Errorf("read %q after %d retries: %w", path, r.maxRetries, lastErr)}
}

func (r *FileReader) ReadWithStats(path string) FileReadResult {

	file, err := os.Open(path)

	if err != nil {
		log.Printf("Error opening file: %v\n", err)
		return FileReadResult{Err: err}
	}

	defer file.Close()

	var sb strings.Builder
	scanner := bufio.NewScanner(file)
	//scanner.Split(bufio.ScanLines)
	lines := 0

	for scanner.Scan() {
		lines++
		sb.WriteString(scanner.Text())
		sb.WriteByte('\n')
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error scan: %v\n", err)
		return FileReadResult{Err: err}
	}

	return FileReadResult{
		Text:      sb.String(),
		LineCount: lines,
	}
}

func (r *FileReader) calcDelay(attempt int) time.Duration {
	delay := float64(r.baseDelay) * math.Pow(2, float64(attempt))
	return time.Duration(delay)
}

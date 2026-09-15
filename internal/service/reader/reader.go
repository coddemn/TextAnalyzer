package reader

import (
	"bufio"
	"log"
	"os"
	"strings"
)

type FileReadResult struct {
	Text      string
	LineCount int
	Err       error
}

type FileReader struct {
}

func NewFileReader() *FileReader {
	return &FileReader{}
}

func (r *FileReader) ReadWithStats(path string) FileReadResult {

	file, err := os.Open(path)

	if err != nil {
		log.Fatalf("Error opening file: %v", err)
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

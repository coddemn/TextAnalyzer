package reader

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/coddemn/TextAnalyzer/internal/metrics"
)

func TestMain(m *testing.M) {
	metrics.Init()
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Вспомогательные функции
// ---------------------------------------------------------------------------

// createTempFile создаёт временный файл с заданным содержимым и возвращает его путь.
// Файл автоматически удаляется через t.Cleanup после завершения теста.
func createTempFile(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "testfile.txt")

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	return path
}

// ---------------------------------------------------------------------------
// Тесты ReadWithStats
// ---------------------------------------------------------------------------

func TestReadWithStats_SingleFile(t *testing.T) {
	content := "Hello, World!\nSecond line.\nThird line.\n"
	path := createTempFile(t, content)

	reader := NewFileReader(3, 10*time.Millisecond)
	result := reader.ReadWithStats(path)

	if result.Err != nil {
		t.Fatalf("expected no error, got: %v", result.Err)
	}
	if result.Text != content {
		t.Errorf("expected text %q, got %q", content, result.Text)
	}
	if result.LineCount != 3 {
		t.Errorf("expected 3 lines, got %d", result.LineCount)
	}
}

func TestReadWithStats_MultipleFiles(t *testing.T) {
	files := []struct {
		name     string
		content  string
		expected int // ожидаемое количество строк
	}{
		{"file1", "line A\nline B\n", 2},
		{"file2", "single line\n", 1},
		{"file3", "one\ntwo\nthree\nfour\n", 4},
		{"file4", "", 0}, // пустой файл — 0 строк
	}

	reader := NewFileReader(3, 10*time.Millisecond)

	for _, f := range files {
		t.Run(f.name, func(t *testing.T) {
			path := createTempFile(t, f.content)
			result := reader.ReadWithStats(path)

			if result.Err != nil {
				t.Fatalf("expected no error, got: %v", result.Err)
			}
			if result.LineCount != f.expected {
				t.Errorf("file %s: expected %d lines, got %d",
					f.name, f.expected, result.LineCount)
			}
		})
	}
}

func TestReadWithStats_NonExistentFile(t *testing.T) {
	reader := NewFileReader(3, 10*time.Millisecond)
	result := reader.ReadWithStats("/nonexistent/path/to/file.txt")

	if result.Err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
	if result.Text != "" {
		t.Errorf("expected empty text on error, got %q", result.Text)
	}
	if result.LineCount != 0 {
		t.Errorf("expected 0 lines on error, got %d", result.LineCount)
	}
}

func TestReadWithStats_EmptyPath(t *testing.T) {
	reader := NewFileReader(3, 10*time.Millisecond)
	result := reader.ReadWithStats("")

	if result.Err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
}

// ---------------------------------------------------------------------------
// Тесты ReadWithRetry
// ---------------------------------------------------------------------------

func TestReadWithRetry_CountAttempts(t *testing.T) {
	attemptsCount := 0
	failuresNeeded := 2

	spyOpen := func(path string) (*os.File, error) {
		attemptsCount++ // Считаем вызов

		if attemptsCount <= failuresNeeded {
			return nil, errors.New("simulated temporary error")
		}

		tmpFile := createTempFile(t, "success content\n")
		f, _ := os.Open(tmpFile)
		return f, nil
	}

	reader := NewReaderForTests(3, 10*time.Millisecond, spyOpen)

	result := reader.ReadWithRetry("/fake/path.txt")

	// Проверки

	// А. Результат должен быть успешным (ретраи сработали)
	if result.Err != nil {
		t.Fatalf("Expected success after retries, got: %v", result.Err)
	}

	// Б. Проверяем счетчик
	// Ожидаем: 2 ошибки + 1 успех = 3 вызова
	expectedAttempts := failuresNeeded + 1
	if attemptsCount != expectedAttempts {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, attemptsCount)
	}

}

// ---------------------------------------------------------------------------
// Тесты calcDelay
// ---------------------------------------------------------------------------

func TestCalcDelay(t *testing.T) {
	tests := []struct {
		name      string
		baseDelay time.Duration
		attempt   int
		expected  time.Duration
	}{
		{"attempt_0", 10 * time.Millisecond, 0, 10 * time.Millisecond},
		{"attempt_1", 10 * time.Millisecond, 1, 20 * time.Millisecond},
		{"attempt_2", 10 * time.Millisecond, 2, 40 * time.Millisecond},
		{"attempt_3", 10 * time.Millisecond, 3, 80 * time.Millisecond},
		{"attempt_4", 1 * time.Millisecond, 4, 16 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewFileReader(5, tt.baseDelay)
			got := reader.calcDelay(tt.attempt)

			if got != tt.expected {
				t.Errorf("calcDelay(%d) = %v, expected %v",
					tt.attempt, got, tt.expected)
			}
		})
	}
}

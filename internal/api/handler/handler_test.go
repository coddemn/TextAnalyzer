package api

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/coddemn/TextAnalyzer/internal/api/dto"
	"github.com/coddemn/TextAnalyzer/internal/domain"
)

func startMockWorker(t *testing.T, jobs <-chan dto.JobRequest) func() {
	t.Helper()

	done := make(chan struct{})

	go func() {
		for {
			select {
			case job := <-jobs:
				if strings.Contains(job.FilePath, "error") {
					job.Result <- domain.AnalysisResult{
						FilePath: job.FilePath,
						Err:      fmt.Errorf("mock analysis error"),
					}
					continue
				}

				job.Result <- domain.AnalysisResult{
					FilePath:      job.FilePath,
					Lines:         2,
					Symbols:       20,
					Sentences:     1,
					WordCount:     5,
					AvgWordLength: 4.0,
					LongestWord: domain.Word{
						Text:     "hello",
						Length:   5,
						Quantity: 1,
					},
					TopFrequents: []*domain.Word{
						{Text: "hello", Length: 5, Quantity: 1},
						{Text: "world", Length: 5, Quantity: 1},
					},
				}
			case <-done:
				return
			}
		}
	}()

	return func() { close(done) }
}

func setupRouter(t *testing.T, topN, maxFiles int) (*gin.Engine, chan dto.JobRequest) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	jobs := make(chan dto.JobRequest, 10)
	handler := NewHandler(jobs, topN, maxFiles)

	r := gin.New()
	r.POST("/analyze/upload", handler.AnalyzeUploadFile)
	r.POST("/analyze/multiple", handler.AnalyzeMultipleFiles)
	r.POST("/analyze", handler.AnalyzeFile)
	r.GET("/health", handler.Health)
	r.GET("/metrics", handler.Metrics)

	return r, jobs
}

func createMultipartBody(t *testing.T, fieldName, fileName, content string) (*strings.Reader, string) {
	t.Helper()

	var buf strings.Builder
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	_, err = io.Copy(part, strings.NewReader(content))
	if err != nil {
		t.Fatalf("failed to write file content: %v", err)
	}
	writer.Close()

	reader := strings.NewReader(buf.String())
	return reader, writer.FormDataContentType()
}

func createMultipartBodyMulti(t *testing.T, fieldName string, files []struct{ Name, Content string }) (*strings.Reader, string) {
	t.Helper()

	var buf strings.Builder
	writer := multipart.NewWriter(&buf)

	for _, f := range files {
		part, err := writer.CreateFormFile(fieldName, f.Name)
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		_, err = io.Copy(part, strings.NewReader(f.Content))
		if err != nil {
			t.Fatalf("failed to write file content: %v", err)
		}
	}
	writer.Close()

	reader := strings.NewReader(buf.String())
	return reader, writer.FormDataContentType()
}

// ---------------------------------------------------------------------------
// Health
// ---------------------------------------------------------------------------

func TestHealth(t *testing.T) {
	r, jobs := setupRouter(t, 10, 10)
	stop := startMockWorker(t, jobs)
	defer stop()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected 'ok', got %q", body["status"])
	}
}

// ---------------------------------------------------------------------------
// AnalyzeUploadFile (POST /analyze/upload)
// ---------------------------------------------------------------------------

func TestAnalyzeUploadFile_Success(t *testing.T) {
	r, jobs := setupRouter(t, 10, 10)
	stop := startMockWorker(t, jobs)
	defer stop()

	body, contentType := createMultipartBody(t, "file", "test.txt", "Hello\nWorld\n")
	req := httptest.NewRequest(http.MethodPost, "/analyze/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var result domain.AnalysisResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result.Err != nil {
		t.Errorf("expected no error, got: %v", result.Err)
	}
	if result.Lines != 2 {
		t.Errorf("expected 2 lines, got %d", result.Lines)
	}
	if result.WordCount != 5 {
		t.Errorf("expected 5 words, got %d", result.WordCount)
	}
}

func TestAnalyzeUploadFile_NoFile(t *testing.T) {
	r, jobs := setupRouter(t, 10, 10)
	stop := startMockWorker(t, jobs)
	defer stop()

	req := httptest.NewRequest(http.MethodPost, "/analyze/upload", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if body["error"] != "no file provided" {
		t.Errorf("expected 'no file provided', got %q", body["error"])
	}
}

func TestAnalyzeUploadFile_WorkerError(t *testing.T) {
	r, jobs := setupRouter(t, 10, 10)
	stop := startMockWorker(t, jobs)
	defer stop()

	reqBody, contentType := createMultipartBody(t, "file", "error_file.txt", "content\n")
	req := httptest.NewRequest(http.MethodPost, "/analyze/upload", reqBody)
	req.Header.Set("Content-Type", contentType)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 1. Проверяем статус
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d, body: %s", w.Code, w.Body.String())
	}

	// 2. Распаковка в map, так как ответ сервера: {"error": "mock analysis error"}
	var respBody map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &respBody); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	// 3. Проверяем сообщение об ошибке
	expectedErr := "mock analysis error"
	if respBody["error"] != expectedErr {
		t.Errorf("expected error message %q, got %q", expectedErr, respBody["error"])
	}
}

// ---------------------------------------------------------------------------
// AnalyzeMultipleFiles (POST /analyze/multiple)
// ---------------------------------------------------------------------------

func TestAnalyzeMultipleFiles_Success(t *testing.T) {
	r, jobs := setupRouter(t, 10, 10)
	stop := startMockWorker(t, jobs)
	defer stop()

	files := []struct{ Name, Content string }{
		{"file1.txt", "line one\nline two\n"},
		{"file2.txt", "another file\n"},
		{"file3.txt", "third\nfile\nhere\n"},
	}

	body, contentType := createMultipartBodyMulti(t, "file", files)
	req := httptest.NewRequest(http.MethodPost, "/analyze/multiple", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var results []domain.AnalysisResult
	if err := json.Unmarshal(w.Body.Bytes(), &results); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for i, res := range results {
		if res.Err != nil {
			t.Errorf("result %d: expected no error, got %v", i, res.Err)
		}
		if res.Lines != 2 {
			t.Errorf("result %d: expected 2 lines, got %d", i, res.Lines)
		}
		if res.WordCount != 5 {
			t.Errorf("result %d: expected 5 words, got %d", i, res.WordCount)
		}
	}
}

func TestAnalyzeMultipleFiles_NoFiles(t *testing.T) {
	r, jobs := setupRouter(t, 10, 10)
	stop := startMockWorker(t, jobs)
	defer stop()

	var buf strings.Builder
	writer := multipart.NewWriter(&buf)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/analyze/multiple", strings.NewReader(buf.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if body["error"] != "no files provided" {
		t.Errorf("expected 'no files provided', got %q", body["error"])
	}
}

func TestAnalyzeMultipleFiles_ExceedMaxFiles(t *testing.T) {
	r, jobs := setupRouter(t, 10, 2)
	stop := startMockWorker(t, jobs)
	defer stop()

	files := []struct{ Name, Content string }{
		{"file1.txt", "a\n"},
		{"file2.txt", "b\n"},
		{"file3.txt", "c\n"},
	}

	body, contentType := createMultipartBodyMulti(t, "file", files)
	req := httptest.NewRequest(http.MethodPost, "/analyze/multiple", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var respBody map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &respBody); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !strings.Contains(respBody["error"], "max") {
		t.Errorf("expected error about max files, got %q", respBody["error"])
	}
}

// ---------------------------------------------------------------------------
// Metrics (GET /metrics)
// ---------------------------------------------------------------------------

func TestMetrics(t *testing.T) {
	r, jobs := setupRouter(t, 10, 10)
	stop := startMockWorker(t, jobs)
	defer stop()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.Len() == 0 {
		t.Error("expected non-empty metrics body")
	}
}

package reader

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/coddemn/TextAnalyzer/internal/api/dto"
	api "github.com/coddemn/TextAnalyzer/internal/api/handler"
	"github.com/gin-gonic/gin"
)

func TestAnalyzeFile_NoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jobs := make(chan dto.JobRequest, 1) // можно сделать буферизированным
	handler := api.NewHandler(jobs, 10, 10)

	r := gin.New()
	r.POST("/analyze/upload", handler.AnalyzeUploadFile)

	req := httptest.NewRequest(http.MethodPost, "/analyze/upload", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAnalyzeFile_WithFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jobs := make(chan dto.JobRequest, 1)
	handler := api.NewHandler(jobs, 10, 10)

	r := gin.New()
	r.POST("/analyze/upload", handler.AnalyzeUploadFile)

	// Создаём временный файл с тестовым текстом
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	testText := "Hello world. This is a test. Another sentence."
	if _, err := tmpFile.WriteString(testText); err != nil {
		t.Fatal(err)
	}

	file, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest("POST", "/analyze/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

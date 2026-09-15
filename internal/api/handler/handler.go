package api

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/coddemn/TextAnalyzer/internal/api/dto"
	"github.com/coddemn/TextAnalyzer/internal/domain"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	jobs     chan dto.JobRequest
	topN     int
	maxFiles int
}

func NewHandler(jobs chan dto.JobRequest, topN int, maxFiles int) *Handler {
	return &Handler{
		jobs:     jobs,
		topN:     topN,
		maxFiles: maxFiles,
	}
}

func (h *Handler) AnalyzeUploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}

	res := h.processSingleFile(file)

	if res.Err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, res)
}

func (h *Handler) AnalyzeMultipleFiles(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil || form == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse multipart form"})
		return
	}

	files := form.File["file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files provided"})
		return
	}

	if len(files) > h.maxFiles {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("max %d files allowed", h.maxFiles)})
		return
	}

	var wg sync.WaitGroup
	results := make([]domain.AnalysisResult, 0, len(files))
	mu := sync.Mutex{}

	for _, fileHeader := range files {
		currentFile := fileHeader

		wg.Add(1)
		go func(f *multipart.FileHeader) {
			defer wg.Done()

			res := h.processSingleFile(f)

			mu.Lock()
			results = append(results, res)
			mu.Unlock()
		}(currentFile)
	}

	wg.Wait()

	c.JSON(http.StatusOK, results)
}

// AnalyzeFile — POST /analyze?file=path
func (h *Handler) AnalyzeFile(c *gin.Context) {
	filePath := c.Query("file")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file parameter is required"})
		return
	}

	resultChan := make(chan domain.AnalysisResult, 1)

	h.jobs <- dto.JobRequest{
		FilePath: filePath,
		Result:   resultChan,
	}

	res := <-resultChan

	if res.Err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, res)
}

func (h *Handler) processSingleFile(f *multipart.FileHeader) domain.AnalysisResult {
	src, err := f.Open()
	if err != nil {
		return domain.AnalysisResult{
			FilePath: f.Filename,
			Err:      fmt.Errorf("open file %q: %w", f.Filename, err),
		}
	}
	defer src.Close()

	tmpPath := filepath.Join(os.TempDir(), f.Filename)
	dst, err := os.Create(tmpPath)
	if err != nil {
		return domain.AnalysisResult{
			FilePath: f.Filename,
			Err:      fmt.Errorf("create temp file %q: %w", tmpPath, err),
		}
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		_ = os.Remove(tmpPath)
		return domain.AnalysisResult{
			FilePath: f.Filename,
			Err:      fmt.Errorf("copy file %q: %w", f.Filename, err),
		}
	}
	dst.Close() // закрываем до отправки в воркер

	resultChan := make(chan domain.AnalysisResult, 1)
	h.jobs <- dto.JobRequest{
		FilePath: tmpPath,
		Result:   resultChan,
	}

	res := <-resultChan
	_ = os.Remove(tmpPath) // удаляем временный файл

	res.FilePath = f.Filename // нормализуем имя в ответе
	return res
}

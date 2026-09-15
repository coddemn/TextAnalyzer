package api

import (
	"net/http"

	"github.com/coddemn/TextAnalyzer/internal/api/dto"
	"github.com/coddemn/TextAnalyzer/internal/domain"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	jobs chan dto.JobRequest
	topN int
}

func NewHandler(jobs chan dto.JobRequest, topN int) *Handler {
	return &Handler{
		jobs: jobs,
		topN: topN,
	}
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

	c.JSON(http.StatusOK, res)
}

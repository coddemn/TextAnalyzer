package dto

import "github.com/coddemn/TextAnalyzer/internal/domain"

type JobRequest struct {
	FilePath string
	Result   chan domain.AnalysisResult
}

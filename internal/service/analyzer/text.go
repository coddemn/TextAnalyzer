package analyzer

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/coddemn/TextAnalyzer/internal/domain"
)

type Analyzer struct {
}

func New() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) Run(filePath, text string, topN int, lineCount int) domain.AnalysisResult {
	symbols := a.CountSymbols(text)
	sentences := a.CountSentences(text)
	words := a.SplitWords(text)

	wordCount := a.CountWords(words)
	avgLen := a.AverageLength(words)
	longest := a.LongestWord(words)
	topFreq := a.TopFrequent(words, topN)

	return domain.AnalysisResult{
		FilePath:      filePath,
		Lines:         lineCount,
		Symbols:       symbols,
		Sentences:     sentences,
		WordCount:     wordCount,
		AvgWordLength: avgLen,
		LongestWord:   *longest,
		TopFrequents:  topFreq,
	}
}

func (a *Analyzer) CountSymbols(text string) int {
	return len([]rune(text))
}

func (a *Analyzer) CountSentences(text string) int {
	sentences := regexp.MustCompile(`[.!?]+`).Split(text, -1)
	count := 0

	for _, s := range sentences {
		if strings.TrimSpace(s) != "" {
			count++
		}
	}
	return count
}

// TODO: Debug counting word - words tipes word-test is one, but now two
func (a *Analyzer) SplitWords(text string) []*domain.Word {
	words := regexp.MustCompile(`[^\p{L}\p{N}]+`).Split(text, -1) // bug
	result := make([]*domain.Word, 0)

	wordMap := make(map[string]int)
	for _, w := range words {
		if strings.TrimSpace(w) == "" {
			continue
		}

		if idx, ok := wordMap[w]; ok {
			result[idx].Quantity++
		} else {
			newIdx := len(result)
			wordMap[w] = newIdx
			result = append(result, &domain.Word{Text: strings.ToLower(w), Length: utf8.RuneCountInString(w), Quantity: 1})
		}
	}

	return result
}

// -Подсчет символов в тексте;
// -Подсчет предложений в тексте;
// -Разделение текста по отдельным словам;

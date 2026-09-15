package analyzer

import (
	"sort"

	"github.com/coddemn/TextAnalyzer/internal/domain"
)

func (a *Analyzer) CountWords(words []*domain.Word) int {
	total := 0
	for _, w := range words {
		total += w.Quantity
	}
	return total
}

func (a *Analyzer) AverageLength(words []*domain.Word) float64 {

	if len(words) == 0 {
		return 0
	}

	countLetters := 0
	countWords := 0

	for _, w := range words {
		countLetters += w.Length * w.Quantity
		countWords += w.Quantity
	}
	//return int(math.Ceil(float64(total) / float64(len(words))))
	return float64(countLetters) / float64(countWords)
}

func (a *Analyzer) LongestWord(words []*domain.Word) *domain.Word {

	if len(words) == 0 {
		return &domain.Word{}
	}

	longest := words[0]
	for _, w := range words {
		if w.Length > longest.Length {
			longest = w
		}
	}

	return longest
}

func (a *Analyzer) TopFrequent(words []*domain.Word, qty int) []*domain.Word {
	sort.Slice(words, func(i, j int) bool {
		return words[i].Quantity > words[j].Quantity
	})

	if qty > len(words) {
		qty = len(words)
	}

	return words[:qty]
}

// -Подсчет количества слов в тексте;
// -Расчет средней длины слов в тексте;
// -Вычисление самого длинного слова;
// -Вывод топ частых слов;

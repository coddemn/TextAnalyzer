package analyzer

import (
	"testing"

	"github.com/coddemn/TextAnalyzer/internal/domain"
)

// func TestLongestWord(t *testing.T) {

// }

func TestCountSentences(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{"empty", "", 0},
		{"one sentence", "Hello world.", 1},
		{"three sentences", "Hello. How are you? I am fine!", 3},
		{"no ending punctuation", "Hello world", 1},
	}

	a := New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := a.CountSentences(tt.text)
			if got != tt.want {
				t.Errorf("CountSentences() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTopFrequent(t *testing.T) {
	a := New()
	words := []*domain.Word{
		{Text: "go", Length: 2, Quantity: 6},
		{Text: "Send", Length: 4, Quantity: 2},
		{Text: "I", Length: 1, Quantity: 1},
		{Text: "rust", Length: 4, Quantity: 1},
		{Text: "python", Length: 6, Quantity: 3},
	}

	top := a.TopFrequent(words, 3)

	if len(top) != 3 {
		t.Fatalf("expected 3 words, got %d", len(top))
	}
	if top[0].Text != "go" || top[0].Quantity != 6 {
		t.Errorf("expected go:6, got %s:%d", top[0].Text, top[0].Quantity)
	}
	if top[1].Text != "python" || top[1].Quantity != 3 {
		t.Errorf("expected python:3, got %s:%d", top[1].Text, top[1].Quantity)
	}
	if top[2].Text != "Send" || top[2].Quantity != 2 {
		t.Errorf("expected Send:2, got %s:%d", top[2].Text, top[2].Quantity)
	}
}

func TestCountWords(t *testing.T) {
	a := New()
	words := []*domain.Word{
		{Text: "go", Length: 2, Quantity: 6},
		{Text: "Send", Length: 4, Quantity: 2},
		{Text: "I", Length: 1, Quantity: 1},
		{Text: "rust", Length: 4, Quantity: 1},
		{Text: "python", Length: 6, Quantity: 3},
	}
	want := 6 + 2 + 1 + 1 + 3
	got := a.CountWords(words)

	if got != want {
		t.Fatalf("expected %d count words, got %d", want, got)
	}
}

func TestCountWords_Zero(t *testing.T) {
	a := New()
	words := []*domain.Word{}
	want := 0

	got := a.CountWords(words)

	if got != want {
		t.Fatalf("expected %d count words, got %d", want, got)
	}
}

func TestAverageLength(t *testing.T) {
	a := New()
	words := []*domain.Word{
		{Text: "go", Length: 2, Quantity: 6},
		{Text: "Send", Length: 4, Quantity: 2},
		{Text: "I", Length: 1, Quantity: 1},
		{Text: "rust", Length: 4, Quantity: 1},
		{Text: "python", Length: 6, Quantity: 3},
	}
	want := float64(12+8+1+4+18) / float64(6+2+1+1+3)
	got := a.AverageLength(words)

	if got != want {
		t.Fatalf("expected %.2f avg, got %.2f", want, got)
	}
}

func TestLongestWord(t *testing.T) {
	a := New()
	words := []*domain.Word{
		{Text: "go", Length: 2, Quantity: 6},
		{Text: "Send", Length: 4, Quantity: 2},
		{Text: "I", Length: 1, Quantity: 1},
		{Text: "rust", Length: 4, Quantity: 1},
		{Text: "python", Length: 6, Quantity: 3},
	}

	longest := a.LongestWord(words)

	if longest.Text != "python" {
		t.Fatalf("expected \"python\" words, got %v", longest)
	}
}

func TestSplitWords(t *testing.T) {
	a := New()
	text := "Hi, my name is... Hi!"

	want := []domain.Word{
		{Text: "hi", Length: 2, Quantity: 2},
		{Text: "my", Length: 2, Quantity: 1},
		{Text: "name", Length: 4, Quantity: 1},
		{Text: "is", Length: 2, Quantity: 1},
	}

	words := a.SplitWords(text)

	if len(words) != 4 {
		t.Fatalf("expected 4 words, got %d", len(words))
	}
	for i, w := range words {
		if w.Text != want[i].Text && w.Quantity != want[i].Quantity {
			t.Fatalf("expected 4 words: \n\twant: %v\n\tgot %v", want, words)
		}

	}
}

func TestCountSymbols(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{"empty", "", 0},
		{"twenty one symbs", "Hi, my name is... Hi!", 21},
		{"sticker (31 symb)", "Hello. How are you? I am fine!🎉", 31},
		{"line break (11symb)", "Hello\nworld", 11},
	}

	a := New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := a.CountSymbols(tt.text)
			if got != tt.want {
				t.Errorf("CountSymbols() = %d, want %d", got, tt.want)
			}
		})
	}
}

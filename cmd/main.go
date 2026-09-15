package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/coddemn/TextAnalyzer/internal/service/analyzer"
	"github.com/coddemn/TextAnalyzer/internal/service/reader"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Ошибка: Не указано имя текстового файла")
		os.Exit(1)
	}

	filePath := os.Args

	r := reader.NewFileReader()
	a := analyzer.New()

	readResult := r.ReadWithStats(filePath[1])
	if readResult.Err != nil {
		log.Fatal(readResult.Err)
	}

	res := a.Run(filePath[1], readResult.Text, 5, readResult.LineCount)

	// Для компактного вывода (без отступов)
	data, err := json.Marshal(res)
	if err != nil {
		fmt.Println("Ошибка маршалинга:", err)
		return
	}
	fmt.Println(string(data))

	fmt.Println()
	fmt.Println()

	// Для читаемого вывода с отступами
	dataIndent, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		fmt.Println("Ошибка маршалинга:", err)
		return
	}
	fmt.Println(string(dataIndent))

	fmt.Println()
	fmt.Println()

	// В консоль
	fmt.Printf("=== %s ===\n", res.FilePath)
	fmt.Printf("Строк: %d\n", res.Lines)
	fmt.Printf("Символов: %d\n", res.Symbols)
	fmt.Printf("Предложений: %d\n", res.Sentences)
	fmt.Printf("Слов: %d\n", res.WordCount)
	fmt.Printf("Средняя длина слова: %.2f\n", res.AvgWordLength)
	fmt.Printf("Самое длинное слово: %s - %d\n", res.LongestWord.Text, res.LongestWord.Length)
	fmt.Println("Топ частых слов:")
	for _, w := range res.TopFrequents {
		fmt.Printf("  %s: %d\n", w.Text, w.Quantity)
	}
	fmt.Println()
}

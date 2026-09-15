package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/coddemn/TextAnalyzer/internal/domain"
	"github.com/coddemn/TextAnalyzer/internal/service/analyzer"
	"github.com/coddemn/TextAnalyzer/internal/service/reader"
)

func main() {

	// if len(os.Args) < 2 {
	// 	fmt.Fprintln(os.Stderr, "Ошибка: Не указано имя текстового файла")
	// 	os.Exit(1)
	// }

	// filePath := os.Args

	files := []string{
		"../example.txt",
		"../ex2.txt",
	}

	workerCount := 4
	topN := 5

	jobs := make(chan string, len(files))
	results := make(chan domain.AnalysisResult, len(files))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	r := reader.NewFileReader()
	a := analyzer.New()

	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for filePath := range jobs { // wait new jobs while chan is open
				select {
				case <-ctx.Done():
					return
				default:
				}

				readRes := r.ReadWithStats(filePath)
				var res domain.AnalysisResult
				res.FilePath = filePath
				if readRes.Err != nil {
					res.Err = fmt.Errorf("Worker %d: read error: %w", workerID, readRes.Err)

					results <- res
					continue
				}

				res = a.Run(filePath, readRes.Text, topN, readRes.LineCount)
				results <- res
			}
		}(w)
	}

	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	var allResults []domain.AnalysisResult
	for i := 0; i < len(files); i++ {
		res := <-results
		allResults = append(allResults, res)
	}

	for _, res := range allResults {
		if res.Err != nil {
			log.Printf("[%s] ERROR: %v\n", res.FilePath, res.Err)
			continue
		}

		// Для читаемого json с отступами
		dataIndent, err := json.MarshalIndent(res, "", " ")
		if err != nil {
			fmt.Println("Ошибка маршалинга:", err)
			return
		}
		fmt.Println(string(dataIndent))

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
		fmt.Println()

	}

	// Graceful shutdown (ctrl+c)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Given shutdown signal. Stopping pool...")
		cancel()
	}()

	// wait workers is finalized
	wg.Wait()
	log.Println("All workers is completed. App correct stopped.")

}

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
	"time"

	"github.com/coddemn/TextAnalyzer/internal/domain"
	"github.com/coddemn/TextAnalyzer/internal/metrics"
	"github.com/coddemn/TextAnalyzer/internal/service/analyzer"
	"github.com/coddemn/TextAnalyzer/internal/service/reader"
	"github.com/prometheus/client_golang/prometheus"
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

	// config
	workerCount := 4
	topN := 5
	maxRetries := 3
	retryDelay := 100 * time.Millisecond

	// initialyze
	metrics.Init()
	r := reader.NewFileReader(maxRetries, retryDelay)
	a := analyzer.New()

	// channels
	jobs := make(chan string, len(files))
	results := make(chan domain.AnalysisResult, len(files))

	// worker pool
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		metrics.ActiveWorkers.Inc()

		go func(workerID int) {
			defer wg.Done()
			defer metrics.ActiveWorkers.Dec()

			for filePath := range jobs { // wait new jobs while chan is open
				select {
				case <-ctx.Done():
					return
				default:
				}

				timer := prometheus.NewTimer(metrics.ProcessDuration)
				res := processFile(r, a, filePath, topN, workerID)
				timer.ObserveDuration()

				if res.Err != nil {
					metrics.FilesProcessed.WithLabelValues("error").Inc()
				} else {
					metrics.FilesProcessed.WithLabelValues("ok").Inc()
				}

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

func processFile(
	r *reader.FileReader,
	a *analyzer.Analyzer,
	filePath string,
	topN int,
	workerID int,
) domain.AnalysisResult {
	res := r.ReadWithRetry(filePath)
	if res.Err != nil {
		return domain.AnalysisResult{
			FilePath: filePath,
			Err:      fmt.Errorf("Worker %d: %w", workerID, res.Err),
		}
	}

	return a.Run(filePath, res.Text, topN, res.LineCount)
}

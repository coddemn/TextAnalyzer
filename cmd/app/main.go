package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/coddemn/TextAnalyzer/internal/api/dto"
	api "github.com/coddemn/TextAnalyzer/internal/api/handler"
	"github.com/coddemn/TextAnalyzer/internal/config"
	"github.com/coddemn/TextAnalyzer/internal/domain"
	"github.com/coddemn/TextAnalyzer/internal/metrics"
	"github.com/coddemn/TextAnalyzer/internal/service/analyzer"
	"github.com/coddemn/TextAnalyzer/internal/service/reader"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	_ = godotenv.Load(".env")

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml" // default
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// if len(os.Args) < 2 {
	// 	fmt.Fprintln(os.Stderr, "Ошибка: Не указано имя текстового файла")
	// 	os.Exit(1)
	// }

	// filePath := os.Args

	// files := []string{
	// 	"../example.txt",
	// 	"../ex2.txt",
	// }

	// config
	workerCount := cfg.App.WorkerCount
	topN := cfg.App.TopN
	maxFiles := cfg.App.MaxFiles
	maxRetries := cfg.App.MaxRetries
	retryDelay := time.Duration(cfg.App.RetryDelay) * time.Millisecond
	httpAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)

	// initialyze
	metrics.Init()
	r := reader.NewFileReader(maxRetries, retryDelay)
	a := analyzer.New()

	// channels
	jobs := make(chan dto.JobRequest, 100)
	//results := make(chan domain.AnalysisResult, len(files))

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

			for job := range jobs { // wait new jobs while chan is open
				select {
				case <-ctx.Done():
					return
				default:
				}

				timer := prometheus.NewTimer(metrics.ProcessDuration)
				res := processFile(r, a, job.FilePath, topN, workerID)
				timer.ObserveDuration()

				if res.Err != nil {
					metrics.FilesProcessed.WithLabelValues("error").Inc()
				} else {
					metrics.FilesProcessed.WithLabelValues("ok").Inc()
				}

				job.Result <- res
			}
		}(w)
	}

	// for _, f := range files {
	// 	jobs <- f
	// }
	// close(jobs)

	// var allResults []domain.AnalysisResult
	// for i := 0; i < len(files); i++ {
	// 	res := <-results
	// 	allResults = append(allResults, res)
	// }

	// for _, res := range allResults {
	// 	if res.Err != nil {
	// 		log.Printf("[%s] ERROR: %v\n", res.FilePath, res.Err)
	// 		continue
	// 	}

	// 	// Для читаемого json с отступами
	// 	dataIndent, err := json.MarshalIndent(res, "", " ")
	// 	if err != nil {
	// 		fmt.Println("Ошибка маршалинга:", err)
	// 		return
	// 	}
	// 	fmt.Println(string(dataIndent))

	// 	fmt.Println()

	// 	// В консоль
	// 	fmt.Printf("=== %s ===\n", res.FilePath)
	// 	fmt.Printf("Строк: %d\n", res.Lines)
	// 	fmt.Printf("Символов: %d\n", res.Symbols)
	// 	fmt.Printf("Предложений: %d\n", res.Sentences)
	// 	fmt.Printf("Слов: %d\n", res.WordCount)
	// 	fmt.Printf("Средняя длина слова: %.2f\n", res.AvgWordLength)
	// 	fmt.Printf("Самое длинное слово: %s - %d\n", res.LongestWord.Text, res.LongestWord.Length)
	// 	fmt.Println("Топ частых слов:")
	// 	for _, w := range res.TopFrequents {
	// 		fmt.Printf("  %s: %d\n", w.Text, w.Quantity)
	// 	}
	// 	fmt.Println()
	// 	fmt.Println()

	// }

	//gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	//router.Use(gin.Recovery())

	handler := api.NewHandler(jobs, topN, maxFiles)
	router.POST("/analyze", handler.AnalyzeFile)
	router.POST("/analyze/upload", handler.AnalyzeUploadFile)
	router.POST("/analyze/multiple", handler.AnalyzeMultipleFiles)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "It`s API from Gin, Demidos!",
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: router,
	}

	go func() {
		log.Printf("HTTP server on %s", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}

	}()

	// Graceful shutdown (ctrl+c)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	log.Printf("Given signal %v. Stopping work...\n", sig)

	shtdCtx, shtdCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shtdCancel()

	if err := httpServer.Shutdown(shtdCtx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	close(jobs)

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

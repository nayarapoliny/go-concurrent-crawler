package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/nayarapoliny/go-concurrent-crawler/internal/crawler"
)

func main() {
	seedURL := flag.String("url", "https://go.dev", "URL inicial para iniciar o crawling")
	maxDepth := flag.Int("depth", 1, "Profundidade máxima de navegação")
	workers := flag.Int("workers", 5, "Número máximo de goroutines simultâneas")
	timeout := flag.Duration("timeout", 30*time.Second, "Tempo limite global da execução")
	sameHost := flag.Bool("same-host", true, "Restringir navegação apenas ao mesmo domínio")
	outputFile := flag.String("out", "report.json", "Ficheiro JSON de saída com os resultados")
	flag.Parse()

	fmt.Printf("🚀 A iniciar Web Crawler Concorrente...\n")
	fmt.Printf("🔗 Alvo: %s | Profundidade: %d | Workers: %d | Timeout: %v\n\n", *seedURL, *maxDepth, *workers, *timeout)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	cfg := crawler.Config{
		MaxDepth:    *maxDepth,
		Concurrency: *workers,
		SameHost:    *sameHost,
	}

	engine := crawler.New(httpClient, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	start := time.Now()
	results, err := engine.Crawl(ctx, *seedURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Erro fatal: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Crawling finalizado em %v! Total de páginas processadas: %d\n\n", time.Since(start).Round(time.Millisecond), len(results))

	for _, r := range results {
		status := fmt.Sprintf("%d", r.StatusCode)
		if r.Error != "" {
			status = "ERRO"
		}
		fmt.Printf("[Depth %d] [%s] (%dms) %s -> %d links (%s)\n", r.Depth, status, r.DurationMs, r.URL, r.LinksCount, r.Title)
	}

	if *outputFile != "" {
		if err := saveReport(*outputFile, results); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️ Erro ao guardar relatório JSON: %v\n", err)
		} else {
			fmt.Printf("\n📄 Relatório completo exportado para: %s\n", *outputFile)
		}
	}
}

func saveReport(filename string, results []crawler.PageResult) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

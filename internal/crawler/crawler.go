package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type PageResult struct {
	URL        string    `json:"url"`
	Title      string    `json:"title"`
	StatusCode int       `json:"status_code"`
	LinksCount int       `json:"links_count"`
	Links      []string  `json:"links"`
	DurationMs int64     `json:"duration_ms"`
	Depth      int       `json:"depth"`
	Error      string    `json:"error,omitempty"`
	CrawledAt  time.Time `json:"crawled_at"`
}

type Config struct {
	MaxDepth    int
	Concurrency int
	SameHost    bool
	UserAgent   string
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Crawler struct {
	client  HTTPClient
	config  Config
	visited map[string]bool
	mu      sync.RWMutex
	results []PageResult
	resMu   sync.Mutex
	sem     chan struct{}
}

func New(client HTTPClient, cfg Config) *Crawler {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 5
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "GoConcurrentCrawler/1.0"
	}
	return &Crawler{
		client:  client,
		config:  cfg,
		visited: make(map[string]bool),
		results: make([]PageResult, 0),
		sem:     make(chan struct{}, cfg.Concurrency),
	}
}

func (c *Crawler) Crawl(ctx context.Context, seedURL string) ([]PageResult, error) {
	parsedSeed, err := url.Parse(seedURL)
	if err != nil || (parsedSeed.Scheme != "http" && parsedSeed.Scheme != "https") {
		return nil, fmt.Errorf("URL inicial inválida: %s", seedURL)
	}

	var wg sync.WaitGroup

	var crawlRecursive func(targetURL string, depth int)
	crawlRecursive = func(targetURL string, depth int) {
		defer wg.Done()

		if depth > c.config.MaxDepth {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		select {
		case c.sem <- struct{}{}:
			defer func() { <-c.sem }()
		case <-ctx.Done():
			return
		}

		result := c.fetchAndParse(ctx, targetURL, depth)
		c.addResult(result)

		if result.Error != "" || depth == c.config.MaxDepth {
			return
		}

		for _, link := range result.Links {
			if c.shouldVisit(link, parsedSeed.Host) {
				wg.Add(1)
				go crawlRecursive(link, depth+1)
			}
		}
	}

	c.markVisited(seedURL)
	wg.Add(1)
	go crawlRecursive(seedURL, 0)

	wg.Wait()
	return c.results, nil
}

func (c *Crawler) fetchAndParse(ctx context.Context, targetURL string, depth int) PageResult {
	start := time.Now()
	res := PageResult{
		URL:       targetURL,
		Depth:     depth,
		CrawledAt: start.UTC(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	req.Header.Set("User-Agent", c.config.UserAgent)

	resp, err := c.client.Do(req)
	res.DurationMs = time.Since(start).Milliseconds()
	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer resp.Body.Close()

	res.StatusCode = resp.StatusCode
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		res.Error = fmt.Sprintf("status HTTP inesperado: %d", resp.StatusCode)
		return res
	}

	baseURL, _ := url.Parse(targetURL)
	title, links, err := ParseHTML(resp.Body, baseURL)
	if err != nil {
		res.Error = fmt.Sprintf("erro ao processar HTML: %v", err)
		return res
	}

	res.Title = title
	res.Links = links
	res.LinksCount = len(links)
	return res
}

func (c *Crawler) shouldVisit(rawURL, baseHost string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	if c.config.SameHost && parsed.Host != baseHost {
		return false
	}
	return c.markVisited(rawURL)
}

func (c *Crawler) markVisited(targetURL string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.visited[targetURL] {
		return false
	}
	c.visited[targetURL] = true
	return true
}

func (c *Crawler) addResult(r PageResult) {
	c.resMu.Lock()
	defer c.resMu.Unlock()
	c.results = append(c.results, r)
}

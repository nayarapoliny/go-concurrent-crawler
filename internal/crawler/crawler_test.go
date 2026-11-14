package crawler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type mockHTTPClient struct {
	pages map[string]string
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	htmlContent, exists := m.pages[req.URL.String()]
	if !exists {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("Not Found")),
		}, nil
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(htmlContent)),
	}, nil
}

func TestCrawler_CrawlSuccessAndDeduplication(t *testing.T) {
	mockPages := map[string]string{
		"https://example.com": `
			<html>
				<head><title>Home Page</title></head>
				<body>
					<a href="/about">About</a>
					<a href="https://example.com/contact">Contact</a>
					<a href="https://external.com/page">External</a>
				</body>
			</html>`,
		"https://example.com/about": `
			<html>
				<head><title>About Us</title></head>
				<body>
					<a href="/">Home</a>
				</body>
			</html>`,
		"https://example.com/contact": `
			<html>
				<head><title>Contact</title></head>
				<body></body>
			</html>`,
	}

	client := &mockHTTPClient{pages: mockPages}
	cfg := Config{
		MaxDepth:    1,
		Concurrency: 4,
		SameHost:    true,
	}

	c := New(client, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results, err := c.Crawl(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("esperava nenhum erro, obteve: %v", err)
	}

	// Deve visitar apenas example.com, /about e /contact (ignorando external.com e evitando loop infinito com /)
	if len(results) != 3 {
		t.Errorf("esperava 3 páginas visitadas, obteve %d", len(results))
	}
}

func TestCrawler_InvalidSeedURL(t *testing.T) {
	client := &mockHTTPClient{}
	c := New(client, Config{MaxDepth: 1, Concurrency: 2})

	_, err := c.Crawl(context.Background(), "ftp://invalid-url")
	if err == nil {
		t.Fatal("esperava erro para URL com esquema inválido, mas obteve nil")
	}
}

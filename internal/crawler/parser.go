package crawler

import (
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func ParseHTML(body io.Reader, baseURL *url.URL) (string, []string, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return "", nil, err
	}

	var title string
	var links []string
	seen := make(map[string]bool)

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "title" && n.FirstChild != nil && title == "" {
				title = strings.TrimSpace(n.FirstChild.Data)
			}
			if n.Data == "a" {
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						resolved := resolveURL(baseURL, attr.Val)
						if resolved != "" && !seen[resolved] {
							seen[resolved] = true
							links = append(links, resolved)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return title, links, nil
}

func resolveURL(base *url.URL, href string) string {
	href = strings.TrimSpace(href)
	if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "javascript:") {
		return ""
	}

	parsed, err := url.Parse(href)
	if err != nil {
		return ""
	}

	resolved := base.ResolveReference(parsed)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}

	resolved.Fragment = ""
	return resolved.String()
}

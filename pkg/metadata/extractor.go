package metadata

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Metadata struct {
	Title       string
	Description string
	Keywords    string
	URL         string
}

func ExtractFromURL(url string) (*Metadata, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Warning: metadata fetch failed for %s: %v", url, err)
		return &Metadata{
			Title: url,
			URL:   url,
		}, nil
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read and parse HTML
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	// Extract metadata
	meta := &Metadata{URL: url}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if n.FirstChild != nil {
					meta.Title = strings.TrimSpace(n.FirstChild.Data)
				}
			case "meta":
				var name, content string
				for _, attr := range n.Attr {
					switch attr.Key {
					case "name", "property":
						name = strings.ToLower(attr.Val)
					case "content":
						content = attr.Val
					}
				}
				switch name {
				case "description", "og:description":
					if meta.Description == "" {
						meta.Description = content
					}
				case "keywords":
					meta.Keywords = content
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	// If no title found in title tag, try extracting from h1
	if meta.Title == "" {
		meta.Title = extractFirstH1(doc)
	}

	// If still no title, use URL
	if meta.Title == "" {
		meta.Title = url
	}

	return meta, nil
}

func extractFirstH1(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "h1" {
		var sb strings.Builder
		extractText(n, &sb)
		return strings.TrimSpace(sb.String())
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := extractFirstH1(c); result != "" {
			return result
		}
	}
	return ""
}

func extractText(n *html.Node, sb *strings.Builder) {
	if n.Type == html.TextNode {
		sb.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, sb)
	}
}

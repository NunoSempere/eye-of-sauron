package dsca

import (
	"encoding/xml"
	"fmt"
	"git.nunosempere.com/NunoSempere/news/lib/types"
	"git.nunosempere.com/NunoSempere/news/lib/web"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const feedURL = "https://www.dsca.mil/DesktopModules/ArticleCS/RSS.ashx?ContentType=700&Site=1509&isdashboardselected=0&max=20"

// getClient returns an http.Client with reasonable timeout settings
func getClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

// makeRequest creates a new request with browser-like headers
func makeRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add headers to look more like a browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	return req, nil
}

func GetArticleContent(url string) (string, error) {
	client := getClient()
	req, err := makeRequest(url)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching article content: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Got non-200 status code: %d", resp.StatusCode)
		return "", fmt.Errorf("got status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	// Find the article body div
	articleBody := doc.Find("div.body[itemprop='articleBody']")
	if articleBody.Length() == 0 {
		// Try alternative selectors if the main one fails
		articleBody = doc.Find(".article-body, .content-body, main article")
		if articleBody.Length() == 0 {
			return "", nil
		}
	}

	// Clean up the HTML and return it
	html, err := articleBody.Html()
	if err != nil {
		return "", err
	}

	cleanHtml, err := web.CompressHtml(html)
	if err != nil {
		return "", err
	}

	return cleanHtml, nil
}

func FetchFeed() ([]types.Source, error) {
	log.Printf("Fetching DSCA feed: %s", feedURL)
	
	client := getClient()
	req, err := makeRequest(feedURL)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching DSCA feed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Got non-200 status code from DSCA feed: %d", resp.StatusCode)
		return nil, fmt.Errorf("got status code %d from feed", resp.StatusCode)
	}

	xml_bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, err
	}

	var feed RSSFeed
	err = xml.Unmarshal(xml_bytes, &feed)
	if err != nil {
		log.Printf("Error unmarshaling XML: %v\n", err)
		return nil, err
	}

	log.Printf("Found %d items in feed", len(feed.Channel.Items))

	var sources []types.Source
	for _, item := range feed.Channel.Items {
		log.Printf("Processing item with date: %s", item.PubDate)
		
		// Try to parse the date
		var pubDate time.Time
		formats := []string{
			time.RFC1123,
			time.RFC1123Z,
			"Mon, 02 Jan 2006 15:04:05 MST",
			"Mon, 2 Jan 2006 15:04:05 MST",
			"Mon, 02 Jan 2006 15:04:05 -0700",
			"2006-01-02T15:04:05-07:00",
		}

		parsed := false
		for _, format := range formats {
			if pd, err := time.Parse(format, item.PubDate); err == nil {
				pubDate = pd
				parsed = true
				break
			}
		}

		if !parsed {
			log.Printf("Could not parse date '%s' with any known format, using current time", item.PubDate)
			pubDate = time.Now()
		}

		// Skip articles older than 24 hours
		age := time.Since(pubDate)
		log.Printf("Article age: %v", age)
		if age > 24*time.Hour {
			log.Printf("Skipping article older than 24 hours")
			continue
		}

		// Clean up the title
		title := strings.TrimSpace(item.Title)

		sources = append(sources, types.Source{
			Title:  title,
			Link:   item.Link,
			Date:   pubDate.Format(time.RFC3339),
			Origin: "DSCA",
		})
	}

	return sources, nil
}

package readability

import (
	"errors"
	"git.nunosempere.com/NunoSempere/news/lib/web"
	"log"
	"net/url"
	"net/http"
	"os/exec"
	"strings"
	"time"
	"github.com/PuerkitoBio/goquery"
)

// LynxDump extracts text content from a URL using the lynx browser
func LynxDump(url string) (string, error) {
	cmd := exec.Command("lynx", "-dump", "-nolist", url)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	result := strings.TrimSpace(string(output))
	if len(result) < 50 {
		return "", errors.New("lynx output too short")
	}
	
	return result, nil
}

func GetReadabilityOutput(article_url string) (string, error) {
	// Helper function to try lynx fallback
	lynxFallback := func(reason string, originalErr error) (string, error) {
		lynx_result, lynx_err := LynxDump(article_url)
		if lynx_err != nil {
			if originalErr != nil {
				log.Printf("%s: %v, lynx also failed: %v", reason, originalErr, lynx_err)
				return "", errors.Join(originalErr, lynx_err)
			} else {
				log.Printf("%s, lynx also failed: %v", reason, lynx_err)
				return "", lynx_err
			}
		}
		log.Printf("%s, using lynx fallback", reason)
		return lynx_result, nil
	}

	readability_url := "https://trastos.nunosempere.com/readability?url=" + article_url // url must start with https
	readability_response, err := web.Get(readability_url)
	if err != nil {
		return lynxFallback("Readability service failed", err)
	}
	readability_result := string(readability_response)

	if len(readability_result) < 200 {
		return lynxFallback("Readability output too short", nil)
	}
	return readability_result, nil
}

func ReplaceWithOSFrontend(u string) (string, error) {
	// Parse the URL
	parsedURL, err := url.Parse(u)
	if err != nil {
		log.Printf("Error parsing url to check for open source alternatives: %v", err)
		return "", err
	}

	// Check if the host is bad_domain.com
	if parsedURL.Host == "www.reuters.com" {
		// Replace the host with open_source_alternative.net
		parsedURL.Host = "neuters.de"

		// Return the updated URL as a string
		return parsedURL.String(), nil
	}
	if parsedURL.Host == "x.com" {
		// Replace the host with open_source_alternative.net
		parsedURL.Host = "nitter.net"

		// Return the updated URL as a string
		return parsedURL.String(), nil
	}
	// Return the original URL if no replacement is necessary
	return u, nil
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

// Try to extract title from HTML
func ExtractTitle(url string) string {
	url_for_title := url 
	oss_url, err := ReplaceWithOSFrontend(url)
	if err == nil {
		url_for_title = oss_url
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := makeRequest(url_for_title)
	if err != nil {
		return ""
	}

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ""
	}

	title := doc.Find("title").Text()
	return strings.TrimSpace(title)
}

func GetArticleContent(init_url string) (string, error) {
	req_url := init_url
	os_url, err0 := ReplaceWithOSFrontend(init_url)
	if err0 == nil {
		req_url = os_url
	}
	readable_text, err1 := GetReadabilityOutput(req_url)
	log.Printf("Req url: %v", req_url)
	if err1 != nil {

		url_content, err2 := web.Get(req_url)
		if err2 != nil {
			log.Println("Errors in both redability AND web.Get")
			err := errors.Join(err1, err2)
			return "", err
		}
		compressed_html, err2 := web.CompressHtml(string(url_content[:]))
		if err2 != nil {
			log.Println("Errors in both redability AND web.Get")
			err := errors.Join(err1, err2)
			return "", err
		}
		return compressed_html, nil

	}
	return readable_text, nil
}

/*
func main() {
	url := "https://www.washingtonpost.com/nation/2024/02/29/ukraine-support-alabama-political-divide/"
	readable_content, err := getReadabilityOutput(url)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(readable_content)
	}

	url = "https://www.vox.com/future-perfect/2024/2/13/24070864/samotsvety-forecasting-superforecasters-tetlock"
	readable_content, err = getReadabilityOutput(url)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(readable_content)
	}
}
*/

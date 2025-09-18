package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/url"
	"time"

	"git.nunosempere.com/NunoSempere/news/lib/types"
	"git.nunosempere.com/NunoSempere/news/lib/web"
)

type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Entries []Entry  `xml:"entry"`
}

type Entry struct {
	Title   string `xml:"title"`
	Link    link   `xml:"link"`
	PubDate string `xml:"published"`
}

type link struct {
	Val string `xml:",cdata"`
	Url string `xml:"href,attr"`
}

func KeywordToRSSFeed(keyword string) (string, error) {

	feedMap := map[string]string{
		"War":                  "https://www.google.com/alerts/feeds/12823167512648692611/14775069330880237129",
		"Emergency":            "https://www.google.com/alerts/feeds/12823167512648692611/10752650238419774131",
		"disaster":             "https://www.google.com/alerts/feeds/12823167512648692611/1731039193900529474",
		"alert":                "https://www.google.com/alerts/feeds/12823167512648692611/1731039193900529512",
		"nuclear":              "https://www.google.com/alerts/feeds/12823167512648692611/16302398974188618783",
		"human-to-human":       "https://www.google.com/alerts/feeds/12823167512648692611/16363046595406729950",
		"pandemic":             "https://www.google.com/alerts/feeds/12823167512648692611/1689941676777225519",
		"blockade":             "https://www.google.com/alerts/feeds/12823167512648692611/8592384102868889996",
		"invasion":             "https://www.google.com/alerts/feeds/12823167512648692611/14400132195257773260",
		"undersea cables":      "https://www.google.com/alerts/feeds/12823167512648692611/16660286502183277886",
		"Carrington event":     "https://www.google.com/alerts/feeds/12823167512648692611/17032681478781817561",
		"mystery pneumonia":    "https://www.google.com/alerts/feeds/12823167512648692611/17032681478781820157",
		"China Taiwan":         "https://www.google.com/alerts/feeds/12823167512648692611/3055804732710246461",
		"Russia Ukraine":       "https://www.google.com/alerts/feeds/12823167512648692611/16094280695893389744",
		"OpenAI announces AGI": "https://www.google.com/alerts/feeds/12823167512648692611/500375972710348852",
		"military exercise":    "https://www.google.com/alerts/feeds/12823167512648692611/267614087142809738",
		"Kessler syndrome":     "https://www.google.com/alerts/feeds/12823167512648692611/14873084661553437561",
		"Cyberattack":          "https://www.google.com/alerts/feeds/12823167512648692611/4267352864131660551",
	}

	if identifier, exists := feedMap[keyword]; exists && identifier != "" {
		return identifier, nil
	}

	return "", fmt.Errorf("identifier not found in alerts map")
}

func extractActualLink(encodedURL string) (string, error) {

	/*
	   https://www.google.com/url?rct=j&sa=t&url=https://www.fema.gov/press-release/20241010/fema-responds-hurricane-milton-florida-it-continues-coordinated-recovery&ct=ga&cd=CAIyGmRmNmMwNjc2YmM0NzgxYTg6Y29tOmVuOlVT&usg=AOvVaw0D3F1Eq05brlgUtbYGPCcl

	   into

	   https://www.fema.gov/press-release/20241010/fema-responds-hurricane-milton-florida-it-continues-coordinated-recovery
	*/
	parsedURL, err := url.Parse(encodedURL)
	if err != nil {
		return "", err
	}

	queryParams := parsedURL.Query()
	targetURL := queryParams.Get("url")

	if targetURL == "" {
		return "", fmt.Errorf("url parameter not found")
	}

	return targetURL, nil
}

func SearchGoogleAlerts(query string) ([]types.Source, error) {
	log.Printf("Making google alerts request for query: %s", query)

	url, err := KeywordToRSSFeed(query)
	if err != nil {
		return nil, err
	}

	xml_bytes, err := web.Get(url)
	if err != nil {
		return nil, err
	}

	var feed Feed
	err = xml.Unmarshal(xml_bytes, &feed)
	if err != nil {
		log.Printf("Error unmarshaling XML: %v\n", err)
		return nil, err
	}

	var sources []types.Source
	for _, entry := range feed.Entries {
		actual_link, err := extractActualLink(entry.Link.Url)
		if err != nil {
			log.Printf("Error parsing url parameter from %v", entry.Link.Url)
			continue
		}
		date, err := time.Parse("2006-01-02T15:04:05Z", entry.PubDate)
		if err != nil {
			date = time.Now()
		}
		sources = append(sources, types.Source{Title: entry.Title, Link: actual_link, Date: date})
	}

	return sources, nil
}

func TestGoogleAlerts() {
	keywords := []string{"War", "Emergency", "disaster"}
	for _, keyword := range keywords {
		log.Printf("Testing Google Alerts for keyword: %s", keyword)
		sources, err := SearchGoogleAlerts(keyword)
		if err != nil {
			log.Printf("Error searching Google Alerts for %s: %v", keyword, err)
			continue
		}
		log.Printf("Found %d sources for keyword %s:", len(sources), keyword)
		for _, source := range sources {
			log.Printf("Title: %s", source.Title)
			log.Printf("Link: %s", source.Link)
			log.Printf("Date: %s", source.Date)
			log.Println("---")
		}
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"git.nunosempere.com/NunoSempere/news/lib/types"
)

type TweetsAPIResponse struct {
	Success bool `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Tweets []Tweet `json:"tweets"`
		Count  int    `json:"count"`
	} `json:"data"`
}

type Tweet struct {
	TweetID   string `json:"tweet_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
	Username  string `json:"username"`
}

// FetchSources retrieves tweets from xAI and elonmusk accounts and groups them into weekly articles
func FetchSources() ([]types.Source, error) {
	// Fetch tweets from both accounts
	xaiTweets, err := fetchTweetsFromAccount("xAI")
	if err != nil {
		return nil, fmt.Errorf("error fetching xAI tweets: %v", err)
	}

	elonTweets, err := fetchTweetsFromAccount("elonmusk")
	if err != nil {
		return nil, fmt.Errorf("error fetching elonmusk tweets: %v", err)
	}

	// Group tweets by week and create articles
	sources := []types.Source{}

	// Create weekly articles for xAI tweets
	xaiWeeklyArticles := groupTweetsByWeek(xaiTweets, "xAI")
	sources = append(sources, xaiWeeklyArticles...)

	// Create weekly articles for Elon tweets
	elonWeeklyArticles := groupTweetsByWeek(elonTweets, "elonmusk")
	sources = append(sources, elonWeeklyArticles...)

	return sources, nil
}

// fetchTweetsFromAccount fetches tweets from a specific account
func fetchTweetsFromAccount(account string) ([]Tweet, error) {
	url := fmt.Sprintf("https://tweets.nunosempere.com/api/tweets/%s?limit=1000", account)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp TweetsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("API returned error: %s", apiResp.Message)
	}

	return apiResp.Data.Tweets, nil
}

// groupTweetsByWeek groups tweets by week and creates articles for recent weeks only
func groupTweetsByWeek(tweets []Tweet, account string) []types.Source {
	// Parse tweets and group by week
	weeklyTweets := make(map[string][]Tweet)

	for _, tweet := range tweets {
		// Parse the tweet date
		tweetTime, err := time.Parse(time.RFC3339, tweet.CreatedAt)
		if err != nil {
			continue // Skip malformed dates
		}

		// Only process tweets from the last 4 weeks
		if time.Since(tweetTime) > 4*7*24*time.Hour {
			continue
		}

		// Get week key (year-week format)
		year, week := tweetTime.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%02d", year, week)

		weeklyTweets[weekKey] = append(weeklyTweets[weekKey], tweet)
	}

	// Create sources for each week that has tweets
	var sources []types.Source
	for weekKey, weekTweets := range weeklyTweets {
		if len(weekTweets) == 0 {
			continue
		}

		// Sort tweets by date (oldest first)
		sort.Slice(weekTweets, func(i, j int) bool {
			time1, _ := time.Parse(time.RFC3339, weekTweets[i].CreatedAt)
			time2, _ := time.Parse(time.RFC3339, weekTweets[j].CreatedAt)
			return time1.Before(time2)
		})

		// Create article content from tweets
		articleContent := createArticleFromTweets(weekTweets, account, weekKey)

		// Use the first tweet's date as the article date
		firstTweetTime, _ := time.Parse(time.RFC3339, weekTweets[0].CreatedAt)

		source := types.Source{
			Title:  fmt.Sprintf("%s Tweets - Week %s", account, weekKey),
			Link:   fmt.Sprintf("https://twitter.com/%s/week/%s", account, weekKey), // Virtual link
			Date:   firstTweetTime,
			Origin: fmt.Sprintf("%s-tweets", account),
		}

		// Store the article content in a way that can be accessed later
		// We'll need to modify the Source struct or use a different approach
		// For now, we'll store it in the title temporarily
		source.Title = articleContent

		sources = append(sources, source)
	}

	return sources
}

// createArticleFromTweets creates a readable article format from a week's tweets
func createArticleFromTweets(tweets []Tweet, account string, weekKey string) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# %s Tweets - Week %s\n\n", account, weekKey))
	content.WriteString(fmt.Sprintf("Collection of tweets from @%s for week %s\n\n", account, weekKey))

	for i, tweet := range tweets {
		tweetTime, _ := time.Parse(time.RFC3339, tweet.CreatedAt)
		content.WriteString(fmt.Sprintf("## Tweet %d - %s\n\n", i+1, tweetTime.Format("Jan 2, 2006 15:04 UTC")))
		content.WriteString(tweet.Text)
		content.WriteString("\n\n")
		content.WriteString(fmt.Sprintf("[Original Tweet](https://twitter.com/%s/status/%s)\n\n", account, tweet.TweetID))
		content.WriteString("---\n\n")
	}

	return content.String()
}
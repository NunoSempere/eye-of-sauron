package search

// BraveSearchResponse represents the response from Brave Search API
type BraveSearchResponse struct {
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			URL        string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

package main

type HNResponse struct {
	Hits        []HNHit `json:"hits"`
	NbHits      int     `json:"nbHits"`      // Total number of hits
	Page        int     `json:"page"`        // Current page
	NbPages     int     `json:"nbPages"`     // Total number of pages
	HitsPerPage int     `json:"hitsPerPage"` // Number of hits per page
}

type HNHit struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	CreatedAt  string `json:"created_at"`
	ObjectID   string `json:"objectID"`
	StoryText  string `json:"story_text"` // Text content for Ask HN posts
	Points     int    `json:"points"`
	NumComments int   `json:"num_comments"`
}

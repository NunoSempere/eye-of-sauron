package hn

type HNResponse struct {
	Hits []HNHit `json:"hits"`
}

type HNHit struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
	ObjectID  string `json:"objectID"`
}

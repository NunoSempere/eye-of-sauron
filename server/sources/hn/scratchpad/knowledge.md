# HackerNews Integration

## API Endpoint
Using the Algolia HN Search API:

The endpoint format is:
`http://hn.algolia.com/api/v1/search_by_date?tags=story<Y&page=P`&numericFilters=created_at_i>X,created_at_i

Where:
- X and Y are Unix timestamps (seconds since epoch)
- `tags=story` filters for stories only (not comments)
- `created_at_i` is the creation timestamp
- `page=P` for pagination (0-based)

## Implementation Details
- Fetches stories from the last hour across all pages
- Filters out low engagement posts (<2 points or comments)
- Skips items without URLs or story text
- Uses optimized filtering pipeline:
  1. Checks title importance first
  2. Only fetches full article if title shows potential importance
  3. Processes story_text for text-only posts
- Saves to output/potpourri/hn/

## Response Format
```json
{
  "hits": [
    {
      "title": "Story title",
      "url": "https://example.com",
      "created_at": "2024-01-10T15:04:05Z",
      "objectID": "12345",
      "story_text": "Hey HN! here is a short explanation of our project",
      "points": 10,
      "num_comments": 5
    }
  ],
  "nbHits": 100,
  "page": 0,
  "nbPages": 5,
  "hitsPerPage": 20
}
```

## Service
Runs as a systemd service (hn.service) checking for new stories every hour.

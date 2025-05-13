# HackerNews Integration

## API Endpoint
Using the Algolia HN Search API:

The endpoint format is:
`http://hn.algolia.com/api/v1/search_by_date?tags=story<Y`

Where&numericFilters=created_at_i>X,created_at_i:
- X and Y are Unix timestamps (seconds since epoch)
- `tags=story` filters for stories only (not comments)
- `created_at_i` is the creation timestamp

## Implementation Details
- Fetches stories from the last hour
- Skips items without URLs (text-only posts)
- Uses standard filtering and importance checking pipeline
- Saves to output/potpourri/hn/

## Response Format
```json
{
  "hits": [
    {
      "title": "Story title",
      "url": "https://example.com",
      "created_at": "2024-01-10T15:04:05Z",
      "objectID": "12345"
    }
  ]
}
```

## Service
Runs as a systemd service (hn.service) checking for new stories every hour.

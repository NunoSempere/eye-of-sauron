# Anthropic News Fetcher

This fetcher monitors Anthropic's news feed through RSShub and processes articles for storage in both the main sources database and the AI-specific sources database.

## Source

- **RSS Feed**: https://rsshub.app/anthropic/news
- **Original**: Anthropic news articles

## Usage

```bash
# Run once
make run

# Watch logs
make listen

# Install as systemd service
make systemd
```

## Database Storage

- **sources-ai**: All articles are saved here
- **sources**: Only articles that pass importance filters are saved here

## Processing Pipeline

1. Fetch articles from RSShub Anthropic feed
2. Check for duplicates
3. Validate host acceptability
4. Extract and clean content
5. Generate summary using LLM
6. Check existential importance
7. Save to appropriate databases

## Dependencies

- Go modules from the main project
- Database connection via `DATABASE_POOL_URL`
- OpenAI API key for LLM processing

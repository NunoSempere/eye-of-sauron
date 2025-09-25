# xAI Tweets Source

This source fetches tweets from xAI and elonmusk accounts and processes them into weekly articles.

## Overview

The source fetches up to 1000 tweets from each account using the endpoints:
- `https://tweets.nunosempere.com/api/tweets/xAI?limit=1000`
- `https://tweets.nunosempere.com/api/tweets/elonmusk?limit=1000`

Tweets are grouped by week (ISO week format) and converted into markdown articles. Only tweets from the last 4 weeks are processed.

## Features

- **Weekly Grouping**: Tweets are automatically grouped by ISO week (YYYY-WXX format)
- **Markdown Format**: Each weekly collection becomes a structured markdown article
- **Standard Pipeline**: Uses the same filtering and importance checking as other sources
- **Dual Account Support**: Processes both xAI and elonmusk accounts
- **Recent Focus**: Only processes tweets from the last 4 weeks

## Files

- `main.go`: Main entry point with weekly processing loop
- `fetch.go`: Tweet fetching and weekly grouping logic
- `filterAndExpandSource.go`: Custom filtering pipeline for tweet articles
- `makefile`: Build and deployment commands
- `xai.service`: Systemd service configuration

## Usage

```bash
# Build
make build

# Run locally (requires .env with OPENAI_KEY and DATABASE_POOL_URL)
make run

# Install as system service
make install
make start
```

## Output Format

Each weekly article includes:
- Title: `{Account} Tweets - Week {YYYY-WXX}`
- Structured markdown with individual tweets
- Tweet timestamps and links to originals
- Processed through standard importance checking

## Processing Schedule

The source runs continuously with a 1-week sleep cycle, making it efficient for weekly tweet collection.


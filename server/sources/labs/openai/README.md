# OpenAI News Source

This source fetches and processes news articles from OpenAI's official RSS feed.

## Overview

- **RSS Feed**: https://openai.com/news/rss.xml
- **Update Frequency**: Every 6 hours
- **Content**: Official OpenAI news, announcements, and updates

## Files

- **main.go**: Main entry point with processing loop
- **fetch.go**: RSS feed fetching and parsing logic
- **filterAndExpandSource.go**: Content filtering and LLM processing
- **saveSource.go**: Database persistence
- **makefile**: Build and service management commands
- **openai-news.service**: Systemd service configuration

## Setup

1. **Environment Variables**:
   - `OPENAI_KEY`: Required for LLM summarization and importance checking
   - `DATABASE_POOL_URL`: PostgreSQL connection string

2. **Test locally**:
   ```bash
   make run
   ```

3. **Install as service**:
   ```bash
   make systemd
   ```

4. **View logs**:
   ```bash
   make listen
   ```

## Features

- Parses OpenAI RSS feed every 6 hours
- Extracts article title, link, and publication date
- Uses readability extraction for full article content
- LLM-powered summarization and importance assessment
- Duplicate detection and filtering
- Automatic date format conversion

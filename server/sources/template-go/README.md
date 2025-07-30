# {{SOURCE_NAME}} Source Template

This template provides a standardized structure for adding new Go-based news sources to the eye-of-sauron system.

## Files Overview

- **main.go**: Main entry point with logging setup, environment loading, and processing loop
- **fetch.go**: Source-specific fetching logic (RSS, API, or web scraping)
- **filterAndExpandSource.go**: Common filtering and content processing pipeline
- **saveSource.go**: Database persistence functionality
- **makefile**: Build and service management commands
- **template.service**: Systemd service configuration template

## Setup Instructions

1. **Copy the template folder**:
   ```bash
   cp -r template-go your-source-name
   cd your-source-name
   ```

2. **Replace placeholders**:
   - `{{SOURCE_NAME}}`: Replace with your source name (e.g., "reuters", "bbc")
   - `{{RSS_URL}}`: Replace with RSS feed URL if applicable
   - `{{API_URL}}`: Replace with API endpoint if applicable  
   - `{{WEBPAGE_URL}}`: Replace with webpage URL if scraping
   - `{{SLEEP_DURATION}}`: Replace with appropriate sleep duration (e.g., "12 * time.Hour")

3. **Implement fetch logic**:
   - Edit `fetch.go` to implement `FetchSources()` function
   - Choose appropriate method: RSS, API, or web scraping
   - Define custom types if needed for API responses

4. **Customize filtering (optional)**:
   - Edit `filterAndExpandSource.go` if source-specific filtering is needed
   - Add freshness checks or custom validation logic

5. **Update service files**:
   - Rename `template.service` to `{source-name}.service`
   - Update makefile with correct service name

6. **Test the implementation**:
   ```bash
   make run
   ```

## Common Patterns

### RSS Feed Sources
- Use `fetchFromRSS()` function in `fetch.go`
- Handle common RSS date formats (RFC1123Z)
- Extract title, link, and publication date

### API Sources  
- Use `fetchFromAPI()` function in `fetch.go`
- Define custom response types
- Handle pagination if needed
- Add authentication headers if required

### Web Scraping Sources
- Use `fetchFromWebpage()` function in `fetch.go`
- Consider using HTML parsing libraries
- Be respectful of rate limits
- Handle dynamic content if needed

## Environment Variables Required

- `OPENAI_KEY`: For LLM summarization and importance checking
- `DATABASE_POOL_URL`: PostgreSQL connection string

## Service Management

```bash
# Test locally
make run

# Install as systemd service
make systemd

# View logs
make listen

# Rotate logs
make rotate
```

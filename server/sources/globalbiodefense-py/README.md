# {{SOURCE_NAME}} Python Source Template

This template provides a standardized Python structure for adding new news sources to the eye-of-sauron system, equivalent to the Go template but designed for Python developers.

## Files Overview

- **main.py**: Main entry point with logging setup, environment loading, and processing loop
- **fetch.py**: Source-specific fetching logic (RSS, API, or web scraping) with comprehensive examples
- **filter_and_expand.py**: Common filtering and content processing pipeline
- **save_source.py**: Database persistence functionality with PostgreSQL
- **requirements.txt**: Python package dependencies
- **Makefile**: Build, setup, and service management commands
- **template.service**: Systemd service configuration template
- **lib_stubs.py**: Placeholder implementations for Go library equivalents

## Quick Start

1. **Copy the template folder**:
   ```bash
   cp -r template-python your-source-name
   cd your-source-name
   ```

2. **Replace placeholders**:
   - `{{SOURCE_NAME}}`: Replace with your source name (e.g., "reuters", "bbc")
   - `{{RSS_URL}}`: Replace with RSS feed URL if applicable
   - `{{API_URL}}`: Replace with API endpoint if applicable  
   - `{{WEBPAGE_URL}}`: Replace with webpage URL if scraping
   - `{{SLEEP_SECONDS}}`: Replace with sleep duration in seconds (e.g., 43200 for 12 hours)

3. **Set up Python environment**:
   ```bash
   make setup  # Uses uv to create virtual environment and install dependencies
   ```

4. **Implement fetch logic**:
   - Edit `fetch.py` to implement `fetch_sources()` function
   - Choose appropriate method: RSS, API, or web scraping
   - Remove unused example functions

5. **Replace library stubs**:
   - Implement actual functionality in `lib_stubs.py` or replace with real libraries
   - See "Library Implementation" section below

6. **Test the implementation**:
   ```bash
   make run
   ```

## Library Implementation

The template includes `lib_stubs.py` with placeholder implementations. You'll need to implement:

### Filters Module
```python
def is_dupe(source, database_url):
    # Query database or use hash-based duplicate detection
    
def is_good_host(source):
    # Check allowlist/blocklist of domains
    
def clean_title(title):
    # Remove unwanted characters, normalize text
```

### Readability Module
```python
def extract_title(url):
    # Use newspaper3k, trafilatura, or BeautifulSoup
    
def get_article_content(url):
    # Extract main article content using readability algorithms
```

Recommended libraries:
- **newspaper3k**: `pip install newspaper3k`
- **trafilatura**: `pip install trafilatura` 
- **readability-lxml**: `pip install readability-lxml`

### LLM Module
```python
def summarize(content, openai_key):
    # Use OpenAI API or other LLM service
    
def check_existential_importance(snippet, openai_key):
    # Analyze content importance using LLM
```

## Fetch Patterns

### RSS Feed Sources
```python
def fetch_sources():
    return fetch_from_rss("https://example.com/rss")
```

### JSON API Sources  
```python
def fetch_sources():
    headers = {"Authorization": "Bearer YOUR_TOKEN"}
    return fetch_from_api("https://api.example.com/articles", headers)
```

### Web Scraping Sources
```python
def fetch_sources():
    return fetch_from_webpage("https://example.com/news")
```

## Environment Variables

Create a `.env` file with:
```bash
OPENAI_KEY=your_openai_api_key_here
DATABASE_POOL_URL=postgresql://user:pass@localhost/dbname
```

## Development Commands

```bash
# Setup environment (uses uv)
make setup

# Add new dependency
make add PACKAGE=requests

# Add development dependency  
make add-dev PACKAGE=pytest

# Run locally
make run

# Test database connection
make test-db

# Watch logs
make listen

# Format code
make format

# Lint code  
make lint

# Clean up
make clean
```

## Service Management

```bash
# Install as systemd service
make systemd

# Check service status
make status

# View service logs
make logs

# Rotate local logs
make rotate
```

## Service Configuration

Before installing as a service:

1. **Update service file**:
   - Rename `template.service` to `{source-name}.service`
   - Update paths and user in service file

2. **Set up environment**:
   - Ensure uv environment is set up: `make setup`
   - Verify database connectivity: `make test-db`

3. **Install service**:
   - Run `make systemd`
   - Service will auto-start and restart on failure

## Common Libraries

### Web Scraping
- **requests**: HTTP client (included)
- **beautifulsoup4**: HTML parsing (included)
- **selenium**: Browser automation (optional)
- **playwright**: Modern browser automation (optional)

### Content Processing  
- **feedparser**: RSS/Atom feed parsing (included)
- **newspaper3k**: Article extraction
- **trafilatura**: Web content extraction
- **python-dateutil**: Advanced date parsing

### Database
- **psycopg2**: PostgreSQL driver (included)
- **sqlalchemy**: ORM alternative

### Development
- **black**: Code formatting
- **flake8**: Linting
- **pytest**: Testing

## Error Handling

The template includes comprehensive error handling:
- Network timeouts and retries
- Database connection failures  
- LLM API failures
- Malformed data parsing
- Service restart on critical failures

## Logging

Logs are written to both console and `v2.log` file:
- Info level for normal operations
- Warning level for recoverable issues
- Error level for failures
- Debug level for detailed troubleshooting

Use `make listen` to watch logs in real-time or `make logs` for systemd journal.
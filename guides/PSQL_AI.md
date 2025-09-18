# PostgreSQL AI Table Setup

This guide covers the setup and management of the `sources-ai` table, which is used specifically for AI-related news sources like OpenAI news.

## Table Creation

### Initial Setup

The `sources-ai` table is created in the same database as the main `sources` table, using the existing `DATABASE_POOL_URL` connection.

## Schema Setup

### Create Sources-AI Table

The `sources-ai` table uses the same schema as the main `sources` table:

```sql
psql $DATABASE_POOL_URL -c "CREATE TABLE IF NOT EXISTS \"sources-ai\" (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    link TEXT NOT NULL UNIQUE,
    date TIMESTAMP NOT NULL,
    summary TEXT,
    importance_bool BOOLEAN,
    importance_reasoning TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed BOOLEAN DEFAULT FALSE,
    relevant_per_human_check TEXT DEFAULT 'maybe'
);"
```



## Database Management

### Connection Commands

```bash
# Connect to database
source .env
psql $DATABASE_POOL_URL
```

### Useful Queries

```bash
# View recent AI sources
psql $DATABASE_POOL_URL -c "SELECT title, link, date FROM \"sources-ai\" WHERE created_at > NOW() - INTERVAL '1 week' ORDER BY date DESC;"

# Count total AI sources
psql $DATABASE_POOL_URL -c "SELECT COUNT(*) FROM \"sources-ai\";"

# Export AI sources to CSV
psql $DATABASE_POOL_URL -c "COPY (SELECT title, link, date, summary FROM \"sources-ai\") TO STDOUT WITH CSV HEADER;"

# View unprocessed AI sources
psql $DATABASE_POOL_URL -c "SELECT id, title, link, date FROM \"sources-ai\" WHERE processed = false ORDER BY date ASC;"

# View important AI sources
psql $DATABASE_POOL_URL -c "SELECT title, link, importance_reasoning FROM \"sources-ai\" WHERE importance_bool = true ORDER BY date DESC;"
```

### Maintenance Commands

```bash
# Clear duplicate links (if any)
psql $DATABASE_POOL_URL -c "DELETE FROM \"sources-ai\" WHERE id NOT IN (SELECT MIN(id) FROM \"sources-ai\" GROUP BY link);"

# Reset processed flag for reprocessing
psql $DATABASE_POOL_URL -c "UPDATE \"sources-ai\" SET processed = false WHERE created_at > NOW() - INTERVAL '1 day';"

# Clear old entries (optional)
psql $DATABASE_POOL_URL -c "DELETE FROM \"sources-ai\" WHERE created_at < NOW() - INTERVAL '6 months';"
```

## Backup and Restore

### Backup

```bash
# Sources-AI table backup
psql $DATABASE_POOL_URL -c "COPY \"sources-ai\" TO STDOUT WITH CSV HEADER" > sources-ai-data-$(date +%Y%m%d).csv
```

### Restore

```bash
# Restore sources-ai table from CSV
psql $DATABASE_POOL_URL -c "COPY \"sources-ai\"(title,link,date,summary,importance_bool,importance_reasoning,created_at,processed,relevant_per_human_check) FROM STDIN WITH CSV HEADER" < sources-ai-data-YYYYMMDD.csv
```

## Separation Rationale

The `sources-ai` table is separate from the main `sources` table to:

1. **Isolate AI-specific content** for specialized processing and analysis
2. **Enable different retention policies** for AI vs. traditional news sources
3. **Facilitate AI-specific queries and analytics** without impacting main operations
4. **Allow independent scaling and optimization** for AI workloads
5. **Provide clear data governance** between human-curated and AI-generated content

## Monitoring

```bash
# Check table size
psql $DATABASE_POOL_URL -c "SELECT pg_size_pretty(pg_total_relation_size('\"sources-ai\"'));"

# Monitor recent activity in sources-ai table
psql $DATABASE_POOL_URL -c "SELECT DATE(created_at) as date, COUNT(*) as sources_added FROM \"sources-ai\" WHERE created_at > NOW() - INTERVAL '7 days' GROUP BY DATE(created_at) ORDER BY date;"

# Compare activity between main and AI tables
psql $DATABASE_POOL_URL -c "SELECT 'main' as table_name, COUNT(*) as total FROM sources UNION ALL SELECT 'ai' as table_name, COUNT(*) as total FROM \"sources-ai\";"
```

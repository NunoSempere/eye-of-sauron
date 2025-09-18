# PostgreSQL Database Management Makefile
# Run these commands from the guides/PSQL directory

.PHONY: help install setup connect backup clear-duplicates clear-duplicates-main clear-duplicates-ai reset-processed reset-processed-main reset-processed-ai clear-old clear-old-main clear-old-ai

# Load environment variables
include ../../server/.env
export

help:
	@echo "PostgreSQL Database Management Commands:"
	@echo ""
	@echo "Setup:"
	@echo "  install              Install PostgreSQL on Debian/Ubuntu"
	@echo "  setup                Create database, user, and all tables"
	@echo "  create-sources-table Create main sources table"
	@echo "  create-sources-ai-table Create AI sources table"
	@echo "  create-flags-table   Create flags table"
	@echo ""
	@echo "Connection:"
	@echo "  connect              Connect to database"
	@echo ""
	@echo "Queries:"
	@echo "  recent-sources       View recent main sources (1 week)"
	@echo "  recent-ai-sources    View recent AI sources (1 week)"
	@echo "  count-sources        Count total sources in both tables"
	@echo "  unprocessed          View unprocessed sources"
	@echo "  important            View important sources"
	@echo "  table-sizes          Check table sizes"
	@echo "  activity             Monitor recent activity"
	@echo ""
	@echo "Maintenance:"
	@echo "  clear-duplicates     Remove duplicate entries"
	@echo "  reset-processed      Reset processed flags for reprocessing"
	@echo "  clear-old            Remove entries older than 6 months"
	@echo "  terminate-connections Drop other database connections"
	@echo ""
	@echo "Backup/Export:"
	@echo "  backup               Backup both tables to CSV"
	@echo "  backup-sources       Backup main sources table"
	@echo "  backup-sources-ai    Backup AI sources table"
	@echo "  export-relevant      Export human-marked relevant sources"
	@echo "  export-all           Export all sources to CSV"
	@echo "  export-ai            Export all AI sources to CSV"

# Installation
install:
	sudo apt install postgresql postgresql-client

# Setup
setup: create-user create-database create-sources-table create-sources-ai-table create-flags-table
	@echo "Database setup complete!"

create-user:
	sudo -u postgres createuser nuno || true

create-database:
	sudo -u postgres createdb -O nuno nuno || true

create-sources-table:
	psql $$DATABASE_POOL_URL -c "CREATE TABLE IF NOT EXISTS sources (id SERIAL PRIMARY KEY, title TEXT NOT NULL, link TEXT NOT NULL UNIQUE, date TIMESTAMP NOT NULL, summary TEXT, importance_bool BOOLEAN, importance_reasoning TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, processed BOOLEAN DEFAULT FALSE, relevant_per_human_check TEXT DEFAULT 'maybe');"

create-sources-ai-table:
	psql $$DATABASE_POOL_URL -c "CREATE TABLE IF NOT EXISTS \"sources-ai\" (id SERIAL PRIMARY KEY, title TEXT NOT NULL, link TEXT NOT NULL UNIQUE, date TIMESTAMP NOT NULL, summary TEXT, importance_bool BOOLEAN, importance_reasoning TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, processed BOOLEAN DEFAULT FALSE, relevant_per_human_check TEXT DEFAULT 'maybe');"

create-flags-table:
	psql $$DATABASE_POOL_URL -c "CREATE TABLE IF NOT EXISTS flags (name VARCHAR(50) PRIMARY KEY, code INTEGER NOT NULL, msg TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);"

# Connection
connect:
	psql $$DATABASE_POOL_URL

# Query commands
recent-sources:
	psql $$DATABASE_POOL_URL -c "SELECT title, link, date FROM sources WHERE created_at > NOW() - INTERVAL '1 week' ORDER BY date DESC;"

recent-ai-sources:
	psql $$DATABASE_POOL_URL -c "SELECT title, link, date FROM \"sources-ai\" WHERE created_at > NOW() - INTERVAL '1 week' ORDER BY date DESC;"

count-sources:
	psql $$DATABASE_POOL_URL -c "SELECT 'main' as table_name, COUNT(*) as total FROM sources UNION ALL SELECT 'ai' as table_name, COUNT(*) as total FROM \"sources-ai\";"

unprocessed:
	psql $$DATABASE_POOL_URL -c "SELECT id, title, link, date FROM sources WHERE processed = false ORDER BY date ASC;"

unprocessed-ai:
	psql $$DATABASE_POOL_URL -c "SELECT id, title, link, date FROM \"sources-ai\" WHERE processed = false ORDER BY date ASC;"

important:
	psql $$DATABASE_POOL_URL -c "SELECT title, link, importance_reasoning FROM sources WHERE importance_bool = true ORDER BY date DESC;"

important-ai:
	psql $$DATABASE_POOL_URL -c "SELECT title, link, importance_reasoning FROM \"sources-ai\" WHERE importance_bool = true ORDER BY date DESC;"

table-sizes:
	psql $$DATABASE_POOL_URL -c "SELECT 'sources' as table_name, pg_size_pretty(pg_total_relation_size('sources')) as size UNION ALL SELECT 'sources-ai' as table_name, pg_size_pretty(pg_total_relation_size('\"sources-ai\"')) as size;"

activity:
	psql $$DATABASE_POOL_URL -c "SELECT DATE(created_at) as date, COUNT(*) as sources_added FROM sources WHERE created_at > NOW() - INTERVAL '7 days' GROUP BY DATE(created_at) ORDER BY date;"

activity-ai:
	psql $$DATABASE_POOL_URL -c "SELECT DATE(created_at) as date, COUNT(*) as sources_added FROM \"sources-ai\" WHERE created_at > NOW() - INTERVAL '7 days' GROUP BY DATE(created_at) ORDER BY date;"

# Maintenance commands
clear-duplicates: clear-duplicates-main clear-duplicates-ai

clear-duplicates-main:
	psql $DATABASE_POOL_URL -c "DELETE FROM sources WHERE id NOT IN (SELECT MIN(id) FROM sources GROUP BY link);"

clear-duplicates-ai:
	psql $DATABASE_POOL_URL -c "DELETE FROM \"sources-ai\" WHERE id NOT IN (SELECT MIN(id) FROM \"sources-ai\" GROUP BY link);"

reset-processed: reset-processed-main reset-processed-ai

reset-processed-main:
	psql $DATABASE_POOL_URL -c "UPDATE sources SET processed = false WHERE created_at > NOW() - INTERVAL '1 day';"

reset-processed-ai:
	psql $DATABASE_POOL_URL -c "UPDATE \"sources-ai\" SET processed = false WHERE created_at > NOW() - INTERVAL '1 day';"

clear-old: clear-old-main clear-old-ai

clear-old-main:
	psql $DATABASE_POOL_URL -c "DELETE FROM sources WHERE created_at < NOW() - INTERVAL '6 months';"

clear-old-ai:
	psql $DATABASE_POOL_URL -c "DELETE FROM \"sources-ai\" WHERE created_at < NOW() - INTERVAL '6 months';"

terminate-connections:
	psql $$DATABASE_POOL_URL -c "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE datname = current_database() AND pid <> pg_backend_pid();"

# Backup and export commands
backup: backup-sources backup-sources-ai
	@echo "Backup complete!"

backup-sources:
	psql $$DATABASE_POOL_URL -c "COPY sources TO STDOUT WITH CSV HEADER" > sources-backup-$$(date +%Y%m%d).csv
	@echo "Sources table backed up to sources-backup-$$(date +%Y%m%d).csv"

backup-sources-ai:
	psql $$DATABASE_POOL_URL -c "COPY \"sources-ai\" TO STDOUT WITH CSV HEADER" > sources-ai-backup-$$(date +%Y%m%d).csv
	@echo "Sources-AI table backed up to sources-ai-backup-$$(date +%Y%m%d).csv"

export-relevant:
	psql $$DATABASE_POOL_URL -c "COPY (SELECT title, link, date FROM sources WHERE relevant_per_human_check = 'yes') TO STDOUT WITH CSV HEADER;" > relevant-sources-$$(date +%Y%m%d).csv
	@echo "Relevant sources exported to relevant-sources-$$(date +%Y%m%d).csv"

export-all:
	psql $$DATABASE_POOL_URL -c "COPY (SELECT title, link, date, summary FROM sources) TO STDOUT WITH CSV HEADER;" > all-sources-$$(date +%Y%m%d).csv
	@echo "All sources exported to all-sources-$$(date +%Y%m%d).csv"

export-ai:
	psql $$DATABASE_POOL_URL -c "COPY (SELECT title, link, date, summary FROM \"sources-ai\") TO STDOUT WITH CSV HEADER;" > ai-sources-$$(date +%Y%m%d).csv
	@echo "AI sources exported to ai-sources-$$(date +%Y%m%d).csv"

# Flag management
clear-openai-flag:
	psql $$DATABASE_POOL_URL -c "UPDATE flags SET code = 0, msg = 'Flag cleared', updated_at = CURRENT_TIMESTAMP WHERE name = 'openai_refill';"

show-flags:
	psql $$DATABASE_POOL_URL -c "SELECT * FROM flags ORDER BY updated_at DESC;"

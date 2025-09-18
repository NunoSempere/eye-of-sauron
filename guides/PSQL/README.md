# PostgreSQL Database Management

This directory contains PostgreSQL database setup, management, and utility commands for the eye-of-sauron project.

## Quick Start

```bash
# Run makefile commands from this directory
cd guides/PSQL

# Install PostgreSQL
make install

# Set up database and tables
make setup

# Connect to database
make connect
```

## Overview

The project uses PostgreSQL with two main tables:
- `sources`: Main news sources table
- `sources-ai`: AI-specific sources (OpenAI, etc.)

## Environment Variables

Required in your `.env` file:
```
DATABASE_POOL_URL=postgresql://doadmin:...
```

## Available Commands

Use `make <command>` to run these operations:

### Setup Commands
- `install`: Install PostgreSQL on Debian/Ubuntu
- `setup`: Create database, user, and tables
- `create-sources-table`: Create main sources table
- `create-sources-ai-table`: Create AI sources table
- `create-flags-table`: Create flags table

### Connection Commands
- `connect`: Connect to database
- `backup`: Backup both tables to CSV
- `backup-sources`: Backup main sources table
- `backup-sources-ai`: Backup AI sources table

### Query Commands
- `recent-sources`: View recent main sources (1 week)
- `recent-ai-sources`: View recent AI sources (1 week)
- `count-sources`: Count total sources in both tables
- `unprocessed`: View unprocessed sources
- `important`: View important sources
- `table-sizes`: Check table sizes
- `activity`: Monitor recent activity

### Maintenance Commands
- `clear-duplicates`: Remove duplicate entries
- `reset-processed`: Reset processed flags for reprocessing
- `clear-old`: Remove entries older than 6 months
- `terminate-connections`: Drop other database connections

### Export Commands
- `export-relevant`: Export human-marked relevant sources
- `export-all`: Export all sources to CSV
- `export-ai`: Export all AI sources to CSV

## Table Schemas

### Sources Table
```sql
CREATE TABLE sources (
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
);
```

### Sources-AI Table
Same schema as sources table, but with quoted table name `"sources-ai"`.

### Flags Table
```sql
CREATE TABLE flags (
    name VARCHAR(50) PRIMARY KEY,
    code INTEGER NOT NULL,
    msg TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## SQL Files

The following SQL files are located in the `sql/` subfolder:
- `sql/create_sources_table.sql`: Creates main sources table
- `sql/create_sources_ai_table.sql`: Creates AI sources table  
- `sql/create_flags_table.sql`: Creates flags table
- `sql/clear_duplicates_main.sql`: Removes duplicates from sources
- `sql/clear_duplicates_ai.sql`: Removes duplicates from sources-ai
- `sql/count_sources.sql`: Counts entries in both tables
- `sql/table_sizes.sql`: Shows table sizes
- `sql/terminate_connections.sql`: Terminates database connections

## Notes

- All commands use `$DATABASE_POOL_URL` from your `.env` file
- The `sources-ai` table requires quoted identifiers due to the hyphen
- Backup files are timestamped with format `YYYYMMDD`
- Use `make help` to see all available commands

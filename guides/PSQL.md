## Installation on debian

Install libraries:

```
sudo apt install postgresql postgresql-client
```


Some local commands:

```
sudo -u postgres createuser nuno
sudo -u postgres createdb -O nuno nuno
# sudo -u postgres createuser galadriel
# sudo -u postgres -i
psql
psql -d nuno -h localhost -U nuno
DATABASE_URL=postgresql://localhost:5432
```

command to log into psql database:


```
source .env 
psql $DATABASE_POOL_URL
```

create and alter tables:

```
psql $DATABASE_URL -c "CREATE TABLE IF NOT EXISTS sources (id SERIAL PRIMARY KEY, title TEXT NOT NULL, link TEXT NOT NULL UNIQUE, date TIMESTAMP NOT NULL, summary TEXT, importance_bool BOOLEAN, importance_reasoning TEXT, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);"
ALTER TABLE sources ADD COLUMN "processed" BOOLEAN DEFAULT FALSE;
ALTER TABLE sources ADD COLUMN "relevant_per_human_check" TEXT DEFAULT 'maybe';
```

To read a command from a file:

```
psql $DATABASE_URL -f filename.sql
```

Some useful commands:

```
psql $DATABASE_POOL_URL -c "SELECT link FROM sources WHERE created_at > NOW() - INTERVAL '2 weeks';"
psql $DATABASE_URL -c "COPY (SELECT link FROM sources WHERE created_at > NOW() - INTERVAL '2 weeks') TO STDOUT WITH CSV;"
source .env && psql $DATABASE_URL -c "COPY (SELECT link FROM sources) TO STDOUT WITH CSV;" | grep gmw.cn
psql $DATABASE_URL -c "COPY (SELECT title, link, date FROM sources WHERE relevant_per_human_check = 'yes') TO STDOUT WITH CSV;"
psql $DATABASE_URL -c "COPY () TO STDOUT WITH CSV;"
psql $DATABASE_URL -c "SELECT id, title, link, date, created_at, processed FROM sources WHERE EXTRACT(MONTH FROM date) = $1 AND processed = false ORDER BY date ASC, id ASC"
psql $DATABASE_URL -c "COPY (SELECT title FROM sources WHERE EXTRACT(MONTH FROM date) = 3) TO STDOUT WITH CSV;"
psql $DATABASE_URL -c "COPY (SELECT title, link, date FROM sources WHERE relevant_per_human_check = 'yes' AND EXTRACT('week' from date) = 22) TO STDOUT WITH CSV;"
```

To drop other connections

```
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE datname = current_database()
  AND pid <> pg_backend_pid();
```

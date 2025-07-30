psql $DATABASE_URL -c "COPY (SELECT title FROM sources WHERE processed = false  AND EXTRACT('week' from date) = 26) TO STDOUT WITH CSV;" > titles.txt

